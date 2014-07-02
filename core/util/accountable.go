package util

/* An object whose RAM usage can be computed. */
type Accountable interface {
	// Return the memory usage of this object in bytes. Negative values are illegal.
	RamBytesUsed() int64
}
