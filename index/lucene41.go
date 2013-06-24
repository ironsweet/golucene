package index

import (
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
)

const (
	LUCENE41_DOC_EXTENSION   = "doc"
	LUCENE41_POS_EXTENSION   = "pos"
	LUCENE41_PAY_EXTENSION   = "pay"
	LUCENE41_DOC_CODEC       = "Lucene41PostingsWriterDoc"
	LUCENE41_POS_CODEC       = "Lucene41PostingsWriterPos"
	LUCENE41_PAY_CODEC       = "Lucene41PostingsWriterPay"
	LUCENE41_VERSION_START   = 0
	LUCENE41_VERSION_CURRENT = LUCENE41_VERSION_START
)

type Lucene41PostingReader struct {
	docIn   *store.IndexInput
	posIn   *store.IndexInput
	payIn   *store.IndexInput
	forUtil ForUtil
}

func NewLucene41PostingReader(dir *store.Directory, fis FieldInfos, si SegmentInfo,
	ctx store.IOContext, segmentSuffix string) PostingsReaderBase {
	success := false
	var docIn, posIn, payIn *store.IndexInput = nil, nil, nil
	defer func() {
		util.CloseWhileSupressingError(docIn, posIn, payIn)
	}()

	docIn = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_DOC_EXTENSION), ctx)
	store.CheckHeader(docIn, LUCENE41_DOC_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
	forUtil := NewForUtil(docIn)

	if fis.hasProx {
		posIn = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_POS_EXTENSION), ctx)
		store.CheckHeader(posIn, LUCENE41_POS_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)

		if fis.hasPayloads || fis.hasOffsets {
			payIn = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			store.CheckHeader(payIn, LUCENE41_PAY_CODED, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
		}
	}

	return Lucene41PostingReader{docIn, posIn, payIn, forUtil}
}
