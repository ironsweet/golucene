package packed

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

const (
	PACKED64_BLOCK_SIZE = 64                      // 32 = int, 64 = long
	PACKED64_BLOCK_BITS = 6                       // The #bits representing BLOCK_SIZE
	PACKED64_MOD_MASK   = PACKED64_BLOCK_SIZE - 1 // x % BLOCK_SIZE
)

type Packed64 struct {
	*MutableImpl
	blocks            []int64
	maskRight         uint64
	bpvMinusBlockSize int32
}

func newPacked64(valueCount int32, bitsPerValue uint32) *Packed64 {
	longCount := PackedFormat(PACKED).longCount(VERSION_CURRENT, valueCount, bitsPerValue)
	ans := &Packed64{
		blocks:            make([]int64, longCount),
		maskRight:         uint64(^(int64(0))<<(PACKED64_BLOCK_SIZE-bitsPerValue)) >> (PACKED64_BLOCK_SIZE - bitsPerValue),
		bpvMinusBlockSize: int32(bitsPerValue) - PACKED64_BLOCK_SIZE}
	ans.MutableImpl = newMutableImpl(ans, int(valueCount), int(bitsPerValue))
	return ans
}

func newPacked64FromInput(version int32, in DataInput, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
	ans := newPacked64(valueCount, bitsPerValue)
	byteCount := PackedFormat(PACKED).ByteCount(version, valueCount, bitsPerValue)
	longCount := PackedFormat(PACKED).longCount(VERSION_CURRENT, valueCount, bitsPerValue)
	ans.blocks = make([]int64, longCount)
	// read as many longs as we can
	for i := int64(0); i < byteCount/8; i++ {
		if ans.blocks[i], err = in.ReadLong(); err != nil {
			break
		}
	}
	if err == nil {
		if remaining := int8(byteCount % 8); remaining != 0 {
			// read the last bytes
			var lastLong int64
			for i := int8(0); i < remaining; i++ {
				b, err := in.ReadByte()
				if err != nil {
					break
				}
				lastLong |= (int64(b) << uint8(56-i*8))
			}
			if err == nil {
				ans.blocks[len(ans.blocks)-1] = lastLong
			}
		}
	}
	return ans, err
}

func (p *Packed64) Get(index int) int64 {
	// The abstract index in a bit stream
	majorBitPos := int64(index) * int64(p.bitsPerValue)
	// The index in the backing long-array
	elementPos := int32(uint64(majorBitPos) >> PACKED64_BLOCK_BITS)
	// The number of value-bits in the second long
	endBits := (majorBitPos & PACKED64_MOD_MASK) + int64(p.bpvMinusBlockSize)

	if endBits <= 0 { // Single block
		return int64(uint64(p.blocks[elementPos])>>uint64(-endBits)) & int64(p.maskRight)
	}
	// Two blocks
	return ((p.blocks[elementPos] << uint64(endBits)) |
		int64(uint64(p.blocks[elementPos+1])>>(PACKED64_BLOCK_SIZE-uint64(endBits)))) &
		int64(p.maskRight)
}

func (p *Packed64) getBulk(index int, arr []int64) int {
	off, length := 0, len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < p.valueCount)
	if p.valueCount-index < length {
		length = p.valueCount - index
	}

	originalIndex := index
	decoder := newBulkOperation(PackedFormat(PACKED), uint32(p.bitsPerValue))

	// go to the next block where the value does not span across two blocks
	offsetInBlocks := index % decoder.LongValueCount()
	if offsetInBlocks != 0 {
		panic("niy")
	}

	// bulk get
	assert(index%decoder.LongValueCount() == 0)
	blockIndex := int(uint64(int64(index)*int64(p.bitsPerValue)) >> PACKED64_BLOCK_BITS)
	assert(((int64(index) * int64(p.bitsPerValue)) & PACKED64_MOD_MASK) == 0)
	iterations := length / decoder.LongValueCount()
	decoder.decodeLongToLong(p.blocks[blockIndex:], arr[off:], iterations)
	gotValues := iterations * decoder.LongValueCount()
	index += gotValues
	length -= gotValues
	assert(length >= 0)

	if index > originalIndex {
		// stay at the block boundry
		return index - originalIndex
	}
	// no progress so far => already at a block boundry but no full block to get
	assert(index == originalIndex)
	return p.MutableImpl.getBulk(index, arr[off:off+length])
}

func (p *Packed64) Set(index int, value int64) {
	// The abstract index in a contiguous bit stream
	majorBitPos := int64(index) * int64(p.bitsPerValue)
	// The index int the backing long-array
	elementPos := int(uint64(majorBitPos) >> PACKED64_BLOCK_BITS) // / BLOCK_SIZE
	// The number of value-bits in the second long
	endBits := (majorBitPos & PACKED64_MOD_MASK) + int64(p.bpvMinusBlockSize)

	if endBits <= 0 { // single block
		p.blocks[elementPos] = p.blocks[elementPos] &
			^int64(p.maskRight<<uint64(-endBits)) |
			(value << uint64(-endBits))
		return
	}
	// two blocks
	p.blocks[elementPos] = p.blocks[elementPos] &
		^int64(p.maskRight>>uint64(endBits)) |
		int64(uint64(value)>>uint64(endBits))
	elementPos++
	p.blocks[elementPos] = p.blocks[elementPos]&
		int64(^uint64(0)>>uint64(endBits)) |
		(value << uint64(PACKED64_BLOCK_SIZE-endBits))
}

func (p *Packed64) setBulk(index int, arr []int64) int {
	off, length := 0, len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < p.valueCount)
	if p.valueCount-index < length {
		length = p.valueCount - index
	}

	originalIndex := index
	encoder := newBulkOperation(PackedFormat(PACKED), uint32(p.bitsPerValue))

	// go to the next block where the value does not span across two blocks
	offsetInBlocks := index % encoder.LongValueCount()
	if offsetInBlocks != 0 {
		panic("niy")
	}

	// bulk set
	assert(index%encoder.LongValueCount() == 0)
	blockIndex := int(uint64(int64(index)*int64(p.bitsPerValue)) >> PACKED64_BLOCK_BITS)
	assert(((int64(index) * int64(p.bitsPerValue)) & PACKED64_MOD_MASK) == 0)
	iterations := length / encoder.LongValueCount()
	encoder.encodeLongToLong(arr[off:], p.blocks[blockIndex:], iterations)
	setValues := iterations * encoder.LongValueCount()
	index += setValues
	length -= setValues
	assert(length >= 0)

	if index > originalIndex {
		// stay at the block boundry
		return index - originalIndex
	}
	// no progress so far => already at a block boundry but no full block to get
	assert(index == originalIndex)
	return p.MutableImpl.setBulk(index, arr[off:off+length])
}

func (p *Packed64) String() string {
	return fmt.Sprintf("Packed64(bitsPerValue=%v, size=%v, elements.length=%v)",
		p.bitsPerValue, p.Size(), len(p.blocks))
}

func (p *Packed64) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			3*util.NUM_BYTES_INT +
			util.NUM_BYTES_LONG +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(p.blocks))
}

func (p *Packed64) fill(from, to int, val int64) {
	panic("niy")
}

func (p *Packed64) Clear() {
	panic("niy")
}
