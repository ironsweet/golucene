// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked23 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked23() BulkOperation {
	return &BulkOperationPacked23{newBulkOperationPacked(23)}
}

func (op *BulkOperationPacked23) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 41)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 18) & 8388607); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 262143) << 5) | (int64(uint64(block1) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 36) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 13) & 8388607); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 8191) << 10) | (int64(uint64(block2) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 31) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 8) & 8388607); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 255) << 15) | (int64(uint64(block3) >> 49))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 26) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 3) & 8388607); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 7) << 20) | (int64(uint64(block4) >> 44))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 21) & 8388607); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 2097151) << 2) | (int64(uint64(block5) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 39) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 16) & 8388607); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 65535) << 7) | (int64(uint64(block6) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 34) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 11) & 8388607); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 2047) << 12) | (int64(uint64(block7) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 29) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 6) & 8388607); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 63) << 17) | (int64(uint64(block8) >> 47))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 24) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 1) & 8388607); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 1) << 22) | (int64(uint64(block9) >> 42))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 19) & 8388607); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 524287) << 4) | (int64(uint64(block10) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 37) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 14) & 8388607); valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block10 & 16383) << 9) | (int64(uint64(block11) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 32) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 9) & 8388607); valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block11 & 511) << 14) | (int64(uint64(block12) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 27) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 4) & 8388607); valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block12 & 15) << 19) | (int64(uint64(block13) >> 45))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 22) & 8388607); valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block13 & 4194303) << 1) | (int64(uint64(block14) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 40) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 17) & 8388607); valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block14 & 131071) << 6) | (int64(uint64(block15) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 35) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 12) & 8388607); valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block15 & 4095) << 11) | (int64(uint64(block16) >> 53))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 30) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 7) & 8388607); valuesOffset++
		block17 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block16 & 127) << 16) | (int64(uint64(block17) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block17) >> 25) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block17) >> 2) & 8388607); valuesOffset++
		block18 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block17 & 3) << 21) | (int64(uint64(block18) >> 43))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block18) >> 20) & 8388607); valuesOffset++
		block19 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block18 & 1048575) << 3) | (int64(uint64(block19) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block19) >> 38) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block19) >> 15) & 8388607); valuesOffset++
		block20 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block19 & 32767) << 8) | (int64(uint64(block20) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block20) >> 33) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block20) >> 10) & 8388607); valuesOffset++
		block21 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block20 & 1023) << 13) | (int64(uint64(block21) >> 51))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block21) >> 28) & 8388607); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block21) >> 5) & 8388607); valuesOffset++
		block22 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block21 & 31) << 18) | (int64(uint64(block22) >> 46))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block22) >> 23) & 8388607); valuesOffset++
		values[valuesOffset] = int32(block22 & 8388607); valuesOffset++
	}
}

func (op *BulkOperationPacked23) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 15) | (int64(byte1) << 7) | int64(uint8(byte2) >> 1))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 1) << 22) | (int64(byte3) << 14) | (int64(byte4) << 6) | int64(uint8(byte5) >> 2))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 3) << 21) | (int64(byte6) << 13) | (int64(byte7) << 5) | int64(uint8(byte8) >> 3))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte8 & 7) << 20) | (int64(byte9) << 12) | (int64(byte10) << 4) | int64(uint8(byte11) >> 4))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte11 & 15) << 19) | (int64(byte12) << 11) | (int64(byte13) << 3) | int64(uint8(byte14) >> 5))
		valuesOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		byte17 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte14 & 31) << 18) | (int64(byte15) << 10) | (int64(byte16) << 2) | int64(uint8(byte17) >> 6))
		valuesOffset++
		byte18 := blocks[blocksOffset]
		blocksOffset++
		byte19 := blocks[blocksOffset]
		blocksOffset++
		byte20 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte17 & 63) << 17) | (int64(byte18) << 9) | (int64(byte19) << 1) | int64(uint8(byte20) >> 7))
		valuesOffset++
		byte21 := blocks[blocksOffset]
		blocksOffset++
		byte22 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte20 & 127) << 16) | (int64(byte21) << 8) | int64(byte22))
		valuesOffset++
	}
}
func (op *BulkOperationPacked23) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 41); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 18) & 8388607; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 262143) << 5) | (int64(uint64(block1) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 36) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 13) & 8388607; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 8191) << 10) | (int64(uint64(block2) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 31) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 8) & 8388607; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 255) << 15) | (int64(uint64(block3) >> 49)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 26) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 3) & 8388607; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 7) << 20) | (int64(uint64(block4) >> 44)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 21) & 8388607; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 2097151) << 2) | (int64(uint64(block5) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 39) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 16) & 8388607; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 65535) << 7) | (int64(uint64(block6) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 34) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 11) & 8388607; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 2047) << 12) | (int64(uint64(block7) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 29) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 6) & 8388607; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 63) << 17) | (int64(uint64(block8) >> 47)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 24) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 1) & 8388607; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 1) << 22) | (int64(uint64(block9) >> 42)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 19) & 8388607; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 524287) << 4) | (int64(uint64(block10) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 37) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 14) & 8388607; valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block10 & 16383) << 9) | (int64(uint64(block11) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 32) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 9) & 8388607; valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block11 & 511) << 14) | (int64(uint64(block12) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 27) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 4) & 8388607; valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block12 & 15) << 19) | (int64(uint64(block13) >> 45)); valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 22) & 8388607; valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block13 & 4194303) << 1) | (int64(uint64(block14) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 40) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 17) & 8388607; valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block14 & 131071) << 6) | (int64(uint64(block15) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 35) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 12) & 8388607; valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block15 & 4095) << 11) | (int64(uint64(block16) >> 53)); valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 30) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 7) & 8388607; valuesOffset++
		block17 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block16 & 127) << 16) | (int64(uint64(block17) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block17) >> 25) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block17) >> 2) & 8388607; valuesOffset++
		block18 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block17 & 3) << 21) | (int64(uint64(block18) >> 43)); valuesOffset++
		values[valuesOffset] = int64(uint64(block18) >> 20) & 8388607; valuesOffset++
		block19 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block18 & 1048575) << 3) | (int64(uint64(block19) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block19) >> 38) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block19) >> 15) & 8388607; valuesOffset++
		block20 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block19 & 32767) << 8) | (int64(uint64(block20) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block20) >> 33) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block20) >> 10) & 8388607; valuesOffset++
		block21 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block20 & 1023) << 13) | (int64(uint64(block21) >> 51)); valuesOffset++
		values[valuesOffset] = int64(uint64(block21) >> 28) & 8388607; valuesOffset++
		values[valuesOffset] = int64(uint64(block21) >> 5) & 8388607; valuesOffset++
		block22 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block21 & 31) << 18) | (int64(uint64(block22) >> 46)); valuesOffset++
		values[valuesOffset] = int64(uint64(block22) >> 23) & 8388607; valuesOffset++
		values[valuesOffset] = block22 & 8388607; valuesOffset++
	}
}

func (op *BulkOperationPacked23) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 15) | (int64(byte1) << 7) | int64(uint8(byte2) >> 1))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 1) << 22) | (int64(byte3) << 14) | (int64(byte4) << 6) | int64(uint8(byte5) >> 2))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 3) << 21) | (int64(byte6) << 13) | (int64(byte7) << 5) | int64(uint8(byte8) >> 3))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte8 & 7) << 20) | (int64(byte9) << 12) | (int64(byte10) << 4) | int64(uint8(byte11) >> 4))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte11 & 15) << 19) | (int64(byte12) << 11) | (int64(byte13) << 3) | int64(uint8(byte14) >> 5))
		valuesOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		byte17 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte14 & 31) << 18) | (int64(byte15) << 10) | (int64(byte16) << 2) | int64(uint8(byte17) >> 6))
		valuesOffset++
		byte18 := blocks[blocksOffset]
		blocksOffset++
		byte19 := blocks[blocksOffset]
		blocksOffset++
		byte20 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte17 & 63) << 17) | (int64(byte18) << 9) | (int64(byte19) << 1) | int64(uint8(byte20) >> 7))
		valuesOffset++
		byte21 := blocks[blocksOffset]
		blocksOffset++
		byte22 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte20 & 127) << 16) | (int64(byte21) << 8) | int64(byte22))
		valuesOffset++
	}
}
