// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked7 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked7() BulkOperation {
	return &BulkOperationPacked7{newBulkOperationPacked(7)}
}

func (op *BulkOperationPacked7) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 57)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 50) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 43) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 36) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 29) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 22) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 15) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 8) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 1) & 127); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 1) << 6) | (int64(uint64(block1) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 51) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 44) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 37) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 30) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 23) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 16) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 9) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 127); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 5) | (int64(uint64(block2) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 52) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 45) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 38) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 31) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 17) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 10) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 3) & 127); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 7) << 4) | (int64(uint64(block3) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 53) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 46) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 39) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 32) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 25) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 18) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 11) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 4) & 127); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 15) << 3) | (int64(uint64(block4) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 54) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 47) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 40) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 33) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 26) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 19) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 12) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 5) & 127); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 31) << 2) | (int64(uint64(block5) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 55) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 48) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 41) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 34) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 27) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 20) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 13) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 6) & 127); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 63) << 1) | (int64(uint64(block6) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 56) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 49) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 42) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 35) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 28) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 21) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 14) & 127); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 7) & 127); valuesOffset++
		values[valuesOffset] = int32(block6 & 127); valuesOffset++
	}
}

func (op *BulkOperationPacked7) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32( int64(uint8(byte0) >> 1))
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0 & 1) << 6) | int64(uint8(byte1) >> 2))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 3) << 5) | int64(uint8(byte2) >> 3))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 7) << 4) | int64(uint8(byte3) >> 4))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 15) << 3) | int64(uint8(byte4) >> 5))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 31) << 2) | int64(uint8(byte5) >> 6))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 63) << 1) | int64(uint8(byte6) >> 7))
		valuesOffset++
		values[valuesOffset] = int32( int64(byte6) & 127)
		valuesOffset++
	}
}
func (op *BulkOperationPacked7) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 57); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 50) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 43) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 36) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 29) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 22) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 15) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 8) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 1) & 127; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 6) | (int64(uint64(block1) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 51) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 44) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 37) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 30) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 23) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 16) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 9) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 127; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 5) | (int64(uint64(block2) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 52) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 45) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 38) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 31) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 17) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 10) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 3) & 127; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 7) << 4) | (int64(uint64(block3) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 53) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 46) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 39) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 32) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 25) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 18) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 11) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 4) & 127; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 3) | (int64(uint64(block4) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 54) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 47) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 40) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 33) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 26) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 19) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 12) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 5) & 127; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 2) | (int64(uint64(block5) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 55) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 48) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 41) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 34) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 27) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 20) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 13) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 6) & 127; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 1) | (int64(uint64(block6) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 56) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 49) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 42) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 35) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 28) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 21) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 14) & 127; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 7) & 127; valuesOffset++
		values[valuesOffset] = block6 & 127; valuesOffset++
	}
}

func (op *BulkOperationPacked7) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64( int64(uint8(byte0) >> 1))
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0 & 1) << 6) | int64(uint8(byte1) >> 2))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 3) << 5) | int64(uint8(byte2) >> 3))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 7) << 4) | int64(uint8(byte3) >> 4))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 15) << 3) | int64(uint8(byte4) >> 5))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 31) << 2) | int64(uint8(byte5) >> 6))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 63) << 1) | int64(uint8(byte6) >> 7))
		valuesOffset++
		values[valuesOffset] = int64( int64(byte6) & 127)
		valuesOffset++
	}
}
