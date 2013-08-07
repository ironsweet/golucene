package store

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/util"
	"hash"
	"hash/crc32"
	"io"
	"log"
	"os"
)

type IndexInput interface {
	io.Closer
	// util.DataInput
	ReadByte() (b byte, err error)
	ReadBytes(buf []byte) error
	ReadShort() (n int16, err error)
	ReadInt() (n int32, err error)
	ReadVInt() (n int32, err error)
	ReadLong() (n int64, err error)
	ReadVLong() (n int64, err error)
	ReadString() (s string, err error)
	ReadStringStringMap() (m map[string]string, err error)
	ReadStringSet() (m map[string]bool, err error)
	// IndexInput
	FilePointer() int64
	Seek(pos int64)
	Length() int64
	// Clone
	Clone() IndexInput
}

type IndexInputImpl struct {
	*util.DataInputImpl
	desc   string
	close  func() error
	length func() int64
}

func newIndexInputImpl(desc string, r util.DataReader) *IndexInputImpl {
	if desc == "" {
		panic("resourceDescription must not be null")
	}
	super := &util.DataInputImpl{r}
	return &IndexInputImpl{DataInputImpl: super, desc: desc}
}

func (in *IndexInputImpl) Length() int64 {
	return in.length()
}

func (in *IndexInputImpl) Close() error {
	return in.close()
}

type BufferedIndexInput struct {
	*IndexInputImpl
	bufferSize     int
	buffer         []byte
	bufferStart    int64
	bufferLength   int
	bufferPosition int
	seekInternal   func(pos int64)
	readInternal   func(buf []byte) error
}

func newBufferedIndexInput(desc string, context IOContext) *BufferedIndexInput {
	return newBufferedIndexInputBySize(desc, bufferSize(context))
}

func newBufferedIndexInputBySize(desc string, bufferSize int) *BufferedIndexInput {
	checkBufferSize(bufferSize)
	ans := &BufferedIndexInput{bufferSize: bufferSize}
	ans.IndexInputImpl = newIndexInputImpl(desc, ans)
	return ans
}

func (in *BufferedIndexInput) ReadByte() (b byte, err error) {
	if in.bufferPosition >= in.bufferLength {
		in.refill()
	}
	in.bufferPosition++
	return in.buffer[in.bufferPosition-1], nil
}

func (in *BufferedIndexInput) newBuffer(newBuffer []byte) {
	// Subclasses can do something here
	in.buffer = newBuffer
}

func checkBufferSize(bufferSize int) {
	if bufferSize <= 0 {
		panic(fmt.Sprintf("bufferSize must be greater than 0 (got %v)", bufferSize))
	}
}

func (in *BufferedIndexInput) ReadBytes(buf []byte) error {
	return in.ReadBytesBuffered(buf, true)
}

func (in *BufferedIndexInput) ReadBytesBuffered(buf []byte, useBuffer bool) error {
	available := in.bufferLength - in.bufferPosition
	if length := len(buf); length <= available {
		// the buffer contains enough data to satisfy this request
		if length > 0 { // to allow b to be null if len is 0...
			copy(buf, in.buffer[in.bufferPosition:in.bufferPosition+length])
		}
		in.bufferPosition += length
	} else {
		// the buffer does not have enough data. First serve all we've got.
		if available > 0 {
			copy(buf, in.buffer[in.bufferPosition:in.bufferPosition+available])
			buf = buf[available:]
			in.bufferPosition += available
		}
		// and now, read the remaining 'len' bytes:
		if length := len(buf); useBuffer && length < in.bufferSize {
			// If the amount left to read is small enough, and
			// we are allowed to use our buffer, do it in the usual
			// buffered way: fill the buffer and copy from it:
			if err := in.refill(); err != nil {
				return err
			}
			if in.bufferLength < length {
				// Throw an exception when refill() could not read len bytes:
				copy(buf, in.buffer[0:in.bufferLength])
				return errors.New(fmt.Sprintf("read past EOF: %v", in))
			} else {
				copy(buf, in.buffer[0:length])
				in.bufferPosition += length
			}
		} else {
			// The amount left to read is larger than the buffer
			// or we've been asked to not use our buffer -
			// there's no performance reason not to read it all
			// at once. Note that unlike the previous code of
			// this function, there is no need to do a seek
			// here, because there's no need to reread what we
			// had in the buffer.
			length := len(buf)
			after := in.bufferStart + int64(in.bufferPosition) + int64(length)
			if after > in.Length() {
				return errors.New(fmt.Sprintf("read past EOF: %v", in))
			}
			if err := in.readInternal(buf); err != nil {
				return err
			}
			in.bufferStart = after
			in.bufferPosition = 0
			in.bufferLength = 0 // trigger refill() on read
		}
	}
	return nil
}

func (in *BufferedIndexInput) ReadShort() (n int16, err error) {
	if 2 <= in.bufferLength-in.bufferPosition {
		in.bufferPosition += 2
		return int16((in.buffer[in.bufferPosition-2] << 8) | (in.buffer[in.bufferPosition-1])), nil
	}
	return in.DataInputImpl.ReadShort()
}

func (in *BufferedIndexInput) ReadInt() (n int32, err error) {
	log.Print("Reading int from buffer...")
	if 4 <= in.bufferLength-in.bufferPosition {
		in.bufferPosition += 4
		return int32((in.buffer[in.bufferPosition-4] << 24) | (in.buffer[in.bufferPosition-3] << 16) |
			(in.buffer[in.bufferPosition-2] << 8) | (in.buffer[in.bufferPosition-1])), nil
	}
	return in.DataInputImpl.ReadInt()
}

func (in *BufferedIndexInput) ReadLong() (n int64, err error) {
	if 8 <= in.bufferLength-in.bufferPosition {
		in.bufferPosition += 4
		i1 := int64((in.buffer[in.bufferPosition-4] << 24) | (in.buffer[in.bufferPosition-3] << 16) |
			(in.buffer[in.bufferPosition-2] << 8) | (in.buffer[in.bufferPosition-1]))
		in.bufferPosition += 4
		i2 := int64((in.buffer[in.bufferPosition-4] << 24) | (in.buffer[in.bufferPosition-3] << 16) |
			(in.buffer[in.bufferPosition-2] << 8) | (in.buffer[in.bufferPosition-1]))
		return (i1 << 32) | i2, nil
	}
	return in.DataInputImpl.ReadLong()
}

func (in *BufferedIndexInput) ReadVInt() (n int32, err error) {
	if 5 <= in.bufferLength-in.bufferPosition {
		b := in.buffer[in.bufferPosition]
		in.bufferPosition++
		if b >= 0 {
			return int32(b), nil
		}
		n := int32(b) & 0x7F
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int32(b) & 0x7F) << 7
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int32(b) & 0x7F) << 14
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int32(b) & 0x7F) << 21
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		// Warning: the next ands use 0x0F / 0xF0 - beware copy/paste errors:
		n |= (int32(b) & 0x0F) << 28
		if (b & 0xF0) == 0 {
			return n, nil
		}
		return 0, errors.New("Invalid vInt detected (too many bits)")
	}
	return in.DataInputImpl.ReadVInt()
}

func (in *BufferedIndexInput) ReadVLong() (n int64, err error) {
	if 9 <= in.bufferLength-in.bufferPosition {
		b := in.buffer[in.bufferPosition]
		in.bufferPosition++
		if b >= 0 {
			return int64(b), nil
		}
		n := int64(b) & 0x7F
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 7
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 14
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 21
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 28
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 35
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 42
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 49
		if b >= 0 {
			return n, nil
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		n |= (int64(b) & 0x7F) << 56
		if b >= 0 {
			return n, nil
		}
		return 0, errors.New("Invalid vLong detected (negative values disallowed)")
	}
	return in.DataInputImpl.ReadVLong()
}

// use panic/recover to handle error
func (in *BufferedIndexInput) refill() error {
	start := in.bufferStart + int64(in.bufferPosition)
	end := start + int64(in.bufferSize)
	if end > in.Length() { // don't read past EOF
		end = in.Length()
	}
	newLength := int(end - start)
	if newLength <= 0 {
		return errors.New(fmt.Sprintf("read past EOF: %v", in))
	}

	if in.buffer == nil {
		in.newBuffer(make([]byte, in.bufferSize)) // allocate buffer lazily
		in.seekInternal(int64(in.bufferStart))
	}
	in.readInternal(in.buffer[0:newLength])
	in.bufferLength = newLength
	in.bufferStart = start
	in.bufferPosition = 0
	return nil
}

func (in *BufferedIndexInput) FilePointer() int64 {
	return in.bufferStart + int64(in.bufferPosition)
}

func (in *BufferedIndexInput) Seek(pos int64) {
	if pos >= in.bufferStart && pos < in.bufferStart+int64(in.bufferLength) {
		in.bufferPosition = int(pos - in.bufferStart) // seek within buffer
	} else {
		in.bufferStart = pos
		in.bufferPosition = 0
		in.bufferLength = 0 // trigger refill() on read()
		in.seekInternal(pos)
	}
}

func (in *BufferedIndexInput) Clone() IndexInput {
	var clone BufferedIndexInput = *in
	clone.buffer = nil
	clone.bufferLength = 0
	clone.bufferPosition = 0
	clone.bufferStart = in.FilePointer()
	return &clone
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

type FSIndexInput struct {
	*BufferedIndexInput
	file      *os.File
	isClone   bool
	chunkSize int
	off       int64
	end       int64
}

func newFSIndexInput(desc, path string, context IOContext, chunkSize int) (in *FSIndexInput, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	super := newBufferedIndexInput(desc, context)
	in = &FSIndexInput{super, f, false, chunkSize, 0, fi.Size()}
	super.length = func() int64 {
		return in.end - in.off
	}
	super.close = func() error {
		// only close the file if this is not a clone
		if !in.isClone {
			in.file.Close()
		}
		return nil
	}
	return in, nil
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

func (in *ChecksumIndexInput) Seek(pos int64) {
	panic("unsupported")
}

func (in *ChecksumIndexInput) Length() int64 {
	return in.main.Length()
}

type ByteArrayDataInput struct {
	bytes []byte
	pos   int
	limit int
}

func NewByteArrayDataInput(bytes []byte) *ByteArrayDataInput {
	return &ByteArrayDataInput{bytes, 0, len(bytes)}
}

func (in *ByteArrayDataInput) Length() int {
	return in.limit
}

func (in *ByteArrayDataInput) ReadShort() (n int16, err error) {
	in.pos += 2
	return (int16(in.bytes[in.pos-2]) << 8) | int16(in.bytes[in.pos-1]), nil
}

func (in *ByteArrayDataInput) ReadInt() (n int32, err error) {
	in.pos += 4
	return (int32(in.bytes[in.pos-4]) << 24) | (int32(in.bytes[in.pos-3]) << 16) |
		(int32(in.bytes[in.pos-2]) << 8) | int32(in.bytes[in.pos-1]), nil
}

func (in *ByteArrayDataInput) ReadLong() (n int64, err error) {
	i1, _ := in.ReadInt()
	i2, _ := in.ReadInt()
	return (int64(i1) << 32) | int64(i2), nil
}

func (in *ByteArrayDataInput) ReadVInt() (n int32, err error) {
	b := in.bytes[in.pos]
	in.pos++
	if b >= 0 {
		return int32(b), nil
	}
	n = int32(b) & 0x7F
	b = in.bytes[in.pos]
	in.pos++
	n |= (int32(b) & 0x7F) << 7
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int32(b) & 0x7F) << 14
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int32(b) & 0x7F) << 21
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	// Warning: the next ands use 0x0F / 0xF0 - beware copy/paste errors:
	n |= (int32(b) & 0x0F) << 28
	if (b & 0xF0) == 0 {
		return n, nil
	}
	return 0, errors.New("Invalid vInt detected (too many bits)")
}

func (in *ByteArrayDataInput) ReadVLong() (n int64, err error) {
	b := in.bytes[in.pos]
	in.pos++
	if b >= 0 {
		return int64(b), nil
	}
	n = int64(b) & 0x7F
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 7
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 14
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 21
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 28
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 35
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 42
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 49
	if b >= 0 {
		return n, nil
	}
	b = in.bytes[in.pos]
	in.pos++
	n |= (int64(b) & 0x7F) << 56
	if b >= 0 {
		return n, nil
	}
	return 0, errors.New("Invalid vLong detected (negative values disallowed)")
}

func (in *ByteArrayDataInput) ReadByte() (b byte, err error) {
	in.pos++
	return in.bytes[in.pos-1], nil
}

func (in *ByteArrayDataInput) ReadBytes(buf []byte) error {
	copy(buf, in.bytes[in.pos:in.pos+len(buf)])
	in.pos += len(buf)
	return nil
}
