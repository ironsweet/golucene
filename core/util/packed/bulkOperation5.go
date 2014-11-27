// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked5 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked5() BulkOperation {
	return &BulkOperationPacked5{newBulkOperationPacked(5)}
}

func (op *BulkOperationPacked5) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 59)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 54) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 49) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 44) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 39) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 34) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 29) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 24) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 19) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 14) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 9) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 31); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 15) << 1) | (int64(uint64(block1) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 58) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 53) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 48) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 43) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 33) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 28) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 23) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 18) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 13) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 3) & 31); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 7) << 2) | (int64(uint64(block2) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 57) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 52) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 47) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 42) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 37) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 32) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 27) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 22) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 17) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 7) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 2) & 31); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 3) << 3) | (int64(uint64(block3) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 56) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 51) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 46) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 41) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 36) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 31) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 26) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 21) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 16) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 11) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 6) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 1) & 31); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 1) << 4) | (int64(uint64(block4) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 55) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 50) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 45) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 40) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 35) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 30) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 25) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 20) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 15) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 10) & 31); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 5) & 31); valuesOffset++
		values[valuesOffset] = int32(block4 & 31); valuesOffset++
	}
}

func (op *BulkOperationPacked5) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32( int64(uint8(byte0) >> 3))
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0 & 7) << 2) | int64(uint8(byte1) >> 6))
		valuesOffset++
		values[valuesOffset] = int32( int64(uint8(byte1) >> 1) & 31)
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 1) << 4) | int64(uint8(byte2) >> 4))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 15) << 1) | int64(uint8(byte3) >> 7))
		valuesOffset++
		values[valuesOffset] = int32( int64(uint8(byte3) >> 2) & 31)
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 3) << 3) | int64(uint8(byte4) >> 5))
		valuesOffset++
		values[valuesOffset] = int32( int64(byte4) & 31)
		valuesOffset++
	}
}
func (op *BulkOperationPacked5) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 59); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 54) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 49) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 44) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 39) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 34) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 29) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 24) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 19) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 14) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 9) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 31; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 1) | (int64(uint64(block1) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 58) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 53) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 48) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 43) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 33) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 28) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 23) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 18) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 13) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 3) & 31; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 7) << 2) | (int64(uint64(block2) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 57) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 52) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 47) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 42) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 37) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 32) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 27) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 22) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 17) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 7) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 2) & 31; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 3) << 3) | (int64(uint64(block3) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 56) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 51) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 46) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 41) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 36) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 31) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 26) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 21) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 16) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 11) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 6) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 1) & 31; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 1) << 4) | (int64(uint64(block4) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 55) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 50) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 45) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 40) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 35) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 30) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 25) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 20) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 15) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 10) & 31; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 5) & 31; valuesOffset++
		values[valuesOffset] = block4 & 31; valuesOffset++
	}
}

func (op *BulkOperationPacked5) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64( int64(uint8(byte0) >> 3))
		valuesOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0 & 7) << 2) | int64(uint8(byte1) >> 6))
		valuesOffset++
		values[valuesOffset] = int64( int64(uint8(byte1) >> 1) & 31)
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 1) << 4) | int64(uint8(byte2) >> 4))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 15) << 1) | int64(uint8(byte3) >> 7))
		valuesOffset++
		values[valuesOffset] = int64( int64(uint8(byte3) >> 2) & 31)
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 3) << 3) | int64(uint8(byte4) >> 5))
		valuesOffset++
		values[valuesOffset] = int64( int64(byte4) & 31)
		valuesOffset++
	}
}
