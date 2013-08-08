package util

import (
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"log"
	"math"
)

const (
	PACKED_CODEC_NAME           = "PackedInts"
	PACKED_VERSION_START        = 0
	PACKED_VERSION_BYTE_ALIGNED = 1
	PACKED_VERSION_CURRENT      = PACKED_VERSION_BYTE_ALIGNED
)

func CheckVersion(version int32) {
	if version < PACKED_VERSION_START {
		panic(fmt.Sprintf("Version is too old, should be at least %v (got %v)", PACKED_VERSION_START, version))
	} else if version > PACKED_VERSION_CURRENT {
		panic(fmt.Sprintf("Version is too new, should be at most %v (got %v)", PACKED_VERSION_CURRENT, version))
	}
}

type PackedFormat int

const (
	PACKED              = 0
	PACKED_SINGLE_BLOCK = 1
)

func (f PackedFormat) ByteCount(packedIntsVersion, valueCount int32, bitsPerValue uint32) int64 {
	switch int(f) {
	case PACKED:
		if packedIntsVersion < PACKED_VERSION_BYTE_ALIGNED {
			return 8 * int64(math.Ceil(float64(valueCount)*float64(bitsPerValue)/64))
		}
		return int64(math.Ceil(float64(valueCount) * float64(bitsPerValue) / 8))
	}
	// assert bitsPerValue >= 0 && bitsPerValue <= 64
	// assume long-aligned
	return 8 * int64(f.longCount(packedIntsVersion, valueCount, bitsPerValue))
}

func (f PackedFormat) longCount(packedIntsVersion, valueCount int32, bitsPerValue uint32) int {
	switch int(f) {
	case PACKED_SINGLE_BLOCK:
		valuesPerBlock := 64 / bitsPerValue
		return int(math.Ceil(float64(valueCount) / float64(valuesPerBlock)))
	}
	// assert bitsPerValue >= 0 && bitsPerValue <= 64
	ans := f.ByteCount(packedIntsVersion, valueCount, bitsPerValue)
	// assert ans < 8 * math.MaxInt32()
	if ans%8 == 0 {
		return int(ans / 8)
	}
	return int(ans/8) + 1
}

type PackedIntsEncoder interface {
}

type PackedIntsDecoder interface {
	ByteValueCount() int
}

func GetPackedIntsEncoder(format PackedFormat, version int32, bitsPerValue uint32) PackedIntsEncoder {
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}

func GetPackedIntsDecoder(format PackedFormat, version int32, bitsPerValue uint32) PackedIntsDecoder {
	log.Printf("Obtaining PackedIntsDecoder(%v, %v), version=%v", format, bitsPerValue, version)
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}

type PackedIntsReader interface {
	Get(index int32) int64
	Size() int32
}

type PackedIntsMutable interface {
	PackedIntsReader
}

func newPackedReaderNoHeader(in DataInput, format PackedFormat, version, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
	CheckVersion(version)
	switch format {
	case PACKED_SINGLE_BLOCK:
		return newPacked64SingleBlockFromInput(in, valueCount, bitsPerValue)
	case PACKED:
		switch bitsPerValue {
		case 8:
			return newDirect8FromInput(version, in, valueCount)
		case 16:
			return newDirect16FromInput(version, in, valueCount)
		case 32:
			return newDirect32FromInput(version, in, valueCount)
		case 64:
			return newDirect64FromInput(version, in, valueCount)
		case 24:
			if valueCount <= PACKED8_THREE_BLOCKS_MAX_SIZE {
				return newPacked8ThreeBlocksFromInput(version, in, valueCount)
			}
		case 48:
			if valueCount <= PACKED16_THREE_BLOCKS_MAX_SIZE {
				return newPacked16ThreeBlocksFromInput(version, in, valueCount)
			}
		}
		return newPacked64FromInput(version, in, valueCount, bitsPerValue)
	default:
		panic(fmt.Sprintf("Unknown Writer foramt: %v", format))
	}
}

func asUint32(n int32, err error) (n2 uint32, err2 error) {
	return uint32(n), err
}

func newPackedReader(in DataInput) (r PackedIntsReader, err error) {
	if version, err := codec.CheckHeader(in, PACKED_CODEC_NAME, PACKED_VERSION_START, PACKED_VERSION_CURRENT); err == nil {
		if bitsPerValue, err := asUint32(in.ReadVInt()); err == nil {
			// assert bitsPerValue > 0 && bitsPerValue <= 64
			if valueCount, err := in.ReadVInt(); err == nil {
				if id, err := in.ReadVInt(); err == nil {
					format := PackedFormat(id)
					return newPackedReaderNoHeader(in, format, version, valueCount, bitsPerValue)
				}
			}
		}
	}
	return
}

type PackedIntsReaderImpl struct {
	bitsPerValue uint32
	valueCount   int32
}

func newPackedIntsReaderImpl(valueCount int32, bitsPerValue uint32) PackedIntsReaderImpl {
	// assert bitsPerValue > 0 && bitsPerValue <= 64
	return PackedIntsReaderImpl{bitsPerValue, valueCount}
}

func (p PackedIntsReaderImpl) Size() int32 {
	return p.valueCount
}

type PackedIntsMutableImpl struct {
}

type Direct8 struct {
	PackedIntsReaderImpl
	values []byte
}

func newDirect8(valueCount int32) Direct8 {
	ans := Direct8{values: make([]byte, valueCount)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 8)
	return ans
}

func newDirect8FromInput(version int32, in DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newDirect8(valueCount)
	if err = in.ReadBytes(ans.values[0:valueCount]); err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, valueCount, 8) - int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return &ans, err
}

func (d *Direct8) Get(index int32) int64 {
	return int64(d.values[index])
}

type Direct16 struct {
	PackedIntsReaderImpl
	values []int16
}

func newDirect16(valueCount int32) Direct16 {
	ans := Direct16{values: make([]int16, valueCount)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 16)
	return ans
}

func newDirect16FromInput(version int32, in DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newDirect16(valueCount)
	for i, _ := range ans.values {
		if ans.values[i], err = in.ReadShort(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, valueCount, 16) - 2*int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return &ans, err
}

func (d *Direct16) Get(index int32) int64 {
	return int64(d.values[index])
}

type Direct32 struct {
	PackedIntsReaderImpl
	values []int32
}

func newDirect32(valueCount int32) Direct32 {
	ans := Direct32{values: make([]int32, valueCount)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 32)
	return ans
}

func newDirect32FromInput(version int32, in DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newDirect32(valueCount)
	for i, _ := range ans.values {
		if ans.values[i], err = in.ReadInt(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining := PackedFormat(PACKED).ByteCount(version, valueCount, 32) - 4*int64(valueCount)
		for i := int64(0); i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return &ans, err
}

func (d *Direct32) Get(index int32) int64 {
	return int64(d.values[index])
}

type Direct64 struct {
	PackedIntsReaderImpl
	values []int64
}

func newDirect64(valueCount int32) Direct64 {
	ans := Direct64{values: make([]int64, valueCount)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 64)
	return ans
}

func newDirect64FromInput(version int32, in DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newDirect64(valueCount)
	for i, _ := range ans.values {
		if ans.values[i], err = in.ReadLong(); err != nil {
			break
		}
	}
	return &ans, err
}

func (d *Direct64) Get(index int32) int64 {
	return d.values[index]
}

var PACKED8_THREE_BLOCKS_MAX_SIZE = int32(math.MaxInt32 / 3)

type Packed8ThreeBlocks struct {
	PackedIntsReaderImpl
	blocks []byte
}

func newPacked8ThreeBlocks(valueCount int32) Packed8ThreeBlocks {
	if valueCount > PACKED8_THREE_BLOCKS_MAX_SIZE {
		panic("MAX_SIZE exceeded")
	}
	ans := Packed8ThreeBlocks{blocks: make([]byte, valueCount*3)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 24)
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
	return &ans, err
}

func (r *Packed8ThreeBlocks) Get(index int32) int64 {
	o := index * 3
	return int64(r.blocks[o])<<16 | int64(r.blocks[o+1])<<8 | int64(r.blocks[o+2])
}

var PACKED16_THREE_BLOCKS_MAX_SIZE = int32(math.MaxInt32 / 3)

type Packed16ThreeBlocks struct {
	PackedIntsReaderImpl
	blocks []int16
}

func newPacked16ThreeBlocks(valueCount int32) Packed16ThreeBlocks {
	if valueCount > PACKED16_THREE_BLOCKS_MAX_SIZE {
		panic("MAX_SIZE exceeded")
	}
	ans := Packed16ThreeBlocks{blocks: make([]int16, valueCount*3)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 48)
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
	return &ans, err
}

func (p *Packed16ThreeBlocks) Get(index int32) int64 {
	o := index * 3
	return int64(p.blocks[o])<<32 | int64(p.blocks[o+1])<<16 | int64(p.blocks[o])
}

const (
	PACKED64_BLOCK_SIZE = 64                      // 32 = int, 64 = long
	PACKED64_BLOCK_BITS = 6                       // The #bits representing BLOCK_SIZE
	PACKED64_MOD_MASK   = PACKED64_BLOCK_SIZE - 1 // x % BLOCK_SIZE
)

type Packed64 struct {
	PackedIntsReaderImpl
	blocks            []int64
	maskRight         uint64
	bpvMinusBlockSize int32
}

func newPacked64(valueCount int32, bitsPerValue uint32) Packed64 {
	longCount := PackedFormat(PACKED).longCount(PACKED_VERSION_CURRENT, valueCount, bitsPerValue)
	ans := Packed64{
		blocks:            make([]int64, longCount),
		maskRight:         uint64(^(int64(0))<<(PACKED64_BLOCK_SIZE-bitsPerValue)) >> (PACKED64_BLOCK_SIZE - bitsPerValue),
		bpvMinusBlockSize: int32(bitsPerValue) - PACKED64_BLOCK_SIZE}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, bitsPerValue)
	return ans
}

func newPacked64FromInput(version int32, in DataInput, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
	ans := newPacked64(valueCount, bitsPerValue)
	byteCount := PackedFormat(PACKED).ByteCount(version, valueCount, bitsPerValue)
	longCount := PackedFormat(PACKED).longCount(PACKED_VERSION_CURRENT, valueCount, bitsPerValue)
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
	return &ans, err
}

func (p *Packed64) Get(index int32) int64 {
	// The abstract index in a bit stream
	majorBitPos := int64(index) * int64(p.bitsPerValue)
	// The index in the backing long-array
	elementPos := int32(uint64(majorBitPos) >> PACKED64_BLOCK_BITS)
	// The number of value-bits in the second long
	endBits := (majorBitPos & PACKED64_MOD_MASK) + int64(p.bpvMinusBlockSize)

	if endBits <= 0 { // Single block
		return int64((uint64(p.blocks[elementPos]) >> uint64(-endBits)) & p.maskRight)
	}
	// Two blocks
	return ((p.blocks[elementPos] << uint64(endBits)) |
		int64(uint64(p.blocks[elementPos+1])>>(PACKED64_BLOCK_SIZE-uint64(endBits)))) &
		int64(p.maskRight)
}

type GrowableWriter struct {
	current PackedIntsMutable
}

func (w *GrowableWriter) Get(index int32) int64 {
	return w.current.Get(index)
}

func (w *GrowableWriter) Size() int32 {
	return w.current.Size()
}
