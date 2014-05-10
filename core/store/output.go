package store

import (
	"github.com/balzaczyy/golucene/core/util"
	"hash"
	"hash/crc32"
	"io"
)

/*
Wrap IndexOutput to allow coding in flow style without worrying about
error check. If error happens in the middle, following calls are just
ignored.
*/
type IndexOutputStream struct {
	err error
	out IndexOutput
}

func Stream(out IndexOutput) *IndexOutputStream {
	return &IndexOutputStream{out: out}
}

func (ios *IndexOutputStream) WriteString(s string) *IndexOutputStream {
	if ios.err == nil {
		ios.err = ios.out.WriteString(s)
	}
	return ios
}

func (ios *IndexOutputStream) WriteInt(i int32) *IndexOutputStream {
	if ios.err == nil {
		ios.err = ios.out.WriteInt(i)
	}
	return ios
}

func (ios *IndexOutputStream) WriteByte(b byte) *IndexOutputStream {
	if ios.err == nil {
		ios.err = ios.out.WriteByte(b)
	}
	return ios
}

func (ios *IndexOutputStream) WriteStringStringMap(m map[string]string) *IndexOutputStream {
	if ios.err == nil {
		ios.err = ios.out.WriteStringStringMap(m)
	}
	return ios
}

func (ios *IndexOutputStream) WriteStringSet(m map[string]bool) *IndexOutputStream {
	if ios.err == nil {
		ios.err = ios.out.WriteStringSet(m)
	}
	return ios
}

func (ios *IndexOutputStream) Close() error {
	return ios.err
}

// store/IndexOutput.java

type IndexOutput interface {
	io.Closer
	util.DataOutput
	// Forces any buffered output to be written.
	Flush() error
	// Returns the current position in this file, where the next write will occur.
	FilePointer() int64
	// The numer of bytes in the file.
	Length() (int64, error)
	// Set the file length. By default, this method does nothing (it's
	// optional for a Directory to implement it). But, certain
	// Directory implementations (for example FSDirectory) can use this
	// to inform the unerlying IO system to pre-allocate the file to
	// the specified size. If the length is longer than the current
	// file length, the bytes added to the file are undefined.
	// Otherwise the file is truncated.
	SetLength(length int64) error
}

type IndexOutputImpl struct {
	*util.DataOutputImpl
}

func NewIndexOutput(part util.DataWriter) *IndexOutputImpl {
	return &IndexOutputImpl{util.NewDataOutput(part)}
}

func (out *IndexOutputImpl) SetLength(length int64) error {
	return nil
}

// store/ChecksumIndexOutput.java

/*
Writes bytes through to  a primary IndexOutput, computing checksum.
Note that you cannot use seek().
*/
type ChecksumIndexOutput struct {
	*IndexOutputImpl
	main   IndexOutput
	digest hash.Hash32
}

func NewChecksumIndexOutput(main IndexOutput) *ChecksumIndexOutput {
	ans := &ChecksumIndexOutput{main: main, digest: crc32.NewIEEE()}
	ans.IndexOutputImpl = NewIndexOutput(ans)
	return ans
}

func (out *ChecksumIndexOutput) WriteByte(b byte) error {
	out.digest.Write([]byte{b})
	return out.main.WriteByte(b)
}

func (out *ChecksumIndexOutput) WriteBytes(buf []byte) error {
	out.digest.Write(buf)
	return out.main.WriteBytes(buf)
}

func (out *ChecksumIndexOutput) Flush() error {
	return out.main.Flush()
}

func (out *ChecksumIndexOutput) Close() error {
	return out.main.Close()
}

func (out *ChecksumIndexOutput) FilePointer() int64 {
	return out.main.FilePointer()
}

func (out *ChecksumIndexOutput) FinishCommit() error {
	return out.main.WriteLong(int64(out.digest.Sum32()))
}

func (out *ChecksumIndexOutput) Length() (int64, error) {
	return out.main.Length()
}
