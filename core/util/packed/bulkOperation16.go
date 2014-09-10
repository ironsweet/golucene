// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked16 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked16() BulkOperation {
	return &BulkOperationPacked16{newBulkOperationPacked(16)}
}

func (op *BulkOperationPacked16) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(48); shift >= 0; shift -= 16 {
			values[valuesOffset] = int32((int64(uint64(block) >> shift)) & 65535); valuesOffset++
		}
	}
}

func (op *BulkOperationPacked16) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		values[valuesOffset] = (int32(blocks[blocksOffset+0]) << 8) | int32(blocks[blocksOffset+1])
		valuesOffset++
		blocksOffset += 2
	}
}
func (op *BulkOperationPacked16) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(48); shift >= 0; shift -= 16 {
			values[valuesOffset] = (int64(uint64(block) >> shift)) & 65535; valuesOffset++
		}
	}
}

func (op *BulkOperationPacked16) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		values[valuesOffset] = (int64(blocks[blocksOffset+0]) << 8) | int64(blocks[blocksOffset+1])
		valuesOffset++
		blocksOffset += 2
	}
}
