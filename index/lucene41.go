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

	LUCENE41_BLOCK_SIZE = 128
)

type Lucene41PostingReader struct {
	docIn   *store.IndexInput
	posIn   *store.IndexInput
	payIn   *store.IndexInput
	forUtil ForUtil
}

func NewLucene41PostingReader(dir *store.Directory, fis FieldInfos, si SegmentInfo,
	ctx store.IOContext, segmentSuffix string) (r PostingsReaderBase, err error) {
	success := false
	var docIn, posIn, payIn *store.IndexInput = nil, nil, nil
	defer func() {
		util.CloseWhileSupressingError(docIn, posIn, payIn)
	}()

	docIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_DOC_EXTENSION), ctx)
	if err != nil {
		return r, err
	}
	store.CheckHeader(docIn.DataInput, LUCENE41_DOC_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
	forUtil, err := NewForUtil(docIn.DataInput)
	if err != nil {
		return r, err
	}

	if fis.hasProx {
		posIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_POS_EXTENSION), ctx)
		if err != nil {
			return r, err
		}
		store.CheckHeader(posIn.DataInput, LUCENE41_POS_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)

		if fis.hasPayloads || fis.hasOffsets {
			payIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			if err != nil {
				return r, err
			}
			store.CheckHeader(payIn.DataInput, LUCENE41_PAY_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
		}
	}

	return &Lucene41PostingReader{docIn, posIn, payIn, forUtil}, nil
}

func (r *Lucene41PostingReader) Close() error {
	return util.Close(r.docIn, r.posIn, r.payIn)
}

type ForUtil struct {
	encodedSizes []int
	encoders     []util.PackedIntsEncoder
	decoders     []util.PackedIntsDecoder
	iterations   []int
}

func NewForUtil(in *store.DataInput) (fu ForUtil, err error) {
	self := ForUtil{}
	packedIntsVersion, err := in.ReadVInt()
	if err != nil {
		return self, err
	}
	util.CheckVersion(packedIntsVersion)
	self.encodedSizes = make([]int, 33)
	self.encoders = make([]util.PackedIntsEncoder, 33)
	self.decoders = make([]util.PackedIntsDecoder, 33)
	self.iterations = make([]int, 33)

	for bpv := 1; bpv <= 32; bpv++ {
		code, err := in.ReadVInt()
		if err != nil {
			return self, err
		}
		formatId := uint32(code) >> 5
		bitsPerValue := (code & 31) + 1

		format := util.PackedFormat(formatId)
		// assert format.isSupported(bitsPerValue)
		self.encodedSizes[bpv] = encodedSize(format, packedIntsVersion, bitsPerValue)
		self.encoders[bpv] = getPackedIntsEncoder(format, packedIntsVersion, bitsPerValue)
		self.docoders[bpv] = getPackedIntsDecoder(format, packedIntsVersion, bitsPerValue)
		self.iterations[bpv] = computeIterations(self.decoders[bpv])
	}
	return self, nil
}

func encodedSize(format PackedFormat, packedIntsVersion, bitsPerValue int) int {
	byteCount := format.byteCount(packedIntsVersion, LUCENE41_BLOCK_SIZE, bitsPerValue)
	// assert byteCount >= 0 && byteCount <= math.MaxInt32()
	return int(byteCount)
}
