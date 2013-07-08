package util

import (
	"fmt"
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
	ByteValueCount() uint32
}

func GetPackedIntsEncoder(format PackedFormat, version int32, bitsPerValue uint32) PackedIntsEncoder {
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}

func GetPackedIntsDecoder(format PackedFormat, version int32, bitsPerValue uint32) PackedIntsDecoder {
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

func newPackedHeaderNoHeader(in *DataInput, format PackedFormat, version, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
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

func newPackedReader(in *DataInput) (r PackedIntsReader, err error) {
	version, err := codec.CheckHeader(in, PACKED_CODEC_NAME, PACKED_VERSION_START, PACKED_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	bitsPerValue, err := in.ReadVInt()
	if err != nil {
		return nil, err
	}
	// assert bitsPerValue > 0 && bitsPerValue <= 64
	valueCount, err := in.ReadVInt()
	if err != nil {
		return nil, err
	}
	id, err := in.ReadVInt()
	if err != nil {
		return nil, err
	}
	format := PackedFormat(id)
	return newPackedReaderNoHeader(in, format, version, valueCount, bitsPerValue)
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
	ans.PackedIntsReaderImpl = newPakedIntsReaderImpl(valueCount, 8)
	return ans
}

func newDirect8FromInput(version int32, in *DataInput, valueCount int32) (r PackedIntsReader, err error) {
	r = newDirect8(valueCount)
	if err = in.ReadBytes(values[0:valueCount]); err == nil {
		// because packed ints have not always been byte-aligned
		remaining = PACKED.ByteCount(version, valueCount, 8) - valueCount
		for i := 0; i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return r, err
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

func newDirect16FromInput(version int32, in *DataInput, valueCount int32) (r PackedIntsReader, err error) {
	r = newDirect16(valueCount)
	for i, _ := range r.values {
		if r.values[i], err = in.ReadShort(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining = PACKED.ByteCount(version, valueCount, 16) - 2*valueCount
		for i := 0; i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return r, err
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

func newDirect32FromInput(version int32, in *DataInput, valueCount int32) (r PackedIntsReader, err error) {
	r = newDirect32(valueCount)
	for i, _ := range r.values {
		if r.values[i], err = in.ReadInt(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining = PACKED.ByteCount(version, valueCount, 32) - 4*valueCount
		for i := 0; i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return r, err
}

type Direct64 struct {
	PackedIntsReaderImpl
	values []int64
}

func newDirect64(valueCount int32) Direct32 {
	ans := Direct64{values: make([]int32, valueCount)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 64)
	return ans
}

func newDirect64FromInput(version int32, in *DataInput, valueCount int32) (r PackedIntsReader, err error) {
	r = newDirect64(valueCount)
	for i, _ := range r.values {
		if r.values[i], err = in.ReadLong(); err != nil {
			break
		}
	}
	return r, err
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

func newPacked8ThreeBlocksFromInput(version int32, in *DataInput, valueCount int32) (r PackedIntsReader, err error) {
	r = newPacked8ThreeBlocks(valueCount)
	if err = in.ReadBytes(r.blocks); err == nil {
		// because packed ints have not always been byte-aligned
		remaining = PACKED.ByteCount(version, valueCount, 24) - 3*valueCount
		for i := 0; i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return r, err
}

func (r *Packed8ThreeBlocks) Get(index int32) int64 {
	o := index * 3
	return blocks[o]<<16 | blocks[o+1]<<8 | blocks[o+2]
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

func newPacked16ThreeBlocksFromInput(version int32, in *DataInput, valueCount int32) (r PackedIntsReader, err error) {
	ans := newPacked16ThreeBlocks(valueCount)
	for i, _ := range ans.blocks {
		if ans.blocks[i], err = in.ReadShort(); err != nil {
			break
		}
	}
	if err == nil {
		// because packed ints have not always been byte-aligned
		remaining = PACKED.ByteCount(version, valueCount, 48) - 3*valueCount*2
		for i := 0; i < remaining; i++ {
			if _, err = in.ReadByte(); err != nil {
				break
			}
		}
	}
	return ans, err
}

type Packed64SingleBlock struct {
	PackedIntsReaderImpl
	get    func(index int32) int64
	blocks []int64
}

func (p *Packed64SingleBlock) Get(index int32) int64 {
	return p.get(index)
}

func newPacked64SingleBlockFromInput(in *DataInput, valueCount int32, bitsPerValue uint32) (reader *Packed64SingleBlock, err error) {
	reader = newPacked64SingleBlockBy(valueCount, bitsPerValue)
	for i, _ := range reader.blocks {
		reader.blocks[i], err = in.ReadLong()
		if err == nil {
			return reader, nil
		}
	}
	return &reader, nil
}

func newPacked64SingleBlockBy(valueCount int32, bitsPerValue uint32) (reader Packed64SingleBlock, err error) {
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

func newPacked64SingleBlock1(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 1)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 6
		b := index & 63
		shift := b << 0
		return int64(uint64(ans.blocks[o])>>shift) & 1
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 6
		b := index & 63
		shift := b << 0
		ans.blocks[o] = (ans.blocks[o] & ^(int64(1) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock2(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 2)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 5
		b := index & 31
		shift := b << 1
		return int64(uint64(ans.blocks[o])>>shift) & 3
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 5
		b := index & 31
		shift := b << 1
		ans.blocks[o] = (ans.blocks[o] & ^(int64(3) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock3(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 3)
	ans.get = func(index int32) int64 {
		o := index / 21
		b := index % 21
		shift := b * 3
		return int64(uint64(ans.blocks[o])>>shift) & 7
	}
	ans.set = func(index int32, value int64) {
		o := index / 21
		b := index % 21
		shift := b * 3
		ans.blocks[o] = (ans.blocks[o] & ^(int64(7) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock4(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 4)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 4
		b := index & 15
		shift := b << 2
		return int64(uint64(ans.blocks[o])>>shift) & 15
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 4
		b := index & 15
		shift := b << 2
		ans.blocks[o] = (ans.blocks[o] & ^(int64(15) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock5(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 5)
	ans.get = func(index int32) int64 {
		o := index / 12
		b := index % 12
		shift := b * 5
		return int64(uint64(ans.blocks[o])>>shift) & 31
	}
	ans.set = func(index int32, value int64) {
		o := index / 12
		b := index % 12
		shift := b * 5
		ans.blocks[o] = (ans.blocks[o] & ^(int64(31) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock6(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 6)
	ans.get = func(index int32) int64 {
		o := index / 10
		b := index % 10
		shift := b * 6
		return int64(uint64(ans.blocks[o])>>shift) & 63
	}
	ans.set = func(index int32, value int64) {
		o := index / 10
		b := index % 10
		shift := b * 6
		ans.blocks[o] = (ans.blocks[o] & ^(int64(63) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock7(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 7)
	ans.get = func(index int32) int64 {
		o := index / 9
		b := index % 9
		shift := b * 7
		return int64(uint64(ans.blocks[o])>>shift) & 127
	}
	ans.set = func(index int32, value int64) {
		o := index / 9
		b := index % 9
		shift := b * 7
		ans.blocks[o] = (ans.blocks[o] & ^(int64(127) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock8(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 8)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 3
		b := index & 7
		shift := b << 3
		return int64(uint64(ans.blocks[o])>>shift) & 255
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 3
		b := index & 7
		shift := b << 3
		ans.blocks[o] = (ans.blocks[o] & ^(int64(255) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock9(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 9)
	ans.get = func(index int32) int64 {
		o := index / 7
		b := index % 7
		shift := b * 9
		return int64(uint64(ans.blocks[o])>>shift) & 511
	}
	ans.set = func(index int32, value int64) {
		o := index / 7
		b := index % 7
		shift := b * 9
		ans.blocks[o] = (ans.blocks[o] & ^(int64(511) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock10(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 10)
	ans.get = func(index int32) int64 {
		o := index / 6
		b := index % 6
		shift := b * 10
		return int64(uint64(ans.blocks[o])>>shift) & 1023
	}
	ans.set = func(index int32, value int64) {
		o := index / 6
		b := index % 6
		shift := b * 10
		ans.blocks[o] = (ans.blocks[o] & ^(int64(1023) << shift)) | (value << shift)
	}
	return ans
}
func newPacked64SingleBlock12(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 12)
	ans.get = func(index int32) int64 {
		o := index / 5
		b := index % 5
		shift := b * 12
		return int64(uint64(ans.blocks[o])>>shift) & 4095
	}
	ans.set = func(index int32, value int64) {
		o := index / 5
		b := index % 5
		shift := b * 12
		ans.blocks[o] = (ans.blocks[o] & ^(int64(4095) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock16(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 16)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 2
		b := index & 3
		shift := b << 4
		return int64(uint64(ans.blocks[o])>>shift) & 65535
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 2
		b := index & 3
		shift := b << 4
		ans.blocks[o] = (ans.blocks[o] & ^(int64(65335) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock21(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 21)
	ans.get = func(index int32) int64 {
		o := index / 3
		b := index % 3
		shift := b * 21
		return int64(uint64(ans.blocks[o])>>shift) & 2097151
	}
	ans.set = func(index int32, value int64) {
		o := index / 3
		b := index % 3
		shift := b * 21
		ans.blocks[o] = (ans.blocks[o] & ^(int64(2097151) << shift)) | (value << shift)
	}
	return ans
}

func newPacked64SingleBlock32(valueCount int32) Packed64SingleBlock {
	ans := newPacked64SingleBlock(valueCount, 32)
	ans.get = func(index int32) int64 {
		o := uint32(index) >> 1
		b := index & 1
		shift := b << 5
		return int64(uint64(ans.blocks[o])>>shift) & 4294967295
	}
	ans.set = func(index int32, value int64) {
		o := uint32(index) >> 1
		b := index & 1
		shift := b << 5
		ans.blocks[o] = (ans.blocks[o] & ^(int64(4294967295) << shift)) | (value << shift)
	}
	return ans
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
