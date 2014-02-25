package packed

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/util"
	"math"
	"strconv"
)

// util/packed/PackedInts.java

/*
Simplistic compression for arrays of unsinged int64 values. Each value
is >= 0 and <= a specified maximum value. The vlues are stored as
packed ints, with each value consuming a fixed number of bits.
*/
var PackedInts = struct {
	FASTEST float32 // At most 700% memory overhead, always select a direct implementation.
	FAST    float32 // At most 50% memory overhead, always elect a reasonable fast implementation.
	DEFAULT float32 // At most 20% memory overhead.
	COMPACT float32 // No memory overhead at all, but hte returned implementation may be slow.
}{7, 0.5, 0.2, 0}

const (
	PACKED_CODEC_NAME           = "PackedInts"
	PACKED_VERSION_START        = 0
	PACKED_VERSION_BYTE_ALIGNED = 1
	PACKED_VERSION_CURRENT      = PACKED_VERSION_BYTE_ALIGNED
)

// Ceck the validity of a version number
func CheckVersion(version int32) {
	if version < PACKED_VERSION_START {
		panic(fmt.Sprintf("Version is too old, should be at least %v (got %v)", PACKED_VERSION_START, version))
	} else if version > PACKED_VERSION_CURRENT {
		panic(fmt.Sprintf("Version is too new, should be at most %v (got %v)", PACKED_VERSION_CURRENT, version))
	}
}

/**
 * A format to write packed ints.
 *
 * @lucene.internal
 */
type PackedFormat int

const (
	PACKED              = 0
	PACKED_SINGLE_BLOCK = 1
)

/**
 * Computes how many byte blocks are needed to store <code>values</code>
 * values of size <code>bitsPerValue</code>.
 */
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

/**
 * Computes how many long blocks are needed to store <code>values</code>
 * values of size <code>bitsPerValue</code>.
 */
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

/**
 * Tests whether the provided number of bits per value is supported by the
 * format.
 */
func (f PackedFormat) IsSupported(bitsPerValue uint32) bool {
	switch int(f) {
	case PACKED_SINGLE_BLOCK:
		return is64Supported(bitsPerValue)
	}
	return bitsPerValue >= 1 && bitsPerValue <= 64
}

type PackedIntsEncoder interface {
}

type PackedIntsDecoder interface {
	// The minum numer of byte blocks to encode in a single iteration, when using byte encoding
	ByteBlockCount() int
	// The number of values that can be stored in byteBlockCount() byte blocks
	ByteValueCount() int
	/*
		Read 8 * iterations * blockCount() blocks from blocks, decodethem and write
		iterations * valueCount() values inot values.
	*/
	decodeByteToint64(blocks []byte, values []int64, iterations int)
}

func GetPackedIntsEncoder(format PackedFormat, version int32, bitsPerValue uint32) PackedIntsEncoder {
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}

func GetPackedIntsDecoder(format PackedFormat, version int32, bitsPerValue uint32) PackedIntsDecoder {
	// log.Printf("Obtaining PackedIntsDecoder(%v, %v), version=%v", format, bitsPerValue, version)
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}

type PackedIntsReader interface {
	Get(index int32) int64
	Size() int32
}

// Run-once iterator interface, to decode previously saved PackedInts
type ReaderIterator interface {
	Next() (v int64, err error)          // next value
	nextN(n int) (vs []int64, err error) // at least 1 and at most n next values, the returned ref MUST NOT be modified
	bitsPerValue() int                   // number of bits per value
	size() int                           // number of values
	ord() int                            // the current position
}

type nextNAction interface {
	nextN(n int) (vs []int64, err error)
}

type ReaderIteratorImpl struct {
	nextNAction
	in            util.DataInput
	_bitsPerValue int
	valueCount    int
}

func newReaderIteratorImpl(sub nextNAction, valueCount, bitsPerValue int, in util.DataInput) *ReaderIteratorImpl {
	return &ReaderIteratorImpl{sub, in, bitsPerValue, valueCount}
}

/*
Lucene(Java) manipulates underlying LongsRef to advance the pointer, which
can not be implemented using Go's slice. Here I have to assume nextN() method
would automatically increment the pointer without next().
*/
func (it *ReaderIteratorImpl) Next() (v int64, err error) {
	nextValues, err := it.nextN(1)
	if err != nil {
		return 0, err
	}
	assert(len(nextValues) > 0)
	return nextValues[0], nil
}

func (it *ReaderIteratorImpl) bitsPerValue() int {
	return it._bitsPerValue
}

func (it *ReaderIteratorImpl) size() int {
	return it.valueCount
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

type PackedIntsMutable interface {
	PackedIntsReader
}

func NewPackedReaderNoHeader(in util.DataInput, format PackedFormat, version, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
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

func NewPackedReader(in util.DataInput) (r PackedIntsReader, err error) {
	if version, err := codec.CheckHeader(in, PACKED_CODEC_NAME, PACKED_VERSION_START, PACKED_VERSION_CURRENT); err == nil {
		if bitsPerValue, err := asUint32(in.ReadVInt()); err == nil {
			// assert bitsPerValue > 0 && bitsPerValue <= 64
			if valueCount, err := in.ReadVInt(); err == nil {
				if id, err := in.ReadVInt(); err == nil {
					format := PackedFormat(id)
					return NewPackedReaderNoHeader(in, format, version, valueCount, bitsPerValue)
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

/*
Expert: Restore a ReaderIterator from a stream without reading metadata at the
beginning of the stream. This method is useful to restore data from streams
which have been created using WriterNoHeader().
*/
func ReaderIteratorNoHeader(in util.DataInput, format PackedFormat, version,
	valueCount, bitsPerValue, mem int) ReaderIterator {
	CheckVersion(int32(version))
	return newPackedReaderIterator(format, version, valueCount, bitsPerValue, in, mem)
}

// Returns how many bits are required to hold values up to and including maxValue
func BitsRequired(maxValue int64) int {
	assert2(maxValue >= 0, fmt.Sprintf("maxValue must be non-negative (got: %v)", maxValue))
	// TODO not efficient
	return len(strconv.FormatInt(maxValue, 2))
}

func assert2(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

// Calculate the maximum unsigned long that can be expressed with the given number of bits
func MaxValue(bitsPerValue int) int64 {
	if bitsPerValue == 64 {
		return math.MaxInt64
	}
	return (1 << uint64(bitsPerValue)) - 1
}

// util/packed/PackedReaderIterator.java

type PackedReaderIterator struct {
	*ReaderIteratorImpl
	packedIntsVersion int
	format            PackedFormat
	bulkOperation     BulkOperation
	nextBlocks        []byte
	nextValues        []int64
	nextValuesOrig    []int64
	_iterations       int
	position          int
}

func newPackedReaderIterator(format PackedFormat, packedIntsVersion, valueCount, bitsPerValue int, in util.DataInput, mem int) *PackedReaderIterator {
	it := &PackedReaderIterator{
		format:            format,
		packedIntsVersion: packedIntsVersion,
		bulkOperation:     newBulkOperation(format, uint32(bitsPerValue)),
		position:          -1,
	}
	assert(it.bulkOperation != nil)
	it.ReaderIteratorImpl = newReaderIteratorImpl(it, valueCount, bitsPerValue, in)
	it._iterations = it.iterations(mem)
	assert(valueCount == 0 || it._iterations > 0)
	it.nextBlocks = make([]byte, it._iterations*it.bulkOperation.ByteBlockCount())
	it.nextValuesOrig = make([]int64, it._iterations*it.bulkOperation.ByteValueCount())
	it.nextValues = nil
	return it
}

func (it *PackedReaderIterator) iterations(mem int) int {
	iterations := it.bulkOperation.computeIterations(it.valueCount, mem)
	if it.packedIntsVersion < PACKED_VERSION_BYTE_ALIGNED {
		// make sure iterations is a multiple of 8
		iterations = int((int64(iterations) + 7) & 0xFFFFFFF8)
	}
	return iterations
}

/*
Go slice is used to mimic Lucene(Java)'s LongsRef and I have to keep the
original slice to avoid re-allocation.
*/
func (it *PackedReaderIterator) nextN(count int) (vs []int64, err error) {
	assert(len(it.nextValues) >= 0)
	assert(count > 0)

	remaining := it.valueCount - it.position - 1
	if remaining <= 0 {
		return nil, errors.New("EOF")
	}
	if remaining < count {
		count = remaining
	}

	if len(it.nextValues) == 0 {
		remainingBlocks := it.format.ByteCount(int32(it.packedIntsVersion), int32(remaining), uint32(it._bitsPerValue))
		blocksToRead := len(it.nextBlocks)
		if remainingBlocks < int64(blocksToRead) {
			blocksToRead = int(remainingBlocks)
		}
		err = it.in.ReadBytes(it.nextBlocks[0:blocksToRead])
		if err != nil {
			return nil, err
		}
		if blocksToRead < len(it.nextBlocks) {
			for i := blocksToRead; i < len(it.nextBlocks); i++ {
				it.nextBlocks[i] = 0
			}
		}

		it.nextValues = it.nextValuesOrig // restore
		it.bulkOperation.decodeByteToint64(it.nextBlocks, it.nextValues, it._iterations)
	}

	if len(it.nextValues) < count {
		count = len(it.nextValues)
	}
	it.position += count
	values := it.nextValues[0:count]
	it.nextValues = it.nextValues[count:]
	return values, nil
}

func (it *PackedReaderIterator) ord() int {
	return it.position
}

// util/packed/Direct8.java

type Direct8 struct {
	PackedIntsReaderImpl
	values []byte
}

func newDirect8(valueCount int32) Direct8 {
	ans := Direct8{values: make([]byte, valueCount)}
	ans.PackedIntsReaderImpl = newPackedIntsReaderImpl(valueCount, 8)
	return ans
}

func newDirect8FromInput(version int32, in util.DataInput, valueCount int32) (r PackedIntsReader, err error) {
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

func newDirect16FromInput(version int32, in util.DataInput, valueCount int32) (r PackedIntsReader, err error) {
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

func newDirect32FromInput(version int32, in util.DataInput, valueCount int32) (r PackedIntsReader, err error) {
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

func newDirect64FromInput(version int32, in util.DataInput, valueCount int32) (r PackedIntsReader, err error) {
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

func newPacked8ThreeBlocksFromInput(version int32, in util.DataInput, valueCount int32) (r PackedIntsReader, err error) {
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

func newPacked16ThreeBlocksFromInput(version int32, in util.DataInput, valueCount int32) (r PackedIntsReader, err error) {
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

func newPacked64FromInput(version int32, in util.DataInput, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
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
