package packed

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math"
)

var PACKED16_THREE_BLOCKS_MAX_SIZE = int32(math.MaxInt32 / 3)

type Packed16ThreeBlocks struct {
	*MutableImpl
	blocks []int16
}

func newPacked16ThreeBlocks(valueCount int32) *Packed16ThreeBlocks {
	assert2(valueCount <= PACKED16_THREE_BLOCKS_MAX_SIZE, "MAX_SIZE exceeded")
	ans := &Packed16ThreeBlocks{blocks: make([]int16, valueCount*3)}
	ans.MutableImpl = newMutableImpl(ans, int(valueCount), 48)
	return ans
}

func newPacked16ThreeBlocksFromInput(version int32, in DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newPacked16ThreeBlocks(valueCount)
	for i, _ := range ans.blocks {
		if ans.blocks[i], err = in.ReadShort(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, valueCount, 48) - 3*int64(valueCount)*2
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return ans, err
}

func (p *Packed16ThreeBlocks) Get(index int) int64 {
	o := index * 3
	return int64(p.blocks[o])<<32 | int64(p.blocks[o+1])<<16 | int64(p.blocks[o])
}

func (r *Packed16ThreeBlocks) getBulk(index int, arr []int64) int {
	panic("niy")
}

func (r *Packed16ThreeBlocks) Set(index int, value int64) {
	panic("not implemented yet")
}

func (r *Packed16ThreeBlocks) setBulk(index int, arr []int64) int {
	panic("not implemented yet")
}

func (r *Packed16ThreeBlocks) fill(from, to int, val int64) {
	panic("niy")
}

func (r *Packed16ThreeBlocks) Clear() {
	panic("niy")
}

func (p *Packed16ThreeBlocks) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(p.blocks))
}

func (p *Packed16ThreeBlocks) String() string {
	return fmt.Sprintf("Packed16ThreeBlocks(bitsPerValue=%v, size=%v, elements.length=%v)",
		p.bitsPerValue, p.Size(), len(p.blocks))
}
