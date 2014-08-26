package packed

// packed/AbstractPagedMutable.java

const MIN_BLOCK_SIZE = 1 << 6
const MAX_BLOCK_SIZE = 1 << 30

type abstractPagedMutableSPI interface {
	newMutable(int, int) Mutable
}

type abstractPagedMutable struct {
	spi          abstractPagedMutableSPI
	size         int64
	pageShift    uint
	pageMask     int
	subMutables  []Mutable
	bitsPerValue int
}

func newAbstractPagedMutable(spi abstractPagedMutableSPI,
	bitsPerValue int, size int64, pageSize int) *abstractPagedMutable {
	numPages := numBlocks(size, pageSize)
	return &abstractPagedMutable{
		spi:          spi,
		bitsPerValue: bitsPerValue,
		size:         size,
		pageShift:    uint(checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE)),
		pageMask:     pageSize - 1,
		subMutables:  make([]Mutable, numPages),
	}
}

func (m *abstractPagedMutable) fillPages() {
	numPages := numBlocks(m.size, m.pageSize())
	for i := 0; i < numPages; i++ {
		// do not allocate for more entries than necessary on the last page
		var valueCount int
		if i == numPages-1 {
			valueCount = m.lastPageSize(m.size)
		} else {
			valueCount = m.pageSize()
		}
		m.subMutables[i] = m.spi.newMutable(valueCount, m.bitsPerValue)
	}
}

func (m *abstractPagedMutable) lastPageSize(size int64) int {
	if sz := m.indexInPage(size); sz != 0 {
		return sz
	}
	return m.pageSize()
}

func (m *abstractPagedMutable) pageSize() int {
	return m.pageMask + 1
}

func (m *abstractPagedMutable) Size() int64 {
	return m.size
}

func (m *abstractPagedMutable) pageIndex(index int64) int {
	return int(uint64(index) >> m.pageShift)
}

func (m *abstractPagedMutable) indexInPage(index int64) int {
	return int(index & int64(m.pageMask))
}

func (m *abstractPagedMutable) Get(index int64) int64 {
	assert(index >= 0 && index < m.size)
	pageIndex := m.pageIndex(index)
	indexInPage := m.indexInPage(index)
	return m.subMutables[pageIndex].Get(indexInPage)
}

/* set value at index. */
func (m *abstractPagedMutable) Set(index, value int64) {
	assert(index >= 0 && index < m.size)
	pageIndex := m.pageIndex(index)
	indexInPage := m.indexInPage(index)
	m.subMutables[pageIndex].Set(indexInPage, value)
}

// packed/PagedGrowableWriter.java

/*
A PagedGrowableWriter. This class slices data into fixed-size blocks
which have independent numbers of bits per value and grow on-demand.

You should use this class instead of the AbstractAppendingLongBuffer
related ones only when you need random write-access. Otherwise this
class will likely be slower and less memory-efficient.
*/
type PagedGrowableWriter struct {
	*abstractPagedMutable
	acceptableOverheadRatio float32
}

/* Create a new PagedGrowableWriter instance. */
func NewPagedGrowableWriter(size int64, pageSize, startBitsPerValue int,
	acceptableOverheadRatio float32) *PagedGrowableWriter {
	return newPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio, true)
}

func newPagedGrowableWriter(size int64, pageSize, startBitsPerValue int,
	acceptableOverheadRatio float32, fillPages bool) *PagedGrowableWriter {
	ans := &PagedGrowableWriter{acceptableOverheadRatio: acceptableOverheadRatio}
	ans.abstractPagedMutable = newAbstractPagedMutable(ans, startBitsPerValue, size, pageSize)
	if fillPages {
		ans.fillPages()
	}
	return ans
}

func (w *PagedGrowableWriter) newMutable(valueCount, bitsPerValue int) Mutable {
	return NewGrowableWriter(bitsPerValue, valueCount, w.acceptableOverheadRatio)
}
