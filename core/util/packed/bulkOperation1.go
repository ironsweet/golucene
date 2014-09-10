// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked1 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked1() BulkOperation {
	return &BulkOperationPacked1{newBulkOperationPacked(1)}
}

func (op *BulkOperationPacked1) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(63); shift >= 0; shift -= 1 {
			values[valuesOffset] = int32((int64(uint64(block) >> shift)) & 1); valuesOffset++
		}
	}
}

func (op *BulkOperationPacked1) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		block := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 7) & 1
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 6) & 1
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 5) & 1
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 4) & 1
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 3) & 1
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 2) & 1
		valuesOffset++
		values[valuesOffset] = int32(byte(uint8(block)) >> 1) & 1
		valuesOffset++
		values[valuesOffset] = int32(block & 1)
		valuesOffset++
	}
}
func (op *BulkOperationPacked1) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block := blocks[blocksOffset]; blocksOffset++
		for shift := uint(63); shift >= 0; shift -= 1 {
			values[valuesOffset] = (int64(uint64(block) >> shift)) & 1; valuesOffset++
		}
	}
}

func (op *BulkOperationPacked1) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for j := 0; j < iterations; j ++ {
		block := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 7) & 1
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 6) & 1
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 5) & 1
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 4) & 1
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 3) & 1
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 2) & 1
		valuesOffset++
		values[valuesOffset] = int64(byte(uint8(block)) >> 1) & 1
		valuesOffset++
		values[valuesOffset] = int64(block & 1)
		valuesOffset++
	}
}
