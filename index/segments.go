package index

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"lucene/store"
	"lucene/util"
	"strconv"
	"strings"
)

type FindSegmentsFile struct {
	directory                *store.Directory
	doBody                   func(segmentFileName string) (obj interface{}, err error)
	defaultGenLookaheadCount int
}

func NewFindSegmentsFile(directory *store.Directory,
	doBody func(segmentFileName string) (obj interface{}, err error)) *FindSegmentsFile {
	return &FindSegmentsFile{directory, doBody, 10}
}

// TODO support IndexCommit
func (fsf *FindSegmentsFile) run() (obj interface{}, err error) {
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
		if useFirstMethod {
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

				version, err := genInput.ReadInt()
				if _, ok := err.(*CorruptIndexError); ok {
					return nil, err
				}
				if version == FORMAT_SEGMENTS_GEN_CURRENT {
					gen0, err := genInput.ReadLong()
					if err != nil {
						if _, ok := err.(*CorruptIndexError); ok {
							return nil, err
						}
					} else {
						gen1, err := genInput.ReadLong()
						if err != nil {
							if _, ok := err.(*CorruptIndexError); ok {
								return nil, err
							}
						} else {
							// if fsf.infoStream != nil {
							log.Printf("fallback check: %v; %v", gen0, gen1)
							// }
							if gen0 == gen1 {
								// The file is consistent.
								genB = gen0
							}
						}
					}
				} else {
					return nil, NewIndexFormatTooNewError(genInput.DataInput, version, FORMAT_SEGMENTS_GEN_CURRENT, FORMAT_SEGMENTS_GEN_CURRENT)
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
		segmentFileName := FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", gen)

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
				prevSegmentFileName := FileNameFromGeneration(INDEX_FILENAME_SEGMENTS, "", gen-1)

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

func FileNameFromGeneration(base, ext string, gen int64) string {
	switch {
	case gen == -1:
		return ""
	case gen == 0:
		return SegmentFileName(base, "", ext)
	default:
		// assert gen > 0
		// The '6' part in the length is: 1 for '.', 1 for '_' and 4 as estimate
		// to the gen length as string (hopefully an upper limit so SB won't
		// expand in the middle.
		var buffer bytes.Buffer
		fmt.Fprintf(&buffer, "%v_%v", base, strconv.FormatInt(gen, 36))
		if len(ext) > 0 {
			buffer.WriteString(".")
			buffer.WriteString(ext)
		}
		return buffer.String()
	}
}

func SegmentFileName(name, suffix, ext string) string {
	if len(ext) > 0 || len(suffix) > 0 {
		// assert ext[0] != '.'
		var buffer bytes.Buffer
		buffer.WriteString(name)
		if len(suffix) > 0 {
			buffer.WriteString("_")
			buffer.WriteString(suffix)
		}
		if len(ext) > 0 {
			buffer.WriteString(".")
			buffer.WriteString(ext)
		}
		return buffer.String()
	}
	return name
}

const (
	INDEX_FILENAME_SEGMENTS     = "segments"
	INDEX_FILENAME_SEGMENTS_GEN = "segments.gen"
	VERSION_40                  = 0
	FORMAT_SEGMENTS_GEN_CURRENT = -2
)

type SegmentInfo struct {
}

type SegmentInfos struct {
	counter        int
	version        int64
	generation     int64
	lastGeneration int64
	userData       map[string]string
	Segments       []SegmentInfoPerCommit
}

func LastCommitGeneration(files []string) int64 {
	if files == nil {
		return int64(-1)
	}
	max := int64(-1)
	for _, file := range files {
		if strings.HasPrefix(file, INDEX_FILENAME_SEGMENTS) && file != INDEX_FILENAME_SEGMENTS {
			gen := GenerationFromSegmentsFileName(file)
			if gen > max {
				max = gen
			}
		}
	}
	return max
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

func (sis *SegmentInfos) Read(directory *store.Directory, segmentFileName string) error {
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
			util.CloseWhileSupressingException(input)
		} else {
			input.Close()
		}
	}()

	format, err := input.ReadInt()
	if err != nil {
		return err
	}
	if format == CODEC_MAGIC {
		// 4.0+
		CheckHeaderNoMagic(input.DataInput, "segments", VERSION_40, VERSION_40)
		sis.version, err = input.ReadLong()
		if err != nil {
			return err
		}
		sis.counter, err = input.ReadInt()
		if err != nil {
			return err
		}
		numSegments, err := input.ReadInt()
		if err != nil {
			return err
		}
		if numSegments < 0 {
			return &CorruptIndexError{fmt.Sprintf("invalid segment count: %v (resource: %v)", numSegments, input)}
		}
		for seg := 0; seg < numSegments; seg++ {
			segName, err := input.ReadString()
			if err != nil {
				return err
			}
			codecName, err := input.ReadString()
			if err != nil {
				return err
			} else if codecName != "lucene42" {
				log.Panicf("Not supported yet: %v", codecName)
			}
			// method := CodecForName(codecName)
			method := NewLucene42Codec()
			info := method.SegmentInfoFormat().Reader().Read(directory, segName, store.IO_CONTEXT_READ)
			info.Codec = method
			delGen, err := input.ReadLong()
			if err != nil {
				return err
			}
			delCount, err := input.ReadInt()
			if err != nil {
				return err
			}
			if delCount < 0 || delCount > info.DocCount() {
				return &CorruptIndexError{fmt.Sprintf("invalid deletion count: %v (resource: %v)", delCount, input)}
			}
			sis.Add(NewSegmentInfoPerCommit(info, delCount, delGen))
		}
		sis.userData = input.readStringStringMap()
	} else {
		// TODO support <4.0 index
		panic("not supported yet")
	}

	if checksumNow, checksumThen = input.Checksum(), input.ReadLong(); checksumNow != checksumThen {
		return &CorruptIndexError{fmt.Sprintf("checksum mismatch in segments file (resource: %v)", input)}
	}

	success = true
	return nil
}

func (sis *SegmentInfos) Clear() {
	sis.segments = make([]SegmentInfoPerCommit, 0)
}

type SegmentInfoPerCommit struct {
}

type SegmentReader struct {
	*AtomicReader
	si       SegmentInfoPerCommit
	liveDocs *util.Bits
	numDocs  int
	core     SegmentCoreReaders
}

func NewSegmentReader(si SegmentInfoPerCommit, termInfosIndexDivisor int, context store.IOContext) *SegmentReader {
	r := &SegmentReader{}
	r.AtomicReader = newAtomicReader(r)
	r.si = si
	r.core = NewSegmentCoreReaders(r, si.info.dir, si, context, termInfosIndexDivisor)
	success := false
	defer func() {
		// With lock-less commits, it's entirely possible (and
		// fine) to hit a FileNotFound exception above.  In
		// this case, we want to explicitly close any subset
		// of things that were opened so that we don't have to
		// wait for a GC to do so.
		if !success {
			core.decRef()
		}
	}()

	if si.hasDeletions() {
		panic("not supported yet")
	} else {
		// assert si.getDelCount() == 0
		r.liveDocs = nil
	}
	r.numDocs = si.info.DocCount() - si.DelCount()
	success = true
}
