package util

import (
	"math"
	"sort"
)

// util/ArrayUtil.java

/* Maximum length for an array */
const MAX_ARRAY_LENGTH = math.MaxInt32 - NUM_BYTES_ARRAY_HEADER

// L152
/** Returns an array size >= minTargetSize, generally
 *  over-allocating exponentially to achieve amortized
 *  linear-time cost as the array grows.
 *
 *  NOTE: this was originally borrowed from Python 2.4.2
 *  listobject.c sources (attribution in LICENSE.txt), but
 *  has now been substantially changed based on
 *  discussions from java-dev thread with subject "Dynamic
 *  array reallocation algorithms", started on Jan 12
 *  2010.
 *
 * @param minTargetSize Minimum required value to be returned.
 * @param bytesPerElement Bytes used by each element of
 * the array.  See constants in {@link RamUsageEstimator}.
 *
 * @lucene.internal
 */
func Oversize(minTargetSize int, bytesPerElement int) int {
	// catch usage that accidentally overflows int
	assert2(minTargetSize >= 0, "invalid array size %v", minTargetSize)

	if minTargetSize == 0 {
		// wait until at least one element is requested
		return 0
	}

	assert2(minTargetSize <= MAX_ARRAY_LENGTH,
		"requested array size %v exceeds maximum array in Go (%v)", MAX_ARRAY_LENGTH)

	// asymptotic exponential growth by 1/8th, favors
	// spending a bit more CPU to not tie up too much wasted
	// RAM:
	extra := minTargetSize >> 3
	if extra < 3 {
		// for very small arrays, where constant overhead of
		// realloc is presumably relatively high, we grow
		// faster
		extra = 3
	}

	newSize := minTargetSize + extra
	// add 7 to allow for worst case byte alignment addition below:
	if n := int32(newSize + 7); n < 0 || n > MAX_ARRAY_LENGTH {
		// int overflowed, or we exceeded the maximum array length
		return MAX_ARRAY_LENGTH
	}

	// Lucene support 32bit/64bit detection
	// However I assume golucene in 64bit only
	// if is64bit {
	// round up to 8 byte alignment in 64bit env
	switch bytesPerElement {
	case 4:
		// round up to multiple of 2
		return (newSize + 1) & 0x7ffffffe
	case 2:
		// round up to multiple of 4
		return (newSize + 3) & 0x7ffffffc
	case 1:
		// round up to multiple of 8
		return (newSize + 7) & 0x7ffffff8
	case 8:
		// no rounding
		return newSize
	default:
		// odd (invalid?) size
		return newSize
	}
	// }
}

// L285
func GrowIntSlice(arr []int, minSize int) []int {
	assert2(minSize >= 0, "size must be positive (got %v): likely integer overflow?", minSize)
	if len(arr) < minSize {
		newArr := make([]int, Oversize(minSize, NUM_BYTES_INT))
		copy(newArr, arr)
		return newArr
	}
	return arr
}

// L343
func GrowByteSlice(arr []byte, minSize int) []byte {
	assert2(minSize >= 0, "size must be positive (got %v): likely integer overflow?", minSize)
	if cap(arr) < minSize {
		newArr := make([]byte, Oversize(minSize, 1))
		copy(newArr, arr)
		return newArr[:minSize]
	}
	for len(arr) < minSize {
		arr = append(arr, 0)
	}
	return arr
}

// L699
/*
Sorts the given array slice in its own order. This method uses the
Tim sort algorithm, but falls back to binary sort for small arrays.
*/
func TimSort(data sort.Interface) {
	newArrayTimSorter(data, data.Len()/64).sort(0, data.Len())
}

// util/ArrayTimSorter.java

// A TimSorter for object arrays
type ArrayTimSorter struct {
	*TimSorter
	arr sort.Interface
	tmp []interface{}
}

// Create a new ArrayTimSorter
func newArrayTimSorter(arr sort.Interface, maxTempSlots int) *ArrayTimSorter {
	ans := &ArrayTimSorter{
		TimSorter: newTimSorter(arr, maxTempSlots),
		arr:       arr,
	}
	if maxTempSlots > 0 {
		ans.tmp = make([]interface{}, maxTempSlots)
	}
	return ans
}
