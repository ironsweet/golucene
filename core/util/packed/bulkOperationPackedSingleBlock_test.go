package packed

import (
	"testing"
)

var values = []int64{
	6, 5, 7, 4, 4, 8, 8, 8, 4, 8, 6, 8, 8, 8, 8, 8, 4, 8, 3, 8, 4, 6,
	8, 5, 8, 7, 8, 6, 5, 4, 8, 3, 6, 8, 7, 8, 7, 8, 6, 8, 6, 6, 8, 6,
	6, 8, 8, 7, 8, 8, 7, 8, 7, 8, 5, 8, 4, 8, 6, 8, 5, 5, 4, 4, 6, 6,
	4, 8, 6, 8, 6, 4, 5, 4, 8, 6, 5, 8, 6, 3, 4, 5, 6, 7, 8, 3, 8, 4,
	8, 5, 8, 5, 8, 4, 8, 7, 8, 7, 8, 4, 6, 6, 3, 8, 8, 8, 6, 8, 4, 7,
	8, 6,
}

func TestEncodeLongToByte(t *testing.T) {
	p := newBulkOperationPackedSingleBlock(4)
	blocks := make([]byte, 56)
	p.encodeLongToByte(values, blocks, 7)
	for _, v := range blocks {
		if v == 0 {
			t.Errorf("should have no zero (got %v)", blocks)
			break
		}
	}
	if blocks[16] != 120 {
		t.Errorf("blocks[16] = 120 (got %v)", blocks[16])
	}
}
