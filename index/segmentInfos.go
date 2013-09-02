package index

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
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
		if strings.HasPrefix(file, INDEX_FILENAME_SEGMENTS) && file != INDEX_FILENAME_SEGMENTS_GEN {
			gen := GenerationFromSegmentsFileName(file)
			if gen > max {
				max = gen
			}
		}
	}
	return max
}

func (sis SegmentInfos) SegmentsFileName() string {
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
			// method := CodecForName(codecName)
			method := NewLucene42Codec()
			info, err := method.ReadSegmentInfo(directory, segName, store.IO_CONTEXT_READ)
			if err != nil {
				return err
			}
			info.codec = method
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

func (sis *SegmentInfos) Clear() {
	sis.Segments = make([]SegmentInfoPerCommit, 0)
}
