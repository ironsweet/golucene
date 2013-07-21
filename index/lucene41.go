package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"math"
)

const (
	LUCENE41_DOC_EXTENSION = "doc"
	LUCENE41_POS_EXTENSION = "pos"
	LUCENE41_PAY_EXTENSION = "pay"

	LUCENE41_BLOCK_SIZE = 128

	LUCENE41_TERMS_CODEC = "Lucene41PostingsWriterTerms"
	LUCENE41_DOC_CODEC   = "Lucene41PostingsWriterDoc"
	LUCENE41_POS_CODEC   = "Lucene41PostingsWriterPos"
	LUCENE41_PAY_CODEC   = "Lucene41PostingsWriterPay"

	LUCENE41_VERSION_START   = 0
	LUCENE41_VERSION_CURRENT = LUCENE41_VERSION_START
)

type Lucene41PostingReader struct {
	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	forUtil ForUtil
}

func NewLucene41PostingReader(dir *store.Directory, fis FieldInfos, si SegmentInfo,
	ctx store.IOContext, segmentSuffix string) (r PostingsReaderBase, err error) {
	success := false
	var docIn, posIn, payIn store.IndexInput = nil, nil, nil
	defer func() {
		util.CloseWhileSuppressingError(docIn, posIn, payIn)
	}()

	docIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_DOC_EXTENSION), ctx)
	if err != nil {
		return r, err
	}
	codec.CheckHeader(docIn, LUCENE41_DOC_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
	forUtil, err := NewForUtil(docIn)
	if err != nil {
		return r, err
	}

	if fis.hasProx {
		posIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_POS_EXTENSION), ctx)
		if err != nil {
			return r, err
		}
		codec.CheckHeader(posIn, LUCENE41_POS_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)

		if fis.hasPayloads || fis.hasOffsets {
			payIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			if err != nil {
				return r, err
			}
			codec.CheckHeader(payIn, LUCENE41_PAY_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
		}
	}

	return &Lucene41PostingReader{docIn, posIn, payIn, forUtil}, nil
}

func (r *Lucene41PostingReader) init(termsIn store.IndexInput) error {
	// Make sure we are talking to the matching postings writer
	_, err := codec.CheckHeader(termsIn, LUCENE41_TERMS_CODEC, LUCENE41_VERSION_START, LUCENE41_VERSION_CURRENT)
	if err != nil {
		return err
	}
	indexBlockSize, err := termsIn.ReadVInt()
	if err != nil {
		return err
	}
	if indexBlockSize != LUCENE41_BLOCK_SIZE {
		panic(fmt.Sprintf("index-time BLOCK_SIZE (%v) != read-time BLOCK_SIZE (%v)", indexBlockSize, LUCENE41_BLOCK_SIZE))
	}
	return nil
}

func (r *Lucene41PostingReader) Close() error {
	return util.Close(r.docIn, r.posIn, r.payIn)
}

type ForUtil struct {
	encodedSizes []int32
	encoders     []util.PackedIntsEncoder
	decoders     []util.PackedIntsDecoder
	iterations   []int32
}

type DataInput interface {
	ReadVInt() (n int32, err error)
}

func NewForUtil(in DataInput) (fu ForUtil, err error) {
	self := ForUtil{}
	packedIntsVersion, err := in.ReadVInt()
	if err != nil {
		return self, err
	}
	util.CheckVersion(packedIntsVersion)
	self.encodedSizes = make([]int32, 33)
	self.encoders = make([]util.PackedIntsEncoder, 33)
	self.decoders = make([]util.PackedIntsDecoder, 33)
	self.iterations = make([]int32, 33)

	for bpv := 1; bpv <= 32; bpv++ {
		code, err := in.ReadVInt()
		if err != nil {
			return self, err
		}
		formatId := uint32(code) >> 5
		bitsPerValue := (uint32(code) & 31) + 1

		format := util.PackedFormat(formatId)
		// assert format.isSupported(bitsPerValue)
		self.encodedSizes[bpv] = encodedSize(format, packedIntsVersion, bitsPerValue)
		self.encoders[bpv] = util.GetPackedIntsEncoder(format, packedIntsVersion, bitsPerValue)
		self.decoders[bpv] = util.GetPackedIntsDecoder(format, packedIntsVersion, bitsPerValue)
		self.iterations[bpv] = computeIterations(self.decoders[bpv])
	}
	return self, nil
}

func encodedSize(format util.PackedFormat, packedIntsVersion int32, bitsPerValue uint32) int32 {
	byteCount := format.ByteCount(packedIntsVersion, LUCENE41_BLOCK_SIZE, bitsPerValue)
	// assert byteCount >= 0 && byteCount <= math.MaxInt32()
	return int32(byteCount)
}

func computeIterations(decoder util.PackedIntsDecoder) int32 {
	return int32(math.Ceil(float64(LUCENE41_BLOCK_SIZE) / float64(decoder.ByteValueCount())))
}
