package lucene40

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"reflect"
)

const (
	CODEC = "BitVector"

	/* Change DGaps to encode gaps between cleared bits, not set: */
	BV_VERSION_DGAPS_CLEARED = 1

	BV_VERSION_CHECKSUM = 2

	/* Imcrement version to change it: */
	BV_VERSION_CURRENT = BV_VERSION_CHECKSUM
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

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (bv *BitVector) Length() int {
	return bv.size
}

/*
Returns the total number of bits in this vector. This is efficiently
computed and cached, so that, if the vector is not changed, no
recomputation is done for repeated calls.
*/
func (bv *BitVector) Count() int {
	// if the vector has been modified
	if bv.count == -1 {
		c := 0
		for _, v := range bv.bits {
			c += util.BitCount(v) // sum bits per byte
		}
		bv.count = c
	}
	assert2(bv.count <= bv.size, "count=%v size=%v", bv.count, bv.size)
	return bv.count
}

/*
Writes this vector to the file name in Directory d, in a format that
can be read by the constructor BitVector(Directory, String, IOContext)
*/
func (bv *BitVector) Write(d store.Directory, name string, ctx store.IOContext) (err error) {
	assert(reflect.TypeOf(d).Name() != "CompoundFileDirectory")
	var output store.IndexOutput
	if output, err = d.CreateOutput(name, ctx); err != nil {
		return err
	}
	defer func() {
		err = mergeError(err, output.Close())
	}()

	if err = output.WriteInt(-2); err != nil {
		return err
	}
	if err = codec.WriteHeader(output, CODEC, BV_VERSION_CURRENT); err != nil {
		return err
	}
	if bv.isSparse() {
		// sparse bit-set more efficiently saved as d-gaps.
		err = bv.writeClearedDgaps(output)
	} else {
		err = bv.writeBits(output)
	}
	if err != nil {
		return err
	}
	if err = codec.WriteFooter(output); err != nil {
		return err
	}
	bv.assertCount()
	return nil
}

func mergeError(err, err2 error) error {
	if err == nil {
		return err2
	} else {
		return errors.New(fmt.Sprintf("%v\n  %v", err, err2))
	}
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

/* Write as a bit set */
func (bv *BitVector) writeBits(output store.IndexOutput) error {
	return store.Stream(output).
		WriteInt(int32(bv.size)).
		WriteInt(int32(bv.Count())).
		WriteBytes(bv.bits).
		Close()
}

/* Write as a d-gaps list */
func (bv *BitVector) writeClearedDgaps(output store.IndexOutput) error {
	err := store.Stream(output).
		WriteInt(-1). // mark using d-gaps
		WriteInt(int32(bv.size)).
		WriteInt(int32(bv.Count())).
		Close()
	if err != nil {
		return err
	}
	last, numCleared := 0, bv.size-bv.Count()
	for i, v := range bv.bits {
		if v == byte(0xff) {
			continue
		}
		err = output.WriteVInt(int32(i - last))
		if err == nil {
			err = output.WriteByte(v)
		}
		if err != nil {
			return err
		}
		last = i
		numCleared -= (8 - util.BitCount(v))
		assert(numCleared >= 0 ||
			i == len(bv.bits)-1 && numCleared == -(8-(bv.size&7)))
		if numCleared <= 0 {
			break
		}
	}
	return nil
}

/*
Indicates if the bit vector is sparse and should be saved as a d-gaps
list, or dense, and should be saved as a bit set.
*/
func (bv *BitVector) isSparse() bool {
	panic("not implemented yet")
}

func (bv *BitVector) assertCount() {
	assert(bv.count != -1)
	countSav := bv.count
	bv.count = -1
	assert2(countSav == bv.Count(), "saved count was %v but recomputed count is %v", countSav, bv.count)
}
