// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked19 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked19() BulkOperation {
	return &BulkOperationPacked19{newBulkOperationPacked(19)}
}

func (op *BulkOperationPacked19) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 45)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 26) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 7) & 524287); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 127) << 12) | (int64(uint64(block1) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 33) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 14) & 524287); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 16383) << 5) | (int64(uint64(block2) >> 59))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 40) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 21) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 2) & 524287); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 3) << 17) | (int64(uint64(block3) >> 47))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 28) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 9) & 524287); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 511) << 10) | (int64(uint64(block4) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 35) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 16) & 524287); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 65535) << 3) | (int64(uint64(block5) >> 61))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 42) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 23) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 4) & 524287); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 15) << 15) | (int64(uint64(block6) >> 49))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 30) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 11) & 524287); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 2047) << 8) | (int64(uint64(block7) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 37) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 18) & 524287); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 262143) << 1) | (int64(uint64(block8) >> 63))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 44) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 25) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 6) & 524287); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 63) << 13) | (int64(uint64(block9) >> 51))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 32) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 13) & 524287); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 8191) << 6) | (int64(uint64(block10) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 39) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 20) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 1) & 524287); valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block10 & 1) << 18) | (int64(uint64(block11) >> 46))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 27) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block11) >> 8) & 524287); valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block11 & 255) << 11) | (int64(uint64(block12) >> 53))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 34) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block12) >> 15) & 524287); valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block12 & 32767) << 4) | (int64(uint64(block13) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 41) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 22) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block13) >> 3) & 524287); valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block13 & 7) << 16) | (int64(uint64(block14) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 29) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block14) >> 10) & 524287); valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block14 & 1023) << 9) | (int64(uint64(block15) >> 55))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 36) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block15) >> 17) & 524287); valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block15 & 131071) << 2) | (int64(uint64(block16) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 43) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 24) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block16) >> 5) & 524287); valuesOffset++
		block17 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block16 & 31) << 14) | (int64(uint64(block17) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block17) >> 31) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block17) >> 12) & 524287); valuesOffset++
		block18 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block17 & 4095) << 7) | (int64(uint64(block18) >> 57))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block18) >> 38) & 524287); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block18) >> 19) & 524287); valuesOffset++
		values[valuesOffset] = int32(block18 & 524287); valuesOffset++
	}
}

func (op *BulkOperationPacked19) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 11) | (int64(byte1) << 3) | int64(uint8(byte2) >> 5))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 31) << 14) | (int64(byte3) << 6) | int64(uint8(byte4) >> 2))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 3) << 17) | (int64(byte5) << 9) | (int64(byte6) << 1) | int64(uint8(byte7) >> 7))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte7 & 127) << 12) | (int64(byte8) << 4) | int64(uint8(byte9) >> 4))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte9 & 15) << 15) | (int64(byte10) << 7) | int64(uint8(byte11) >> 1))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte11 & 1) << 18) | (int64(byte12) << 10) | (int64(byte13) << 2) | int64(uint8(byte14) >> 6))
		valuesOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte14 & 63) << 13) | (int64(byte15) << 5) | int64(uint8(byte16) >> 3))
		valuesOffset++
		byte17 := blocks[blocksOffset]
		blocksOffset++
		byte18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte16 & 7) << 16) | (int64(byte17) << 8) | int64(byte18))
		valuesOffset++
	}
}
func (op *BulkOperationPacked19) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 45); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 26) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 7) & 524287; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 127) << 12) | (int64(uint64(block1) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 33) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 14) & 524287; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 16383) << 5) | (int64(uint64(block2) >> 59)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 40) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 21) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 2) & 524287; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 3) << 17) | (int64(uint64(block3) >> 47)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 28) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 9) & 524287; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 511) << 10) | (int64(uint64(block4) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 35) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 16) & 524287; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 65535) << 3) | (int64(uint64(block5) >> 61)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 42) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 23) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 4) & 524287; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 15) << 15) | (int64(uint64(block6) >> 49)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 30) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 11) & 524287; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 2047) << 8) | (int64(uint64(block7) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 37) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 18) & 524287; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 262143) << 1) | (int64(uint64(block8) >> 63)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 44) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 25) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 6) & 524287; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 63) << 13) | (int64(uint64(block9) >> 51)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 32) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 13) & 524287; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 8191) << 6) | (int64(uint64(block10) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 39) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 20) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 1) & 524287; valuesOffset++
		block11 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block10 & 1) << 18) | (int64(uint64(block11) >> 46)); valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 27) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block11) >> 8) & 524287; valuesOffset++
		block12 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block11 & 255) << 11) | (int64(uint64(block12) >> 53)); valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 34) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block12) >> 15) & 524287; valuesOffset++
		block13 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block12 & 32767) << 4) | (int64(uint64(block13) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 41) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 22) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block13) >> 3) & 524287; valuesOffset++
		block14 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block13 & 7) << 16) | (int64(uint64(block14) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 29) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block14) >> 10) & 524287; valuesOffset++
		block15 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block14 & 1023) << 9) | (int64(uint64(block15) >> 55)); valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 36) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block15) >> 17) & 524287; valuesOffset++
		block16 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block15 & 131071) << 2) | (int64(uint64(block16) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 43) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 24) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block16) >> 5) & 524287; valuesOffset++
		block17 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block16 & 31) << 14) | (int64(uint64(block17) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block17) >> 31) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block17) >> 12) & 524287; valuesOffset++
		block18 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block17 & 4095) << 7) | (int64(uint64(block18) >> 57)); valuesOffset++
		values[valuesOffset] = int64(uint64(block18) >> 38) & 524287; valuesOffset++
		values[valuesOffset] = int64(uint64(block18) >> 19) & 524287; valuesOffset++
		values[valuesOffset] = block18 & 524287; valuesOffset++
	}
}

func (op *BulkOperationPacked19) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 11) | (int64(byte1) << 3) | int64(uint8(byte2) >> 5))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 31) << 14) | (int64(byte3) << 6) | int64(uint8(byte4) >> 2))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 3) << 17) | (int64(byte5) << 9) | (int64(byte6) << 1) | int64(uint8(byte7) >> 7))
		valuesOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte7 & 127) << 12) | (int64(byte8) << 4) | int64(uint8(byte9) >> 4))
		valuesOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		byte11 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte9 & 15) << 15) | (int64(byte10) << 7) | int64(uint8(byte11) >> 1))
		valuesOffset++
		byte12 := blocks[blocksOffset]
		blocksOffset++
		byte13 := blocks[blocksOffset]
		blocksOffset++
		byte14 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte11 & 1) << 18) | (int64(byte12) << 10) | (int64(byte13) << 2) | int64(uint8(byte14) >> 6))
		valuesOffset++
		byte15 := blocks[blocksOffset]
		blocksOffset++
		byte16 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte14 & 63) << 13) | (int64(byte15) << 5) | int64(uint8(byte16) >> 3))
		valuesOffset++
		byte17 := blocks[blocksOffset]
		blocksOffset++
		byte18 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte16 & 7) << 16) | (int64(byte17) << 8) | int64(byte18))
		valuesOffset++
	}
}
