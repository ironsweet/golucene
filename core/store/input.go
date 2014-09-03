package store

import (
	"errors"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

type IndexInputService interface {
	io.Closer
	// Returns the current position in this file, where the next read will occur.
	FilePointer() int64
	// Sets current position in this file, where the next read will occur.
	Seek(pos int64) error
	Length() int64
}

type IndexInputSub interface {
	io.Closer
	util.DataReader
	FilePointer() int64
	Seek(pos int64) error
	Length() int64
}

type IndexInput interface {
	util.DataInput
	IndexInputService
	ReadBytesBuffered(buf []byte, useBuffer bool) error
	// Clone
	Clone() IndexInput
	// Creates a slice of this index input, with the given description,
	// offset, and length. The slice is seeked to the beginning.
	Slice(desc string, offset, length int64) (IndexInput, error)
}

type IndexInputImpl struct {
	*util.DataInputImpl
	desc string
}

func NewIndexInputImpl(desc string, r util.DataReader) *IndexInputImpl {
	assert2(desc != "", "resourceDescription must not be null")
	return &IndexInputImpl{
		DataInputImpl: &util.DataInputImpl{Reader: r},
		desc:          desc,
	}
}

func (in *IndexInputImpl) String() string {
	return in.desc
}

const (
	BUFFER_SIZE       = 1024
	MERGE_BUFFER_SIZE = 4096
)

func bufferSize(context IOContext) int {
	switch context.context {
	case IO_CONTEXT_TYPE_MERGE:
		// The normal read buffer size defaults to 1024, but
		// increasing this during merging seems to yield
		// performance gains.  However we don't want to increase
		// it too much because there are quite a few
		// BufferedIndexInputs created during merging.  See
		// LUCENE-888 for details.
		return MERGE_BUFFER_SIZE
	default:
		return BUFFER_SIZE
	}
}

// store/ByteArrayDataInput.java

// DataInput backed by a byte array.
// Warning: this class omits all low-level checks.
type ByteArrayDataInput struct {
	*util.DataInputImpl
	bytes []byte
	Pos   int
	limit int
}

func NewByteArrayDataInput(bytes []byte) *ByteArrayDataInput {
	ans := &ByteArrayDataInput{}
	ans.DataInputImpl = util.NewDataInput(ans)
	ans.Reset(bytes)
	return ans
}

func NewEmptyByteArrayDataInput() *ByteArrayDataInput {
	ans := &ByteArrayDataInput{}
	ans.DataInputImpl = util.NewDataInput(ans)
	ans.Reset(make([]byte, 0))
	return ans
}

func (in *ByteArrayDataInput) Reset(bytes []byte) {
	in.bytes = bytes
	in.Pos = 0
	in.limit = len(bytes)
}

func (in *ByteArrayDataInput) Position() int {
	return in.Pos
}

// NOTE: sets pos to 0, which is not right if you had
// called reset w/ non-zero offset!!
func (in *ByteArrayDataInput) Rewind() {
	in.Pos = 0
}

func (in *ByteArrayDataInput) Length() int {
	return in.limit
}

func (in *ByteArrayDataInput) SkipBytes(count int64) {
	in.Pos += int(count)
}

func (in *ByteArrayDataInput) ReadShort() (n int16, err error) {
	in.Pos += 2
	return (int16(in.bytes[in.Pos-2]) << 8) | int16(in.bytes[in.Pos-1]), nil
}

func (in *ByteArrayDataInput) ReadInt() (n int32, err error) {
	in.Pos += 4
	return (int32(in.bytes[in.Pos-4]) << 24) | (int32(in.bytes[in.Pos-3]) << 16) |
		(int32(in.bytes[in.Pos-2]) << 8) | int32(in.bytes[in.Pos-1]), nil
}

func (in *ByteArrayDataInput) ReadLong() (n int64, err error) {
	i1, _ := in.ReadInt()
	i2, _ := in.ReadInt()
	return (int64(i1) << 32) | int64(i2), nil
}

func (in *ByteArrayDataInput) ReadVInt() (n int32, err error) {
	b := in.bytes[in.Pos]
	in.Pos++
	if b < 128 {
		return int32(b), nil
	}
	n = int32(b) & 0x7F
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int32(b) & 0x7F) << 7
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int32(b) & 0x7F) << 14
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int32(b) & 0x7F) << 21
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	// Warning: the next ands use 0x0F / 0xF0 - beware copy/paste errors:
	n |= (int32(b) & 0x0F) << 28
	if (b & 0xF0) == 0 {
		return n, nil
	}
	return 0, errors.New("Invalid vInt detected (too many bits)")
}

func (in *ByteArrayDataInput) ReadVLong() (n int64, err error) {
	b := in.bytes[in.Pos]
	in.Pos++
	if b < 128 {
		return int64(b), nil
	}
	n = int64(b & 0x7F)
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 7)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 14)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 21)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 28)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 35)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 42)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 49)
	if b < 128 {
		return n, nil
	}
	b = in.bytes[in.Pos]
	in.Pos++
	n |= (int64(b&0x7F) << 56)
	if b < 128 {
		return n, nil
	}
	return 0, errors.New("Invalid vLong detected (negative values disallowed)")
}

func (in *ByteArrayDataInput) ReadByte() (b byte, err error) {
	in.Pos++
	return in.bytes[in.Pos-1], nil
}

func (in *ByteArrayDataInput) ReadBytes(buf []byte) error {
	copy(buf, in.bytes[in.Pos:in.Pos+len(buf)])
	in.Pos += len(buf)
	return nil
}
