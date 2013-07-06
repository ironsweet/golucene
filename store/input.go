package store

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
)

type DataInput struct {
	/* Reads and returns a single byte.	*/
	ReadByte func() (b byte, err error)
	/* Reads a specified number of bytes into an array */
	ReadBytes func(buf []byte) error
}

func (in *DataInput) ReadInt() (n int32, err error) {
	if b1, err := in.ReadByte(); err == nil {
		if b2, err := in.ReadByte(); err == nil {
			if b3, err := in.ReadByte(); err == nil {
				if b4, err := in.ReadByte(); err == nil {
					return (int32(b1) << 24) | (int32(b2) << 16) | (int32(b3) << 8) | int32(b4), nil
				}
			}
		}
	}
	return 0, err
}

func (in *DataInput) ReadShort() (n int16, err error) {
	if b1, err := in.ReadByte(); err == nil {
		if b2, err := in.ReadByte(); err == nil {
			return (int16(b1) << 8) | int16(b2), nil
		}
	}
	return 0, err
}

func (in *DataInput) ReadVInt() (n int32, err error) {
	if b, err := in.ReadByte(); err == nil {
		n = int32(b) & 0x7F
		if b >= 0 {
			return n, nil
		}
		if b, err = in.ReadByte(); err == nil {
			n |= (int32(b) & 0x7F) << 7
			if b >= 0 {
				return n, nil
			}
			if b, err = in.ReadByte(); err == nil {
				n |= (int32(b) & 0x7F) << 14
				if b >= 0 {
					return n, nil
				}
				if b, err = in.ReadByte(); err == nil {
					n |= (int32(b) & 0x7F) << 21
					if b >= 0 {
						return n, nil
					}
					if b, err = in.ReadByte(); err == nil {
						// Warning: the next ands use 0x0F / 0xF0 - beware copy/paste errors:
						n |= (int32(b) & 0x0F) << 28
						if int32(b)&0x0F == 0 {
							return n, nil
						}
						return 0, errors.New("Invalid vInt detected (too many bits)")
					}
				}
			}
		}
	}
	return 0, err
}

func (in *DataInput) ReadLong() (n int64, err error) {
	d1, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	d2, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	return (int64(d1) << 32) | (int64(d2) & 0xFFFFFFFF), nil
}

func (in *DataInput) ReadVLong() (n int64, err error) {
	if b, err := in.ReadByte(); err == nil {
		n = int64(b) & 0x7F
		if b >= 0 {
			return n, nil
		}
		if b, err = in.ReadByte(); err == nil {
			n |= (int64(b) & 0x7F) << 7
			if b >= 0 {
				return n, nil
			}
			if b, err = in.ReadByte(); err == nil {
				n |= (int64(b) & 0x7F) << 14
				if b >= 0 {
					return n, nil
				}
				if b, err = in.ReadByte(); err == nil {
					n |= (int64(b) & 0x7F) << 21
					if b >= 0 {
						return n, nil
					}
					if b, err = in.ReadByte(); err == nil {
						n |= (int64(b) & 0x7F) << 28
						if b >= 0 {
							return n, nil
						}
						if b, err = in.ReadByte(); err == nil {
							n |= (int64(b) & 0x7F) << 35
							if b >= 0 {
								return n, nil
							}
							if b, err = in.ReadByte(); err == nil {
								n |= (int64(b) & 0x7F) << 42
								if b >= 0 {
									return n, nil
								}
								if b, err = in.ReadByte(); err == nil {
									n |= (int64(b) & 0x7F) << 49
									if b >= 0 {
										return n, nil
									}
									if b, err = in.ReadByte(); err == nil {
										n |= (int64(b) & 0x7F) << 56
										if b >= 0 {
											return n, nil
										}
										return 0, errors.New("Invalid vLong detected (negative values disallowed)")
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return 0, err
}

func (in *DataInput) ReadString() (s string, err error) {
	length, err := in.ReadVInt()
	if err != nil {
		return "", err
	}
	bytes := make([]byte, length)
	in.ReadBytes(bytes)
	return string(bytes), nil
}

func (in *DataInput) ReadStringStringMap() (m map[string]string, err error) {
	count, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	m = make(map[string]string)
	for i := int32(0); i < count; i++ {
		key, err := in.ReadString()
		if err != nil {
			return nil, err
		}
		value, err := in.ReadString()
		if err != nil {
			return nil, err
		}
		m[key] = value
	}
	return m, nil
}

func (in *DataInput) ReadStringSet() (s map[string]bool, err error) {
	count, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	s = make(map[string]bool)
	for i := int32(0); i < count; i++ {
		key, err := in.ReadString()
		if err != nil {
			return nil, err
		}
		s[key] = true
	}
	return s, nil
}

type IndexInput struct {
	*DataInput
	desc        string
	close       func() error
	FilePointer func() int64
	Seek        func(pos int64)
	Length      func() int64
}

func newIndexInput(desc string) *IndexInput {
	if desc == "" {
		panic("resourceDescription must not be null")
	}
	super := &DataInput{}
	return &IndexInput{DataInput: super, desc: desc}
}

func (in *IndexInput) Close() error {
	return in.close()
}

type BufferedIndexInput struct {
	*IndexInput
	bufferSize     int
	buffer         []byte
	bufferStart    int64
	bufferLength   int
	bufferPosition int
	seekInternal   func(pos int64)
	readInternal   func(b []byte, offset, length int) error
}

func newBufferedIndexInput(desc string, context IOContext) *BufferedIndexInput {
	return newBufferedIndexInputBySize(desc, bufferSize(context))
}

func newBufferedIndexInputBySize(desc string, bufferSize int) *BufferedIndexInput {
	super := newIndexInput(desc)
	checkBufferSize(bufferSize)
	in := &BufferedIndexInput{IndexInput: super, bufferSize: bufferSize}
	super.ReadByte = func() (b byte, err error) {
		if in.bufferPosition >= in.bufferLength {
			in.refill()
		}
		b = in.buffer[in.bufferPosition]
		in.bufferPosition++
		return b, nil
	}
	super.FilePointer = func() int64 {
		return in.bufferStart + int64(in.bufferPosition)
	}
	super.Seek = func(pos int64) {
		if pos >= in.bufferStart && pos < (in.bufferStart+int64(in.bufferLength)) {
			in.bufferPosition = int(pos - in.bufferStart)
		} else {
			in.bufferStart = pos
			in.bufferPosition = 0
			in.bufferLength = 0 // trigger refill() on read()
			in.seekInternal(pos)
		}
	}
	return in
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
	in.readInternal(in.buffer, 0, newLength)
	in.bufferLength = newLength
	in.bufferStart = start
	in.bufferPosition = 0
	return nil
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
	super.Length = func() int64 {
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
	*IndexInput
	main   *IndexInput
	digest hash.Hash32
}

func NewChecksumIndexInput(main *IndexInput) *ChecksumIndexInput {
	super := newIndexInput(fmt.Sprintf("ChecksumIndexInput(%v)", main))
	digest := crc32.NewIEEE()
	super.ReadByte = func() (b byte, err error) {
		if b, err = main.ReadByte(); err == nil {
			digest.Write([]byte{b})
		}
		return b, err
	}
	super.ReadBytes = func(buf []byte) error {
		err := main.ReadBytes(buf)
		if err == nil {
			digest.Write(buf)
		}
		return err
	}
	super.FilePointer = main.FilePointer
	super.Length = main.Length
	return &ChecksumIndexInput{super, main, digest}
}

func (in *ChecksumIndexInput) Checksum() int64 {
	return int64(in.digest.Sum32())
}

type IndexInputSlicer interface {
	io.Closer
	openSlice(desc string, offset, length int64) IndexInput
	openFullSlice() IndexInput
}

type SlicedIndexInput struct {
	*BufferedIndexInput
	base       IndexInput
	fileOffset int64
	length     int64
}

func newSlicedIndexInput(desc string, base *IndexInput, fileOffset, length int64) SlicedIndexInput {
	return newSlicedIndexInputBySize(desc, base, fileOffset, length, BUFFER_SIZE)
}

func newSlicedIndexInputBySize(desc string, base *IndexInput, fileOffset, length int64, bufferSize int) SlicedIndexInput {
	return SlicedIndexInput{
		BufferedIndexInput: newBufferedIndexInputBySize(fmt.Sprintf(
			"SlicedIndexInput(%v in %v slice=%v:%v)", desc, base, fileOffset, fileOffset+length), bufferSize),
		base: *base, fileOffset: fileOffset, length: length}
}
