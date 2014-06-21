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
}

func (r *ByteSliceReader) init(pool *util.ByteBlockPool, startIndex, endIndex int) {
	panic("not implemented yet")
}

func (r *ByteSliceReader) eof() bool {
	panic("not implemented yet")
}
