// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked2 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked2() BulkOperation {
	return &BulkOperationPacked2{newBulkOperationPacked(2)}
}

func (op *BulkOperationPacked2) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(62); shift >= 0; shift -= 2 {
			values[valuesOffset] = int32((int64(uint64(block) >> shift)) & 3); valuesOffset++
		}
	}
}

func (op *BulkOperationPacked2) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		block := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 6) & 3
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 4) & 3
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 2) & 3
		valuesOffset++
		values[valuesOffset] = int32(block & 3)
		valuesOffset++
	}
}
func (op *BulkOperationPacked2) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(62); shift >= 0; shift -= 2 {
			values[valuesOffset] = (int64(uint64(block) >> shift)) & 3; valuesOffset++
		}
	}
}

func (op *BulkOperationPacked2) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		block := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 6) & 3
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 4) & 3
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 2) & 3
		valuesOffset++
		values[valuesOffset] = int64(block & 3)
		valuesOffset++
	}
}
