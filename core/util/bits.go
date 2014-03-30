package util

// util/Bits.java

/**
 * Interface for Bitset-like structures.
 * @lucene.experimental
 */
type Bits interface {
	/**
	 * Returns the value of the bit with the specified <code>index</code>.
	 * @param index index, should be non-negative and &lt; {@link #length()}.
	 *        The result of passing negative or out of bounds values is undefined
	 *        by this interface, <b>just don't do it!</b>
	 * @return <code>true</code> if the bit is set, <code>false</code> otherwise.
	 */
	At(index int) bool

	// Returns the number of bits in the set
	Length() int
}

// util/MutableBits.java

/* Extension of Bits for live documents. */
type MutableBits interface {
	Bits
	// Sets the bit specified by index to false.
	Clear(index int)
}
