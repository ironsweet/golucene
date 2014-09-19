package packed

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

func is64Supported(bitsPerValue int) bool {
	// Lucene use binary-search which is unnecessary
	assert(bitsPerValue > 0 && bitsPerValue <= 64)
	return bitsPerValue <= len(packedSingleBlockBulkOps) &&
		packedSingleBlockBulkOps[bitsPerValue-1] != nil
}

func requiredCapacity(valueCount, valuesPerBlock int32) int32 {
	fit := (valueCount % valuesPerBlock) != 0
	ans := valueCount / valuesPerBlock
	if fit {
		ans++
	}
	return ans
}

type Packed64SingleBlock struct {
	*MutableImpl
	get    func(int) int64
	set    func(int, int64)
	blocks []int64
}

func newPacked64SingleBlock(valueCount int32, bitsPerValue uint32) *Packed64SingleBlock {
	// assert isSupported(bitsPerValue)
	valuesPerBlock := int32(64 / bitsPerValue)
	ans := &Packed64SingleBlock{blocks: make([]int64, requiredCapacity(valueCount, valuesPerBlock))}
	ans.MutableImpl = newMutableImpl(ans, int(valueCount), int(bitsPerValue))
	return ans
}

func (p *Packed64SingleBlock) Clear() {
	for i, _ := range p.blocks {
		p.blocks[i] = 0
	}
}

func (p *Packed64SingleBlock) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(p.blocks))
}

func (p *Packed64SingleBlock) Get(index int) int64 {
	return p.get(index)
}

func (p *Packed64SingleBlock) Set(index int, value int64) {
	p.set(index, value)
}

func (p *Packed64SingleBlock) getBulk(index int, arr []int64) int {
	off, length := 0, len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < p.valueCount)
	if p.valueCount-index < length {
		length = p.valueCount - index
	}

	originalIndex := index

	// go to the next block boundry
	valuesPerBlock := 64 / p.bitsPerValue
	offsetInBlock := index % valuesPerBlock
	if offsetInBlock != 0 {
		for i := offsetInBlock; i < valuesPerBlock && length > 0; i++ {
			arr[off] = p.Get(index)
			off++
			index++
			length--
		}
		if length == 0 {
			return index - originalIndex
		}
	}

	// bulk get
	assert(index%valuesPerBlock == 0)
	op := newBulkOperation(PackedFormat(PACKED_SINGLE_BLOCK), uint32(p.bitsPerValue))
	assert(op.LongBlockCount() == 1)
	assert(op.LongValueCount() == valuesPerBlock)
	blockIndex := index / valuesPerBlock
	nBlocks := (index+length)/valuesPerBlock - blockIndex
	op.decodeLongToLong(p.blocks[blockIndex:], arr[off:], nBlocks)
	diff := nBlocks * valuesPerBlock
	index += diff
	length -= diff

	if index > originalIndex {
		// stay at the block boundry
		return index - originalIndex
	}
	// no progress so far => already at a block boundry but no full block to set
	assert(index == originalIndex)
	return p.MutableImpl.getBulk(index, arr[off:off+length])
}

func (p *Packed64SingleBlock) setBulk(index int, arr []int64) int {
	off, length := 0, len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < p.valueCount)
	if p.valueCount-index < length {
		length = p.valueCount - index
	}

	originalIndex := index

	// go to the next block boundry
	valuesPerBlock := 64 / p.bitsPerValue
	offsetInBlock := index % valuesPerBlock
	if offsetInBlock != 0 {
		for i := offsetInBlock; i < valuesPerBlock && length > 0; i++ {
			p.Set(index, arr[off])
			index++
			off++
			length--
		}
		if length == 0 {
			return index - originalIndex
		}
	}

	// bulk set
	assert(index%valuesPerBlock == 0)
	op := newBulkOperation(PackedFormat(PACKED_SINGLE_BLOCK), uint32(p.bitsPerValue))
	assert(op.LongBlockCount() == 1)
	assert(op.LongValueCount() == valuesPerBlock)
	blockIndex := index / valuesPerBlock
	nBlocks := (index+length)/valuesPerBlock - blockIndex
	op.encodeLongToLong(arr[off:], p.blocks[blockIndex:], nBlocks)
	diff := nBlocks * valuesPerBlock
	index += diff
	length -= diff

	if index > originalIndex {
		// stay at the block boundry
		return index - originalIndex
	}
	// no progress so far => already at a block boundry but no full block to set
	assert(index == originalIndex)
	return p.MutableImpl.setBulk(index, arr[off:off+length])
}

func (p *Packed64SingleBlock) fill(from, to int, val int64) {
	panic("niy")
}

func (p *Packed64SingleBlock) Format() PackedFormat {
	return PackedFormat(PACKED_SINGLE_BLOCK)
}

func (p *Packed64SingleBlock) String() string {
	return fmt.Sprintf("Packed64SingleBlock(bitsPerValue=%v, size=%v, elements.length=%v)",
		p.bitsPerValue, p.Size(), len(p.blocks))
}

func newPacked64SingleBlockFromInput(in DataInput, valueCount int32, bitsPerValue uint32) (reader PackedIntsReader, err error) {
	ans := newPacked64SingleBlockBy(valueCount, bitsPerValue)
	for i, _ := range ans.blocks {
		if ans.blocks[i], err = in.ReadLong(); err != nil {
			break
		}
	}
	return ans, err
}

func newPacked64SingleBlockBy(valueCount int32, bitsPerValue uint32) *Packed64SingleBlock {
	switch bitsPerValue {
	case 1:
		return newPacked64SingleBlock1(valueCount)
	case 2:
		return newPacked64SingleBlock2(valueCount)
	case 3:
		return newPacked64SingleBlock3(valueCount)
	case 4:
		return newPacked64SingleBlock4(valueCount)
	case 5:
		return newPacked64SingleBlock5(valueCount)
	case 6:
		return newPacked64SingleBlock6(valueCount)
	case 7:
		return newPacked64SingleBlock7(valueCount)
	case 8:
		return newPacked64SingleBlock8(valueCount)
	case 9:
		return newPacked64SingleBlock9(valueCount)
	case 10:
		return newPacked64SingleBlock10(valueCount)
	case 12:
		return newPacked64SingleBlock12(valueCount)
	case 16:
		return newPacked64SingleBlock16(valueCount)
	case 21:
		return newPacked64SingleBlock21(valueCount)
	case 32:
		return newPacked64SingleBlock32(valueCount)
	default:
		panic(fmt.Sprintf("Unsuppoted number of bits per value: %v", bitsPerValue))
	}
}

func newPacked64SingleBlock1(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 1)
	ans.get = func(index int) int64 {
		o := uint32(index) >> 6
		b := index & 63
		shift := uint32(b << 0)
		return int64(uint64(ans.blocks[o])>>shift) & 1
	}
	ans.set = func(index int, value int64) {
		o := uint32(index) >> 6
		b := uint32(index & 63)
		shift := b << 0
		ans.blocks[o] = (ans.blocks[o] & ^(int64(1) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock2(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 2)
	ans.get = func(index int) int64 {
		o := uint32(index) >> 5
		b := index & 31
		shift := uint32(b << 1)
		return int64(uint64(ans.blocks[o])>>shift) & 3
	}
	ans.set = func(index int, value int64) {
		o := uint32(index) >> 5
		b := index & 31
		shift := uint32(b << 1)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(3) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock3(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 3)
	ans.get = func(index int) int64 {
		o := index / 21
		b := index % 21
		shift := uint32(b * 3)
		return int64(uint64(ans.blocks[o])>>shift) & 7
	}
	ans.set = func(index int, value int64) {
		o := index / 21
		b := index % 21
		shift := uint32(b * 3)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(7) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock4(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 4)
	ans.get = func(index int) int64 {
		o := uint32(index) >> 4
		b := index & 15
		shift := uint32(b << 2)
		return int64(uint64(ans.blocks[o])>>shift) & 15
	}
	ans.set = func(index int, value int64) {
		o := uint32(index) >> 4
		b := index & 15
		shift := uint32(b << 2)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(15) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock5(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 5)
	ans.get = func(index int) int64 {
		o := index / 12
		b := index % 12
		shift := uint32(b * 5)
		return int64(uint64(ans.blocks[o])>>shift) & 31
	}
	ans.set = func(index int, value int64) {
		o := index / 12
		b := index % 12
		shift := uint32(b * 5)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(31) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock6(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 6)
	ans.get = func(index int) int64 {
		o := index / 10
		b := index % 10
		shift := uint32(b * 6)
		return int64(uint64(ans.blocks[o])>>shift) & 63
	}
	ans.set = func(index int, value int64) {
		o := index / 10
		b := index % 10
		shift := uint32(b * 6)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(63) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock7(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 7)
	ans.get = func(index int) int64 {
		o := index / 9
		b := index % 9
		shift := uint32(b * 7)
		return int64(uint64(ans.blocks[o])>>shift) & 127
	}
	ans.set = func(index int, value int64) {
		o := index / 9
		b := index % 9
		shift := uint32(b * 7)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(127) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock8(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 8)
	ans.get = func(index int) int64 {
		o := uint32(index) >> 3
		b := index & 7
		shift := uint32(b << 3)
		return int64(uint64(ans.blocks[o])>>shift) & 255
	}
	ans.set = func(index int, value int64) {
		o := uint32(index) >> 3
		b := index & 7
		shift := uint32(b << 3)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(255) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock9(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 9)
	ans.get = func(index int) int64 {
		o := index / 7
		b := index % 7
		shift := uint32(b * 9)
		return int64(uint64(ans.blocks[o])>>shift) & 511
	}
	ans.set = func(index int, value int64) {
		o := index / 7
		b := index % 7
		shift := uint32(b * 9)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(511) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock10(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 10)
	ans.get = func(index int) int64 {
		o := index / 6
		b := index % 6
		shift := uint32(b * 10)
		return int64(uint64(ans.blocks[o])>>shift) & 1023
	}
	ans.set = func(index int, value int64) {
		o := index / 6
		b := index % 6
		shift := uint32(b * 10)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(1023) << shift)) | (value << shift)
	}
	return ans
}
func newPacked64SingleBlock12(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 12)
	ans.get = func(index int) int64 {
		o := index / 5
		b := index % 5
		shift := uint32(b * 12)
		return int64(uint64(ans.blocks[o])>>shift) & 4095
	}
	ans.set = func(index int, value int64) {
		o := index / 5
		b := index % 5
		shift := uint32(b * 12)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(4095) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock16(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 16)
	ans.get = func(index int) int64 {
		o := uint32(index) >> 2
		b := index & 3
		shift := uint32(b << 4)
		return int64(uint64(ans.blocks[o])>>shift) & 65535
	}
	ans.set = func(index int, value int64) {
		o := uint32(index) >> 2
		b := index & 3
		shift := uint32(b << 4)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(65335) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock21(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 21)
	ans.get = func(index int) int64 {
		o := index / 3
		b := index % 3
		shift := uint32(b * 21)
		return int64(uint64(ans.blocks[o])>>shift) & 2097151
	}
	ans.set = func(index int, value int64) {
		o := index / 3
		b := index % 3
		shift := uint32(b * 21)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(2097151) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock32(valueCount int32) *Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 32)
	ans.get = func(index int) int64 {
		o := uint32(index) >> 1
		b := index & 1
		shift := uint32(b << 5)
		return int64(uint64(ans.blocks[o])>>shift) & 4294967295
	}
	ans.set = func(index int, value int64) {
		o := uint32(index) >> 1
		b := index & 1
		shift := uint32(b << 5)
		ans.blocks[o] = (ans.blocks[o] & ^(int64(4294967295) << shift)) | (value << shift)
	}
	return ans
}
