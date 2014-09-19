package lucene41

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"math"
)

// codecs/lucene41/ForUtil.java

const (
	/**
	 * Special number of bits per value used whenever all values to encode are equal.
	 */
	ALL_VALUES_EQUAL = 0
	/**
	 * Upper limit of the number of bytes that might be required to stored
	 * <code>BLOCK_SIZE</code> encoded values.
	 */
	MAX_ENCODED_SIZE = LUCENE41_BLOCK_SIZE * 4
)

/**
 * Upper limit of the number of values that might be decoded in a single call to
 * {@link #readBlock(IndexInput, byte[], int[])}. Although values after
 * <code>BLOCK_SIZE</code> are garbage, it is necessary to allocate value buffers
 * whose size is >= MAX_DATA_SIZE to avoid {@link ArrayIndexOutOfBoundsException}s.
 */
var MAX_DATA_SIZE int = computeMaxDataSize()

func computeMaxDataSize() int {
	maxDataSize := 0
	// for each version
	for version := packed.PACKED_VERSION_START; version <= packed.VERSION_CURRENT; version++ {
		// for each packed format
		for format := packed.PACKED; format <= packed.PACKED_SINGLE_BLOCK; format++ {
			// for each bit-per-value
			for bpv := 1; bpv <= 32; bpv++ {
				if !packed.PackedFormat(format).IsSupported(bpv) {
					continue
				}
				decoder := packed.GetPackedIntsDecoder(packed.PackedFormat(format), int32(version), uint32(bpv))
				iterations := int(computeIterations(decoder))
				if n := iterations * decoder.ByteValueCount(); n > maxDataSize {
					maxDataSize = n
				}
			}
		}
	}
	return maxDataSize
}

/**
 * Encode all values in normal area with fixed bit width,
 * which is determined by the max value in this block.
 */
type ForUtil struct {
	encodedSizes []int32
	encoders     []packed.PackedIntsEncoder
	decoders     []packed.PackedIntsDecoder
	iterations   []int32
}

/* Create a new ForUtil instance and save state into out. */
func NewForUtilInto(accetableOverheadRatio float32, out util.DataOutput) (*ForUtil, error) {
	ans, err := &ForUtil{}, out.WriteVInt(packed.VERSION_CURRENT)
	if err != nil {
		return ans, err
	}
	ans.encodedSizes = make([]int32, 33)
	ans.encoders = make([]packed.PackedIntsEncoder, 33)
	ans.decoders = make([]packed.PackedIntsDecoder, 33)
	ans.iterations = make([]int32, 33)

	packedIntsVersion := int32(packed.VERSION_CURRENT)
	for bpv := 1; bpv <= 32; bpv++ {
		formatAndBits := packed.FastestFormatAndBits(
			LUCENE41_BLOCK_SIZE, bpv, accetableOverheadRatio)
		format := formatAndBits.Format
		bitsPerValue := formatAndBits.BitsPerValue
		assert(format.IsSupported(bitsPerValue))
		assert(bitsPerValue <= 32)
		ans.encodedSizes[bpv] = encodedSize(format, packedIntsVersion, uint32(bitsPerValue))
		ans.encoders[bpv] = packed.GetPackedIntsEncoder(format, packedIntsVersion, uint32(bitsPerValue))
		ans.decoders[bpv] = packed.GetPackedIntsDecoder(format, packedIntsVersion, uint32(bitsPerValue))
		ans.iterations[bpv] = computeIterations(ans.decoders[bpv])

		err = out.WriteVInt(int32(format.Id()<<5 | (bitsPerValue - 1)))
		if err != nil {
			return ans, err
		}
	}
	return ans, err
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

type DataInput interface {
	ReadVInt() (n int32, err error)
}

/* Restore a ForUtil from a DataInput. */
func NewForUtilFrom(in DataInput) (fu *ForUtil, err error) {
	self := &ForUtil{}
	packedIntsVersion, err := in.ReadVInt()
	if err != nil {
		return self, err
	}
	packed.CheckVersion(packedIntsVersion)
	self.encodedSizes = make([]int32, 33)
	self.encoders = make([]packed.PackedIntsEncoder, 33)
	self.decoders = make([]packed.PackedIntsDecoder, 33)
	self.iterations = make([]int32, 33)

	for bpv := 1; bpv <= 32; bpv++ {
		code, err := in.ReadVInt()
		if err != nil {
			return self, err
		}
		formatId := uint32(code) >> 5
		bitsPerValue := (uint32(code) & 31) + 1

		format := packed.PackedFormat(formatId)
		// assert format.isSupported(bitsPerValue)
		self.encodedSizes[bpv] = encodedSize(format, packedIntsVersion, bitsPerValue)
		self.encoders[bpv] = packed.GetPackedIntsEncoder(format, packedIntsVersion, bitsPerValue)
		self.decoders[bpv] = packed.GetPackedIntsDecoder(format, packedIntsVersion, bitsPerValue)
		self.iterations[bpv] = computeIterations(self.decoders[bpv])
	}
	return self, nil
}

type IndexOutput interface {
	WriteByte(byte) error
	WriteBytes([]byte) error
	WriteVInt(int32) error
}

func (u *ForUtil) writeBlock(data []int, encoded []byte, out IndexOutput) error {
	if isAllEqual(data) {
		var err error
		if err = out.WriteByte(byte(ALL_VALUES_EQUAL)); err == nil {
			err = out.WriteVInt(int32(data[0]))
		}
		return err
	}

	numBits := bitsRequired(data)
	assert2(numBits > 0 && numBits <= 32, "%v", numBits)
	encoder := u.encoders[numBits]
	iters := int(u.iterations[numBits])
	assert(iters*encoder.ByteValueCount() >= LUCENE41_BLOCK_SIZE)
	encodedSize := int(u.encodedSizes[numBits])
	assert(iters*encoder.ByteBlockCount() >= encodedSize)

	if err := out.WriteByte(byte(numBits)); err != nil {
		return err
	}

	encoder.EncodeIntToByte(data, encoded, iters)
	return out.WriteBytes(encoded[:encodedSize])
}

func encodedSize(format packed.PackedFormat, packedIntsVersion int32, bitsPerValue uint32) int32 {
	byteCount := format.ByteCount(packedIntsVersion, LUCENE41_BLOCK_SIZE, bitsPerValue)
	// assert byteCount >= 0 && byteCount <= math.MaxInt32()
	return int32(byteCount)
}

func computeIterations(decoder packed.PackedIntsDecoder) int32 {
	return int32(math.Ceil(float64(LUCENE41_BLOCK_SIZE) / float64(decoder.ByteValueCount())))
}

func isAllEqual(data []int) bool {
	first := data[0]
	for _, v := range data[1:LUCENE41_BLOCK_SIZE] {
		if v != first {
			return false
		}
	}
	return true
}

func bitsRequired(data []int) int {
	or := int64(0)
	for _, v := range data[:LUCENE41_BLOCK_SIZE] {
		assert(v >= 0)
		or |= int64(v)
	}
	return packed.BitsRequired(or)
}
