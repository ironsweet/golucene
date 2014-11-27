// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked6 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked6() BulkOperation {
	return &BulkOperationPacked6{newBulkOperationPacked(6)}
}

func (op *BulkOperationPacked6) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 58)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 52) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 46) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 40) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 34) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 28) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 22) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 16) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 10) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 63); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 15) << 2) | (int64(uint64(block1) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 56) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 50) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 44) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 32) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 26) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 20) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 14) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 63); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 4) | (int64(uint64(block2) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 54) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 48) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 42) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 36) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 30) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 18) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 63); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 6) & 63); valuesOffset++
		values[valuesOffset] = int32(block2 & 63); valuesOffset++
	}
}

func (op *BulkOperationPacked6) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32( int64(uint8(byte0) >> 2))
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0 & 3) << 4) | int64(uint8(byte1) >> 4))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 15) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		values[valuesOffset] = int32( int64(byte2) & 63)
		valuesOffset++
	}
}
func (op *BulkOperationPacked6) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 58); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 52) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 46) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 40) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 34) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 28) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 22) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 16) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 10) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 63; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 2) | (int64(uint64(block1) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 56) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 50) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 44) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 32) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 26) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 20) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 14) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 63; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 4) | (int64(uint64(block2) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 54) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 48) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 42) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 36) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 30) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 18) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 63; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 6) & 63; valuesOffset++
		values[valuesOffset] = block2 & 63; valuesOffset++
	}
}

func (op *BulkOperationPacked6) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64( int64(uint8(byte0) >> 2))
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0 & 3) << 4) | int64(uint8(byte1) >> 4))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 15) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		values[valuesOffset] = int64( int64(byte2) & 63)
		valuesOffset++
	}
}
