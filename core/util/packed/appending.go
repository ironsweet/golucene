package packed

// packed/AbstractAppendingLongBuffer.java

/* Common functionality shared by AppendingDeltaPackedLongBuffer and MonotonicAppendingLongBuffer. */
type abstractAppendingLongBuffer struct {
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
}

/*
Create an AppendingDeltaPackedLongBuffer with initialPageCount=16,
pageSize=1024
*/
func NewAppendingDeltaPackedLongBufferWithOverhead(acceptableOverheadRatio float32) *AppendingDeltaPackedLongBuffer {
	panic("not implemented yet")
}

func (buf *AppendingDeltaPackedLongBuffer) RamBytesUsed() int64 {
	panic("not implemented yet")
}
