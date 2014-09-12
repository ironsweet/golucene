package packed

import (
	"github.com/balzaczyy/golucene/core/util"
)

type GrowableWriter struct {
	*abstractMutable
	currentMask             int64
	current                 Mutable
	acceptableOverheadRatio float32
}

func NewGrowableWriter(startBitsPerValue, valueCount int,
	acceptableOverheadRatio float32) *GrowableWriter {
	m := MutableFor(valueCount, startBitsPerValue, acceptableOverheadRatio)
	ans := &GrowableWriter{
		acceptableOverheadRatio: acceptableOverheadRatio,
		current:                 m,
		currentMask:             mask(m.BitsPerValue()),
	}
	ans.abstractMutable = newMutable(ans)
	return ans
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

func (w *GrowableWriter) Size() int {
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

func (w *GrowableWriter) Clear() {
	w.current.Clear()
}

func (w *GrowableWriter) getBulk(index int, arr []int64) int {
	panic("niy")
}

func (w *GrowableWriter) setBulk(index int, arr []int64) int {
	panic("niy")
}

func (w *GrowableWriter) fill(from, to int, val int64) {
	panic("niy")
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
