package packed

import (
	"github.com/balzaczyy/golucene/core/util"
)

// packed/AbstractAppendingLongBuffer.java

const MIN_PAGE_SIZE = 64

// More than 1M doesn't really makes sense with these appending buffers
// since their goal is to try to have small number of bits per value
const MAX_PAGE_SIZE = 1 << 20

type abstractAppendingLongBufferSPI interface {
	packPendingValues()
	grow(int)
	baseRamBytesUsed() int64
}

/* Common functionality shared by AppendingDeltaPackedLongBuffer and MonotonicAppendingLongBuffer. */
type abstractAppendingLongBuffer struct {
	spi                     abstractAppendingLongBufferSPI
	pageShift, pageMask     int
	values                  []PackedIntsReader
	valuesBytes             int64
	valuesOff               int
	pending                 []int64
	pendingOff              int
	acceptableOverheadRatio float32
}

func newAbstractAppendingLongBuffer(spi abstractAppendingLongBufferSPI,
	initialPageCount, pageSize int, acceptableOverheadRatio float32) *abstractAppendingLongBuffer {
	ps := checkBlockSize(pageSize, MIN_PAGE_SIZE, MAX_PAGE_SIZE)
	return &abstractAppendingLongBuffer{
		spi:                     spi,
		values:                  make([]PackedIntsReader, initialPageCount),
		pending:                 make([]int64, pageSize),
		pageShift:               ps,
		pageMask:                ps - 1,
		acceptableOverheadRatio: acceptableOverheadRatio,
	}
}

func (buf *abstractAppendingLongBuffer) pageSize() int {
	return buf.pageMask + 1
}

/* Get the numbre of values that have been added to the buffer. */
func (buf *abstractAppendingLongBuffer) Size() int64 {
	size := int64(buf.pendingOff)
	if buf.valuesOff > 0 {
		size += int64(buf.values[buf.valuesOff-1].Size())
	}
	if buf.valuesOff > 1 {
		size += int64((buf.valuesOff - 1) * buf.pageSize())
	}
	return size
}

/* Append a value to this buffer. */
func (buf *abstractAppendingLongBuffer) Add(l int64) {
	assert2(buf.pending != nil, "This buffer is frozen")
	if buf.pendingOff == len(buf.pending) {
		// check size
		if len(buf.values) == buf.valuesOff {
			newLength := util.Oversize(buf.valuesOff+1, 8)
			buf.spi.grow(newLength)
		}
		buf.spi.packPendingValues()
		buf.valuesBytes += buf.values[buf.valuesOff].RamBytesUsed()
		buf.valuesOff++
		// reset pending buffer
		buf.pendingOff = 0
	}
	buf.pending[buf.pendingOff] = 1
	buf.pendingOff++
}

func (buf *abstractAppendingLongBuffer) grow(newBlockCount int) {
	arr := make([]PackedIntsReader, newBlockCount)
	copy(arr, buf.values)
	buf.values = arr
}

func (buf *abstractAppendingLongBuffer) baseRamBytesUsed() int64 {
	return util.NUM_BYTES_OBJECT_HEADER +
		2*util.NUM_BYTES_OBJECT_REF + // the 2 arrays
		2*util.NUM_BYTES_INT + // the 2 offsets
		2*util.NUM_BYTES_INT + // pageShift, pageMask
		util.NUM_BYTES_FLOAT + // acceptable overhead
		util.NUM_BYTES_LONG // valuesBytes
}

/* Return the number of bytes used by this instance. */
func (buf *abstractAppendingLongBuffer) RamBytesUsed() int64 {
	// TODO: this is called per-doc-per-norm/dv-field, can we optimize this?
	return util.AlignObjectSize(buf.spi.baseRamBytesUsed()) +
		util.SizeOf(buf.pending) +
		util.AlignObjectSize(util.NUM_BYTES_ARRAY_HEADER+util.NUM_BYTES_OBJECT_REF*int64(len(buf.values))) +
		buf.valuesBytes
}

/*
Utility class to buffer a list of signed int64 in memory. This class
only supports appending and is optimized for the case where values
are close to each other.
*/
type AppendingDeltaPackedLongBuffer struct {
	*abstractAppendingLongBuffer
	minValues []int64
}

func NewAppendingDeltaPackedLongBuffer(initialPageCount,
	pageSize int, acceptableOverheadRatio float32) *AppendingDeltaPackedLongBuffer {
	ans := &AppendingDeltaPackedLongBuffer{minValues: make([]int64, initialPageCount)}
	ans.abstractAppendingLongBuffer = newAbstractAppendingLongBuffer(ans, initialPageCount, pageSize, acceptableOverheadRatio)
	return ans
}

/*
Create an AppendingDeltaPackedLongBuffer with initialPageCount=16,
pageSize=1024
*/
func NewAppendingDeltaPackedLongBufferWithOverhead(acceptableOverheadRatio float32) *AppendingDeltaPackedLongBuffer {
	return NewAppendingDeltaPackedLongBuffer(16, 1024, acceptableOverheadRatio)
}

func (buf *AppendingDeltaPackedLongBuffer) packPendingValues() {
	panic("not implemented yet")
}

func (buf *AppendingDeltaPackedLongBuffer) grow(newBlockCount int) {
	panic("not implemented yet")
}

func (buf *AppendingDeltaPackedLongBuffer) baseRamBytesUsed() int64 {
	return buf.abstractAppendingLongBuffer.baseRamBytesUsed() +
		util.NUM_BYTES_OBJECT_REF
}

func (buf *AppendingDeltaPackedLongBuffer) RamBytesUsed() int64 {
	return buf.abstractAppendingLongBuffer.RamBytesUsed() + util.SizeOf(buf.minValues)
}
