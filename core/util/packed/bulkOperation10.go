// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked10 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked10() BulkOperation {
	return &BulkOperationPacked10{newBulkOperationPacked(10)}
}

func (op *BulkOperationPacked10) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 54)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 44) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 34) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 24) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 14) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 1023); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 15) << 6) | (int64(uint64(block1) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 48) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 28) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 18) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 1023); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 255) << 2) | (int64(uint64(block2) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 52) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 42) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 32) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 22) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 2) & 1023); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 3) << 8) | (int64(uint64(block3) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 46) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 36) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 26) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 16) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 6) & 1023); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 63) << 4) | (int64(uint64(block4) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 50) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 40) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 30) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 20) & 1023); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 10) & 1023); valuesOffset++
		values[valuesOffset] = int32(block4 & 1023); valuesOffset++
	}
}

func (op *BulkOperationPacked10) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 2) | int64(uint8(byte1) >> 6))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 63) << 4) | int64(uint8(byte2) >> 4))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 15) << 6) | int64(uint8(byte3) >> 2))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 3) << 8) | int64(byte4))
		valuesOffset++
	}
}
func (op *BulkOperationPacked10) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 54); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 44) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 34) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 24) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 14) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 1023; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 6) | (int64(uint64(block1) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 48) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 28) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 18) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 1023; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 2) | (int64(uint64(block2) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 52) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 42) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 32) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 22) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 2) & 1023; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 3) << 8) | (int64(uint64(block3) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 46) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 36) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 26) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 16) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 6) & 1023; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 63) << 4) | (int64(uint64(block4) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 50) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 40) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 30) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 20) & 1023; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 10) & 1023; valuesOffset++
		values[valuesOffset] = block4 & 1023; valuesOffset++
	}
}

func (op *BulkOperationPacked10) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 2) | int64(uint8(byte1) >> 6))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 63) << 4) | int64(uint8(byte2) >> 4))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 15) << 6) | int64(uint8(byte3) >> 2))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 3) << 8) | int64(byte4))
		valuesOffset++
	}
}
