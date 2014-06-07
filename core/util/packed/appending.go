package packed

// packed/AbstractAppendingLongBuffer.java

const MIN_PAGE_SIZE = 64

// More than 1M doesn't really makes sense with these appending buffers
// since their goal is to try to have small number of bits per value
const MAX_PAGE_SIZE = 1 << 20

/* Common functionality shared by AppendingDeltaPackedLongBuffer and MonotonicAppendingLongBuffer. */
type abstractAppendingLongBuffer struct {
	pageShift, pageMask     int
	values                  []PackedIntsReader
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

/* Return the number of bytes used by this instance. */
func (buf *abstractAppendingLongBuffer) RamBytesUsed() int64 {
	panic("not implemented yet")
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
	panic("not implemented yet")
}
