package lucene40

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// lucene40/Lucene40SegmentInfoReader.java

type Lucene40SegmentInfoReader struct{}

func (r *Lucene40SegmentInfoReader) Read(dir store.Directory,
	segment string, context store.IOContext) (si *SegmentInfo, err error) {

	si = new(SegmentInfo)
	fileName := util.SegmentFileName(segment, "", LUCENE40_SI_EXTENSION)
	input, err := dir.OpenInput(fileName, context)
	if err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(input)
		} else {
			input.Close()
		}
	}()

	_, err = codec.CheckHeader(input, LUCENE40_CODEC_NAME, LUCENE40_VERSION_START, LUCENE40_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	versionStr, err := input.ReadString()
	if err != nil {
		return nil, err
	}
	version, err := util.ParseVersion(versionStr)
	if err != nil {
		return nil, err
	}

	docCount, err := input.ReadInt()
	if err != nil {
		return nil, err
	}
	if docCount < 0 {
		return nil, errors.New(fmt.Sprintf("invalid docCount: %v (resource=%v)", docCount, input))
	}
	sicf, err := input.ReadByte()
	if err != nil {
		return nil, err
	}
	isCompoundFile := (sicf == SEGMENT_INFO_YES)
	diagnostics, err := input.ReadStringStringMap()
	if err != nil {
		return nil, err
	}
	_, err = input.ReadStringStringMap() // read deprecated attributes
	if err != nil {
		return nil, err
	}
	files, err := input.ReadStringSet()
	if err != nil {
		return nil, err
	}

	if err = codec.CheckEOF(input); err != nil {
		return nil, err
	}

	si = NewSegmentInfo(dir, version, segment, int(docCount), isCompoundFile, nil, diagnostics)
	si.SetFiles(files)

	success = true
	return si, nil
}
