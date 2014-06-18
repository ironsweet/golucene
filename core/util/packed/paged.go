package packed

// packed/AbstractPagedMutable.java

// packed/PagedGrowableWriter.java

/*
A PagedGrowableWriter. This class slices data into fixed-size blocks
which have independent numbers of bits per value and grow on-demand.

You should use this class instead of the AbstractAppendingLongBuffer
related ones only when you need random write-access. Otherwise this
class will likely be slower and less memory-efficient.
*/
type PagedGrowableWriter struct {
}

/* Create a new PagedGrowableWriter instance. */
func NewPagedGrowableWriter(size int64, pageSize, startBitsPerValue int,
	acceptableOverheadRatio float32) *PagedGrowableWriter {
	panic("not implemented yet")
}
