// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked11 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked11() BulkOperation {
	return &BulkOperationPacked11{newBulkOperationPacked(11)}
}

func (op *BulkOperationPacked11) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 53)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 42) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 31) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 20) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 9) & 2047); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 511) << 2) | (int64(uint64(block1) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 51) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 40) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 29) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 18) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 7) & 2047); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 127) << 4) | (int64(uint64(block2) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 49) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 38) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 27) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 16) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 5) & 2047); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 31) << 6) | (int64(uint64(block3) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 47) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 36) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 25) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 14) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 3) & 2047); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 7) << 8) | (int64(uint64(block4) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 45) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 34) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 23) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 12) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 1) & 2047); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 1) << 10) | (int64(uint64(block5) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 43) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 32) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 21) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 10) & 2047); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 1023) << 1) | (int64(uint64(block6) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 52) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 41) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 30) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 19) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 8) & 2047); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 255) << 3) | (int64(uint64(block7) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 50) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 39) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 28) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 17) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 6) & 2047); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 63) << 5) | (int64(uint64(block8) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 48) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 37) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 26) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 15) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 4) & 2047); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 15) << 7) | (int64(uint64(block9) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 46) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 35) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 24) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 13) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 2) & 2047); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 3) << 9) | (int64(uint64(block10) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 44) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 33) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 22) & 2047); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 11) & 2047); valuesOffset++
		values[valuesOffset] = int32(block10 & 2047); valuesOffset++
	}
}

func (op *BulkOperationPacked11) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 3) | int64(uint8(byte1) >> 5))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 31) << 6) | int64(uint8(byte2) >> 2))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 3) << 9) | (int64(byte3) << 1) | int64(uint8(byte4) >> 7))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 127) << 4) | int64(uint8(byte5) >> 4))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 15) << 7) | int64(uint8(byte6) >> 1))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte6 & 1) << 10) | (int64(byte7) << 2) | int64(uint8(byte8) >> 6))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte8 & 63) << 5) | int64(uint8(byte9) >> 3))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte9 & 7) << 8) | int64(byte10))
		valuesOffset++
	}
}
func (op *BulkOperationPacked11) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 53); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 42) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 31) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 20) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 9) & 2047; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 511) << 2) | (int64(uint64(block1) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 51) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 40) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 29) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 18) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 7) & 2047; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 127) << 4) | (int64(uint64(block2) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 49) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 38) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 27) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 16) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 5) & 2047; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 31) << 6) | (int64(uint64(block3) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 47) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 36) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 25) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 14) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 3) & 2047; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 7) << 8) | (int64(uint64(block4) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 45) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 34) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 23) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 12) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 1) & 2047; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 1) << 10) | (int64(uint64(block5) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 43) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 32) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 21) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 10) & 2047; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 1023) << 1) | (int64(uint64(block6) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 52) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 41) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 30) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 19) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 8) & 2047; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 255) << 3) | (int64(uint64(block7) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 50) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 39) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 28) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 17) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 6) & 2047; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 63) << 5) | (int64(uint64(block8) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 48) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 37) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 26) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 15) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 4) & 2047; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 15) << 7) | (int64(uint64(block9) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 46) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 35) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 24) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 13) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 2) & 2047; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 3) << 9) | (int64(uint64(block10) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 44) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 33) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 22) & 2047; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 11) & 2047; valuesOffset++
		values[valuesOffset] = block10 & 2047; valuesOffset++
	}
}

func (op *BulkOperationPacked11) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 3) | int64(uint8(byte1) >> 5))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 31) << 6) | int64(uint8(byte2) >> 2))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 3) << 9) | (int64(byte3) << 1) | int64(uint8(byte4) >> 7))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 127) << 4) | int64(uint8(byte5) >> 4))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 15) << 7) | int64(uint8(byte6) >> 1))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte6 & 1) << 10) | (int64(byte7) << 2) | int64(uint8(byte8) >> 6))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte8 & 63) << 5) | int64(uint8(byte9) >> 3))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte9 & 7) << 8) | int64(byte10))
		valuesOffset++
	}
}
