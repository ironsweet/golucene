// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked22 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked22() BulkOperation {
	return &BulkOperationPacked22{newBulkOperationPacked(22)}
}

func (op *BulkOperationPacked22) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 42)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 20) & 4194303); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 1048575) << 2) | (int64(uint64(block1) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 40) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 18) & 4194303); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 262143) << 4) | (int64(uint64(block2) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 38) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 16) & 4194303); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 65535) << 6) | (int64(uint64(block3) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 36) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 14) & 4194303); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 16383) << 8) | (int64(uint64(block4) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 34) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 12) & 4194303); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 4095) << 10) | (int64(uint64(block5) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 32) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 10) & 4194303); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 1023) << 12) | (int64(uint64(block6) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 30) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 8) & 4194303); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 255) << 14) | (int64(uint64(block7) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 28) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 6) & 4194303); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 63) << 16) | (int64(uint64(block8) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 26) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 4) & 4194303); valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block8 & 15) << 18) | (int64(uint64(block9) >> 46))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 24) & 4194303); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block9) >> 2) & 4194303); valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block9 & 3) << 20) | (int64(uint64(block10) >> 44))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block10) >> 22) & 4194303); valuesOffset++
		values[valuesOffset] = int32(block10 & 4194303); valuesOffset++
	}
}

func (op *BulkOperationPacked22) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 14) | (int64(byte1) << 6) | int64(uint8(byte2) >> 2))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 3) << 20) | (int64(byte3) << 12) | (int64(byte4) << 4) | int64(uint8(byte5) >> 4))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte5 & 15) << 18) | (int64(byte6) << 10) | (int64(byte7) << 2) | int64(uint8(byte8) >> 6))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte8 & 63) << 16) | (int64(byte9) << 8) | int64(byte10))
		valuesOffset++
	}
}
func (op *BulkOperationPacked22) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 42); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 20) & 4194303; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 1048575) << 2) | (int64(uint64(block1) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 40) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 18) & 4194303; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 262143) << 4) | (int64(uint64(block2) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 38) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 16) & 4194303; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 65535) << 6) | (int64(uint64(block3) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 36) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 14) & 4194303; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 16383) << 8) | (int64(uint64(block4) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 34) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 12) & 4194303; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 4095) << 10) | (int64(uint64(block5) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 32) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 10) & 4194303; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 1023) << 12) | (int64(uint64(block6) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 30) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 8) & 4194303; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 255) << 14) | (int64(uint64(block7) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 28) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 6) & 4194303; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 63) << 16) | (int64(uint64(block8) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 26) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 4) & 4194303; valuesOffset++
		block9 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block8 & 15) << 18) | (int64(uint64(block9) >> 46)); valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 24) & 4194303; valuesOffset++
		values[valuesOffset] = int64(uint64(block9) >> 2) & 4194303; valuesOffset++
		block10 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block9 & 3) << 20) | (int64(uint64(block10) >> 44)); valuesOffset++
		values[valuesOffset] = int64(uint64(block10) >> 22) & 4194303; valuesOffset++
		values[valuesOffset] = block10 & 4194303; valuesOffset++
	}
}

func (op *BulkOperationPacked22) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 14) | (int64(byte1) << 6) | int64(uint8(byte2) >> 2))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 3) << 20) | (int64(byte3) << 12) | (int64(byte4) << 4) | int64(uint8(byte5) >> 4))
		valuesOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte5 & 15) << 18) | (int64(byte6) << 10) | (int64(byte7) << 2) | int64(uint8(byte8) >> 6))
		valuesOffset++
		byte9 := blocks[blocksOffset]
		blocksOffset++
		byte10 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte8 & 63) << 16) | (int64(byte9) << 8) | int64(byte10))
		valuesOffset++
	}
}
