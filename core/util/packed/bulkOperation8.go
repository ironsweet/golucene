// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked8 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked8() BulkOperation {
	return &BulkOperationPacked8{newBulkOperationPacked(8)}
}

func (op *BulkOperationPacked8) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(56); shift >= 0; shift -= 8 {
			values[valuesOffset] = int32((int64(uint64(block) >> shift)) & 255); valuesOffset++
		}
	}
}

func (op *BulkOperationPacked8) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		values[valuesOffset] = int32(blocks[blocksOffset]); valuesOffset++; blocksOffset++
	}
}
func (op *BulkOperationPacked8) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(56); shift >= 0; shift -= 8 {
			values[valuesOffset] = (int64(uint64(block) >> shift)) & 255; valuesOffset++
		}
	}
}

func (op *BulkOperationPacked8) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		values[valuesOffset] = int64(blocks[blocksOffset]); valuesOffset++; blocksOffset++
	}
}
