package packed

import (
	// "fmt"
	"github.com/balzaczyy/golucene/core/util"
)

type AppendingLongBuffer interface {
	Size() int64
	Get(int64) int64
	GetBulk(int64, []int64) int
	Add(int64)
	Iterator() func() (int64, bool)
	freeze()
	pageSize() int
}

// packed/AbstractAppendingLongBuffer.java

const MIN_PAGE_SIZE = 64

// More than 1M doesn't really makes sense with these appending buffers
// since their goal is to try to have small number of bits per value
const MAX_PAGE_SIZE = 1 << 20

type abstractAppendingLongBufferSPI interface {
	packPendingValues()
	get(int, int) int64
	get2Bulk(int, int, []int64) int
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
	pageShift := checkBlockSize(pageSize, MIN_PAGE_SIZE, MAX_PAGE_SIZE)
	return &abstractAppendingLongBuffer{
		spi:                     spi,
		values:                  make([]PackedIntsReader, initialPageCount),
		pending:                 make([]int64, pageSize),
		pageShift:               pageShift,
		pageMask:                pageSize - 1,
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
		size += int64(buf.valuesOff-1) * int64(buf.pageSize())
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
	buf.pending[buf.pendingOff] = l
	buf.pendingOff++
}

func (buf *abstractAppendingLongBuffer) grow(newBlockCount int) {
	arr := make([]PackedIntsReader, newBlockCount)
	copy(arr, buf.values)
	buf.values = arr
}

func (buf *abstractAppendingLongBuffer) Get(index int64) int64 {
	assert(index >= 0 && index < buf.Size())
	block := int(index >> uint(buf.pageShift))
	element := int(index & int64(buf.pageMask))
	return buf.spi.get(block, element)
}

func (buf *abstractAppendingLongBuffer) GetBulk(index int64, arr []int64) int {
	assert2(len(arr) > 0, "len must be > 0 (got %v)", len(arr))
	assert(index >= 0 && index < buf.Size())

	block := int(index >> uint(buf.pageShift))
	element := int(index & int64(buf.pageMask))
	return buf.spi.get2Bulk(block, element, arr)
}

func (buf *abstractAppendingLongBuffer) Iterator() func() (int64, bool) {
	var currentValues []int64
	vOff, pOff := 0, 0
	var currentCount int // number of entries of the current page

	fillValues := func() {
		if vOff == buf.valuesOff {
			currentValues = buf.pending
			currentCount = buf.pendingOff
		} else {
			currentCount = int(buf.values[vOff].Size())
			for k := 0; k < currentCount; {
				k += buf.spi.get2Bulk(vOff, k, currentValues[k:currentCount])
			}
		}
	}

	if buf.valuesOff == 0 {
		currentValues = buf.pending
		currentCount = buf.pendingOff
	} else {
		currentValues = make([]int64, int(buf.values[0].Size()))
		fillValues()
	}

	return func() (int64, bool) {
		if pOff >= currentCount {
			return 0, false
		}
		result := currentValues[pOff]
		pOff++
		if pOff == currentCount {
			vOff++
			pOff = 0
			if vOff <= buf.valuesOff {
				fillValues()
			} else {
				currentCount = 0
			}
		}
		return result, true
	}
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
		util.ShallowSizeOf(buf.values) +
		buf.valuesBytes
}

/* Pack all pending values in this buffer. Subsequent calls to add() will fail. */
func (buf *abstractAppendingLongBuffer) freeze() {
	if buf.pendingOff > 0 {
		if len(buf.values) == buf.valuesOff {
			buf.spi.grow(buf.valuesOff + 1) // don't oversize
		}
		buf.spi.packPendingValues()
		buf.valuesBytes += buf.values[buf.valuesOff].RamBytesUsed()
		buf.valuesOff++
		buf.pendingOff = 0
	}
	buf.pending = nil
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

func (buf *AppendingDeltaPackedLongBuffer) get(block, element int) int64 {
	if block == buf.valuesOff {
		return buf.pending[element]
	}
	if buf.values[block] == nil {
		return buf.minValues[block]
	}
	return buf.minValues[block] + buf.values[block].Get(element)
}

func (buf *AppendingDeltaPackedLongBuffer) get2Bulk(block, element int, arr []int64) int {
	if block == buf.valuesOff {
		sysCopyToRead := len(arr)
		if buf.pendingOff-element < sysCopyToRead {
			sysCopyToRead = buf.pendingOff - element
		}
		copy(arr, buf.pending[element:element+sysCopyToRead])
		return sysCopyToRead
	} else {
		// packed block
		read := buf.values[block].getBulk(element, arr)
		d := buf.minValues[block]
		for r := 0; r < read; r++ {
			arr[r] += d
		}
		return read
	}
}

func (buf *AppendingDeltaPackedLongBuffer) packPendingValues() {
	// compute max delta
	minValue, maxValue := buf.pending[0], buf.pending[0]
	for _, v := range buf.pending[1:buf.pendingOff] {
		if v < minValue {
			minValue = v
		} else if v > maxValue {
			maxValue = v
		}
	}
	delta := maxValue - minValue

	buf.minValues[buf.valuesOff] = minValue
	if delta == 0 {
		buf.values[buf.valuesOff] = newNilReader(buf.pendingOff)
	} else {
		// build a new packed reader
		bitsRequired := UnsignedBitsRequired(delta)
		for i := 0; i < buf.pendingOff; i++ {
			buf.pending[i] -= minValue
		}
		mutable := MutableFor(buf.pendingOff, bitsRequired, buf.acceptableOverheadRatio)
		for i := 0; i < buf.pendingOff; {
			i += mutable.setBulk(i, buf.pending[i:buf.pendingOff])
		}
		buf.values[buf.valuesOff] = mutable
	}
}

func (buf *AppendingDeltaPackedLongBuffer) grow(newBlockCount int) {
	buf.abstractAppendingLongBuffer.grow(newBlockCount)
	arr := make([]int64, newBlockCount)
	copy(arr, buf.minValues)
	buf.minValues = arr
}

func (buf *AppendingDeltaPackedLongBuffer) baseRamBytesUsed() int64 {
	return buf.abstractAppendingLongBuffer.baseRamBytesUsed() +
		util.NUM_BYTES_OBJECT_REF
}

func (buf *AppendingDeltaPackedLongBuffer) RamBytesUsed() int64 {
	return buf.abstractAppendingLongBuffer.RamBytesUsed() + util.SizeOf(buf.minValues)
}
