package index

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"strconv"
	"strings"
)

/* Prints the given message to the infoStream. */
func message(format string, args ...interface{}) {
	fmt.Printf("SIS: %v\n", fmt.Sprintf(format, args...))
}

type FindSegmentsFile struct {
	directory                store.Directory
	doBody                   func(segmentFileName string) (interface{}, error)
	defaultGenLookaheadCount int
}

func NewFindSegmentsFile(directory store.Directory,
	doBody func(segmentFileName string) (interface{}, error)) *FindSegmentsFile {
	return &FindSegmentsFile{directory, doBody, 10}
}

// TODO support IndexCommit
func (fsf *FindSegmentsFile) run(commit IndexCommit) (interface{}, error) {
	// fmt.Println("Finding segments file...")
	if commit != nil {
		if fsf.directory != commit.Directory() {
			return nil, errors.New("the specified commit does not match the specified Directory")
		}
		return fsf.doBody(commit.SegmentsFileName())
	}

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
		// fmt.Println("Trying...")
		if useFirstMethod {
			// fmt.Println("Trying first method...")
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

			// message("directory listing genA=%v", genA)

			// Also open segments.gen and read its
			// contents.  Then we take the larger of the two
			// gens.  This way, if either approach is hitting
			// a stale cache (NFS) we have a better chance of
			// getting the right generation.
			genB := int64(-1)
			genInput, err := fsf.directory.OpenChecksumInput(INDEX_FILENAME_SEGMENTS_GEN, store.IO_CONTEXT_READ)
			if err != nil {
				message("segments.gen open: %v", err)
			} else {
				defer genInput.Close()
				// fmt.Println("Reading segments info...")

				var version int32
				if version, err = genInput.ReadInt(); err != nil {
					return nil, err
				}
				// fmt.Printf("Version: %v\n", version)
				if version == FORMAT_SEGMENTS_GEN_47 || version == FORMAT_SEGMENTS_GEN_CURRENT {
					// fmt.Println("Version is current.")
					var gen0, gen1 int64
					if gen0, err = genInput.ReadLong(); err != nil {
						return nil, err
					}
					if gen1, err = genInput.ReadLong(); err != nil {
						return nil, err
					}
					message("fallback check: %v; %v", gen0, gen1)
					if version == FORMAT_SEGMENTS_GEN_CHECKSUM {
						if _, err = codec.CheckFooter(genInput); err != nil {
							return nil, err
						}
					} else {
						if err = codec.CheckEOF(genInput); err != nil {
							return nil, err
						}
					}
					if gen0 == gen1 {
						// The file is consistent.
						genB = gen0
					}
				} else {
					return nil, codec.NewIndexFormatTooNewError(genInput, version,
						FORMAT_SEGMENTS_GEN_CURRENT, FORMAT_SEGMENTS_GEN_CURRENT)
				}
			}

			message("%v check: genB=%v", INDEX_FILENAME_SEGMENTS_GEN, genB)

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
				message("look ahead increment gen to %v", gen)
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
		// fmt.Printf("SegmentFileName: %v\n", segmentFileName)

		var v interface{}
		var err error
		if v, err = fsf.doBody(segmentFileName); err == nil {
			message("success on %v", segmentFileName)
			return v, nil
		}
		// Save the original root cause:
		if exc == nil {
			exc = err
		}

		message("primary Exception on '%v': %v; will retry: retryCount = %v; gen = %v",
			segmentFileName, err, retryCount, gen)

		if gen > 1 && useFirstMethod && retryCount == 1 {
			// This is our second time trying this same segments
			// file (because retryCount is 1), and, there is
			// possibly a segments_(N-1) (because gen > 1).
			// So, check if the segments_(N-1) exists and
			// try it if so:
			prevSegmentFileName := util.FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", gen-1)

			if prevExists := fsf.directory.FileExists(prevSegmentFileName); prevExists {
				message("fallback to prior segment file '%v'", prevSegmentFileName)
				if v, err = fsf.doBody(prevSegmentFileName); err != nil {
					message("secondary Exception on '%v': %v; will retry", prevSegmentFileName, err)
				} else {
					message("success on fallback %v", prevSegmentFileName)
					return v, nil
				}
			}
		}
	}
}

// index/SegmentInfos.java

const (
	VERSION_40 = 0
	VERSION_46 = 1
	VERSION_48 = 2
	VERSION_49 = 3

	// Used for the segments.gen file only!
	// Whenver you add a new format, make it 1 smaller (negative version logic)!
	FORMAT_SEGMENTS_GEN_47       = -2
	FORMAT_SEGMENTS_GEN_CHECKSUM = -3
	FORMAT_SEGMENTS_GEN_START    = FORMAT_SEGMENTS_GEN_47
	// Current format of segments.gen
	FORMAT_SEGMENTS_GEN_CURRENT = FORMAT_SEGMENTS_GEN_CHECKSUM
)

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

- segments.gen: GenHeader, Generation, Generation, Footer
- segments_N: Header, Version, NameCounter, SegCount,
  <SegName, SegCodec, DelGen, DeletionCount, FieldInfosGen,
  DocValuesGen, UpdatesFiles>^SegCount, CommitUserData, Footer

Data types:

- Header --> CodecHeader
- Genheader, NameCounter, SegCount, DeletionCount --> int32
- Generation, Version, DelGen, Checksum --> int64
- SegName, SegCodec --> string
- CommitUserData --> map[string]string
- UpdatesFiles --> map[int32]map[string]bool>
- Footer --> CodecFooter

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
- SegCodec is the nme of the Codec that encoded this segment.
- CommitUserData stores an optional user-spplied opaue
  map[string]string that was passed to SetCommitData().
- FieldInfosGen is the generation count of the fieldInfos file. If
	this is -1, there are no updates to the fieldInfos in that segment.
	Anything above zero means there are updates to the fieldInfos
	stored by FieldInfosFormat.
- DocValuesGen is the generation count of the updatable DocValues. If
	this is -1, there are no udpates to DocValues in that segment.
	Anything above zero means there are updates to DocValues stored by
	DocvaluesFormat.
- UpdatesFiles stores the set of files that were updated in that
	segment per file.
*/
type SegmentInfos struct {
	counter        int
	version        int64
	generation     int64
	lastGeneration int64
	userData       map[string]string
	Segments       []*SegmentCommitInfo

	// Only non-nil after prepareCommit has been called and before
	// finishCommit is called
	pendingSegnOutput store.IndexOutput
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

/*
A utility for writing the SEGMENTS_GEN file to a Directory.

NOTE: this is an internal utility which is kept public so that it's
accessible by code from other packages. You should avoid calling this
method unless you're absolutely sure what you're doing!
*/
func writeSegmentsGen(dir store.Directory, generation int64) {
	if err := func() (err error) {
		var genOutput store.IndexOutput
		genOutput, err = dir.CreateOutput(INDEX_FILENAME_SEGMENTS_GEN, store.IO_CONTEXT_READONCE)
		if err != nil {
			return err
		}

		defer func() {
			err = mergeError(err, genOutput.Close())
			err = mergeError(err, dir.Sync([]string{INDEX_FILENAME_SEGMENTS_GEN}))
		}()

		if err = genOutput.WriteInt(FORMAT_SEGMENTS_GEN_CURRENT); err == nil {
			if err = genOutput.WriteLong(generation); err == nil {
				if err = genOutput.WriteLong(generation); err == nil {
					err = codec.WriteFooter(genOutput)
				}
			}
		}
		return err
	}(); err != nil {
		// It's OK if we fail to write this file since it's used only as
		// one of the retry fallbacks.
		dir.DeleteFile(INDEX_FILENAME_SEGMENTS_GEN)
		// Ignore error; this file is only used in a retry fallback on init
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
func (sis *SegmentInfos) Read(directory store.Directory, segmentFileName string) (err error) {
	// fmt.Printf("Reading segment info from %v...\n", segmentFileName)

	// Clear any previous segments:
	sis.Clear()

	sis.generation = GenerationFromSegmentsFileName(segmentFileName)
	sis.lastGeneration = sis.generation

	var input store.ChecksumIndexInput
	if input, err = directory.OpenChecksumInput(segmentFileName, store.IO_CONTEXT_READ); err != nil {
		return
	}

	var success = false
	defer func() {
		if !success {
			// Clear any segment infos we had loaded so we
			// have a clean slate on retry:
			sis.Clear()
			util.CloseWhileSuppressingError(input)
		} else {
			err = input.Close()
		}
	}()

	var format int
	if format, err = asInt(input.ReadInt()); err != nil {
		return
	}

	var actualFormat int
	if format == codec.CODEC_MAGIC {
		// 4.0+
		if actualFormat, err = asInt(codec.CheckHeaderNoMagic(input, "segments", VERSION_40, VERSION_49)); err != nil {
			return
		}
		if sis.version, err = input.ReadLong(); err != nil {
			return
		}
		if sis.counter, err = asInt(input.ReadInt()); err != nil {
			return
		}
		var numSegments int
		if numSegments, err = asInt(input.ReadInt()); err != nil {
			return
		} else if numSegments < 0 {
			return errors.New(fmt.Sprintf("invalid segment count: %v (resource: %v)", numSegments, input))
		}
		var segName, codecName string
		var fCodec Codec
		var delGen, fieldInfosGen, dvGen int64
		var delCount int
		for seg := 0; seg < numSegments; seg++ {
			if segName, err = input.ReadString(); err != nil {
				return
			}
			if codecName, err = input.ReadString(); err != nil {
				return
			}
			fCodec = LoadCodec(codecName)
			assert2(fCodec != nil, "Invalid codec name: %v", codecName)
			// fmt.Printf("SIS.read seg=%v codec=%v\n", seg, fCodec)
			var info *SegmentInfo
			if info, err = fCodec.SegmentInfoFormat().SegmentInfoReader().Read(directory, segName, store.IO_CONTEXT_READ); err != nil {
				return
			}
			info.SetCodec(fCodec)
			if delGen, err = input.ReadLong(); err != nil {
				return
			}
			if delCount, err = asInt(input.ReadInt()); err != nil {
				return
			} else if delCount < 0 || delCount > info.DocCount() {
				return errors.New(fmt.Sprintf(
					"invalid deletion count: %v vs docCount=%v (resource: %v)",
					delCount, info.DocCount(), input))
			}
			fieldInfosGen = -1
			if actualFormat >= VERSION_46 {
				if fieldInfosGen, err = input.ReadLong(); err != nil {
					return
				}
			}
			dvGen = -1
			if actualFormat >= VERSION_49 {
				if dvGen, err = input.ReadLong(); err != nil {
					return
				}
			} else {
				dvGen = fieldInfosGen
			}
			siPerCommit := NewSegmentCommitInfo(info, delCount, delGen, fieldInfosGen, dvGen)
			if actualFormat >= VERSION_46 {
				if actualFormat < VERSION_49 {
					panic("not implemented yet")
				} else {
					var ss map[string]bool
					if ss, err = input.ReadStringSet(); err != nil {
						return err
					}
					siPerCommit.SetFieldInfosFiles(ss)
					var dvUpdatesFiles map[int]map[string]bool
					var numDVFields int
					if numDVFields, err = asInt(input.ReadInt()); err != nil {
						return err
					}
					if numDVFields == 0 {
						dvUpdatesFiles = make(map[int]map[string]bool)
					} else {
						panic("not implemented yet")
					}
					siPerCommit.SetDocValuesUpdatesFiles(dvUpdatesFiles)
				}
			}
			sis.Segments = append(sis.Segments, siPerCommit)
		}
		if sis.userData, err = input.ReadStringStringMap(); err != nil {
			return err
		}
	} else {
		// TODO support <4.0 index
		panic("Index format pre-4.0 not supported yet")
	}

	if actualFormat >= VERSION_48 {
		if _, err = codec.CheckFooter(input); err != nil {
			return
		}
	} else {
		var checksumNow = int64(input.Checksum())
		var checksumThen int64
		if checksumThen, err = input.ReadLong(); err != nil {
			return
		}
		if checksumNow != checksumThen {
			return errors.New(fmt.Sprintf(
				"checksum mismatch in segments file: %v vs %v (resource: %v)",
				checksumNow, checksumThen, input))
		}
		if err = codec.CheckEOF(input); err != nil {
			return
		}
	}

	success = true
	return nil
}

func asInt(n int32, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (sis *SegmentInfos) ReadAll(directory store.Directory) error {
	sis.generation, sis.lastGeneration = -1, -1
	_, err := NewFindSegmentsFile(directory, func(segmentFileName string) (obj interface{}, err error) {
		err = sis.Read(directory, segmentFileName)
		return nil, err
	}).run(nil)
	return err
}

func (sis *SegmentInfos) write(directory store.Directory) (err error) {
	segmentsFilename := sis.nextSegmentFilename()

	// Always advance the generation on write:
	if sis.generation == -1 {
		sis.generation = 1
	} else {
		sis.generation++
	}

	var segnOutput store.IndexOutput
	var success = false
	// var upgradedSIFiles = make(map[string]bool)

	defer func() {
		if !success {
			// We hit an error above; try to close the file but suppress
			// any errors
			util.CloseWhileSuppressingError(segnOutput)

			// for filename, _ := range upgradedSIFiles {
			// 	directory.DeleteFile(filename) // ignore error
			// }

			// Try not to leave a truncated segments_N fle in the index:
			directory.DeleteFile(segmentsFilename) // ignore error
		}
	}()

	if segnOutput, err = directory.CreateOutput(segmentsFilename, store.IO_CONTEXT_DEFAULT); err != nil {
		return
	}
	if err = codec.WriteHeader(segnOutput, "segments", VERSION_49); err != nil {
		return
	}
	if err = segnOutput.WriteLong(sis.version); err == nil {
		if err = segnOutput.WriteInt(int32(sis.counter)); err == nil {
			err = segnOutput.WriteInt(int32(len(sis.Segments)))
		}
	}
	if err != nil {
		return
	}
	for _, siPerCommit := range sis.Segments {
		si := siPerCommit.Info
		if err = segnOutput.WriteString(si.Name); err == nil {
			if err = segnOutput.WriteString(si.Codec().(Codec).Name()); err == nil {
				if err = segnOutput.WriteLong(siPerCommit.DelGen()); err == nil {
					assert2(siPerCommit.DelCount() >= 0 && siPerCommit.DelCount() <= si.DocCount(),
						"cannot write segment: invalid docCount segment=%v docCount=%v delCount=%v",
						si.Name, si.DocCount(), siPerCommit.DelCount())
					if err = segnOutput.WriteInt(int32(siPerCommit.DelCount())); err == nil {
						if err = segnOutput.WriteLong(siPerCommit.FieldInfosGen()); err == nil {
							if err = segnOutput.WriteLong(siPerCommit.DocValuesGen()); err == nil {
								if err = segnOutput.WriteStringSet(siPerCommit.FieldInfosFiles()); err == nil {
									dvUpdatesFiles := siPerCommit.DocValuesUpdatesFiles()
									if err = segnOutput.WriteInt(int32(len(dvUpdatesFiles))); err == nil {
										for k, v := range dvUpdatesFiles {
											if err = segnOutput.WriteInt(int32(k)); err != nil {
												break
											}
											if err = segnOutput.WriteStringSet(v); err != nil {
												break
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		if err != nil {
			return
		}
		assert(si.Dir == directory)

		// If this segment is pre-4.x, perform a one-time "upgrade" to
		// write the .si file for it:
		if version := si.Version(); len(version) == 0 || !version.OnOrAfter(util.VERSION_4_0) {
			panic("not implemented yet")
		}
	}
	if err = segnOutput.WriteStringStringMap(sis.userData); err != nil {
		return
	}
	sis.pendingSegnOutput = segnOutput
	success = true
	return nil
}

// func versionLess(a, b string) bool {
// 	parts1 := strings.Split(a, ".")
// 	parts2 := strings.Split(b, ".")
// 	for i, v := range parts1 {
// 		n1, _ := strconv.Atoi(v)
// 		if i < len(parts2) {
// 			if n2, _ := strconv.Atoi(parts2[i]); n1 != n2 {
// 				return n1 < n2
// 			}
// 		} else if n1 != 0 {
// 			// a has some extra trailing tokens.
// 			// if these are all zeroes, that's ok.
// 			return false
// 		}
// 	}

// 	// b has some extra trailing tokens.
// 	// if these are all zeroes, that's ok.
// 	for i := len(parts1); i < len(parts2); i++ {
// 		if n, _ := strconv.Atoi(parts2[i]); n != 0 {
// 			return true
// 		}
// 	}

// 	return false
// }

/*
Returns a copy of this instance, also copying each SegmentInfo.
*/
func (sis *SegmentInfos) Clone() *SegmentInfos {
	return sis.clone(false)
}

func (sis *SegmentInfos) clone(cloneSegmentInfo bool) *SegmentInfos {
	clone := &SegmentInfos{
		counter:        sis.counter,
		version:        sis.version,
		generation:     sis.generation,
		lastGeneration: sis.lastGeneration,
		userData:       make(map[string]string),
		Segments:       nil,
	}
	for _, info := range sis.Segments {
		assert(info.Info.Codec() != nil)
		clone.Segments = append(clone.Segments, info.CloneDeep(cloneSegmentInfo))
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

func (sis *SegmentInfos) toString(dir store.Directory) string {
	var buffer bytes.Buffer
	buffer.WriteString(sis.SegmentsFileName())
	buffer.WriteString(":")
	for _, info := range sis.Segments {
		buffer.WriteString(info.StringOf(dir, 0))
	}
	return buffer.String()
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
		assert(info.Info.Dir == dir)
		// if info.Info.dir == dir {
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

func (sis *SegmentInfos) finishCommit(dir store.Directory) (fileName string, err error) {
	assert(dir != nil)
	assert2(sis.pendingSegnOutput != nil, "prepareCommit was not called")
	if err = func() error {
		var success = false
		defer func() {
			if !success {
				// Closes pendingSegnOutput & delets partial segments_N:
				sis.rollbackCommit(dir)
			} else {
				err := func() error {
					var success = false
					defer func() {
						if !success {
							// Closes pendingSegnOutput & delets partial segments_N:
							sis.rollbackCommit(dir)
						} else {
							sis.pendingSegnOutput = nil
						}
					}()

					err := sis.pendingSegnOutput.Close()
					success = err == nil
					return err
				}()
				assertn(err == nil, "%v", err) // no shadow
			}
		}()

		if err := codec.WriteFooter(sis.pendingSegnOutput); err != nil {
			return err
		}
		success = true
		return nil
	}(); err != nil {
		return
	}

	// NOTE: if we crash here, we have left a segments_N file in the
	// directory in a possibly corrupt state (if some bytes made it to
	// stable storage and others didn't). But, the segments_N file
	// includes checksum at the end, which should catch this case. So
	// when a reader tries to read it, it will return an index corrupt
	// error, which should cause the retry logic in SegmentInfos to
	// kick in and load the last good (previous) segments_N-1 file.

	fileName = util.FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", sis.generation)
	if err = func() error {
		var success = false
		defer func() {
			if !success {
				dir.DeleteFile(fileName)
				// suppress error so we keep returning the original error
			}
		}()

		err := dir.Sync([]string{fileName})
		success = err == nil
		return err
	}(); err != nil {
		return
	}

	sis.lastGeneration = sis.generation
	writeSegmentsGen(dir, sis.generation)
	return
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

func (sis *SegmentInfos) changed() {
	sis.version++
}

func (sis *SegmentInfos) createBackupSegmentInfos() []*SegmentCommitInfo {
	ans := make([]*SegmentCommitInfo, len(sis.Segments))
	for i, info := range sis.Segments {
		assert(info.Info.Codec() != nil)
		ans[i] = info.Clone()
	}
	return ans
}

// L1104
func (sis *SegmentInfos) rollbackSegmentInfos(infos []*SegmentCommitInfo) {
	if cap(sis.Segments) < len(infos) {
		sis.Segments = make([]*SegmentCommitInfo, len(infos))
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
Remove the provided SegmentCommitInfo.

WARNING: O(N) cost
*/
func (sis *SegmentInfos) remove(si *SegmentCommitInfo) {
	panic("not implemented yet")
}
