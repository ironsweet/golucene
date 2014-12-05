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

/* Returns true or false for the specified bit index */
func (b *OpenBitSet) Get(index int64) bool {
	i := int(index >> 6) // div 64
	if i >= len(b.bits) {
		return false
	}
	bit := uint(index & 0x3f) // mod 64
	bitmask := int64(1) << bit
	return (b.bits[i] & bitmask) != 0
}

/* Sets a bit, expanding the set size if necessary */
func (b *OpenBitSet) Set(index int64) {
	panic("not implemented yet")
}

/* Returns the number of 64 bit words it would take to hold numBits */
func bits2words(numBits int64) int {
	return int((uint64(numBits-1) >> 6) + 1)
}

/* Expert: returns the []int64 storing the bits */
func (b *OpenBitSet) RealBits() []int64 { return b.bits }
