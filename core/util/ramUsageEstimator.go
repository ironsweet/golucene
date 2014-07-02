package util

// util.RamUsageEstimator.java

// amd64 system
const (
	NUM_BYTES_CHAR  = 2 // UTF8 uses 1-4 bytes to represent each rune
	NUM_BYTES_INT   = 8
	NUM_BYTES_FLOAT = 4
	NUM_BYTES_LONG  = 8

	/* Number of bytes to represent an object reference */
	NUM_BYTES_OBJECT_REF = 8

	// Number of bytes to represent an object header (no fields, no alignments).
	NUM_BYTES_OBJECT_HEADER = 16

	// Number of bytes to represent an array header (no content, but with alignments).
	NUM_BYTES_ARRAY_HEADER = 24

	// A constant specifying the object alignment boundary inside the
	// JVM. Objects will always take a full multiple of this constant,
	// possibly wasting some space.
	NUM_BYTES_OBJECT_ALIGNMENT = 8
)

/* Aligns an object size to be the next multiple of NUM_BYTES_OBJECT_ALIGNMENT */
func AlignObjectSize(size int64) int64 {
	size += NUM_BYTES_OBJECT_ALIGNMENT - 1
	return size - (size & NUM_BYTES_OBJECT_ALIGNMENT)
}

/* Returns the size in bytes of the object. */
func SizeOf(arr interface{}) int64 {
	if arr == nil {
		return 0
	}
	switch arr.(type) {
	case []int64:
		return AlignObjectSize(NUM_BYTES_ARRAY_HEADER + NUM_BYTES_LONG*int64(len(arr.([]int64))))
	}
	panic("not supported yet")
}

/*
Estimates a "shallow" memory usage of the given object. For slices,
this will be the memory taken by slice storage (no subreferences will
be followed). For objects, this will be the memory taken by the fields.
*/
func ShallowSizeOf(obj interface{}) int64 {
	panic("not implemented yet")
}
