package packed

import ()

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

func (p *BulkOperationPacked) LongBlockCount() int {
	return p.longBlockCount
}

func (p *BulkOperationPacked) LongValueCount() int {
	return p.longValueCount
}

func (p *BulkOperationPacked) ByteBlockCount() int {
	return p.byteBlockCount
}

func (p *BulkOperationPacked) ByteValueCount() int {
	return p.byteValueCount
}

func (p *BulkOperationPacked) decodeLongToLong(blocks, values []int64, iterations int) {
	blocksOff, valuesOff := 0, 0
	bitsLeft := 64
	for i := 0; i < p.longValueCount*iterations; i++ {
		bitsLeft -= p.bitsPerValue
		if bitsLeft < 0 {
			values[valuesOff] =
				((blocks[blocksOff] & ((int64(1) << uint(p.bitsPerValue+bitsLeft)) - 1)) << uint(-bitsLeft)) |
					int64(uint64(blocks[blocksOff+1])>>uint(64+bitsLeft))
			valuesOff++
			blocksOff++
			bitsLeft += 64
		} else {
			values[valuesOff] = int64(uint64(blocks[blocksOff])>>uint(bitsLeft)) & p.mask
			valuesOff++
		}
	}
}

func (p *BulkOperationPacked) decodeByteToLong(blocks []byte, values []int64, iterations int) {
	panic("niy")
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
			valuesOffset++
			bitsLeft += 64
		}
	}
}

func (p *BulkOperationPacked) encodeLongToByte(values []int64, blocks []byte, iterations int) {
	var nextBlock int = 0
	var bitsLeft int = 8
	valuesOffset, blocksOffset := 0, 0
	for i, limit := 0, p.byteValueCount*iterations; i < limit; i++ {
		v := values[valuesOffset]
		valuesOffset++
		assert(UnsignedBitsRequired(v) <= p.bitsPerValue)
		if p.bitsPerValue < bitsLeft { // just buffer
			nextBlock |= int(v << uint(bitsLeft-p.bitsPerValue))
			bitsLeft -= p.bitsPerValue
		} else { // flush as many blocks as possible
			bits := uint(p.bitsPerValue - bitsLeft)
			blocks[blocksOffset] = byte(nextBlock | int(uint64(v)>>bits))
			blocksOffset++
			for bits >= 8 {
				bits -= 8
				blocks[blocksOffset] = byte(uint64(v) >> bits)
				blocksOffset++
			}
			// then buffer
			bitsLeft = int(8 - bits)
			nextBlock = int((v & ((1 << bits) - 1)) << uint(bitsLeft))
		}
	}
	assert(bitsLeft == 8)
}

func (p *BulkOperationPacked) EncodeIntToByte(values []int, blocks []byte, iterations int) {
	valuesOff, blocksOff := 0, 0
	nextBlock, bitsLeft := 0, 8
	for i := 0; i < p.byteValueCount*iterations; i++ {
		v := values[valuesOff]
		valuesOff++
		assert(BitsRequired(int64(v)) <= p.bitsPerValue)
		if p.bitsPerValue < bitsLeft {
			// just buffer
			nextBlock |= (v << uint(bitsLeft-p.bitsPerValue))
			bitsLeft -= p.bitsPerValue
		} else {
			// flush as many blocks as possible
			bits := p.bitsPerValue - bitsLeft
			blocks[blocksOff] = byte(nextBlock | int(uint(v)>>uint(bits)))
			blocksOff++
			for bits >= 8 {
				bits -= 8
				blocks[blocksOff] = byte(uint(v) >> uint(bits))
				blocksOff++
			}
			// then buffer
			bitsLeft = 8 - bits
			nextBlock = (v & ((1 << uint(bits)) - 1)) << uint(bitsLeft)
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

func (p *BulkOperationPackedSingleBlock) LongBlockCount() int {
	return BLOCK_COUNT
}

func (p *BulkOperationPackedSingleBlock) ByteBlockCount() int {
	return BLOCK_COUNT * 8
}

func (p *BulkOperationPackedSingleBlock) LongValueCount() int {
	return p.valueCount
}

func (p *BulkOperationPackedSingleBlock) ByteValueCount() int {
	return p.valueCount
}

func (p *BulkOperationPackedSingleBlock) decodeLongs(block int64, values []int64) int {
	off := 0
	values[off] = block & p.mask
	off++
	for j := 1; j < p.valueCount; j++ {
		block = int64(uint64(block) >> uint(p.bitsPerValue))
		values[off] = block & p.mask
		off++
	}
	return off
}

func (p *BulkOperationPackedSingleBlock) encodeLongs(values []int64) int64 {
	off := 0
	block := values[off]
	off++
	for j := 1; j < p.valueCount; j++ {
		block |= (values[off] << uint(j*p.bitsPerValue))
		off++
	}
	return block
}

func (p *BulkOperationPackedSingleBlock) encodeInts(values []int) int64 {
	off := 0
	block := int64(values[off])
	off++
	for j := 1; j < p.valueCount; j++ {
		block |= int64(values[off]) << uint(j*p.bitsPerValue)
		off++
	}
	return block
}

func (p *BulkOperationPackedSingleBlock) decodeLongToLong(blocks, values []int64, iterations int) {
	blocksOffset, valuesOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := blocks[blocksOffset]
		blocksOffset++
		valuesOffset = p.decodeLongs(block, values[valuesOffset:])
	}
}

func (p *BulkOperationPackedSingleBlock) decodeByteToLong(blocks []byte,
	values []int64, iterations int) {
	panic("niy")
}

func (p *BulkOperationPackedSingleBlock) encodeLongToLong(values,
	blocks []int64, iterations int) {
	valuesOffset, blocksOffset := 0, 0
	for i, limit := 0, iterations; i < limit; i++ {
		blocks[blocksOffset] = p.encodeLongs(values[valuesOffset:])
		blocksOffset++
		valuesOffset += p.valueCount
	}
}

func (p *BulkOperationPackedSingleBlock) encodeLongToByte(values []int64,
	blocks []byte, iterations int) {

	valuesOffset, blocksOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := p.encodeLongs(values[valuesOffset:])
		valuesOffset += p.valueCount
		blocksOffset += p.writeLong(block, blocks[blocksOffset:])
	}
}

func (p *BulkOperationPackedSingleBlock) EncodeIntToByte(values []int,
	blocks []byte, iterations int) {

	valuesOffset, blocksOffset := 0, 0
	for i := 0; i < iterations; i++ {
		block := p.encodeInts(values[valuesOffset:])
		valuesOffset += p.valueCount
		blocksOffset += p.writeLong(block, blocks[blocksOffset:])
	}
}
