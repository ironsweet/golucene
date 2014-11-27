// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked21 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked21() BulkOperation {
	return &BulkOperationPacked21{newBulkOperationPacked(21)}
}

func (op *BulkOperationPacked21) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 43)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 22) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 1) & 2097151); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 1) << 20) | (int64(uint64(block1) >> 44))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 23) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 2097151); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 19) | (int64(uint64(block2) >> 45))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 24) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 3) & 2097151); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 7) << 18) | (int64(uint64(block3) >> 46))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 25) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 4) & 2097151); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 15) << 17) | (int64(uint64(block4) >> 47))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 26) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 5) & 2097151); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 31) << 16) | (int64(uint64(block5) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 27) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 6) & 2097151); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 63) << 15) | (int64(uint64(block6) >> 49))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 28) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 7) & 2097151); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 127) << 14) | (int64(uint64(block7) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 29) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 8) & 2097151); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 255) << 13) | (int64(uint64(block8) >> 51))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 30) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 9) & 2097151); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 511) << 12) | (int64(uint64(block9) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 31) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 10) & 2097151); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 1023) << 11) | (int64(uint64(block10) >> 53))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 32) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 11) & 2097151); valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block10 & 2047) << 10) | (int64(uint64(block11) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 33) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 12) & 2097151); valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block11 & 4095) << 9) | (int64(uint64(block12) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 34) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 13) & 2097151); valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block12 & 8191) << 8) | (int64(uint64(block13) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 35) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 14) & 2097151); valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block13 & 16383) << 7) | (int64(uint64(block14) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 36) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 15) & 2097151); valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block14 & 32767) << 6) | (int64(uint64(block15) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 37) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 16) & 2097151); valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block15 & 65535) << 5) | (int64(uint64(block16) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 38) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 17) & 2097151); valuesOffset++
		block17 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block16 & 131071) << 4) | (int64(uint64(block17) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block17) >> 39) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block17) >> 18) & 2097151); valuesOffset++
		block18 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block17 & 262143) << 3) | (int64(uint64(block18) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block18) >> 40) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block18) >> 19) & 2097151); valuesOffset++
		block19 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block18 & 524287) << 2) | (int64(uint64(block19) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block19) >> 41) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block19) >> 20) & 2097151); valuesOffset++
		block20 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block19 & 1048575) << 1) | (int64(uint64(block20) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block20) >> 42) & 2097151); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block20) >> 21) & 2097151); valuesOffset++
		values[valuesOffset] = int32(block20 & 2097151); valuesOffset++
	}
}

func (op *BulkOperationPacked21) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 13) | (int64(byte1) << 5) | int64(uint8(byte2) >> 3))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 7) << 18) | (int64(byte3) << 10) | (int64(byte4) << 2) | int64(uint8(byte5) >> 6))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 63) << 15) | (int64(byte6) << 7) | int64(uint8(byte7) >> 1))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte7 & 1) << 20) | (int64(byte8) << 12) | (int64(byte9) << 4) | int64(uint8(byte10) >> 4))
		valuesOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte10 & 15) << 17) | (int64(byte11) << 9) | (int64(byte12) << 1) | int64(uint8(byte13) >> 7))
		valuesOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte13 & 127) << 14) | (int64(byte14) << 6) | int64(uint8(byte15) >> 2))
		valuesOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		byte17 := blocks[blocksOffset]
		blocksOffset++
		byte18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte15 & 3) << 19) | (int64(byte16) << 11) | (int64(byte17) << 3) | int64(uint8(byte18) >> 5))
		valuesOffset++
		byte19 := blocks[blocksOffset]
		blocksOffset++
		byte20 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte18 & 31) << 16) | (int64(byte19) << 8) | int64(byte20))
		valuesOffset++
	}
}
func (op *BulkOperationPacked21) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 43); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 22) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 1) & 2097151; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 1) << 20) | (int64(uint64(block1) >> 44)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 23) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 2097151; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 19) | (int64(uint64(block2) >> 45)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 24) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 3) & 2097151; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 7) << 18) | (int64(uint64(block3) >> 46)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 25) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 4) & 2097151; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 17) | (int64(uint64(block4) >> 47)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 26) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 5) & 2097151; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 31) << 16) | (int64(uint64(block5) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 27) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 6) & 2097151; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 15) | (int64(uint64(block6) >> 49)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 28) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 7) & 2097151; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 127) << 14) | (int64(uint64(block7) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 29) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 8) & 2097151; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 255) << 13) | (int64(uint64(block8) >> 51)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 30) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 9) & 2097151; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 511) << 12) | (int64(uint64(block9) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 31) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 10) & 2097151; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 1023) << 11) | (int64(uint64(block10) >> 53)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 32) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 11) & 2097151; valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block10 & 2047) << 10) | (int64(uint64(block11) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 33) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 12) & 2097151; valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block11 & 4095) << 9) | (int64(uint64(block12) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 34) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 13) & 2097151; valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block12 & 8191) << 8) | (int64(uint64(block13) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 35) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 14) & 2097151; valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block13 & 16383) << 7) | (int64(uint64(block14) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 36) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 15) & 2097151; valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block14 & 32767) << 6) | (int64(uint64(block15) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 37) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 16) & 2097151; valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block15 & 65535) << 5) | (int64(uint64(block16) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 38) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 17) & 2097151; valuesOffset++
		block17 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block16 & 131071) << 4) | (int64(uint64(block17) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block17) >> 39) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block17) >> 18) & 2097151; valuesOffset++
		block18 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block17 & 262143) << 3) | (int64(uint64(block18) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block18) >> 40) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block18) >> 19) & 2097151; valuesOffset++
		block19 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block18 & 524287) << 2) | (int64(uint64(block19) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block19) >> 41) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block19) >> 20) & 2097151; valuesOffset++
		block20 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block19 & 1048575) << 1) | (int64(uint64(block20) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block20) >> 42) & 2097151; valuesOffset++
		values[valuesOffset] = int64(uint64(block20) >> 21) & 2097151; valuesOffset++
		values[valuesOffset] = block20 & 2097151; valuesOffset++
	}
}

func (op *BulkOperationPacked21) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 13) | (int64(byte1) << 5) | int64(uint8(byte2) >> 3))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 7) << 18) | (int64(byte3) << 10) | (int64(byte4) << 2) | int64(uint8(byte5) >> 6))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 63) << 15) | (int64(byte6) << 7) | int64(uint8(byte7) >> 1))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte7 & 1) << 20) | (int64(byte8) << 12) | (int64(byte9) << 4) | int64(uint8(byte10) >> 4))
		valuesOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte10 & 15) << 17) | (int64(byte11) << 9) | (int64(byte12) << 1) | int64(uint8(byte13) >> 7))
		valuesOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte13 & 127) << 14) | (int64(byte14) << 6) | int64(uint8(byte15) >> 2))
		valuesOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		byte17 := blocks[blocksOffset]
		blocksOffset++
		byte18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte15 & 3) << 19) | (int64(byte16) << 11) | (int64(byte17) << 3) | int64(uint8(byte18) >> 5))
		valuesOffset++
		byte19 := blocks[blocksOffset]
		blocksOffset++
		byte20 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte18 & 31) << 16) | (int64(byte19) << 8) | int64(byte20))
		valuesOffset++
	}
}
