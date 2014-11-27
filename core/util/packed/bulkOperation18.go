// This file has been automatically generated, DO NOT EDIT

		package packed

		// Efficient sequential read/write of packed integers.
type BulkOperationPacked18 struct {
	*BulkOperationPacked
}

func newBulkOperationPacked18() BulkOperation {
	return &BulkOperationPacked18{newBulkOperationPacked(18)}
}

func (op *BulkOperationPacked18) decodeLongToInt(blocks []int64, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 46)); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 28) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block0) >> 10) & 262143); valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block0 & 1023) << 8) | (int64(uint64(block1) >> 56))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 38) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 20) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block1) >> 2) & 262143); valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block1 & 3) << 16) | (int64(uint64(block2) >> 48))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 30) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block2) >> 12) & 262143); valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block2 & 4095) << 6) | (int64(uint64(block3) >> 58))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 40) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 22) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block3) >> 4) & 262143); valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block3 & 15) << 14) | (int64(uint64(block4) >> 50))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 32) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block4) >> 14) & 262143); valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block4 & 16383) << 4) | (int64(uint64(block5) >> 60))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 42) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 24) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block5) >> 6) & 262143); valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block5 & 63) << 12) | (int64(uint64(block6) >> 52))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 34) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block6) >> 16) & 262143); valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block6 & 65535) << 2) | (int64(uint64(block7) >> 62))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 44) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 26) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block7) >> 8) & 262143); valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int32(((block7 & 255) << 10) | (int64(uint64(block8) >> 54))); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 36) & 262143); valuesOffset++
		values[valuesOffset] = int32(int64(uint64(block8) >> 18) & 262143); valuesOffset++
		values[valuesOffset] = int32(block8 & 262143); valuesOffset++
	}
}

func (op *BulkOperationPacked18) decodeByteToInt(blocks []byte, values []int32, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte0) << 10) | (int64(byte1) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte2 & 63) << 12) | (int64(byte3) << 4) | int64(uint8(byte4) >> 4))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte4 & 15) << 14) | (int64(byte5) << 6) | int64(uint8(byte6) >> 2))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int32((int64(byte6 & 3) << 16) | (int64(byte7) << 8) | int64(byte8))
		valuesOffset++
	}
}
func (op *BulkOperationPacked18) decodeLongToLong(blocks []int64, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		block0 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = int64(uint64(block0) >> 46); valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 28) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block0) >> 10) & 262143; valuesOffset++
		block1 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block0 & 1023) << 8) | (int64(uint64(block1) >> 56)); valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 38) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 20) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block1) >> 2) & 262143; valuesOffset++
		block2 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block1 & 3) << 16) | (int64(uint64(block2) >> 48)); valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 30) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block2) >> 12) & 262143; valuesOffset++
		block3 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block2 & 4095) << 6) | (int64(uint64(block3) >> 58)); valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 40) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 22) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block3) >> 4) & 262143; valuesOffset++
		block4 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block3 & 15) << 14) | (int64(uint64(block4) >> 50)); valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 32) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block4) >> 14) & 262143; valuesOffset++
		block5 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block4 & 16383) << 4) | (int64(uint64(block5) >> 60)); valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 42) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 24) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block5) >> 6) & 262143; valuesOffset++
		block6 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block5 & 63) << 12) | (int64(uint64(block6) >> 52)); valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 34) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block6) >> 16) & 262143; valuesOffset++
		block7 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block6 & 65535) << 2) | (int64(uint64(block7) >> 62)); valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 44) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 26) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block7) >> 8) & 262143; valuesOffset++
		block8 := blocks[blocksOffset]; blocksOffset++
		values[valuesOffset] = ((block7 & 255) << 10) | (int64(uint64(block8) >> 54)); valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 36) & 262143; valuesOffset++
		values[valuesOffset] = int64(uint64(block8) >> 18) & 262143; valuesOffset++
		values[valuesOffset] = block8 & 262143; valuesOffset++
	}
}

func (op *BulkOperationPacked18) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i ++ {
		byte0 := blocks[blocksOffset]
		blocksOffset++
		byte1 := blocks[blocksOffset]
		blocksOffset++
		byte2 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte0) << 10) | (int64(byte1) << 2) | int64(uint8(byte2) >> 6))
		valuesOffset++
		byte3 := blocks[blocksOffset]
		blocksOffset++
		byte4 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte2 & 63) << 12) | (int64(byte3) << 4) | int64(uint8(byte4) >> 4))
		valuesOffset++
		byte5 := blocks[blocksOffset]
		blocksOffset++
		byte6 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte4 & 15) << 14) | (int64(byte5) << 6) | int64(uint8(byte6) >> 2))
		valuesOffset++
		byte7 := blocks[blocksOffset]
		blocksOffset++
		byte8 := blocks[blocksOffset]
		blocksOffset++
		values[valuesOffset] = int64((int64(byte6 & 3) << 16) | (int64(byte7) << 8) | int64(byte8))
		valuesOffset++
	}
}
