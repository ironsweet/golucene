// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked14 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked14() BulkOperation {
	return &BulkOperationPacked14{newBulkOperationPacked(14)}
}

func (op *BulkOperationPacked14) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 50)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 36) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 22) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 8) & 16383); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 255) << 6) | (int64(uint64(block1) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 44) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 30) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 16) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 16383); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 12) | (int64(uint64(block2) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 38) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 10) & 16383); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 1023) << 4) | (int64(uint64(block3) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 46) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 32) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 18) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 4) & 16383); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 15) << 10) | (int64(uint64(block4) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 40) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 26) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 12) & 16383); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 4095) << 2) | (int64(uint64(block5) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 48) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 34) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 20) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 6) & 16383); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 63) << 8) | (int64(uint64(block6) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 42) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 28) & 16383); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 14) & 16383); valuesOffset++
		values[valuesOffset] = int32(block6 & 16383); valuesOffset++
	}
}

func (op *BulkOperationPacked14) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 6) | int64(uint8(byte1) >> 2))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 3) << 12) | (int64(byte2) << 4) | int64(uint8(byte3) >> 4))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 15) << 10) | (int64(byte4) << 2) | int64(uint8(byte5) >> 6))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 63) << 8) | int64(byte6))
		valuesOffset++
	}
}
func (op *BulkOperationPacked14) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 50); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 36) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 22) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 8) & 16383; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 255) << 6) | (int64(uint64(block1) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 44) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 30) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 16) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 16383; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 12) | (int64(uint64(block2) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 38) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 10) & 16383; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 1023) << 4) | (int64(uint64(block3) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 46) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 32) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 18) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 4) & 16383; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 10) | (int64(uint64(block4) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 40) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 26) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 12) & 16383; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 4095) << 2) | (int64(uint64(block5) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 48) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 34) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 20) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 6) & 16383; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 8) | (int64(uint64(block6) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 42) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 28) & 16383; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 14) & 16383; valuesOffset++
		values[valuesOffset] = block6 & 16383; valuesOffset++
	}
}

func (op *BulkOperationPacked14) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 6) | int64(uint8(byte1) >> 2))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 3) << 12) | (int64(byte2) << 4) | int64(uint8(byte3) >> 4))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 15) << 10) | (int64(byte4) << 2) | int64(uint8(byte5) >> 6))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 63) << 8) | int64(byte6))
		valuesOffset++
	}
}
