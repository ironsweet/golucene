package util

import (
	"fmt"
	"math"
	"sort"
)

// util/ArrayUtil.java

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
	if minTargetSize < 0 {
		// catch usage that accidentally overflows int
		panic(fmt.Sprintf("invalid array size %v", minTargetSize))
	}

	if minTargetSize == 0 {
		// wait until at least one element is requested
		return 0
	}

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
	if newSize+7 < 0 {
		// int overflowed -- return max allowed array size
		return int(math.MaxInt32)
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

// L699
/*
Sorts the given array slice in natural order. This method uses the
Tim sort algorithm, but falls back to binary sort for small arrays.
*/
func TimSort(data sort.Interface) {
	panic("not implemented yet")
}
