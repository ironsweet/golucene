package util

/*
BitSet of fixed length (numBits), backed by accessible bits() []int64,
accessed with an int index, implementing Bits and DocIdSet. Unlike
OpenBitSet, this bit set does not auto-expand, cannot handle long
index, and does not have fastXX/XX variants (just X).
*/
type FixedBitSet struct {
	bits       []int64
	numBits    int
	wordLength int
}

/* returns the number of 64 bit words it would take to hold numBits */
func fbits2words(numBits int) int {
	numLong := int(uint(numBits) >> 6)
	if (numBits & 63) != 0 {
		numLong++
	}
	return numLong
}

func NewFixedBitSetOf(numBits int) *FixedBitSet {
	wordLength := fbits2words(numBits)
	return &FixedBitSet{
		numBits:    numBits,
		bits:       make([]int64, wordLength),
		wordLength: wordLength,
	}
}

/*
Returns number of set bits. NOTE: this visits every int64 in the
backing bits slice, and the result is not internaly cached!
*/
func (b *FixedBitSet) Cardinality() int {
	panic("not implemented yet")
}

func (b *FixedBitSet) Set(index int) {
	assert2(index >= 0 && index < b.numBits, "index=%v numBits=%v", index, b.numBits)
	wordNum := index >> 6     // div 64
	bit := uint(index & 0x3f) // mod 64
	bitmask := int64(1 << bit)
	b.bits[wordNum] |= bitmask
}
