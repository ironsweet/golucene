package store

import (
	"github.com/balzaczyy/golucene/core/util"
	"hash"
	"hash/crc32"
	"io"
)

// store/IndexOutput.java

type IndexOutput interface {
	io.Closer
	util.DataOutput
	// Returns the current position in this file, where the next write will occur.
	FilePointer() int64
	// The numer of bytes in the file.
	Length() (int64, error)
}

type IndexOutputImpl struct {
	*util.DataOutputImpl
}

func NewIndexOutput(part util.DataWriter) *IndexOutputImpl {
	return &IndexOutputImpl{util.NewDataOutput(part)}
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
	return &ChecksumIndexOutput{
		IndexOutputImpl: NewIndexOutput(main),
		main:            main,
		digest:          crc32.NewIEEE(),
	}
}

func (out *ChecksumIndexOutput) WriteByte(b byte) error {
	out.digest.Write([]byte{b})
	return out.main.WriteByte(b)
}

func (out *ChecksumIndexOutput) WriteBytes(buf []byte) error {
	out.digest.Write(buf)
	return out.main.WriteBytes(buf)
}

func (out *ChecksumIndexOutput) Close() error {
	return out.main.Close()
}

func (out *ChecksumIndexOutput) FinishCommit() error {
	return out.main.WriteLong(int64(out.digest.Sum32()))
}
