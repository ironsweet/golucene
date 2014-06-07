package util

// util.RamUsageEstimator.java

// amd64 system
const (
	NUM_BYTES_CHAR = 2 // UTF8 uses 1-4 bytes to represent each rune
	NUM_BYTES_INT  = 8
	NUM_BYTES_LONG = 8

	/* Number of bytes to represent an object reference */
	NUM_BYTES_OBJECT_REF = 8

	// Number of bytes to represent an array header (no content, but with alignments).
	NUM_BYTES_ARRAY_HEADER = 24

	// A constant specifying the object alignment boundary inside the
	// JVM. Objects will always take a full multiple of this constant,
	// possibly wasting some space.
	NUM_BYTES_OBJECT_ALIGNMENT = 8
)

/* Aligns an object size to be the next multiple of NUM_BYTES_OBJECT_ALIGNMENT */
func alignObjectSize(size int64) int64 {
	size += NUM_BYTES_OBJECT_ALIGNMENT - 1
	return size - (size & NUM_BYTES_OBJECT_ALIGNMENT)
}

/* Returns the size in bytes of the object. */
func SizeOf(arr interface{}) int64 {
	switch arr.(type) {
	case []int64:
		return alignObjectSize(NUM_BYTES_ARRAY_HEADER + NUM_BYTES_LONG*int64(len(arr.([]int64))))
	}
	panic("not supported yet")
}
