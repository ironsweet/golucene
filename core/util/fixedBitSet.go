package util

/*
BitSet of fixed length (numBits), backed by accessible bits() []int64,
accessed with an int index, implementing Bits and DocIdSet. Unlike
OpenBitSet, this bit set does not auto-expand, cannot handle long
index, and does not have fastXX/XX variants (just X).
*/
type FixedBitSet struct {
	bits     []int64
	numBits  int
	numWords int
}

/*
If the given FixedBitSet is large enough to hold numBits, returns the
given bits, otherwise returns a new FixedBitSet which can hold the
rquired number of bits.

NOTE: the returned bitset reuses the underlying []int64 of the given
bits if possible. Also, calling length() on the returned bits may
return a value greater than numBits.
*/
func EnsureFixedBitSet(bits *FixedBitSet, numBits int) *FixedBitSet {
	panic("not implemented yet")
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
		numBits:  numBits,
		bits:     make([]int64, wordLength),
		numWords: wordLength,
	}
}

func (b *FixedBitSet) Bits() Bits {
	return b
}

func (b *FixedBitSet) Length() int {
	return b.numBits
}

func (b *FixedBitSet) IsCacheable() bool {
	return true
}

func (b *FixedBitSet) RamBytesUsed() int64 {
	panic("not implemented yet")
}

/*
Returns number of set bits. NOTE: this visits every int64 in the
backing bits slice, and the result is not internaly cached!
*/
func (b *FixedBitSet) Cardinality() int {
	return int(pop_array(b.bits))
}

func (b *FixedBitSet) At(index int) bool {
	panic("not implemented yet")
}

func (b *FixedBitSet) Set(index int) {
	assert2(index >= 0 && index < b.numBits, "index=%v, numBits=%v", index, b.numBits)
	wordNum := index >> 6 // div 64
	bitmask := int64(1 << uint(index))
	b.bits[wordNum] |= bitmask
}
