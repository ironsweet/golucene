// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked24 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked24() BulkOperation {
	return &BulkOperationPacked24{newBulkOperationPacked(24)}
}

func (op *BulkOperationPacked24) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 40)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 16) & 16777215); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 65535) << 8) | (int64(uint64(block1) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 32) & 16777215); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 16777215); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 255) << 16) | (int64(uint64(block2) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 16777215); valuesOffset++
		values[valuesOffset] = int32(block2 & 16777215); valuesOffset++
	}
}

func (op *BulkOperationPacked24) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 16) | (int64(byte1) << 8) | int64(byte2))
		valuesOffset++
	}
}
func (op *BulkOperationPacked24) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 40); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 16) & 16777215; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 65535) << 8) | (int64(uint64(block1) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 32) & 16777215; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 16777215; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 16) | (int64(uint64(block2) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 16777215; valuesOffset++
		values[valuesOffset] = block2 & 16777215; valuesOffset++
	}
}

func (op *BulkOperationPacked24) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 16) | (int64(byte1) << 8) | int64(byte2))
		valuesOffset++
	}
}
