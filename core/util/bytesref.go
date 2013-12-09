package util

type BytesRefs [][]byte

func (br BytesRefs) Len() int {
	return len(br)
}

func (br BytesRefs) Less(i, j int) bool {
	aBytes, bBytes := br[i], br[j]
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

func (br BytesRefs) Swap(i, j int) {
	br[i], br[j] = br[j], br[i]
}
