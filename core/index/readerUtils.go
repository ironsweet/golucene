package index

/*
 * Returns index of the searcher/reader for document n in the
 * array used to construct this searcher/reader.
 */
func subIndex(n int, docStarts []int) int {
	// searcher/reader for doc n:
	size := len(docStarts)
	lo := 0        // search starts array
	hi := size - 1 // for first element less than n, return its index
	for hi >= lo {
		mid := int(uint(lo+hi) >> 1)
		midValue := docStarts[mid]
		if n < midValue {
			hi = mid - 1
		} else if n > midValue {
			lo = mid + 1
		} else { // found a match
			for mid+1 < size && docStarts[mid+1] == midValue {
				mid++ // scan to last match
			}
			return mid
		}
	}
	return hi
}
