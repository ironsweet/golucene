package lucene40

import (
	"fmt"
)

type BitVector struct {
	bits  []byte
	size  int
	count int
}

func NewBitVector(n int) *BitVector {
	return &BitVector{
		size: n,
		bits: make([]byte, numBytes(n)),
	}
}

func numBytes(size int) int {
	bytesLength := int(uint(size) >> 3)
	if (size & 7) != 0 {
		bytesLength++
	}
	return bytesLength
}

func (bv *BitVector) Clear(bit int) {
	assert2(bit >= 0 && bit < bv.size, "bit %v is out of bounds 0..%v", bit, bv.size-1)
	bv.bits[bit>>3] &= ^(1 << (uint(bit) & 7))
	bv.count = -1
}

func (bv *BitVector) At(bit int) bool {
	assert2(bit >= 0 && bit < bv.size, "bit %v is out of bounds 0..%v", bit, bv.size-1)
	return (bv.bits[bit>>3] & (1 << (uint(bit) & 7))) != 0
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (bv *BitVector) Length() int {
	return bv.size
}

/* Invert all bits */
func (bv *BitVector) InvertAll() {
	if bv.count != -1 {
		bv.count = bv.size - bv.count
	}
	if len(bv.bits) > 0 {
		for idx, v := range bv.bits {
			bv.bits[idx] = byte(^v)
		}
	}
}
