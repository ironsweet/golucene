// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked15 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked15() BulkOperation {
	return &BulkOperationPacked15{newBulkOperationPacked(15)}
}

func (op *BulkOperationPacked15) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 49)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 34) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 19) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 4) & 32767); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 15) << 11) | (int64(uint64(block1) >> 53))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 23) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 8) & 32767); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 255) << 7) | (int64(uint64(block2) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 42) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 27) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 32767); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 4095) << 3) | (int64(uint64(block3) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 46) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 31) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 16) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 1) & 32767); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 1) << 14) | (int64(uint64(block4) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 35) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 20) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 5) & 32767); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 31) << 10) | (int64(uint64(block5) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 39) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 24) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 9) & 32767); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 511) << 6) | (int64(uint64(block6) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 43) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 28) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 13) & 32767); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 8191) << 2) | (int64(uint64(block7) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 47) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 32) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 17) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 2) & 32767); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 3) << 13) | (int64(uint64(block8) >> 51))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 36) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 21) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 6) & 32767); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 63) << 9) | (int64(uint64(block9) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 40) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 25) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 10) & 32767); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 1023) << 5) | (int64(uint64(block10) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 44) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 29) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 14) & 32767); valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block10 & 16383) << 1) | (int64(uint64(block11) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 48) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 33) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 18) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 3) & 32767); valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block11 & 7) << 12) | (int64(uint64(block12) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 37) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 22) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 7) & 32767); valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block12 & 127) << 8) | (int64(uint64(block13) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 41) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 26) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 11) & 32767); valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block13 & 2047) << 4) | (int64(uint64(block14) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 45) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 30) & 32767); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 15) & 32767); valuesOffset++
		values[valuesOffset] = int32(block14 & 32767); valuesOffset++
	}
}

func (op *BulkOperationPacked15) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 7) | int64(uint8(byte1) >> 1))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte1 & 1) << 14) | (int64(byte2) << 6) | int64(uint8(byte3) >> 2))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte3 & 3) << 13) | (int64(byte4) << 5) | int64(uint8(byte5) >> 3))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 7) << 12) | (int64(byte6) << 4) | int64(uint8(byte7) >> 4))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte7 & 15) << 11) | (int64(byte8) << 3) | int64(uint8(byte9) >> 5))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte9 & 31) << 10) | (int64(byte10) << 2) | int64(uint8(byte11) >> 6))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte11 & 63) << 9) | (int64(byte12) << 1) | int64(uint8(byte13) >> 7))
		valuesOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte13 & 127) << 8) | int64(byte14))
		valuesOffset++
	}
}
func (op *BulkOperationPacked15) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 49); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 34) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 19) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 4) & 32767; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 15) << 11) | (int64(uint64(block1) >> 53)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 23) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 8) & 32767; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 255) << 7) | (int64(uint64(block2) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 42) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 27) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 32767; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 3) | (int64(uint64(block3) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 46) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 31) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 16) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 1) & 32767; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 1) << 14) | (int64(uint64(block4) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 35) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 20) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 5) & 32767; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 10) | (int64(uint64(block5) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 39) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 24) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 9) & 32767; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 511) << 6) | (int64(uint64(block6) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 43) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 28) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 13) & 32767; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 8191) << 2) | (int64(uint64(block7) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 47) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 32) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 17) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 2) & 32767; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 3) << 13) | (int64(uint64(block8) >> 51)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 36) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 21) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 6) & 32767; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 63) << 9) | (int64(uint64(block9) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 40) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 25) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 10) & 32767; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 1023) << 5) | (int64(uint64(block10) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 44) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 29) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 14) & 32767; valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block10 & 16383) << 1) | (int64(uint64(block11) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 48) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 33) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 18) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 3) & 32767; valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block11 & 7) << 12) | (int64(uint64(block12) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 37) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 22) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 7) & 32767; valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block12 & 127) << 8) | (int64(uint64(block13) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 41) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 26) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 11) & 32767; valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block13 & 2047) << 4) | (int64(uint64(block14) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 45) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 30) & 32767; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 15) & 32767; valuesOffset++
		values[valuesOffset] = block14 & 32767; valuesOffset++
	}
}

func (op *BulkOperationPacked15) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 7) | int64(uint8(byte1) >> 1))
		valuesOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte1 & 1) << 14) | (int64(byte2) << 6) | int64(uint8(byte3) >> 2))
		valuesOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte3 & 3) << 13) | (int64(byte4) << 5) | int64(uint8(byte5) >> 3))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 7) << 12) | (int64(byte6) << 4) | int64(uint8(byte7) >> 4))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte7 & 15) << 11) | (int64(byte8) << 3) | int64(uint8(byte9) >> 5))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte9 & 31) << 10) | (int64(byte10) << 2) | int64(uint8(byte11) >> 6))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte11 & 63) << 9) | (int64(byte12) << 1) | int64(uint8(byte13) >> 7))
		valuesOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte13 & 127) << 8) | int64(byte14))
		valuesOffset++
	}
}
