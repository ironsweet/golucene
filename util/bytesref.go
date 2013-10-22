package util

type BytesRef struct {
	// The contents of the BytesRef. Should never be nil.
	Bytes []byte
	// Offset of first valid byte.
	Offset int
	// Length of used bytes.
	Length int
}

// const EMPTY_BYTES = make([]byte, 0)

// Create a BytesRef with empty bytes.
func NewBytesRef() *BytesRef {
	return &BytesRef{Bytes: make([]byte, 0)}
}

type BytesRefs []BytesRef

func (br BytesRefs) Len() int {
	return len(br)
}

func (br BytesRefs) Less(i, j int) bool {
	aBytes, bBytes := br[i], br[j]
	aLen, bLen := aBytes.Length, bBytes.Length

	for i, v := range aBytes.Bytes {
		if i >= bLen {
			break
		}
		if diff := v - bBytes.Bytes[i]; diff != 0 {
			return diff < 0
		}
	}

	// One is a prefix of the other, or, they are equal:
	return aLen < bLen
}

func (br BytesRefs) Swap(i, j int) {
	br[i], br[j] = br[j], br[i]
}
