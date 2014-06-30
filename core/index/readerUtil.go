package index

// index/ReaderUtil.java

/*
Returns index of the searcher/reader for document n in the slice used
to construct this searcher/reader.
*/
func SubIndex(n int, leaves []*AtomicReaderContext) int {
	// find searcher/reader for doc n:
	size := len(leaves)
	lo, hi := 0, size-1
	// bi-search starts array, for first element less thann, return its index
	for lo <= hi {
		mid := (lo + hi) >> 1
		midValue := leaves[mid].DocBase
		if n < midValue {
			hi = mid - 1
		} else if n > midValue {
			lo = mid + 1
		} else { // found a match
			for mid+1 < size && leaves[mid+1].DocBase == midValue {
				mid++ // scan to last match
			}
			return mid
		}
	}
	return hi
}
