package packed

import (
	"github.com/balzaczyy/golucene/core/util"
	"reflect"
)

// util/packed/PackedLongValues.java

const DEFAULT_PAGE_SIZE = 1024
const MIN_PAGE_SIZE = 64
const MAX_PAGE_SIZE = 1 << 20

type PackedLongValues interface {
	Size() int64
	Iterator() func() (interface{}, bool)
}

type PackedLongValuesBuilder interface {
	util.Accountable
	Build() PackedLongValues
	Size() int64
	Add(int64) PackedLongValuesBuilder
}

func DeltaPackedBuilder(acceptableOverheadRatio float32) PackedLongValuesBuilder {
	return NewDeltaPackedLongValuesBuilder(DEFAULT_PAGE_SIZE, acceptableOverheadRatio)
}

type PackedLongValuesImpl struct {
	values              []PackedIntsReader
	pageShift, pageMask int
	size                int64
	ramBytesUsed        int64
}

func newPackedLongValues(pageShift, pageMask int,
	values []PackedIntsReader,
	size, ramBytesUsed int64) *PackedLongValuesImpl {

	return &PackedLongValuesImpl{
		pageShift:    pageShift,
		pageMask:     pageMask,
		values:       values,
		size:         size,
		ramBytesUsed: ramBytesUsed,
	}
}

func (p *PackedLongValuesImpl) Size() int64 {
	return p.size
}

func (p *PackedLongValuesImpl) decodeBlock(block int, dest []int64) int {
	vals := p.values[block]
	size := vals.Size()
	for k := 0; k < size; {
		k += vals.getBulk(k, dest[k:size])
	}
	return size
}

func (p *PackedLongValuesImpl) Iterator() func() (interface{}, bool) {
	currentValues := make([]int64, p.pageMask+1)
	vOff := 0
	pOff := 0
	currentCount := 0
	fillBlock := func() {
		if vOff == len(p.values) {
			currentCount = 0
		} else {
			currentCount = p.decodeBlock(vOff, currentValues)
			assert(currentCount > 0)
		}
	}
	fillBlock()
	return func() (v interface{}, ok bool) {
		if pOff < currentCount {
			v = currentValues[pOff]
			if pOff++; pOff == currentCount {
				vOff++
				pOff = 0
				fillBlock()
			}
		}
		return v, pOff < currentCount
	}
}

const INITIAL_PAGE_COUNT = 16

type PackedLongValuesBuilderImpl struct {
	pageShift, pageMask     int
	acceptableOverheadRatio float32
	pending                 []int64
	size                    int64

	values       []PackedIntsReader
	ramBytesUsed int64
	valuesOff    int
	pendingOff   int
}

func newPackedLongValuesBuilder(pageSize int,
	acceptableOverheadRatio float32) *PackedLongValuesBuilderImpl {

	ans := &PackedLongValuesBuilderImpl{
		pageShift:               checkBlockSize(pageSize, MIN_PAGE_SIZE, MAX_PAGE_SIZE),
		pageMask:                pageSize - 1,
		acceptableOverheadRatio: acceptableOverheadRatio,
		values:                  make([]PackedIntsReader, INITIAL_PAGE_COUNT),
		pending:                 make([]int64, pageSize),
	}
	ans.ramBytesUsed = util.ShallowSizeOfInstance(reflect.TypeOf(&PackedLongValuesBuilderImpl{})) +
		util.SizeOf(ans.pending) + util.ShallowSizeOf(ans.values)
	return ans
}

/*
Build a PackedLongValues instance that contains values that have been
added to this builder. This operation is destructive.
*/
func (b *PackedLongValuesBuilderImpl) Build() PackedLongValues {
	b.finish()
	b.pending = nil
	values := make([]PackedIntsReader, b.valuesOff)
	copy(values, b.values[:b.valuesOff])
	ramBytesUsed := util.ShallowSizeOfInstance(reflect.TypeOf(&PackedLongValuesImpl{})) +
		util.SizeOf(values)
	return newPackedLongValues(b.pageShift, b.pageMask, values, b.size, ramBytesUsed)
}

func (b *PackedLongValuesBuilderImpl) RamBytesUsed() int64 {
	return b.ramBytesUsed
}

/* Return the number of elements that have been added to this builder */
func (b *PackedLongValuesBuilderImpl) Size() int64 {
	return b.size
}

/* Add a new element to this builder. */
func (b *PackedLongValuesBuilderImpl) Add(l int64) PackedLongValuesBuilder {
	assert2(b.pending != nil, "Cannot be reused after build()")
	if b.pendingOff == len(b.pending) { // check size
		if b.valuesOff == len(b.values) {
			newLength := util.Oversize(b.valuesOff+1, 8)
			b.grow(newLength)
		}
		b.pack()
	}
	b.pending[b.pendingOff] = l
	b.pendingOff++
	b.size++
	return b
}

func (b *PackedLongValuesBuilderImpl) finish() {
	if b.pendingOff > 0 {
		if len(b.values) == b.valuesOff {
			b.grow(b.valuesOff + 1)
		}
		b.pack()
	}
}

func (b *PackedLongValuesBuilderImpl) pack() {
	b.packFrom(b.pending[:b.pendingOff], b.valuesOff, b.acceptableOverheadRatio)
	b.ramBytesUsed += b.values[b.valuesOff].RamBytesUsed()
	b.valuesOff++
	b.pendingOff = 0 // reset pending buffer
}

func (b *PackedLongValuesBuilderImpl) packFrom(values []int64,
	block int, acceptableOverheadRatio float32) {

	assert(len(values) > 0)

	// compute max delta
	minValue := values[0]
	maxValue := values[0]
	for _, v := range values[1:] {
		if v < minValue {
			minValue = v
		} else if v > maxValue {
			maxValue = v
		}
	}

	// build a new packed reader
	if minValue == 0 && maxValue == 0 {
		b.values[block] = newNilReader(len(values))
	} else {
		bitsRequired := 64
		if minValue >= 0 {
			bitsRequired = BitsRequired(maxValue)
		}
		mutable := MutableFor(len(values), bitsRequired, acceptableOverheadRatio)
		for i := 0; i < len(values); i++ {
			i += mutable.setBulk(i, values[i:])
		}
		b.values[block] = mutable
	}
}

func (b *PackedLongValuesBuilderImpl) grow(newBlockCount int) {
	panic("niy")
}

// util/packed/DeltaPackedLongValues.java

type DeltaPackedLongValuesBuilderImpl struct {
	*PackedLongValuesBuilderImpl
	mins []int64
}

func NewDeltaPackedLongValuesBuilder(pageSize int,
	acceptableOverheadRatio float32) *DeltaPackedLongValuesBuilderImpl {

	super := newPackedLongValuesBuilder(pageSize, acceptableOverheadRatio)
	ans := &DeltaPackedLongValuesBuilderImpl{
		PackedLongValuesBuilderImpl: super,
		mins: make([]int64, len(super.values)),
	}
	ans.ramBytesUsed += util.ShallowSizeOfInstance(reflect.TypeOf(&DeltaPackedLongValuesBuilderImpl{})) +
		util.SizeOf(ans.mins)
	return ans
}
