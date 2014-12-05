// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked9 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked9() BulkOperation {
	return &BulkOperationPacked9{newBulkOperationPacked(9)}
}

func (op *BulkOperationPacked9) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 55)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 46) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 37) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 28) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 19) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 10) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 1) & 511); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 1) << 8) | (int64(uint64(block1) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 47) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 29) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 20) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 11) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 511); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 7) | (int64(uint64(block2) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 48) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 39) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 30) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 21) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 3) & 511); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 7) << 6) | (int64(uint64(block3) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 49) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 40) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 31) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 22) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 13) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 4) & 511); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 15) << 5) | (int64(uint64(block4) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 50) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 41) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 32) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 23) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 14) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 5) & 511); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 31) << 4) | (int64(uint64(block5) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 51) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 42) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 33) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 24) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 15) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 6) & 511); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 63) << 3) | (int64(uint64(block6) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 52) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 43) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 34) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 25) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 16) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 7) & 511); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 127) << 2) | (int64(uint64(block7) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 53) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 44) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 35) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 26) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 17) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 8) & 511); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 255) << 1) | (int64(uint64(block8) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 54) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 45) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 36) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 27) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 18) & 511); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 9) & 511); valuesOffset++
		values[valuesOffset] = int32(block8 & 511); valuesOffset++
	}
}

func (op *BulkOperationPacked9) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 1) | int64(uint8(byte1) >> 7))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 127) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 63) << 3) | int64(uint8(byte3) >> 5))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 31) << 4) | int64(uint8(byte4) >> 4))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 15) << 5) | int64(uint8(byte5) >> 3))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 7) << 6) | int64(uint8(byte6) >> 2))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte6 & 3) << 7) | int64(uint8(byte7) >> 1))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte7 & 1) << 8) | int64(byte8))
		valuesOffset++
	}
}
func (op *BulkOperationPacked9) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 55); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 46) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 37) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 28) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 19) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 10) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 1) & 511; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 8) | (int64(uint64(block1) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 47) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 29) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 20) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 11) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 511; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 7) | (int64(uint64(block2) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 48) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 39) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 30) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 21) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 3) & 511; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 7) << 6) | (int64(uint64(block3) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 49) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 40) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 31) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 22) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 13) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 4) & 511; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 5) | (int64(uint64(block4) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 50) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 41) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 32) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 23) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 14) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 5) & 511; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 4) | (int64(uint64(block5) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 51) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 42) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 33) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 24) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 15) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 6) & 511; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 3) | (int64(uint64(block6) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 52) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 43) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 34) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 25) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 16) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 7) & 511; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 127) << 2) | (int64(uint64(block7) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 53) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 44) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 35) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 26) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 17) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 8) & 511; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 255) << 1) | (int64(uint64(block8) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 54) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 45) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 36) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 27) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 18) & 511; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 9) & 511; valuesOffset++
		values[valuesOffset] = block8 & 511; valuesOffset++
	}
}

func (op *BulkOperationPacked9) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 1) | int64(uint8(byte1) >> 7))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 127) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 63) << 3) | int64(uint8(byte3) >> 5))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 31) << 4) | int64(uint8(byte4) >> 4))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 15) << 5) | int64(uint8(byte5) >> 3))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 7) << 6) | int64(uint8(byte6) >> 2))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte6 & 3) << 7) | int64(uint8(byte7) >> 1))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte7 & 1) << 8) | int64(byte8))
		valuesOffset++
	}
}
