package index

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type FindSegmentsFile struct {
	directory                store.Directory
	doBody                   func(segmentFileName string) (obj interface{}, err error)
	defaultGenLookaheadCount int
}

func NewFindSegmentsFile(directory store.Directory,
	doBody func(segmentFileName string) (obj interface{}, err error)) *FindSegmentsFile {
	return &FindSegmentsFile{directory, doBody, 10}
}

// TODO support IndexCommit
func (fsf *FindSegmentsFile) run() (obj interface{}, err error) {
	log.Print("Finding segments file...")
	// if commit != nil {
	// 	if fsf.directory != commit.Directory {
	// 		return nil, errors.New("the specified commit does not match the specified Directory")
	// 	}
	// 	return fsf.doBody(commit.SegmentsFileName)
	// }

	lastGen := int64(-1)
	gen := int64(0)
	genLookaheadCount := 0
	var exc error
	retryCount := 0

	useFirstMethod := true

	// Loop until we succeed in calling doBody() without
	// hitting an IOException.  An IOException most likely
	// means a commit was in process and has finished, in
	// the time it took us to load the now-old infos files
	// (and segments files).  It's also possible it's a
	// true error (corrupt index).  To distinguish these,
	// on each retry we must see "forward progress" on
	// which generation we are trying to load.  If we
	// don't, then the original error is real and we throw
	// it.

	// We have three methods for determining the current
	// generation.  We try the first two in parallel (when
	// useFirstMethod is true), and fall back to the third
	// when necessary.

	for {
		log.Print("Trying...")
		if useFirstMethod {
			log.Print("Trying first method...")
			// List the directory and use the highest
			// segments_N file.  This method works well as long
			// as there is no stale caching on the directory
			// contents (NOTE: NFS clients often have such stale
			// caching):
			genA := int64(-1)

			files, err := fsf.directory.ListAll()
			if err != nil {
				return nil, err
			}
			if files != nil {
				genA = LastCommitGeneration(files)
			}

			// TODO support info stream
			// if fsf.infoStream != nil {
			// 	message("directory listing genA=" + genA)
			// }
			log.Printf("directory listing genA=%v", genA)

			// Also open segments.gen and read its
			// contents.  Then we take the larger of the two
			// gens.  This way, if either approach is hitting
			// a stale cache (NFS) we have a better chance of
			// getting the right generation.
			genB := int64(-1)
			genInput, err := fsf.directory.OpenInput(INDEX_FILENAME_SEGMENTS_GEN, store.IO_CONTEXT_READ)
			if err != nil {
				// if fsf.infoStream != nil {
				log.Printf("segments.gen open: %v", err)
				// }
			} else {
				defer genInput.Close()
				log.Print("Reading segments info...")

				version, err := genInput.ReadInt()
				if err != nil {
					return nil, err
				}
				log.Printf("Version: %v", version)
				if version == FORMAT_SEGMENTS_GEN_CURRENT {
					log.Print("Version is current.")
					gen0, err := genInput.ReadLong()
					if err != nil {
						return nil, err
					}
					gen1, err := genInput.ReadLong()
					if err != nil {
						return nil, err
					} else {
						// if fsf.infoStream != nil {
						log.Printf("fallback check: %v; %v", gen0, gen1)
						// }
						if gen0 == gen1 {
							// The file is consistent.
							genB = gen0
						}
					}
				} else {
					return nil, codec.NewIndexFormatTooNewError(genInput, version, FORMAT_SEGMENTS_GEN_CURRENT, FORMAT_SEGMENTS_GEN_CURRENT)
				}
			}

			// if fsf.infoStream != nil {
			log.Printf("%v check: genB=%v", INDEX_FILENAME_SEGMENTS_GEN, genB)
			// }

			// Pick the larger of the two gen's:
			gen = genA
			if genB > gen {
				gen = genB
			}

			if gen == -1 {
				// Neither approach found a generation
				return nil, errors.New(fmt.Sprintf("no segments* file found in %v: files: %#v", fsf.directory, files))
			}
		}

		if useFirstMethod && lastGen == gen && retryCount >= 2 {
			// Give up on first method -- this is 3rd cycle on
			// listing directory and checking gen file to
			// attempt to locate the segments file.
			useFirstMethod = false
		}

		// Second method: since both directory cache and
		// file contents cache seem to be stale, just
		// advance the generation.
		if !useFirstMethod {
			if genLookaheadCount < fsf.defaultGenLookaheadCount {
				gen++
				genLookaheadCount++
				// if fsf.infoStream != nil {
				log.Printf("look ahead increment gen to %v", gen)
				// }
			} else {
				// All attempts have failed -- throw first exc:
				return nil, exc
			}
		} else if lastGen == gen {
			// This means we're about to try the same
			// segments_N last tried.
			retryCount++
		} else {
			// Segment file has advanced since our last loop
			// (we made "progress"), so reset retryCount:
			retryCount = 0
		}

		lastGen = gen
		segmentFileName := util.FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", gen)
		log.Printf("SegmentFileName: %v", segmentFileName)

		v, err := fsf.doBody(segmentFileName)
		if err != nil {
			// Save the original root cause:
			if exc == nil {
				exc = err
			}

			// if fsf.infoStream != nil {
			log.Printf("primary Exception on '%v': %v; will retry: retryCount = %v; gen = %v",
				segmentFileName, err, retryCount, gen)
			// }

			if gen > 1 && useFirstMethod && retryCount == 1 {
				// This is our second time trying this same segments
				// file (because retryCount is 1), and, there is
				// possibly a segments_(N-1) (because gen > 1).
				// So, check if the segments_(N-1) exists and
				// try it if so:
				prevSegmentFileName := util.FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", gen-1)

				if prevExists := fsf.directory.FileExists(prevSegmentFileName); prevExists {
					// if fsf.infoStream != nil {
					log.Printf("fallback to prior segment file '%v'", prevSegmentFileName)
					// }
					v, err = fsf.doBody(prevSegmentFileName)
					if err != nil {
						// if fsf.infoStream != nil {
						log.Printf("secondary Exception on '%v': %v; will retry", prevSegmentFileName, err)
						//}
					} else {
						// if fsf.infoStream != nil {
						log.Printf("success on fallback %v", prevSegmentFileName)
						//}
						return v, nil
					}
				}
			}
		} else {
			// if fsf.infoStream != nil {
			log.Printf("success on %v", segmentFileName)
			// }
			return v, nil
		}
	}
}

const (
	INDEX_FILENAME_SEGMENTS     = "segments"
	INDEX_FILENAME_SEGMENTS_GEN = "segments.gen"
	COMOPUND_FILE_EXTENSION     = "cfs"
	VERSION_40                  = 0
	FORMAT_SEGMENTS_GEN_CURRENT = -2
)

type SegmentInfo struct {
	dir            store.Directory
	version        string
	name           string
	docCount       int32
	isCompoundFile bool
	codec          Codec
	diagnostics    map[string]string
	attributes     map[string]string
	Files          map[string]bool // must use CheckFileNames()
}

// seprate norms are not supported in >= 4.0
func (si *SegmentInfo) hasSeparateNorms() bool {
	return false
}

func (si *SegmentInfo) String() string {
	return si.StringOf(si.dir, 0)
}

func (si *SegmentInfo) StringOf(dir store.Directory, delCount int) string {
	var buf bytes.Buffer
	buf.WriteString(si.name)
	buf.WriteString("(")
	if si.version == "" {
		buf.WriteString("?")
	} else {
		buf.WriteString(si.version)
	}
	buf.WriteString("):")
	if si.isCompoundFile {
		buf.WriteString("c")
	} else {
		buf.WriteString("C")
	}

	if si.dir != dir {
		buf.WriteString("x")
	}
	buf.WriteString(strconv.Itoa(int(si.docCount)))

	if delCount != 0 {
		buf.WriteString("/")
		buf.WriteString(strconv.Itoa(delCount))
	}

	// TODO: we could append toString of attributes() here?

	return buf.String()
}

var CODEC_FILE_PATTERN = regexp.MustCompile("_[a-z0-9]+(_.*)?\\..*")

func (si *SegmentInfo) CheckFileNames(files map[string]bool) {
	for file, _ := range files {
		if !CODEC_FILE_PATTERN.MatchString(file) {
			panic(fmt.Sprintf("invalid codec filename '%v', must match: %v", file, CODEC_FILE_PATTERN))
		}
	}
}

// index/SegmentInfos.java

/*
A collection of segmentInfo objects with methods for operating on
those segments in relation to the file system.

The active segments in the index are stored in the segment into file,
segments_N. There may be one or more segments_N files in the index;
however, hte one with the largest generation is the activbe one (when
older segments_N files are present it's because they temporarily
cannot be deleted, or, a writer is in the process of committing, or a
custom IndexDeletionPolicy is in use). This file lists each segment
by name and has details about the codec and generation of deletes.

There is also a file segments.gen. This file contains the current
generation (the _N in segments_N) of the index. This is used only as
a fallback in case the current generation cannot be accurately
determined by directory listing alone (as is the case for some NFS
clients with time-based directory cache expiration). This file simply
contains an Int32 version header (FORMAT_SEGMENTS_GEN_CURRENT),
followed by the generation recorded as int64, written twice.

Files:

- segments.gen: GenHeader, Generation, Generation
- segments_N: Header, Version, NameCounter, SegCount,
  <SegName, SegCodec, DelGen, DeletionCount>^SegCount, CommitUserData, Checksum

Data types:

- Header --> CodecHeader
- Genheader, NameCounter, SegCount, DeletionCount --> int32
- Generation, Version, DelGen, Checksum --> int64
- SegName, SegCodec --> string
- CommitUserData --> map[string]string

Field Descriptions:

- Version counts how often the index has been changed by adding or
  deleting docments.
- NameCounter is used to generate names for new segment files.
- SegName is the name of the segment, and is used as the file name
  prefix for all of the files that compose the segment's index.
- DelGen is the generation count of the deletes file. If this is -1,
  there are no deletes. Anything above zero means there are deletes
  stored by LiveDocsFormat.
- DeletionCount records the number of deleted documents in this segment.
- Checksum contains the CRC32 checksum of all bytes in the segments_N
  file up until the checksum. This is used to verify integrity of the
  file on opening the index.
- SegCodec is the nme of the Codec that encoded this segment.
- CommitUserData stores an optional user-spplied opaue
  map[string]string that was passed to SetCommitData().
*/
type SegmentInfos struct {
	counter        int
	version        int64
	generation     int64
	lastGeneration int64
	userData       map[string]string
	Segments       []*SegmentInfoPerCommit

	// Only non-nil after prepareCommit has been called and before
	// finishCommit is called
	pendingSegnOutput *store.ChecksumIndexOutput
}

func LastCommitGeneration(files []string) int64 {
	if files == nil {
		return int64(-1)
	}
	max := int64(-1)
	for _, file := range files {
		if strings.HasPrefix(file, INDEX_FILENAME_SEGMENTS) && file != INDEX_FILENAME_SEGMENTS_GEN {
			gen := GenerationFromSegmentsFileName(file)
			if gen > max {
				max = gen
			}
		}
	}
	return max
}

func (sis *SegmentInfos) SegmentsFileName() string {
	return util.FileNameFromGeneration(util.SEGMENTS, "", sis.lastGeneration)
}

func GenerationFromSegmentsFileName(fileName string) int64 {
	switch {
	case fileName == INDEX_FILENAME_SEGMENTS:
		return int64(0)
	case strings.HasPrefix(fileName, INDEX_FILENAME_SEGMENTS):
		d, err := strconv.ParseInt(fileName[1+len(INDEX_FILENAME_SEGMENTS):], 36, 64)
		if err != nil {
			panic(err)
		}
		return d
	default:
		panic(fmt.Sprintf("filename %v is not a segments file", fileName))
	}
}

/* Get the next segments_N filename that will be written. */
func (sis *SegmentInfos) nextSegmentFilename() string {
	var nextGeneration int64
	if sis.generation == -1 {
		nextGeneration = 1
	} else {
		nextGeneration = sis.generation + 1
	}
	return util.FileNameFromGeneration(util.SEGMENTS, "", nextGeneration)
}

/*
Read a particular segmentFileName. Note that this may return IO error
if a commit is in process.
*/
func (sis *SegmentInfos) Read(directory store.Directory, segmentFileName string) error {
	log.Printf("Reading segment info from %v...", segmentFileName)
	success := false

	// Clear any previous segments:
	sis.Clear()

	sis.generation = GenerationFromSegmentsFileName(segmentFileName)
	sis.lastGeneration = sis.generation

	main, err := directory.OpenInput(segmentFileName, store.IO_CONTEXT_READ)
	if err != nil {
		return err
	}
	input := store.NewChecksumIndexInput(main)
	defer func() {
		if !success {
			// Clear any segment infos we had loaded so we
			// have a clean slate on retry:
			sis.Clear()
			util.CloseWhileSuppressingError(input)
		} else {
			input.Close()
		}
	}()

	format, err := input.ReadInt()
	if err != nil {
		return err
	}
	if format == codec.CODEC_MAGIC {
		// 4.0+
		_, err = codec.CheckHeaderNoMagic(input, "segments", VERSION_40, VERSION_40)
		if err != nil {
			return err
		}
		sis.version, err = input.ReadLong()
		if err != nil {
			return err
		}
		sis.counter, err = asInt(input.ReadInt())
		if err != nil {
			return err
		}
		numSegments, err := asInt(input.ReadInt())
		if err != nil {
			return err
		}
		if numSegments < 0 {
			return errors.New(fmt.Sprintf("invalid segment count: %v (resource: %v)", numSegments, input))
		}
		for seg := 0; seg < numSegments; seg++ {
			segName, err := input.ReadString()
			if err != nil {
				return err
			}
			codecName, err := input.ReadString()
			if err != nil {
				return err
			}
			if codecName != "Lucene42" {
				log.Panicf("Not supported yet: %v", codecName)
			}
			fCodec := LoadCodec(codecName)
			log.Printf("SIS.read seg=%v codec=%v", seg, fCodec)
			info, err := fCodec.SegmentInfoFormat().SegmentInfoReader()(directory, segName, store.IO_CONTEXT_READ)
			// method := NewLucene42Codec()
			// info, err := method.ReadSegmentInfo(directory, segName, store.IO_CONTEXT_READ)
			if err != nil {
				return err
			}
			// info.codec = method
			info.codec = fCodec
			delGen, err := input.ReadLong()
			if err != nil {
				return err
			}
			delCount, err := asInt(input.ReadInt())
			if err != nil {
				return err
			}
			if delCount < 0 || delCount > int(info.docCount) {
				return errors.New(fmt.Sprintf("invalid deletion count: %v (resource: %v)", delCount, input))
			}
			sis.Segments = append(sis.Segments, NewSegmentInfoPerCommit(info, delCount, delGen))
		}
		sis.userData, err = input.ReadStringStringMap()
		if err != nil {
			return err
		}
	} else {
		// TODO support <4.0 index
		panic("Index format pre-4.0 not supported yet")
	}

	checksumNow := int64(input.Checksum())
	checksumThen, err := input.ReadLong()
	if err != nil {
		return err
	}
	if checksumNow != checksumThen {
		return errors.New(fmt.Sprintf("checksum mismatch in segments file (resource: %v)", input))
	}

	success = true
	return nil
}

func (sis *SegmentInfos) ReadAll(directory store.Directory) error {
	sis.generation, sis.lastGeneration = -1, -1
	_, err := NewFindSegmentsFile(directory, func(segmentFileName string) (obj interface{}, err error) {
		err = sis.Read(directory, segmentFileName)
		return nil, err
	}).run()
	return err
}

func (sis *SegmentInfos) write(directory store.Directory) error {
	segmentsFilename := sis.nextSegmentFilename()

	// Always advance the generation on write:
	if sis.generation == -1 {
		sis.generation = 1
	} else {
		sis.generation++
	}

	var segnOutput *store.ChecksumIndexOutput
	var success = false
	var upgradedSIFiles = make(map[string]bool)

	defer func() {
		if !success {
			// We hit an error above; try to close the file but suppress
			// any errors
			util.CloseWhileSuppressingError(segnOutput)

			for filename, _ := range upgradedSIFiles {
				directory.DeleteFile(filename) // ignore error
			}

			// Try not to leave a truncated segments_N fle in the index:
			directory.DeleteFile(segmentsFilename) // ignore error
		}
	}()

	out, err := directory.CreateOutput(segmentsFilename, store.IO_CONTEXT_DEFAULT)
	if err != nil {
		return err
	}
	segnOutput = store.NewChecksumIndexOutput(out)
	err = codec.WriteHeader(segnOutput, "segments", VERSION_40)
	if err != nil {
		return err
	}
	err = segnOutput.WriteLong(sis.version)
	if err == nil {
		err = segnOutput.WriteInt(int32(sis.counter))
		if err == nil {
			err = segnOutput.WriteInt(int32(len(sis.Segments)))
		}
	}
	if err != nil {
		return err
	}
	for _, siPerCommit := range sis.Segments {
		si := siPerCommit.info
		err = segnOutput.WriteString(si.name)
		if err == nil {
			err = segnOutput.WriteString(si.codec.Name())
			if err == nil {
				err = segnOutput.WriteLong(siPerCommit.delGen)
				if err == nil {
					err = segnOutput.WriteInt(int32(siPerCommit.delCount))
				}
			}
		}
		if err != nil {
			return err
		}
		assert(si.dir == directory)

		assert(siPerCommit.delCount <= int(si.docCount))

		// If this segment is pre-4.x, perform a one-time "upgrade" to
		// write the .si file for it:
		if version := si.version; version == "" || versionLess(version, "4.0") {
			panic("not implemented yet")
		}
	}
	err = segnOutput.WriteStringStringMap(sis.userData)
	if err != nil {
		return err
	}
	sis.pendingSegnOutput = segnOutput
	success = true
	return nil
}

func versionLess(a, b string) bool {
	parts1 := strings.Split(a, ".")
	parts2 := strings.Split(b, ".")
	for i, v := range parts1 {
		n1, _ := strconv.Atoi(v)
		if i < len(parts2) {
			if n2, _ := strconv.Atoi(parts2[i]); n1 != n2 {
				return n1 < n2
			}
		} else if n1 != 0 {
			// a has some extra trailing tokens.
			// if these are all zeroes, that's ok.
			return false
		}
	}

	// b has some extra trailing tokens.
	// if these are all zeroes, that's ok.
	for i := len(parts1); i < len(parts2); i++ {
		if n, _ := strconv.Atoi(parts2[i]); n != 0 {
			return true
		}
	}

	return false
}

/*
Returns a copy of this instance, also copying each SegmentInfo.
*/
func (sis *SegmentInfos) Clone() *SegmentInfos {
	clone := &SegmentInfos{
		counter:        sis.counter,
		version:        sis.version,
		generation:     sis.generation,
		lastGeneration: sis.lastGeneration,
		userData:       make(map[string]string),
		Segments:       nil,
	}
	for _, info := range sis.Segments {
		assert(info.info.codec != nil)
		clone.Segments = append(clone.Segments, info.Clone())
	}
	for k, v := range sis.userData {
		clone.userData[k] = v
	}
	return clone
}

// L873
/* Carry over generation numbers from another SegmentInfos */
func (sis *SegmentInfos) updateGeneration(other *SegmentInfos) {
	sis.lastGeneration = other.lastGeneration
	sis.generation = other.generation
}

func (sis *SegmentInfos) rollbackCommit(dir store.Directory) {
	if sis.pendingSegnOutput != nil {
		// Suppress so we keep throwing the original error in our caller
		util.CloseWhileSuppressingError(sis.pendingSegnOutput)
		sis.pendingSegnOutput = nil

		// Must carefully compute filename from "generation" since
		// lastGeneration isn't incremented:
		segmentFilename := util.FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", sis.generation)

		// Suppress so we keep throwing the original error in our caller
		util.DeleteFilesIgnoringErrors(dir, segmentFilename)
	}
}

/*
Call this to start a commit. This writes the new segments file, but
writes an invalid checksum at the end, so that it is not visible to
readers. Once this is called you must call finishCommit() to complete
the commit or rollbackCommit() to abort it.

Note: changed() should be called prior to this method if changes have
been made to this SegmentInfos instance.
*/
func (sis *SegmentInfos) prepareCommit(dir store.Directory) error {
	assert2(sis.pendingSegnOutput == nil, "prepareCommit was already called")
	return sis.write(dir)
}

/*
Returns all file names referenced by SegmentInfo instances matching
the provided Directory (ie files associated with any "external"
segments are skipped). The returned collection is recomputed on each
invocation.
*/
func (sis *SegmentInfos) files(dir store.Directory, includeSegmentsFile bool) []string {
	files := make(map[string]bool)
	if includeSegmentsFile {
		if segmentFileName := sis.SegmentsFileName(); segmentFileName != "" {
			files[segmentFileName] = true
		}
	}
	for _, info := range sis.Segments {
		assert(info.info.dir == dir)
		// if info.info.dir == dir {
		for _, file := range info.Files() {
			files[file] = true
		}
		// }
	}
	var res = make([]string, 0, len(files))
	for file, _ := range files {
		res = append(res, file)
	}
	return res
}

func (sis *SegmentInfos) finishCommit(dir store.Directory) error {
	assert2(sis.pendingSegnOutput != nil, "prepareCommit was not called")
	var success = false
	defer func() {
		if !success {
			// Closes pendingSegnOutput & delets partial segments_N:
			sis.rollbackCommit(dir)
		} else {
			sis.pendingSegnOutput = nil
		}
	}()

	sis.pendingSegnOutput.Close()
	success = true
	return nil
}

// L1041
/*
Replaces all segments in this instance in this instance, but keeps
generation, version, counter so that future commits remain write once.
*/
func (sis *SegmentInfos) replace(other *SegmentInfos) {
	sis.rollbackSegmentInfos(other.Segments)
	sis.lastGeneration = other.lastGeneration
}

func (sis *SegmentInfos) createBackupSegmentInfos() []*SegmentInfoPerCommit {
	ans := make([]*SegmentInfoPerCommit, len(sis.Segments))
	for i, info := range sis.Segments {
		assert(info.info.codec != nil)
		ans[i] = info.Clone()
	}
	return ans
}

// L1104
func (sis *SegmentInfos) rollbackSegmentInfos(infos []*SegmentInfoPerCommit) {
	if cap(sis.Segments) < len(infos) {
		sis.Segments = make([]*SegmentInfoPerCommit, len(infos))
		copy(sis.Segments, infos)
	} else {
		n := len(sis.Segments)
		copy(sis.Segments, infos)
		for i, limit := len(infos), n; i < limit; i++ {
			sis.Segments[i] = nil
		}
		sis.Segments = sis.Segments[0:len(infos)]
	}
}

func (sis *SegmentInfos) Clear() {
	for i, _ := range sis.Segments {
		sis.Segments[i] = nil
	}
	sis.Segments = sis.Segments[:0] // reuse existing space
}

/*
Remove the provided SegmentInfoPerCommit.

WARNING: O(N) cost
*/
func (sis *SegmentInfos) remove(si *SegmentInfoPerCommit) {

}
