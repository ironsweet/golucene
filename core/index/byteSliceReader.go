package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

/*
IndexInput that knows how to read the byte slices written by Posting
and PostingVector. We read the bytes in each slice until we hit the
end of that slice at which point we read the forwarding address of
the next slice and then jump to it.
*/
type ByteSliceReader struct {
	*util.DataInputImpl
	pool         *util.ByteBlockPool
	bufferUpto   int
	buffer       []byte
	upto         int
	limit        int
	level        int
	bufferOffset int

	endIndex int
}

func newByteSliceReader() *ByteSliceReader {
	ans := new(ByteSliceReader)
	ans.DataInputImpl = &util.DataInputImpl{ans}
	return ans
}

func (r *ByteSliceReader) init(pool *util.ByteBlockPool, startIndex, endIndex int) {
	assert(endIndex-startIndex >= 0)
	assert(startIndex >= 0)
	assert(endIndex >= 0)

	r.pool = pool
	r.endIndex = endIndex

	r.level = 0
	r.bufferUpto = startIndex / util.BYTE_BLOCK_SIZE
	r.bufferOffset = r.bufferUpto * util.BYTE_BLOCK_SIZE
	r.buffer = pool.Buffers[r.bufferUpto]
	r.upto = startIndex & util.BYTE_BLOCK_MASK

	firstSize := util.LEVEL_SIZE_ARRAY[0]

	if startIndex+firstSize >= endIndex {
		// there is only this one slice to read
		r.limit = endIndex & util.BYTE_BLOCK_MASK
	} else {
		r.limit = r.upto + firstSize - 4
	}
}

func (r *ByteSliceReader) eof() bool {
	assert(r.upto+r.bufferOffset <= r.endIndex)
	return r.upto+r.bufferOffset == r.endIndex
}

func (r *ByteSliceReader) ReadByte() (byte, error) {
	assert(!r.eof())
	assert(r.upto <= r.limit)
	if r.upto == r.limit {
		r.nextSlice()
	}
	b := r.buffer[r.upto]
	r.upto++
	return b, nil
}

func (r *ByteSliceReader) nextSlice() {
	panic("not implemented yet")
}

func (r *ByteSliceReader) ReadBytes(buf []byte) error {
	panic("not implemented yet")
}
