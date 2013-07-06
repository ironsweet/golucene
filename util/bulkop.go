package util

import (
	"fmt"
)

type BulkOperation struct {
	// PackedIntsEncoder
	PackedIntsDecoder
}

type BulkOperationPacked struct {
	*BulkOperation
	bitsPerValue   uint32
	longBlockCount uint32
	longValueCount uint32
	byteBlockCount uint32
	byteValueCount uint32
	mask           int64
	intMask        int
}

func newBulkOperationPacked(bitsPerValue uint32) *BulkOperation {
	self := BulkOperationPacked{}
	self.bitsPerValue = bitsPerValue
	// assert bitsPerValue > 0 && bitsPerValue <= 64
	blocks := uint32(bitsPerValue)
	for (blocks & 1) == 0 {
		blocks = (blocks >> 1)
	}
	self.longBlockCount = blocks
	self.longValueCount = 64 * self.longBlockCount / bitsPerValue
	byteBlockCount := 8 * self.longBlockCount
	byteValueCount := self.longValueCount
	for (byteBlockCount&1) == 0 && (byteValueCount&1) == 0 {
		byteBlockCount = (byteBlockCount >> 1)
		byteValueCount = (byteValueCount >> 1)
	}
	self.byteBlockCount = byteBlockCount
	self.byteValueCount = byteValueCount
	if bitsPerValue == 64 {
		self.mask = ^int64(0)
	} else {
		self.mask = (int64(1) << bitsPerValue) - 1
	}
	self.intMask = int(self.mask)
	// assert self.longValueCount * bitsPerValue == 64 * self.longBlockCount
	return self.BulkOperation
}

func newBulkOperationPacked1() *BulkOperation {
	ans := newBulkOperationPacked(1)
	return ans
}

func newBulkOperationPacked2() *BulkOperation {
	ans := newBulkOperationPacked(2)
	return ans
}

func newBulkOperationPacked3() *BulkOperation {
	ans := newBulkOperationPacked(3)
	return ans
}

func newBulkOperationPacked4() *BulkOperation {
	ans := newBulkOperationPacked(4)
	return ans
}

func newBulkOperationPacked5() *BulkOperation {
	ans := newBulkOperationPacked(4)
	return ans
}

func newBulkOperationPacked6() *BulkOperation {
	ans := newBulkOperationPacked(6)
	return ans
}

func newBulkOperationPacked7() *BulkOperation {
	ans := newBulkOperationPacked(7)
	return ans
}

func newBulkOperationPacked8() *BulkOperation {
	ans := newBulkOperationPacked(8)
	return ans
}

func newBulkOperationPacked9() *BulkOperation {
	ans := newBulkOperationPacked(9)
	return ans
}

func newBulkOperationPacked10() *BulkOperation {
	ans := newBulkOperationPacked(10)
	return ans
}

func newBulkOperationPacked11() *BulkOperation {
	ans := newBulkOperationPacked(11)
	return ans
}

func newBulkOperationPacked12() *BulkOperation {
	ans := newBulkOperationPacked(12)
	return ans
}

func newBulkOperationPacked13() *BulkOperation {
	ans := newBulkOperationPacked(13)
	return ans
}

func newBulkOperationPacked14() *BulkOperation {
	ans := newBulkOperationPacked(14)
	return ans
}

func newBulkOperationPacked15() *BulkOperation {
	ans := newBulkOperationPacked(15)
	return ans
}

func newBulkOperationPacked16() *BulkOperation {
	ans := newBulkOperationPacked(16)
	return ans
}

func newBulkOperationPacked17() *BulkOperation {
	ans := newBulkOperationPacked(17)
	return ans
}

func newBulkOperationPacked18() *BulkOperation {
	ans := newBulkOperationPacked(18)
	return ans
}

func newBulkOperationPacked19() *BulkOperation {
	ans := newBulkOperationPacked(19)
	return ans
}

func newBulkOperationPacked20() *BulkOperation {
	ans := newBulkOperationPacked(20)
	return ans
}

func newBulkOperationPacked21() *BulkOperation {
	ans := newBulkOperationPacked(21)
	return ans
}

func newBulkOperationPacked22() *BulkOperation {
	ans := newBulkOperationPacked(22)
	return ans
}

func newBulkOperationPacked23() *BulkOperation {
	ans := newBulkOperationPacked(23)
	return ans
}

func newBulkOperationPacked24() *BulkOperation {
	ans := newBulkOperationPacked(24)
	return ans
}

type BulkOperationPackedSingleBlock struct {
	*BulkOperation
	bitsPerValue uint32
	valueCount   uint32
	mask         int64
}

const BLOCK_COUNT = 1

func newBulkOperationPackedSingleBlock(bitsPerValue uint32) *BulkOperation {
	self := BulkOperationPackedSingleBlock{
		bitsPerValue: bitsPerValue,
		valueCount:   64 / bitsPerValue,
		mask:         (int64(1) << bitsPerValue) - 1}
	return self.BulkOperation
}

var (
	packedBulkOps = []*BulkOperation{
		newBulkOperationPacked1(),
		newBulkOperationPacked2(),
		newBulkOperationPacked3(),
		newBulkOperationPacked4(),
		newBulkOperationPacked5(),
		newBulkOperationPacked6(),
		newBulkOperationPacked7(),
		newBulkOperationPacked8(),
		newBulkOperationPacked9(),
		newBulkOperationPacked10(),
		newBulkOperationPacked11(),
		newBulkOperationPacked12(),
		newBulkOperationPacked13(),
		newBulkOperationPacked14(),
		newBulkOperationPacked15(),
		newBulkOperationPacked16(),
		newBulkOperationPacked17(),
		newBulkOperationPacked18(),
		newBulkOperationPacked19(),
		newBulkOperationPacked20(),
		newBulkOperationPacked21(),
		newBulkOperationPacked22(),
		newBulkOperationPacked23(),
		newBulkOperationPacked24(),
		newBulkOperationPacked(25),
		newBulkOperationPacked(26),
		newBulkOperationPacked(27),
		newBulkOperationPacked(28),
		newBulkOperationPacked(29),
		newBulkOperationPacked(30),
		newBulkOperationPacked(31),
		newBulkOperationPacked(32),
		newBulkOperationPacked(33),
		newBulkOperationPacked(34),
		newBulkOperationPacked(35),
		newBulkOperationPacked(36),
		newBulkOperationPacked(37),
		newBulkOperationPacked(38),
		newBulkOperationPacked(39),
		newBulkOperationPacked(40),
		newBulkOperationPacked(41),
		newBulkOperationPacked(42),
		newBulkOperationPacked(43),
		newBulkOperationPacked(44),
		newBulkOperationPacked(45),
		newBulkOperationPacked(46),
		newBulkOperationPacked(47),
		newBulkOperationPacked(48),
		newBulkOperationPacked(49),
		newBulkOperationPacked(50),
		newBulkOperationPacked(51),
		newBulkOperationPacked(52),
		newBulkOperationPacked(53),
		newBulkOperationPacked(54),
		newBulkOperationPacked(55),
		newBulkOperationPacked(56),
		newBulkOperationPacked(57),
		newBulkOperationPacked(58),
		newBulkOperationPacked(59),
		newBulkOperationPacked(60),
		newBulkOperationPacked(61),
		newBulkOperationPacked(62),
		newBulkOperationPacked(63),
		newBulkOperationPacked(64),
	}

	packedSingleBlockBulkOps = []*BulkOperation{
		newBulkOperationPackedSingleBlock(1),
		newBulkOperationPackedSingleBlock(2),
		newBulkOperationPackedSingleBlock(3),
		newBulkOperationPackedSingleBlock(4),
		newBulkOperationPackedSingleBlock(5),
		newBulkOperationPackedSingleBlock(6),
		newBulkOperationPackedSingleBlock(7),
		newBulkOperationPackedSingleBlock(8),
		newBulkOperationPackedSingleBlock(9),
		newBulkOperationPackedSingleBlock(10),
		nil,
		newBulkOperationPackedSingleBlock(12),
		nil,
		nil,
		nil,
		newBulkOperationPackedSingleBlock(16),
		nil,
		nil,
		nil,
		nil,
		newBulkOperationPackedSingleBlock(21),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		newBulkOperationPackedSingleBlock(32),
	}
)

func newBulkOperation(format PackedFormat, bitsPerValue uint32) *BulkOperation {
	switch int(format) {
	case PACKED:
		// assert packedBulkOps[bitsPerValue - 1] != nil
		return packedBulkOps[bitsPerValue-1]
	case PACKED_SINGLE_BLOCK:
		// assert packedSingleBlockBulkOps[bitsPerValue - 1] != nil
		return packedSingleBlockBulkOps[bitsPerValue-1]
	}
	panic(fmt.Sprintf("invalid packed format: %v", format))
}
