package packed

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math"
)

var PACKED8_THREE_BLOCKS_MAX_SIZE = int32(math.MaxInt32 / 3)

type Packed8ThreeBlocks struct {
	*MutableImpl
	blocks []byte
}

func newPacked8ThreeBlocks(valueCount int32) *Packed8ThreeBlocks {
	assert2(valueCount <= PACKED8_THREE_BLOCKS_MAX_SIZE, "MAX_SIZE exceeded")
	ans := &Packed8ThreeBlocks{blocks: make([]byte, valueCount*3)}
	ans.MutableImpl = newMutableImpl(ans, int(valueCount), 24)
	return ans
}

func newPacked8ThreeBlocksFromInput(version int32, in DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newPacked8ThreeBlocks(valueCount)
	if err = in.ReadBytes(ans.blocks); err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, valueCount, 24) - 3*int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return ans, err
}

func (r *Packed8ThreeBlocks) Get(index int) int64 {
	o := index * 3
	return int64(r.blocks[o])<<16 | int64(r.blocks[o+1])<<8 | int64(r.blocks[o+2])
}

func (r *Packed8ThreeBlocks) getBulk(index int, arr []int64) int {
	panic("niy")
}

func (r *Packed8ThreeBlocks) Set(index int, value int64) {
	panic("not implemented yet")
}

func (r *Packed8ThreeBlocks) setBulk(idnex int, arr []int64) int {
	panic("niy")
}

func (r *Packed8ThreeBlocks) fill(from, to int, val int64) {
	panic("niy")
}

func (r *Packed8ThreeBlocks) Clear() {
	panic("niy")
}

func (r *Packed8ThreeBlocks) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER +
			2*util.NUM_BYTES_INT +
			util.NUM_BYTES_OBJECT_REF +
			util.SizeOf(r.blocks))
}

func (r *Packed8ThreeBlocks) String() string {
	return fmt.Sprintf("Packed8ThreeBlocks(bitsPerValue=%v, size=%v, elements.length=%v)",
		r.bitsPerValue, r.Size(), len(r.blocks))
}
