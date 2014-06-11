package util

/*
BitSet of fixed length (numBits), backed by accessible bits() []int64,
accessed with an int index, implementing Bits and DocIdSet. Unlike
OpenBitSet, this bit set does not auto-expand, cannot handle long
index, and does not have fastXX/XX variants (just X).
*/
type FixedBitSet struct {
}

func NewFixedBitSetOf(numBits int) *FixedBitSet {
	panic("not implemented yet")
}

/*
Returns number of set bits. NOTE: this visits every int64 in the
backing bits slice, and the result is not internaly cached!
*/
func (b *FixedBitSet) Cardinality() int {
	panic("not implemented yet")
}

func (b *FixedBitSet) Set(index int) {
	panic("not implemented yet")
}
