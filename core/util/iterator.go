package util

import "sort"

// BytesRefIterator.java
// A simple iterator interface for []byte iteration.
type BytesRefIterator interface {
	/* Increments the iteration to the next []byte in the iterator. Returns the
	resulting []byte or nil if the end of the iterator is reached. The returned
	[]byte may be re-used across calls to the next. After this method returns
	nil, do not call it again: the results are undefined. */
	Next() (buf []byte, err error)

	/* Return the []byte Comparator used to sort terms provided by the iterator.
	This may return nil if there are no items or the iterator is not sorted.
	Callers may invoke this method many times, so it's best to cache a single
	instance & reuse it. */
	Comparator() sort.Interface
}

var EMPTY_BYTES_REF_ITERATOR = &emptyBytesRefIterator{}

type emptyBytesRefIterator struct{}

func (iter *emptyBytesRefIterator) Next() (buf []byte, err error) {
	return nil, nil
}

func (iter *emptyBytesRefIterator) Comparator() sort.Interface {
	return nil
}
