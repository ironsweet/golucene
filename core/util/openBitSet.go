package util

type OpenBitSet struct {
	bits    []int64
	wlen    int   // number of words (elements) used in the array
	numBits int64 // for assert only
}

/* Constructs an OpenBitSet large enough to hold numBits. */
func NewOpenBitSetOf(numBits int64) *OpenBitSet {
	assert(numBits > 0)
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
	bitmask := int64(1) << uint(index)
	return (b.bits[i] & bitmask) != 0
}

/* Sets a bit, expanding the set size if necessary */
func (b *OpenBitSet) Set(index int64) {
	wordNum := b.expandingWordNum(index)
	bitmask := int64(1) << uint64(index)
	b.bits[wordNum] |= bitmask
}

func (b *OpenBitSet) expandingWordNum(index int64) int {
	wordNum := int(index >> 6)
	if wordNum >= b.wlen {
		b.ensureCapacity(index + 1)
	}
	return wordNum
}

/* Clears a bit, allowing access beyond the current set size without changing the size. */
func (b *OpenBitSet) Clear(index int64) {
	panic("niy")
}

/* Returns the number of 64 bit words it would take to hold numBits */
func bits2words(numBits int64) int {
	return int((uint64(numBits-1) >> 6) + 1)
}

/* Expert: returns the []int64 storing the bits */
// func (b *OpenBitSet) RealBits() []int64 { return b.bits }

func (b *OpenBitSet) And(other *OpenBitSet) {
	panic("niy")
}

/* Expand the []int64 with the size given as a number of words (64 bits long). */
func (b *OpenBitSet) ensureCapacityWords(numWords int) {
	if len(b.bits) < numWords {
		arr := make([]int64, numWords)
		copy(arr, b.bits)
		b.bits = arr
	}
	b.wlen = numWords
	if n := int64(numWords) << 6; n > b.numBits {
		b.numBits = n
	}
}

/* Ensure that the []int64 is big enough to hold numBits, expanding it if necessary. */
func (b *OpenBitSet) ensureCapacity(numBits int64) {
	b.ensureCapacityWords(bits2words(numBits))
	// ensureCapacityWords sets numBits to a multiple of 64, but we
	// want to set it to exactly what the app asked.
	if numBits > b.numBits {
		b.numBits = numBits
	}
}
