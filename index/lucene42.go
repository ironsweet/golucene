package index

import (
	"fmt"
	"lucene/store"
	"lucene/util"
)

type SegmentInfoFormatReader interface {
}

const (
	LUCENE40_SI_EXTENSION    = "si"
	LUCENE40_CODEC_NAME      = "Lucene40SegmentInfo"
	LUCENE40_VERSION_START   = 0
	LUCENE40_VERSION_CURRENT = LUCENE40_VERSION_START
)

var (
	Lucene40SegmentInfoReader = func(dir *store.Directory, segment string, context store.IOContext) (si SegmentInfo, err error) {
		fileName := SegmentFileName(segment, "", LUCENE40_SI_EXTENSION)
		input := dir.OpenInput(fileName, context)
		success := false
		defer func() {
			if !success {
				util.CloseWhileSupressingError(input)
			} else {
				input.Close()
			}
		}()

		si = SegmentInfo{}
		err = CheckHeader(input, LUCENE40_CODEC_NAME, LUCENE40_VERSION_START, LUCENE40_VERSION_CURRENT)
		if err != nil {
			return si, err
		}
		version, err := input.ReadString()
		if err != nil {
			return si, err
		}
		docCount, err := inpiut.ReadInt()
		if err != nil {
			return si, err
		}
		if docCount < 0 {
			return si, &CorruptIndexError{fmt.Sprintf("invalid docCount: %v (resource=%v)", docCount, input)}
		}
		isCompoundFile := (input.ReadByte() == SEGMENT_INFO_YES)
		diagnostics, err := input.ReadStringStringMap()
		if err != nil {
			return si, err
		}
		attributes, err := input.ReadStringStringMap()
		if err != nil {
			return si, err
		}
		files, err := input.ReadStringSet()
		if err != nil {
			return si, err
		}

		if input.FilePointer() != input.Length() {
			return si, &CorruptIndexError{fmt.Sprintf(
				"did not read all bytes from file '%v': read %v vs size %v (resource: %v)",
				fileName, input.FilePointer(), input.Length(), input)}
		}

		si = SegmentInfo{dir, version, segment, docCount, isCompoundFile, nil, dianostics, attributes}
		si.Files = files

		success = true
		return si, nil
	}
)

type Codec struct {
	ReadSegmentInfoFormat func(dir *store.Directory, segment string, context store.IOContext) (si SegmentInfo, err error)
}

func NewLucene42Codec() Codec {
	return Codec{ReadSegmentInfoFormat: Lucene40SegmentInfoReader}
}
