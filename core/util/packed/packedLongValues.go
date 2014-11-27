package packed

// util/packed/PackedLongValues.java

func NewDeltaPackedBuilder(acceptableOverheadRatio float32) *PackedLongValuesBuilder {
	panic("niy")
}

type PackedLongValues struct {
}

func (p *PackedLongValues) Size() int64 {
	panic("niy")
}

func (p *PackedLongValues) Iterator() func() (interface{}, bool) {
	panic("niy")
}

type PackedLongValuesBuilder struct {
}

/*
Build a PackedLongValues instance that contains values that have been
added to this builder. This operation is destructive.
*/
func (b *PackedLongValuesBuilder) Build() *PackedLongValues {
	panic("niy")
}

func (b *PackedLongValuesBuilder) RamBytesUsed() int64 {
	panic("niy")
}

/* Return the number of elements that have been added to this builder */
func (b *PackedLongValuesBuilder) Size() int64 {
	panic("niy")
}

/* Add a new element to this builder. */
func (b *PackedLongValuesBuilder) Add(l int64) *PackedLongValuesBuilder {
	panic("niy")
}
