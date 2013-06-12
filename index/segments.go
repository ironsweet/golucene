package index

import (
	"fmt"
	"lucene/store"
	"strconv"
	"strings"
)

type FindSegmentsFile struct {
	directory store.Directory
}

func NewFindSegmentsFile(directory store.Directory) {
	return &FindSegmentsFile{directory}
}

const (
	INDEX_FILENAME_SEGMENTS = "segments"
	VERSION_40              = 0
)

type SegmentInfos struct {
	counter        int
	version        int64
	generation     int64
	lastGeneration int64
	userData       map[string]string
}

func GenerationFromSegmentsFileName(fileName string) int {
	switch {
	case fileName == INDEX_FILENAME_SEGMENTS:
		return 0
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

func (sis *SegmentInfos) Read(directory store.Directory, segmentFileName string) {
	success := false

	// Clear any previous segments:
	sis.clear()

	sis.generation = GenerationFromSegmentsFileName(segmentFileName)
	sis.lastGeneration = sis.generation

	input := NewChecksumIndexInput(directory.OpenInput(segmentFileName, store.IO_CONTEXT_READ))
	defer func() {
		if !success {
			// Clear any segment infos we had loaded so we
			// have a clean slate on retry:
			sis.Clear()
			util.CloseWhileHandlingException(input)
		} else {
			input.Close()
		}
	}()

	format := input.ReadInt()
	if format == codec.CODEC_MAGIC {
		// 4.0+
		codec.CheckHeaderNoMagic(input, "segments", VERSION_40, VERSION_40)
		sis.version = input.ReadLong()
		sis.counter = input.ReadInt()
		numSegments := input.ReadInt()
		if numSegments < 0 {
			return nil, errors.New(fmt.Sprintf("invalid segment count: %v (resource: %v)", numSegments, input))
		}
		for seg := 0; seg < numSegments; seg++ {
			segName := input.ReadString()
			method := codec.ForName(input.ReadString())
			info := method.SegmentInfoFormat().Reader().Read(directory, segName, store.IO_CONTEXT_READ)
			info.Codec = method
			delGen := input.ReadLong()
			delCount := input.ReadInt()
			if delCount < 0 || delCount > info.DocCount() {
				return nil, errors.New("invalid deletion count: %v (resource: %v)", delCount, input)
			}
			sis.Add(NewSegmentInfoPerCommit(info, delCount, delGen))
		}
		sis.userData = input.readStringStringMap()
	} else {
		// TODO support <4.0 index
		panic("not supported yet")
	}

	if checksumNow, checksumThen = input.Checksum(), input.ReadLong(); checksumNow != checksumThen {
		return nil, errors.New(fmt.Sprintf("checksum mismatch in segments file (resource: %v)", input))
	}

	success = true
}

func (sis *SegmentInfos) Clear() {
	sis.segments.Clear()
}
