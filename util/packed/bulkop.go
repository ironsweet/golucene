package packed

// util/packed/BulkOperationPacked.java

// Non-specialized BulkOperation for Packed format
type BulkOperationPacked struct {
	*BulkOperationImpl
	bitsPerValue   uint32
	longBlockCount int
	longValueCount int
	byteBlockCount int
	byteValueCount int
	mask           int64
	intMask        int
}

func newBulkOperationPacked(bitsPerValue uint32) *BulkOperationPacked {
	self := &BulkOperationPacked{}
	self.bitsPerValue = bitsPerValue
	assert(bitsPerValue > 0 && bitsPerValue <= 64)
	blocks := uint32(bitsPerValue)
	for (blocks & 1) == 0 {
		blocks = (blocks >> 1)
	}
	self.longBlockCount = int(blocks)
	self.longValueCount = 64 * self.longBlockCount / int(bitsPerValue)
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
	assert(self.longValueCount*int(bitsPerValue) == 64*self.longBlockCount)
	self.BulkOperationImpl = newBulkOperationImpl(self)
	return self
}

func (p *BulkOperationPacked) ByteBlockCount() int {
	return p.byteBlockCount
}

func (p *BulkOperationPacked) ByteValueCount() int {
	return p.byteValueCount
}

// util/packed/BulkOperationPackedSingleBlock.java

// Non-specialized BulkOperation for PACKED_SINGLE_BLOCK format
type BulkOperationPackedSingleBlock struct {
	*BulkOperationImpl
	bitsPerValue uint32
	valueCount   int
	mask         int64
}

const BLOCK_COUNT = 1

func newBulkOperationPackedSingleBlock(bitsPerValue uint32) BulkOperation {
	// log.Printf("Initializing BulkOperationPackedSingleBlock(%v)", bitsPerValue)
	self := &BulkOperationPackedSingleBlock{
		bitsPerValue: bitsPerValue,
		valueCount:   64 / int(bitsPerValue),
		mask:         (int64(1) << bitsPerValue) - 1,
	}
	self.BulkOperationImpl = newBulkOperationImpl(self)
	return self
}

func (p *BulkOperationPackedSingleBlock) ByteValueCount() int {
	return p.valueCount
}
