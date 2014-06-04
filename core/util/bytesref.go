package util

// util/BytesRef.java

/*
Represents []byte, as a slice (offset + length) into an existing
[]byte, similar to Go's byte slice.

Important note: Unless otherwise noted, GoLucene uses []byte directly
to represent terms that are encoded as UTF8 bytes in the index. It
uses this class in cases when caller needs to hold a reference, while
allowing underlying []byte to change.
*/
type BytesRef struct {
	// The contents of the BytesRef.
	Value []byte
}

func NewBytesRef(bytes []byte) *BytesRef {
	return &BytesRef{bytes}
}

func UTF8SortedAsUnicodeLess(aBytes, bBytes []byte) bool {
	aLen, bLen := len(aBytes), len(bBytes)

	for i, _ := range aBytes {
		if i >= bLen {
			break
		}
		if diff := aBytes[i] - bBytes[i]; diff != 0 {
			return diff < 0
		}
	}

	// One is a prefix of the other, or, they are equal:
	return aLen < bLen
}

type BytesRefs [][]byte

func (br BytesRefs) Len() int {
	return len(br)
}

func (br BytesRefs) Less(i, j int) bool {
	aBytes, bBytes := br[i], br[j]
	return UTF8SortedAsUnicodeLess(aBytes, bBytes)
}

func (br BytesRefs) Swap(i, j int) {
	br[i], br[j] = br[j], br[i]
}
