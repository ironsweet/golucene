// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked17 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked17() BulkOperation {
	return &BulkOperationPacked17{newBulkOperationPacked(17)}
}

func (op *BulkOperationPacked17) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 47)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 30) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 13) & 131071); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 8191) << 4) | (int64(uint64(block1) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 43) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 26) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 9) & 131071); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 511) << 8) | (int64(uint64(block2) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 39) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 22) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 5) & 131071); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 31) << 12) | (int64(uint64(block3) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 35) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 18) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 1) & 131071); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 1) << 16) | (int64(uint64(block4) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 31) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 14) & 131071); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 16383) << 3) | (int64(uint64(block5) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 44) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 27) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 10) & 131071); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 1023) << 7) | (int64(uint64(block6) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 40) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 23) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 6) & 131071); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 63) << 11) | (int64(uint64(block7) >> 53))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 36) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 19) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 2) & 131071); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 3) << 15) | (int64(uint64(block8) >> 49))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 32) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 15) & 131071); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 32767) << 2) | (int64(uint64(block9) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 45) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 28) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 11) & 131071); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 2047) << 6) | (int64(uint64(block10) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 41) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 24) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 7) & 131071); valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block10 & 127) << 10) | (int64(uint64(block11) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 37) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 20) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 3) & 131071); valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block11 & 7) << 14) | (int64(uint64(block12) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 33) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 16) & 131071); valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block12 & 65535) << 1) | (int64(uint64(block13) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 46) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 29) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 12) & 131071); valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block13 & 4095) << 5) | (int64(uint64(block14) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 42) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 25) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 8) & 131071); valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block14 & 255) << 9) | (int64(uint64(block15) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 38) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 21) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 4) & 131071); valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block15 & 15) << 13) | (int64(uint64(block16) >> 51))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 34) & 131071); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 17) & 131071); valuesOffset++
		values[valuesOffset] = int32(block16 & 131071); valuesOffset++
	}
}

func (op *BulkOperationPacked17) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 9) | (int64(byte1) << 1) | int64(uint8(byte2) >> 7))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 127) << 10) | (int64(byte3) << 2) | int64(uint8(byte4) >> 6))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 63) << 11) | (int64(byte5) << 3) | int64(uint8(byte6) >> 5))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte6 & 31) << 12) | (int64(byte7) << 4) | int64(uint8(byte8) >> 4))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte8 & 15) << 13) | (int64(byte9) << 5) | int64(uint8(byte10) >> 3))
		valuesOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte10 & 7) << 14) | (int64(byte11) << 6) | int64(uint8(byte12) >> 2))
		valuesOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte12 & 3) << 15) | (int64(byte13) << 7) | int64(uint8(byte14) >> 1))
		valuesOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte14 & 1) << 16) | (int64(byte15) << 8) | int64(byte16))
		valuesOffset++
	}
}
func (op *BulkOperationPacked17) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 47); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 30) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 13) & 131071; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 8191) << 4) | (int64(uint64(block1) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 43) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 26) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 9) & 131071; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 511) << 8) | (int64(uint64(block2) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 39) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 22) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 5) & 131071; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 31) << 12) | (int64(uint64(block3) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 35) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 18) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 1) & 131071; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 1) << 16) | (int64(uint64(block4) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 31) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 14) & 131071; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 16383) << 3) | (int64(uint64(block5) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 44) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 27) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 10) & 131071; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 1023) << 7) | (int64(uint64(block6) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 40) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 23) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 6) & 131071; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 63) << 11) | (int64(uint64(block7) >> 53)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 36) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 19) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 2) & 131071; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 3) << 15) | (int64(uint64(block8) >> 49)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 32) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 15) & 131071; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 32767) << 2) | (int64(uint64(block9) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 45) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 28) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 11) & 131071; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 2047) << 6) | (int64(uint64(block10) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 41) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 24) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 7) & 131071; valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block10 & 127) << 10) | (int64(uint64(block11) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 37) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 20) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 3) & 131071; valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block11 & 7) << 14) | (int64(uint64(block12) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 33) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 16) & 131071; valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block12 & 65535) << 1) | (int64(uint64(block13) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 46) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 29) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 12) & 131071; valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block13 & 4095) << 5) | (int64(uint64(block14) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 42) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 25) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 8) & 131071; valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block14 & 255) << 9) | (int64(uint64(block15) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 38) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 21) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 4) & 131071; valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block15 & 15) << 13) | (int64(uint64(block16) >> 51)); valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 34) & 131071; valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 17) & 131071; valuesOffset++
		values[valuesOffset] = block16 & 131071; valuesOffset++
	}
}

func (op *BulkOperationPacked17) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 9) | (int64(byte1) << 1) | int64(uint8(byte2) >> 7))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 127) << 10) | (int64(byte3) << 2) | int64(uint8(byte4) >> 6))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 63) << 11) | (int64(byte5) << 3) | int64(uint8(byte6) >> 5))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte6 & 31) << 12) | (int64(byte7) << 4) | int64(uint8(byte8) >> 4))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte8 & 15) << 13) | (int64(byte9) << 5) | int64(uint8(byte10) >> 3))
		valuesOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte10 & 7) << 14) | (int64(byte11) << 6) | int64(uint8(byte12) >> 2))
		valuesOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte12 & 3) << 15) | (int64(byte13) << 7) | int64(uint8(byte14) >> 1))
		valuesOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte14 & 1) << 16) | (int64(byte15) << 8) | int64(byte16))
		valuesOffset++
	}
}
