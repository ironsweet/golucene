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

type DataInput interface {
	ReadByte() (byte, error)
	ReadBytes(buf []byte) error
	ReadShort() (int16, error)
	ReadInt() (int32, error)
	ReadVInt() (int32, error)
	ReadLong() (int64, error)
	ReadString() (string, error)
}

type DataOutput interface {
	WriteBytes(buf []byte) error
	WriteInt(int32) error
	WriteVInt(int32) error
	WriteString(string) error
}

/*
Simplistic compression for arrays of unsinged int64 values. Each value
is >= 0 and <= a specified maximum value. The vlues are stored as
packed ints, with each value consuming a fixed number of bits.
*/
var PackedInts = struct {
	FASTEST float32 // At most 700% memory overhead, always select a direct implementation.
	FAST    float32 // At most 50% memory overhead, always elect a reasonable fast implementation.
	DEFAULT float32 // At most 25% memory overhead.
	COMPACT float32 // No memory overhead at all, but hte returned implementation may be slow.
}{7, 0.5, 0.25, 0}

const (
	/* Default amount of memory to use for bulk operations. */
	DEFAULT_BUFFER_SIZE = 1024 // 1K

	PACKED_CODEC_NAME                = "PackedInts"
	PACKED_VERSION_START             = 0
	PACKED_VERSION_BYTE_ALIGNED      = 1
	VERSION_MONOTONIC_WITHOUT_ZIGZAG = 2
	VERSION_CURRENT                  = VERSION_MONOTONIC_WITHOUT_ZIGZAG
)

// Ceck the validity of a version number
func CheckVersion(version int32) {
	if version < PACKED_VERSION_START {
		panic(fmt.Sprintf("Version is too old, should be at least %v (got %v)", PACKED_VERSION_START, version))
	} else if version > VERSION_CURRENT {
		panic(fmt.Sprintf("Version is too new, should be at most %v (got %v)", VERSION_CURRENT, version))
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

func (f PackedFormat) Id() int {
	return int(f)
}

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

/* Returns the overhead per value, in bits. */
func (f PackedFormat) OverheadPerValue(bitsPerValue int) float32 {
	switch int(f) {
	case PACKED_SINGLE_BLOCK:
		assert(f.IsSupported(uint32(bitsPerValue)))
		valuesPerBlock := 64 / bitsPerValue
		overhead := 64 % bitsPerValue
		return float32(overhead) / float32(valuesPerBlock)
	}
	return 0
}

/* Simple class that holds a format and a number of bits per value. */
type FormatAndBits struct {
	Format       PackedFormat
	BitsPerValue int
}

func (v FormatAndBits) String() string {
	return fmt.Sprintf("FormatAndBits(format=%v bitsPerValue=%v)", v.Format, v.BitsPerValue)
}

/*
Try to find the Format and number of bits per value that would
restore from disk the fastest reader whose overhead is less than
acceptableOverheadRatio.

The acceptableOverheadRatio parameter makes sense for random-access
Readers. In case you only plan to perform sequential access on this
stream later on, you should probably use COMPACT.

If you don't know how many values you are going to write, use
valueCount = -1.
*/
func FastestFormatAndBits(valueCount, bitsPerValue int,
	acceptableOverheadRatio float32) FormatAndBits {
	if valueCount == -1 {
		valueCount = int(math.MaxInt32)
	}

	if acceptableOverheadRatio < PackedInts.COMPACT {
		acceptableOverheadRatio = PackedInts.COMPACT
	}
	if acceptableOverheadRatio < PackedInts.FASTEST {
		acceptableOverheadRatio = PackedInts.FASTEST
	}

	maxBitsPerValue := bitsPerValue + int(acceptableOverheadRatio)

	actualBitsPerValue := -1
	format := PACKED

	if bitsPerValue <= 8 && maxBitsPerValue >= 8 {
		actualBitsPerValue = 8
	} else if bitsPerValue <= 16 && maxBitsPerValue >= 16 {
		actualBitsPerValue = 16
	} else if bitsPerValue <= 32 && maxBitsPerValue >= 32 {
		actualBitsPerValue = 32
	} else if bitsPerValue <= 64 && maxBitsPerValue >= 64 {
		actualBitsPerValue = 64
	} else if valueCount <= int(PACKED8_THREE_BLOCKS_MAX_SIZE) && bitsPerValue <= 24 && maxBitsPerValue >= 24 {
		actualBitsPerValue = 24
	} else if valueCount <= int(PACKED16_THREE_BLOCKS_MAX_SIZE) && bitsPerValue <= 48 && maxBitsPerValue >= 48 {
		actualBitsPerValue = 48
	} else {
		for bpv := bitsPerValue; bpv <= maxBitsPerValue; bpv++ {
			if PackedFormat(PACKED_SINGLE_BLOCK).IsSupported(uint32(bpv)) {
				overhead := PackedFormat(PACKED_SINGLE_BLOCK).OverheadPerValue(bpv)
				acceptableOverhead := acceptableOverheadRatio + float32(bitsPerValue-bpv)
				if overhead <= acceptableOverhead {
					actualBitsPerValue = bpv
					format = PACKED_SINGLE_BLOCK
					break
				}
			}
		}
		if actualBitsPerValue < 0 {
			actualBitsPerValue = bitsPerValue
		}
	}

	return FormatAndBits{PackedFormat(format), actualBitsPerValue}
}

type PackedIntsEncoder interface {
	// Read iterations * valueCount() values from values, encode them
	// and write iterations * blockCount() blocks into blocks.
	// encodeLongToLong(values, blocks []int64, iterations int)
	// Read iterations * valueCount() values from values, encode them
	// and write 8 * iterations * blockCount() blocks into blocks.
	encodeLongToByte(values []int64, blocks []byte, iterations int)
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

/* A read-only random access array of positive integers. */
type PackedIntsReader interface {
	util.Accountable
	Get(index int) int64
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
	in            DataInput
	_bitsPerValue int
	valueCount    int
}

func newReaderIteratorImpl(sub nextNAction, valueCount, bitsPerValue int, in DataInput) *ReaderIteratorImpl {
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

/* A packed integer array that can be modified. */
type Mutable interface {
	PackedIntsReader
	// Returns the number of bits used to store any given value. Note:
	// this does not imply that memory usage is bpv * values() as
	// implementations are free to use non-space-optimal packing of
	// bits.
	BitsPerValue() int
	// Set the value at the given index in the array.
	Set(index int, value int64)
	// Save this mutable into out. Instantiating a reader from the
	// generated data will return a reader with the same number of bits
	// per value.
	Save(out util.DataOutput) error
}

type PackedIntsReaderImpl struct {
	valueCount int32
}

func newPackedIntsReaderImpl(valueCount int) PackedIntsReaderImpl {
	return PackedIntsReaderImpl{int32(valueCount)}
}

func (p PackedIntsReaderImpl) Size() int32 {
	return p.valueCount
}

type MutableImplSPI interface {
	Get(int) int64
}

type MutableImpl struct {
	PackedIntsReaderImpl
	spi          MutableImplSPI
	format       PackedFormat
	bitsPerValue int
}

func newMutableImpl(valueCount, bitsPerValue int, spi MutableImplSPI) *MutableImpl {
	assert(bitsPerValue > 0 && bitsPerValue <= 64)
	return &MutableImpl{newPackedIntsReaderImpl(valueCount), spi, PACKED, bitsPerValue}
}

func (p *MutableImpl) BitsPerValue() int {
	return p.bitsPerValue
}

func (m *MutableImpl) Save(out util.DataOutput) error {
	writer := WriterNoHeader(out, m.format, int(m.valueCount), m.bitsPerValue, DEFAULT_BUFFER_SIZE)
	err := writer.writeHeader()
	if err != nil {
		return err
	}
	for i := 0; i < int(m.valueCount); i++ {
		err = writer.Add(m.spi.Get(i))
		if err != nil {
			return err
		}
	}
	return writer.Finish()
}

func NewPackedReaderNoHeader(in DataInput, format PackedFormat, version, valueCount int32, bitsPerValue uint32) (r PackedIntsReader, err error) {
	CheckVersion(version)
	switch format {
	case PACKED_SINGLE_BLOCK:
		return newPacked64SingleBlockFromInput(in, valueCount, bitsPerValue)
	case PACKED:
		switch bitsPerValue {
		/* [[[gocog
		package main

		import (
			"fmt"
			"os"
		)

		const HEADER = `// This file has been automatically generated, DO NOT EDIT

		package packed

		import (
			"github.com/balzaczyy/golucene/core/util"
		)

		`

		var (
			TYPES = map[int]string{8: "byte", 16: "int16", 32: "int32", 64: "int64"}
			NAMES = map[int]string{8: "Byte", 16: "Short", 32: "Int", 64: "Long"}
			MASKS = map[int]string{8: " & 0xFF", 16: " & 0xFFFF", 32: " & 0xFFFFFFFF", 64: ""}
			CASTS = map[int]string{8: "byte(", 16: "int16(", 32: "int32(", 64: "("}
		)

		func main() {
			w := fmt.Fprintf
			for bpv, typ := range TYPES {
				f, err := os.Create(fmt.Sprintf("direct%d.go", bpv))
				if err != nil {
					panic(err)
				}
				defer f.Close()

				w(f, HEADER)
				w(f, "// Direct wrapping of %d-bits values to a backing array.\n", bpv)
				w(f, "type Direct%d struct {\n", bpv)
				w(f, "	*MutableImpl\n")
				w(f, "	values []%v\n", typ)
				w(f, "}\n\n")

				w(f, "func newDirect%d(valueCount int) *Direct%d {\n", bpv, bpv)
				w(f, "	ans := &Direct%d{\n", bpv)
				w(f, "		values: make([]%v, valueCount),\n", typ)
				w(f, "	}\n")
				w(f, "	ans.MutableImpl = newMutableImpl(valueCount, %v, ans)\n", bpv)
				w(f, "  return ans\n")
				w(f, "}\n\n")

				w(f, "func newDirect%dFromInput(version int32, in DataInput, valueCount int) (r PackedIntsReader, err error) {\n", bpv)
				w(f, "	ans := newDirect%v(valueCount)\n", bpv)
				if bpv == 8 {
					w(f, "	if err = in.ReadBytes(ans.values[:valueCount]); err == nil {\n")
				} else {
					w(f, "	for i, _ := range ans.values {\n")
					w(f, " 		if ans.values[i], err = in.Read%v(); err != nil {\n", NAMES[bpv])
					w(f, "			break\n")
					w(f, "		}\n")
					w(f, "	}\n")
					w(f, "	if err == nil {\n")
				}
				if bpv != 64 {
					w(f, "		// because packed ints have not always been byte-aligned\n")
					w(f, "		remaining := PackedFormat(PACKED).ByteCount(version, int32(valueCount), %v) - %v*int64(valueCount)\n", bpv, bpv/8)
					w(f, "		for i := int64(0); i < remaining; i++ {\n")
					w(f, "			if _, err = in.ReadByte(); err != nil {\n")
					w(f, "				break\n")
					w(f, "			}\n")
					w(f, "		}\n")
				}
				w(f, "	}\n")
				w(f, "	return ans, err\n")
				w(f, "}\n\n")

				w(f, "func (d *Direct%v) Get(index int) int64 {\n", bpv)
				w(f, "	return int64(d.values[index])%s\n", MASKS[bpv])
				w(f, "}\n\n")

				w(f, "func (d *Direct%v) Set(index int, value int64) {\n", bpv)
				w(f, "	d.values[index] = %v(value)\n", typ)
				w(f, "}\n\n")

				w(f, "func (d *Direct%v) RamBytesUsed() int64 {\n", bpv)
				w(f, "	return util.AlignObjectSize(\n")
				w(f, "		util.NUM_BYTES_OBJECT_HEADER +\n")
				w(f, "			2*util.NUM_BYTES_INT +\n")
				w(f, "			util.NUM_BYTES_OBJECT_REF +\n")
				w(f, "			util.SizeOf(d.values))\n")
				w(f, "}\n")

				fmt.Printf("		case %v:\n", bpv)
				fmt.Printf("			return newDirect%vFromInput(version, in, int(valueCount))\n", bpv)
			}
		}
				gocog]]] */
		case 16:
			return newDirect16FromInput(version, in, int(valueCount))
		case 32:
			return newDirect32FromInput(version, in, int(valueCount))
		case 64:
			return newDirect64FromInput(version, in, int(valueCount))
		case 8:
			return newDirect8FromInput(version, in, int(valueCount))
		// [[[end]]]
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
		panic(fmt.Sprintf("Unknown Writer format: %v", format))
	}
}

func asUint32(n int32, err error) (n2 uint32, err2 error) {
	return uint32(n), err
}

func NewPackedReader(in DataInput) (r PackedIntsReader, err error) {
	if version, err := codec.CheckHeader(in, PACKED_CODEC_NAME, PACKED_VERSION_START, VERSION_CURRENT); err == nil {
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

/*
Expert: Restore a ReaderIterator from a stream without reading metadata at the
beginning of the stream. This method is useful to restore data from streams
which have been created using WriterNoHeader().
*/
func ReaderIteratorNoHeader(in DataInput, format PackedFormat, version,
	valueCount, bitsPerValue, mem int) ReaderIterator {
	CheckVersion(int32(version))
	return newPackedReaderIterator(format, version, valueCount, bitsPerValue, in, mem)
}

/*
Create a packed integer slice with the given amount of values
initialized to 0. The valueCount and the bitsPerValue cannot be
changed after creation. All Mutables known by this factory are kept
fully in RAM.

Positive values of acceptableOverheadRatio will trade space for speed
by selecting a faster but potentially less memory-efficient
implementation. An acceptableOverheadRatio of COMPACT will make sure
that the most memory-efficient implementation is selected whereas
FASTEST will make sure that the fastest implementation is selected.
*/
func MutableFor(valueCount, bitsPerValue int, acceptableOverheadRatio float32) Mutable {
	formatAndBits := FastestFormatAndBits(valueCount, bitsPerValue, acceptableOverheadRatio)
	return MutableForFormat(valueCount, formatAndBits.BitsPerValue, formatAndBits.Format)
}

/* Same as MutableFor() with a pre-computed number of bits per value and format. */
func MutableForFormat(vc, bpv int, format PackedFormat) Mutable {
	valueCount := int32(vc)
	bitsPerValue := uint32(bpv)
	assert(valueCount >= 0)
	switch int(format) {
	case PACKED_SINGLE_BLOCK:
		return newPacked64SingleBlockBy(valueCount, bitsPerValue)
	case PACKED:
		switch bitsPerValue {
		case 8:
			return newDirect8(vc)
		case 16:
			return newDirect16(vc)
		case 32:
			return newDirect32(vc)
		case 64:
			return newDirect64(vc)
		case 24:
			if valueCount <= PACKED8_THREE_BLOCKS_MAX_SIZE {
				return newPacked8ThreeBlocks(valueCount)
			}
		case 48:
			if valueCount <= PACKED16_THREE_BLOCKS_MAX_SIZE {
				return newPacked16ThreeBlocks(valueCount)
			}
		}
		return newPacked64(valueCount, bitsPerValue)
	}
	panic(fmt.Sprintf("Invalid format: %v", format))
}

/*
Expert: Create a packed integer array writer for the given output,
format, value count, and number of bits per value.

The resulting stream will be long-aligned. This means that depending
on the format which is used, up to 63 bits will be wasted. An easy
way to make sure tha tno space is lost is to always use a valueCount
that is a multiple of 64.

This method does not write any metadata to the stream, meaning that
it is your responsibility to store it somewhere else in order to be
able to recover data from the stream later on:

- format (using Format.Id()),
- valueCount,
- bitsPerValue,
- VERSION_CURRENT.

It is possible to start writing values without knowing how many of
them you are actually going to write. To do this, just pass -1 as
valueCount. On the other hand, for any positive value of valueCount,
the returned writer will make sure that you don't write more values
than expected and pad the end of stream with zeros in case you have
writen less than valueCount when calling Writer.Finish().

The mem parameter lets your control how much memory can be used to
buffer changes in memory before finishing to disk. High values of mem
are likely to improve throughput. On the other hand, if speed is not
that important to you, a value of 0 will use as little memory as
possible and should already offer reasonble throughput.
*/
func WriterNoHeader(out DataOutput, format PackedFormat,
	valueCount, bitsPerValue, mem int) Writer {
	return newPackedWriter(format, out, valueCount, bitsPerValue, mem)
}

/*
Returns how many bits are required to hold values up to and including maxValue
NOTE: This method returns at least 1.
*/
func BitsRequired(maxValue int64) int {
	assert2(maxValue >= 0, fmt.Sprintf("maxValue must be non-negative (got: %v)", maxValue))
	return UnsignedBitsRequired(maxValue)
}

/*
Returns how many bits are required to store bits, interpreted as an
unsigned value.
NOTE: This method returns at least 1.
*/
func UnsignedBitsRequired(bits int64) int {
	// TODO not efficient
	return len(strconv.FormatInt(bits, 2))
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

// Calculate the maximum unsigned long that can be expressed with the given number of bits
func MaxValue(bitsPerValue int) int64 {
	if bitsPerValue == 64 {
		return math.MaxInt64
	}
	return (1 << uint64(bitsPerValue)) - 1
}

/* Copy src[srcPos:srcPos+len] into dest[destPos:destPos+len] using at most mem bytes. */
func Copy(src PackedIntsReader, srcPos int, dest Mutable, destPos, length, mem int) {
	panic("not implemented yet")
}

var TrailingZeros = func() map[int]int {
	ans := make(map[int]int)
	var n = 1
	for i := 0; i < 32; i++ {
		ans[n] = i
		n <<= 1
	}
	return ans
}()

/* Check that the block size is a power of 2, in the right bounds, and return its log in base 2. */
func checkBlockSize(blockSize, minBlockSize, maxBlockSize int) int {
	assert2(blockSize >= minBlockSize && blockSize <= maxBlockSize,
		"blockSize must be >= %v and <= %v, got %v",
		minBlockSize, maxBlockSize, blockSize)
	assert2((blockSize&(blockSize-1)) == 0,
		"blockSIze must be a power of 2, got %v", blockSize)
	return TrailingZeros[blockSize]
}

/* Return the number of blocks required to store size values on blockSize. */
func numBlocks(size int64, blockSize int) int {
	numBlocks := int(size / int64(blockSize))
	if size%int64(blockSize) != 0 {
		numBlocks++
	}
	assert2(int64(numBlocks)*int64(blockSize) >= size, "size is too large for this block size")
	return numBlocks
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

func newPackedReaderIterator(format PackedFormat, packedIntsVersion, valueCount, bitsPerValue int, in DataInput, mem int) *PackedReaderIterator {
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

var PACKED8_THREE_BLOCKS_MAX_SIZE = int32(math.MaxInt32 / 3)

type Packed8ThreeBlocks struct {
	*MutableImpl
	blocks []byte
}

func newPacked8ThreeBlocks(valueCount int32) *Packed8ThreeBlocks {
	assert2(valueCount <= PACKED8_THREE_BLOCKS_MAX_SIZE, "MAX_SIZE exceeded")
	ans := &Packed8ThreeBlocks{blocks: make([]byte, valueCount*3)}
	ans.MutableImpl = newMutableImpl(int(valueCount), 24, ans)
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

func (r *Packed8ThreeBlocks) Set(index int, value int64) {
	panic("not implemented yet")
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

var PACKED16_THREE_BLOCKS_MAX_SIZE = int32(math.MaxInt32 / 3)

type Packed16ThreeBlocks struct {
	*MutableImpl
	blocks []int16
}

func newPacked16ThreeBlocks(valueCount int32) *Packed16ThreeBlocks {
	assert2(valueCount <= PACKED16_THREE_BLOCKS_MAX_SIZE, "MAX_SIZE exceeded")
	ans := &Packed16ThreeBlocks{blocks: make([]int16, valueCount*3)}
	ans.MutableImpl = newMutableImpl(int(valueCount), 48, ans)
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

func (r *Packed16ThreeBlocks) Set(index int, value int64) {
	panic("not implemented yet")
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
	ans.MutableImpl = newMutableImpl(int(valueCount), int(bitsPerValue), ans)
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
		return int64((uint64(p.blocks[elementPos]) >> uint64(-endBits)) & p.maskRight)
	}
	// Two blocks
	return ((p.blocks[elementPos] << uint64(endBits)) |
		int64(uint64(p.blocks[elementPos+1])>>(PACKED64_BLOCK_SIZE-uint64(endBits)))) &
		int64(p.maskRight)
}

func (p *Packed64) Set(index int, value int64) {
	panic("not implemented yet")
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

type GrowableWriter struct {
	currentMask             int64
	current                 Mutable
	acceptableOverheadRatio float32
}

func NewGrowableWriter(startBitsPerValue, valueCount int,
	acceptableOverheadRatio float32) *GrowableWriter {
	m := MutableFor(valueCount, startBitsPerValue, acceptableOverheadRatio)
	return &GrowableWriter{
		acceptableOverheadRatio: acceptableOverheadRatio,
		current:                 m,
		currentMask:             mask(m.BitsPerValue()),
	}
}

func mask(bitsPerValue int) int64 {
	if bitsPerValue == 64 {
		return ^0
	}
	return MaxValue(bitsPerValue)
}

func (w *GrowableWriter) Get(index int) int64 {
	return w.current.Get(index)
}

func (w *GrowableWriter) Size() int32 {
	return w.current.Size()
}

func (w *GrowableWriter) BitsPerValue() int {
	return w.current.BitsPerValue()
}

func (w *GrowableWriter) ensureCapacity(value int64) {
	if (value & w.currentMask) == value {
		return
	}
	var bitsRequired int
	if value < 0 {
		bitsRequired = 64
	} else {
		bitsRequired = UnsignedBitsRequired(value)
	}
	assert(bitsRequired > w.current.BitsPerValue())
	valueCount := int(w.Size())
	next := MutableFor(valueCount, bitsRequired, w.acceptableOverheadRatio)
	Copy(w.current, 0, next, 0, valueCount, DEFAULT_BUFFER_SIZE)
	w.current = next
	w.currentMask = mask(w.current.BitsPerValue())
}

func (w *GrowableWriter) Set(index int, value int64) {
	w.ensureCapacity(value)
	w.current.Set(index, value)
}

func (w *GrowableWriter) RamBytesUsed() int64 {
	return util.AlignObjectSize(
		util.NUM_BYTES_OBJECT_HEADER+
			util.NUM_BYTES_OBJECT_REF+
			util.NUM_BYTES_LONG+
			util.NUM_BYTES_FLOAT) +
		w.current.RamBytesUsed()
}

func (w *GrowableWriter) Save(out util.DataOutput) error {
	return w.current.Save(out)
}
