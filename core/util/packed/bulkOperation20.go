// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked20 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked20() BulkOperation {
	return &BulkOperationPacked20{newBulkOperationPacked(20)}
}

func (op *BulkOperationPacked20) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 44)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 24) & 1048575); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 1048575); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 15) << 16) | (int64(uint64(block1) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 28) & 1048575); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 1048575); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 255) << 12) | (int64(uint64(block2) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 32) & 1048575); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 1048575); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 4095) << 8) | (int64(uint64(block3) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 36) & 1048575); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 16) & 1048575); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 65535) << 4) | (int64(uint64(block4) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 40) & 1048575); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 20) & 1048575); valuesOffset++
		values[valuesOffset] = int32(block4 & 1048575); valuesOffset++
	}
}

func (op *BulkOperationPacked20) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 12) | (int64(byte1) << 4) | int64(uint8(byte2) >> 4))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 15) << 16) | (int64(byte3) << 8) | int64(byte4))
		valuesOffset++
	}
}
func (op *BulkOperationPacked20) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 44); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 24) & 1048575; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 1048575; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 16) | (int64(uint64(block1) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 28) & 1048575; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 1048575; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 12) | (int64(uint64(block2) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 32) & 1048575; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 1048575; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 8) | (int64(uint64(block3) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 36) & 1048575; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 16) & 1048575; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 65535) << 4) | (int64(uint64(block4) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 40) & 1048575; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 20) & 1048575; valuesOffset++
		values[valuesOffset] = block4 & 1048575; valuesOffset++
	}
}

func (op *BulkOperationPacked20) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 12) | (int64(byte1) << 4) | int64(uint8(byte2) >> 4))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 15) << 16) | (int64(byte3) << 8) | int64(byte4))
		valuesOffset++
	}
}
