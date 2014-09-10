// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked4 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked4() BulkOperation {
	return &BulkOperationPacked4{newBulkOperationPacked(4)}
}

func (op *BulkOperationPacked4) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(60); shift >= 0; shift -= 4 {
			values[valuesOffset] = int32((int64(uint64(block) >> shift)) & 15); valuesOffset++
		}
	}
}

func (op *BulkOperationPacked4) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		block := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 4) & 15
		valuesOffset++
		values[valuesOffset] = int32(block & 15)
		valuesOffset++
	}
}
func (op *BulkOperationPacked4) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(60); shift >= 0; shift -= 4 {
			values[valuesOffset] = (int64(uint64(block) >> shift)) & 15; valuesOffset++
		}
	}
}

func (op *BulkOperationPacked4) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		block := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 4) & 15
		valuesOffset++
		values[valuesOffset] = int64(block & 15)
		valuesOffset++
	}
}
