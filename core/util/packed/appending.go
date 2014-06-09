package packed

import (
	"github.com/balzaczyy/golucene/core/util"
)

// packed/AbstractAppendingLongBuffer.java

const MIN_PAGE_SIZE = 64

// More than 1M doesn't really makes sense with these appending buffers
// since their goal is to try to have small number of bits per value
const MAX_PAGE_SIZE = 1 << 20

/* Common functionality shared by AppendingDeltaPackedLongBuffer and MonotonicAppendingLongBuffer. */
type abstractAppendingLongBuffer struct {
	pageShift, pageMask     int
	values                  []PackedIntsReader
	valuesBytes             int64
	pending                 []int64
	acceptableOverheadRatio float32
}

func newAbstractAppendingLongBuffer(initialPageCount,
	pageSize int, acceptableOverheadRatio float32) *abstractAppendingLongBuffer {
	ps := checkBlockSize(pageSize, MIN_PAGE_SIZE, MAX_PAGE_SIZE)
	return &abstractAppendingLongBuffer{
		values:                  make([]PackedIntsReader, initialPageCount),
		pending:                 make([]int64, pageSize),
		pageShift:               ps,
		pageMask:                ps - 1,
		acceptableOverheadRatio: acceptableOverheadRatio,
	}
}

/* Get the numbre of values that have been added to the buffer. */
func (buf *abstractAppendingLongBuffer) Size() int64 {
	panic("not implemented yet")
}

/* Append a value to this buffer. */
func (buf *abstractAppendingLongBuffer) Add(l int64) {
	panic("not implemented yet")
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
	return util.AlignObjectSize(buf.baseRamBytesUsed()) +
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
	return &AppendingDeltaPackedLongBuffer{
		newAbstractAppendingLongBuffer(initialPageCount, pageSize, acceptableOverheadRatio),
		make([]int64, initialPageCount),
	}
}

/*
Create an AppendingDeltaPackedLongBuffer with initialPageCount=16,
pageSize=1024
*/
func NewAppendingDeltaPackedLongBufferWithOverhead(acceptableOverheadRatio float32) *AppendingDeltaPackedLongBuffer {
	return NewAppendingDeltaPackedLongBuffer(16, 1024, acceptableOverheadRatio)
}

func (buf *AppendingDeltaPackedLongBuffer) RamBytesUsed() int64 {
	return buf.abstractAppendingLongBuffer.RamBytesUsed() + util.SizeOf(buf.minValues)
}
