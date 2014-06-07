package util

type OpenBitSet struct {
	bits    []int64
	wlen    int   // number of words (elements) used in the array
	numBits int64 // for assert only
}

/* Constructs an OpenBitSet large enough to hold numBits. */
func NewOpenBitSetOf(numBits int64) *OpenBitSet {
	bits := make([]int64, bits2words(numBits))
	return &OpenBitSet{
		numBits: numBits,
		bits:    bits,
		wlen:    len(bits),
	}
}

func NewOpenBitSet() *OpenBitSet {
	return NewOpenBitSetOf(64)
}

/* Returns the number of 64 bit words it would take to hold numBits */
func bits2words(numBits int64) int {
	return int((uint64(numBits-1) >> 6) + 1)
}

/* Expert: returns the []int64 storing the bits */
func (b *OpenBitSet) RealBits() []int64 { return b.bits }
