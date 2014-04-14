package packed

// util/packed/BulkOperationPacked.java

// Non-specialized BulkOperation for Packed format
type BulkOperationPacked struct {
	*BulkOperationImpl
	bitsPerValue   int
	longBlockCount int
	longValueCount int
	byteBlockCount int
	byteValueCount int
	mask           int64
	intMask        int
}

func newBulkOperationPacked(bitsPerValue uint32) *BulkOperationPacked {
	self := &BulkOperationPacked{}
	self.bitsPerValue = int(bitsPerValue)
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

func (p *BulkOperationPacked) encodeLongToLong(values, blocks []int64, iterations int) {
	var nextBlock int64 = 0
	var bitsLeft int = 64
	valuesOffset, blocksOffset := 0, 0
	for i, limit := 0, p.longValueCount*iterations; i < limit; i++ {
		bitsLeft -= p.bitsPerValue
		switch {
		case bitsLeft > 0:
			nextBlock |= (values[valuesOffset] << uint(bitsLeft))
			valuesOffset++
		case bitsLeft == 0:
			nextBlock |= values[valuesOffset]
			valuesOffset++
			blocks[blocksOffset] = nextBlock
			blocksOffset++
			nextBlock = 0
			bitsLeft = 64
		default: // bitsLeft < 0
			nextBlock |= int64(uint64(values[valuesOffset]) >> uint(-bitsLeft))
			blocks[blocksOffset] = nextBlock
			blocksOffset++
			nextBlock = (values[valuesOffset] & ((1 << uint(-bitsLeft)) - 1) << uint(64+bitsLeft))
			bitsLeft += 64
		}
	}
}

func (p *BulkOperationPacked) encodeLongToByte(values []int64, blocks []byte, iterations int) {
	var nextBlock int = 0
	var bitsLeft int = 0
	valuesOffset := 0
	// valuesOffset, blocksOffset := 0, 0
	for i, limit := 0, p.byteValueCount*iterations; i < limit; i++ {
		v := values[valuesOffset]
		valuesOffset++
		assert(p.bitsPerValue == 64 || BitsRequired(v) <= p.bitsPerValue)
		if p.bitsPerValue < bitsLeft { // just buffer
			nextBlock |= int(v << uint(bitsLeft-p.bitsPerValue))
			bitsLeft -= p.bitsPerValue
		} else { // flush as many blocks as possible
			panic("not implemented yet")
		}
	}
	assert(bitsLeft == 8)
}

// util/packed/BulkOperationPackedSingleBlock.java

// Non-specialized BulkOperation for PACKED_SINGLE_BLOCK format
type BulkOperationPackedSingleBlock struct {
	*BulkOperationImpl
	bitsPerValue int
	valueCount   int
	mask         int64
}

const BLOCK_COUNT = 1

func newBulkOperationPackedSingleBlock(bitsPerValue uint32) BulkOperation {
	// log.Printf("Initializing BulkOperationPackedSingleBlock(%v)", bitsPerValue)
	self := &BulkOperationPackedSingleBlock{
		bitsPerValue: int(bitsPerValue),
		valueCount:   64 / int(bitsPerValue),
		mask:         (int64(1) << bitsPerValue) - 1,
	}
	self.BulkOperationImpl = newBulkOperationImpl(self)
	return self
}

func (p *BulkOperationPackedSingleBlock) ByteValueCount() int {
	return p.valueCount
}

func (p *BulkOperationPackedSingleBlock) longToLong(values []int64) int64 {
	off := 0
	block := values[off]
	off++
	for j := 1; j < p.valueCount; j++ {
		block |= values[off] << uint(j*p.bitsPerValue)
		off++
	}
	return block
}

func (p *BulkOperationPackedSingleBlock) encodeLongToLong(values,
	blocks []int64, iterations int) {
	valuesOffset, blocksOffset := 0, 0
	for i, limit := 0, iterations; i < limit; i++ {
		blocks[blocksOffset] = p.longToLong(values[valuesOffset:])
		blocksOffset++
		valuesOffset += p.valueCount
	}
}

func (p *BulkOperationPackedSingleBlock) encodeLongToByte(values []int64,
	blocks []byte, iterations int) {
	panic("not implemented yet")
}
