package index

import (
	"github.com/balzaczyy/golucene/util"
	"math"
)

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
