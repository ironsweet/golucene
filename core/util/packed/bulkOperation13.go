// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked13 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked13() BulkOperation {
	return &BulkOperationPacked13{newBulkOperationPacked(13)}
}

func (op *BulkOperationPacked13) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 51)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 38) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 25) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 12) & 8191); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 4095) << 1) | (int64(uint64(block1) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 50) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 37) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 24) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 11) & 8191); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 2047) << 2) | (int64(uint64(block2) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 49) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 36) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 23) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 10) & 8191); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 1023) << 3) | (int64(uint64(block3) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 48) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 35) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 22) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 9) & 8191); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 511) << 4) | (int64(uint64(block4) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 47) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 34) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 21) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 8) & 8191); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 255) << 5) | (int64(uint64(block5) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 46) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 33) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 20) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 7) & 8191); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 127) << 6) | (int64(uint64(block6) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 45) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 32) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 19) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 6) & 8191); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 63) << 7) | (int64(uint64(block7) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 44) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 31) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 18) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 5) & 8191); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 31) << 8) | (int64(uint64(block8) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 43) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 30) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 17) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 4) & 8191); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 15) << 9) | (int64(uint64(block9) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 42) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 29) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 16) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 3) & 8191); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 7) << 10) | (int64(uint64(block10) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 41) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 28) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 15) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 2) & 8191); valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block10 & 3) << 11) | (int64(uint64(block11) >> 53))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 40) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 27) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 14) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 1) & 8191); valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block11 & 1) << 12) | (int64(uint64(block12) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 39) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 26) & 8191); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 13) & 8191); valuesOffset++
		values[valuesOffset] = int32(block12 & 8191); valuesOffset++
	}
}

func (op *BulkOperationPacked13) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 5) | int64(uint8(byte1) >> 3))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 7) << 10) | (int64(byte2) << 2) | int64(uint8(byte3) >> 6))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 63) << 7) | int64(uint8(byte4) >> 1))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 1) << 12) | (int64(byte5) << 4) | int64(uint8(byte6) >> 4))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte6 & 15) << 9) | (int64(byte7) << 1) | int64(uint8(byte8) >> 7))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte8 & 127) << 6) | int64(uint8(byte9) >> 2))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte9 & 3) << 11) | (int64(byte10) << 3) | int64(uint8(byte11) >> 5))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte11 & 31) << 8) | int64(byte12))
		valuesOffset++
	}
}
func (op *BulkOperationPacked13) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 51); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 38) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 25) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 12) & 8191; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 4095) << 1) | (int64(uint64(block1) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 50) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 37) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 24) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 11) & 8191; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 2047) << 2) | (int64(uint64(block2) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 49) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 36) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 23) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 10) & 8191; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 1023) << 3) | (int64(uint64(block3) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 48) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 35) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 22) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 9) & 8191; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 511) << 4) | (int64(uint64(block4) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 47) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 34) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 21) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 8) & 8191; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 255) << 5) | (int64(uint64(block5) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 46) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 33) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 20) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 7) & 8191; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 127) << 6) | (int64(uint64(block6) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 45) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 32) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 19) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 6) & 8191; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 63) << 7) | (int64(uint64(block7) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 44) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 31) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 18) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 5) & 8191; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 31) << 8) | (int64(uint64(block8) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 43) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 30) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 17) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 4) & 8191; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 15) << 9) | (int64(uint64(block9) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 42) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 29) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 16) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 3) & 8191; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 7) << 10) | (int64(uint64(block10) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 41) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 28) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 15) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 2) & 8191; valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block10 & 3) << 11) | (int64(uint64(block11) >> 53)); valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 40) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 27) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 14) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 1) & 8191; valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block11 & 1) << 12) | (int64(uint64(block12) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 39) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 26) & 8191; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 13) & 8191; valuesOffset++
		values[valuesOffset] = block12 & 8191; valuesOffset++
	}
}

func (op *BulkOperationPacked13) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 5) | int64(uint8(byte1) >> 3))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 7) << 10) | (int64(byte2) << 2) | int64(uint8(byte3) >> 6))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 63) << 7) | int64(uint8(byte4) >> 1))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 1) << 12) | (int64(byte5) << 4) | int64(uint8(byte6) >> 4))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte6 & 15) << 9) | (int64(byte7) << 1) | int64(uint8(byte8) >> 7))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte8 & 127) << 6) | int64(uint8(byte9) >> 2))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte9 & 3) << 11) | (int64(byte10) << 3) | int64(uint8(byte11) >> 5))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte11 & 31) << 8) | int64(byte12))
		valuesOffset++
	}
}
