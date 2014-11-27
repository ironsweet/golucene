// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked12 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked12() BulkOperation {
	return &BulkOperationPacked12{newBulkOperationPacked(12)}
}

func (op *BulkOperationPacked12) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 52)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 40) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 28) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 16) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 4095); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 15) << 8) | (int64(uint64(block1) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 44) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 32) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 20) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 4095); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 255) << 4) | (int64(uint64(block2) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 48) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 36) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 4095); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 4095); valuesOffset++
		values[valuesOffset] = int32(block2 & 4095); valuesOffset++
	}
}

func (op *BulkOperationPacked12) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 4) | int64(uint8(byte1) >> 4))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 15) << 8) | int64(byte2))
		valuesOffset++
	}
}
func (op *BulkOperationPacked12) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 52); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 40) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 28) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 16) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 4095; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 8) | (int64(uint64(block1) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 44) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 32) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 20) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 4095; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 4) | (int64(uint64(block2) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 48) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 36) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 4095; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 4095; valuesOffset++
		values[valuesOffset] = block2 & 4095; valuesOffset++
	}
}

func (op *BulkOperationPacked12) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 4) | int64(uint8(byte1) >> 4))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 15) << 8) | int64(byte2))
		valuesOffset++
	}
}
