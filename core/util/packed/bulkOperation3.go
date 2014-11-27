// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked3 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked3() BulkOperation {
	return &BulkOperationPacked3{newBulkOperationPacked(3)}
}

func (op *BulkOperationPacked3) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 61)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 58) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 55) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 52) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 49) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 46) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 43) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 40) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 37) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 34) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 31) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 28) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 25) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 22) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 19) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 16) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 13) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 10) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 7) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 1) & 7); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 1) << 2) | (int64(uint64(block1) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 59) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 56) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 53) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 50) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 47) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 44) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 41) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 35) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 32) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 29) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 26) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 23) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 20) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 17) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 14) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 11) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 5) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 7); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 1) | (int64(uint64(block2) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 60) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 57) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 54) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 51) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 48) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 45) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 42) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 39) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 36) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 33) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 30) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 27) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 21) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 18) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 15) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 9) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 6) & 7); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 3) & 7); valuesOffset++
		values[valuesOffset] = int32(block2 & 7); valuesOffset++
	}
}

func (op *BulkOperationPacked3) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32( int64(uint8(byte0) >> 5))
		valuesOffset++
		values[valuesOffset] = int32( int64(uint8(byte0) >> 2) & 7)
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0 & 3) << 1) | int64(uint8(byte1) >> 7))
		valuesOffset++
		values[valuesOffset] = int32( int64(uint8(byte1) >> 4) & 7)
		valuesOffset++
		values[valuesOffset] = int32( int64(uint8(byte1) >> 1) & 7)
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 1) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		values[valuesOffset] = int32( int64(uint8(byte2) >> 3) & 7)
		valuesOffset++
		values[valuesOffset] = int32( int64(byte2) & 7)
		valuesOffset++
	}
}
func (op *BulkOperationPacked3) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 61); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 58) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 55) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 52) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 49) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 46) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 43) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 40) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 37) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 34) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 31) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 28) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 25) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 22) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 19) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 16) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 13) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 10) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 7) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 1) & 7; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 2) | (int64(uint64(block1) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 59) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 56) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 53) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 50) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 47) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 44) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 41) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 35) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 32) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 29) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 26) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 23) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 20) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 17) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 14) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 11) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 5) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 7; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 1) | (int64(uint64(block2) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 60) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 57) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 54) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 51) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 48) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 45) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 42) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 39) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 36) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 33) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 30) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 27) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 21) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 18) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 15) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 9) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 6) & 7; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 3) & 7; valuesOffset++
		values[valuesOffset] = block2 & 7; valuesOffset++
	}
}

func (op *BulkOperationPacked3) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64( int64(uint8(byte0) >> 5))
		valuesOffset++
		values[valuesOffset] = int64( int64(uint8(byte0) >> 2) & 7)
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0 & 3) << 1) | int64(uint8(byte1) >> 7))
		valuesOffset++
		values[valuesOffset] = int64( int64(uint8(byte1) >> 4) & 7)
		valuesOffset++
		values[valuesOffset] = int64( int64(uint8(byte1) >> 1) & 7)
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 1) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		values[valuesOffset] = int64( int64(uint8(byte2) >> 3) & 7)
		valuesOffset++
		values[valuesOffset] = int64( int64(byte2) & 7)
		valuesOffset++
	}
}
