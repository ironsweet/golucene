package store

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"hash"
	"hash/crc32"
	"io"
)

type IndexInput interface {
	io.Closer
	util.DataInput
	ReadBytesBuffered(buf []byte, useBuffer bool) error
	// IndexInput
	FilePointer() int64
	/** Sets current position in this file, where the next read will occur.
	 * @see #getFilePointer()
	 */
	Seek(pos int64) error
	Length() int64
	// Clone
	Clone() IndexInput
}

type LengthCloser interface {
	Close() error
	Length() int64
}

type IndexInputImpl struct {
	*util.DataInputImpl
	LengthCloser
	desc string
}

func newIndexInputImpl(desc string, r util.DataReader) *IndexInputImpl {
	assert2(desc != "", "resourceDescription must not be null")
	super := &util.DataInputImpl{Reader: r}
	return &IndexInputImpl{DataInputImpl: super, desc: desc}
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

type ChecksumIndexInput struct {
	*IndexInputImpl
	main   IndexInput
	digest hash.Hash32
}

func NewChecksumIndexInput(main IndexInput) *ChecksumIndexInput {
	ans := &ChecksumIndexInput{main: main, digest: crc32.NewIEEE()}
	ans.IndexInputImpl = newIndexInputImpl(fmt.Sprintf("ChecksumIndexInput(%v)", main), ans)
	return ans
}

func (in *ChecksumIndexInput) ReadByte() (b byte, err error) {
	if b, err = in.main.ReadByte(); err == nil {
		in.digest.Write([]byte{b})
	}
	return b, err
}

func (in *ChecksumIndexInput) ReadBytes(buf []byte) error {
	err := in.main.ReadBytes(buf)
	if err == nil {
		in.digest.Write(buf)
	}
	return err
}

func (in *ChecksumIndexInput) Checksum() int64 {
	return int64(in.digest.Sum32())
}

func (in *ChecksumIndexInput) Close() error {
	return in.main.Close()
}

func (in *ChecksumIndexInput) FilePointer() int64 {
	return in.main.FilePointer()
}

func (in *ChecksumIndexInput) Seek(pos int64) error {
	panic("unsupported")
}

func (in *ChecksumIndexInput) Length() int64 {
	return in.main.Length()
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
	ans.DataInputImpl = &util.DataInputImpl{ans}
	ans.Reset(bytes)
	return ans
}

func NewEmptyByteArrayDataInput() *ByteArrayDataInput {
	ans := &ByteArrayDataInput{}
	ans.DataInputImpl = &util.DataInputImpl{ans}
	ans.Reset(make([]byte, 0))
	return ans
}

func (in *ByteArrayDataInput) Reset(bytes []byte) {
	in.bytes = bytes
	in.Pos = 0
	in.limit = len(bytes)
}

// NOTE: sets pos to 0, which is not right if you had
// called reset w/ non-zero offset!!
func (in *ByteArrayDataInput) Rewind() {
	in.Pos = 0
}

func (in *ByteArrayDataInput) Length() int {
	return in.limit
}

func (in *ByteArrayDataInput) SkipBytes(count int) {
	in.Pos += count
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
