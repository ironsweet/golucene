package packed

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/util"
	"math"
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
func (f PackedFormat) IsSupported(bitsPerValue int) bool {
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
		assert(f.IsSupported(bitsPerValue))
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
	if acceptableOverheadRatio > PackedInts.FASTEST {
		acceptableOverheadRatio = PackedInts.FASTEST
	}
	acceptableOverheadRatioValue := acceptableOverheadRatio * float32(bitsPerValue) // in bits

	maxBitsPerValue := bitsPerValue + int(acceptableOverheadRatioValue)

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
			if PackedFormat(PACKED_SINGLE_BLOCK).IsSupported(bpv) {
				overhead := PackedFormat(PACKED_SINGLE_BLOCK).OverheadPerValue(bpv)
				acceptableOverhead := acceptableOverheadRatioValue + float32(bitsPerValue-bpv)
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
	ByteValueCount() int
	ByteBlockCount() int
	// Read iterations * valueCount() values from values, encode them
	// and write iterations * blockCount() blocks into blocks.
	// encodeLongToLong(values, blocks []int64, iterations int)
	// Read iterations * valueCount() values from values, encode them
	// and write 8 * iterations * blockCount() blocks into blocks.
	encodeLongToByte(values []int64, blocks []byte, iterations int)
	EncodeIntToByte(values []int, blocks []byte, iterations int)
}

type PackedIntsDecoder interface {
	// The minum numer of byte blocks to encode in a single iteration, when using byte encoding
	ByteBlockCount() int
	// The number of values that can be stored in byteBlockCount() byte blocks
	ByteValueCount() int
	decodeLongToLong(blocks, values []int64, iterations int)
	// Read 8 * iterations * blockCount() blocks from blocks, decodethem and write
	// iterations * valueCount() values inot values.
	decodeByteToLong(blocks []byte, values []int64, iterations int)
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
	Get(index int) int64 // NumericDocValue
	getBulk(int, []int64) int
	Size() int
}

type abstractReaderSPI interface {
	Get(index int) int64
	Size() int
}

type abstractReader struct {
	spi abstractReaderSPI
}

func newReader(spi abstractReaderSPI) *abstractReader {
	return &abstractReader{
		spi: spi,
	}
}

func (r *abstractReader) getBulk(index int, arr []int64) int {
	length := len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < r.spi.Size())

	gets := r.spi.Size() - index
	if length < gets {
		gets = length
	}
	for i, _ := range arr {
		arr[i] = r.spi.Get(index + i)
	}
	return gets
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
	setBulk(int, []int64) int
	Clear()
	// Save this mutable into out. Instantiating a reader from the
	// generated data will return a reader with the same number of bits
	// per value.
	Save(out util.DataOutput) error
}

type abstractMutableSPI interface {
	Get(index int) int64
	Set(index int, value int64)
	Size() int
}

type abstractMutable struct {
	*abstractReader
	spi abstractMutableSPI
}

func newMutable(spi abstractMutableSPI) *abstractMutable {
	return &abstractMutable{
		abstractReader: newReader(spi),
		spi:            spi,
	}
}

func (m *abstractMutable) setBulk(index int, arr []int64) int {
	length := len(arr)
	assert2(length > 0, "len must be > 0 (got %v)", length)
	assert(index >= 0 && index < m.spi.Size())

	for i, v := range arr {
		m.spi.Set(index+i, v)
	}
	return length
}

/* Fill the mutable [from,to) with val. */
func (m *abstractMutable) fill(from, to int, val int64) {
	panic("niy")
	// assert(val < MaxValue(m.BitsPerValue()))
	// assert(from <= to)
	// for i := from; i < to; i++ {
	// 	m.spi.Set(i, val)
	// }
}

/* Sets all values to 0 */
func (m *abstractMutable) Clear() {
	panic("niy")
	// m.fill(0, int(m.Size()), 0)
}

func (m *abstractMutable) Save(out util.DataOutput) error {
	panic("niy")
	// writer := WriterNoHeader(out, m.format, int(m.valueCount), m.bitsPerValue, DEFAULT_BUFFER_SIZE)
	// err := writer.writeHeader()
	// if err != nil {
	// 	return err
	// }
	// for i := 0; i < int(m.valueCount); i++ {
	// 	err = writer.Add(m.spi.Get(i))
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// return writer.Finish()
}

func (m *abstractMutable) Format() PackedFormat {
	return PackedFormat(PACKED)
}

type ReaderImpl struct {
	valueCount int
}

func newReaderImpl(valueCount int) *ReaderImpl {
	return &ReaderImpl{valueCount}
}

func (p *ReaderImpl) Size() int {
	return p.valueCount
}

type MutableImpl struct {
	*abstractMutable
	valueCount   int
	bitsPerValue int
}

func newMutableImpl(spi abstractMutableSPI,
	valueCount, bitsPerValue int) *MutableImpl {

	assert2(bitsPerValue > 0 && bitsPerValue <= 64, "%v", bitsPerValue)
	return &MutableImpl{
		abstractMutable: newMutable(spi),
		valueCount:      valueCount,
		bitsPerValue:    bitsPerValue,
	}
}

func (m *MutableImpl) BitsPerValue() int {
	return m.bitsPerValue
}

func (m *MutableImpl) Size() int {
	return m.valueCount
}

func ReaderNoHeader(in DataInput, format PackedFormat, version, valueCount int32,
	bitsPerValue uint32) (r PackedIntsReader, err error) {

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
				w(f, "	ans.MutableImpl = newMutableImpl(ans, valueCount, %v)\n", bpv)
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

				w(f, `
		func (d *Direct%v) Clear() {
			for i, _ := range d.values {
				d.values[i] = 0
			}
		}

		func (d *Direct%v) getBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %%v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			gets := d.valueCount - index
			if len(arr) < gets {
				gets = len(arr)
			}
			for i, _ := range arr[:gets] {
				arr[i] = int64(d.values[index+i])%v
			}
			return gets
		}

		func (d *Direct%v) setBulk(index int, arr []int64) int {
			assert2(len(arr) > 0, "len must be > 0 (got %%v)", len(arr))
			assert(index >= 0 && index < d.valueCount)

			sets := d.valueCount - index
			if len(arr) < sets {
				sets = len(arr)
			}
			for i, _ := range arr {
				d.values[index+i] = %varr[i])
			}
			return sets
		}

		func (d *Direct%v) fill(from, to int, val int64) {
			assert(val == val%v)
			for i := from; i < to; i ++ {
				d.values[i] = %vval)
			}
		}
				`, bpv, bpv, MASKS[bpv], bpv, CASTS[bpv], bpv, MASKS[bpv], CASTS[bpv])

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
					return ReaderNoHeader(in, format, version, valueCount, bitsPerValue)
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
	if bits == 0 {
		return 1
	}
	n := uint64(bits)
	ans := 0
	for n != 0 {
		n >>= 1
		ans++
	}
	return ans
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
	assert(srcPos+length <= int(src.Size()))
	assert(destPos+length <= int(dest.Size()))
	capacity := int(uint(mem) >> 3)
	if capacity == 0 {
		panic("niy")
	} else if length > 0 {
		// use bulk operations
		if length < capacity {
			capacity = length
		}
		buf := make([]int64, capacity)
		copyWith(src, srcPos, dest, destPos, length, buf)
	}
}

func copyWith(src PackedIntsReader, srcPos int, dest Mutable, destPos, length int, buf []int64) {
	assert(len(buf) > 0)
	remaining := 0
	for length > 0 {
		limit := remaining + length
		if limit > len(buf) {
			limit = len(buf)
		}
		read := src.getBulk(srcPos, buf[remaining:limit])
		assert(read > 0)
		srcPos += read
		length -= read
		remaining += read
		written := dest.setBulk(destPos, buf[:remaining])
		assert(written > 0)
		destPos += written
		if written < remaining {
			copy(buf, buf[written:remaining])
		}
		remaining -= written
	}
	for remaining > 0 {
		written := dest.setBulk(destPos, buf[:remaining])
		destPos += written
		remaining -= written
		copy(buf, buf[written:written+remaining])
	}
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
		it.bulkOperation.decodeByteToLong(it.nextBlocks, it.nextValues, it._iterations)
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
