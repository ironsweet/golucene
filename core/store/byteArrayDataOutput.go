package store

import (
	"github.com/balzaczyy/golucene/core/util"
)

type ByteArrayDataOutput struct {
	*util.DataOutputImpl
	data  []byte
	pos   int
	limit int
}

func NewByteArrayDataOutput(data []byte) *ByteArrayDataOutput {
	ans := &ByteArrayDataOutput{}
	ans.DataOutputImpl = util.NewDataOutput(ans)
	ans.reset(data)
	return ans
}

func (o *ByteArrayDataOutput) reset(data []byte) {
	o.data = data
	o.pos = 0
	o.limit = len(data)
}

func (o *ByteArrayDataOutput) Position() int {
	return o.pos
}

func (o *ByteArrayDataOutput) WriteByte(b byte) error {
	assert(o.pos < o.limit)
	o.data[o.pos] = b
	o.pos++
	return nil
}

func (o *ByteArrayDataOutput) WriteBytes(b []byte) error {
	assert(o.pos+len(b) <= o.limit)
	copy(o.data[o.pos:], b)
	o.pos += len(b)
	return nil
}
