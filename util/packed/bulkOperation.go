package packed

import (
	"fmt"
	"log"
)

// util/packed/BulkOperation.java

// Efficient sequential read/write of packed integers.
type BulkOperation struct {
	// PackedIntsEncoder
	PackedIntsDecoder
}

func newBulkOperationPacked1() *BulkOperation {
	log.Print("Initializng BulkOperationPacked1...")
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
	// log.Printf("Initializing BulkOperation(%v,%v)", format, bitsPerValue)
	switch int(format) {
	case PACKED:
		assert2(packedBulkOps[bitsPerValue-1] != nil, fmt.Sprintf("bpv=%v", bitsPerValue))
		return packedBulkOps[bitsPerValue-1]
	case PACKED_SINGLE_BLOCK:
		assert2(packedSingleBlockBulkOps[bitsPerValue-1] != nil, fmt.Sprintf("bpv=%v", bitsPerValue))
		return packedSingleBlockBulkOps[bitsPerValue-1]
	}
	panic(fmt.Sprintf("invalid packed format: %v", format))
}

/*
For every number of bits per value, there is a minumum number of
blocks (b) / values (v) you need to write an order to reach the next block
boundary:
- 16 bits per value -> b=2, v=1
- 24 bits per value -> b=3, v=1
- 50 bits per value -> b=25, v=4
- 63 bits per value -> b=63, v=8
- ...

A bulk read consists in copying iterations*v vlaues that are contained in
iterations*b blocks into a []int64 (higher values of iterations are likely to
yield a better throughput) => this requires n * (b + 8v) bytes of memory.

This method computes iterations as ramBudget / (b + 8v) (since an int64 is
8 bytes).
*/
func (op *BulkOperation) computeIterations(valueCount, ramBudget int) int {
	iterations := ramBudget / (op.ByteBlockCount() + 8*op.ByteValueCount())
	if iterations == 0 {
		// at least 1
		return 1
	} else if (iterations-1)*op.ByteValueCount() >= valueCount {
		// don't allocate for more than the size of the reader
		panic("not implemented yet")
	} else {
		return iterations
	}
}
