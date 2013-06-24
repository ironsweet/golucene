package index

import (
	"errors"
	"github.com/balzaczyy/golucene/store"
)

const (
	LUCENE40_SI_EXTENSION    = "si"
	LUCENE40_CODEC_NAME      = "Lucene40SegmentInfo"
	LUCENE40_VERSION_START   = 0
	LUCENE40_VERSION_CURRENT = LUCENE40_VERSION_START

	SEGMENT_INFO_YES = 1
)

var (
	Lucene40SegmentInfoReader = func(dir *store.Directory, segment string, context store.IOContext) (si SegmentInfo, err error) {
		si = SegmentInfo{}
		fileName := SegmentFileName(segment, "", LUCENE40_SI_EXTENSION)
		input, err := dir.OpenInput(fileName, context)
		if err != nil {
			return si, err
		}

		success := false
		defer func() {
			if !success {
				util.CloseWhileSupressingException(input)
			} else {
				input.Close()
			}
		}()

		_, err = CheckHeader(input.DataInput, LUCENE40_CODEC_NAME, LUCENE40_VERSION_START, LUCENE40_VERSION_CURRENT)
		if err != nil {
			return si, err
		}
		version, err := input.ReadString()
		if err != nil {
			return si, err
		}
		docCount, err := input.ReadInt()
		if err != nil {
			return si, err
		}
		if docCount < 0 {
			return si, errors.New(fmt.Sprintf("invalid docCount: %v (resource=%v)", docCount, input))
		}
		sicf, err := input.ReadByte()
		if err != nil {
			return si, err
		}
		isCompoundFile := (sicf == SEGMENT_INFO_YES)
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
			return si, errors.New(fmt.Sprintf(
				"did not read all bytes from file '%v': read %v vs size %v (resource: %v)",
				fileName, input.FilePointer(), input.Length(), input))
		}

		si = SegmentInfo{dir, version, segment, docCount, isCompoundFile, Codec{}, diagnostics, attributes, nil}
		si.CheckFileNames(files)
		si.Files = files

		success = true
		return si, nil
	}
)
