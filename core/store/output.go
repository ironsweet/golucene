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
	IndexOutput
	digest hash.Hash32
}

func NewChecksumIndexOutput(main IndexOutput) *ChecksumIndexOutput {
	return &ChecksumIndexOutput{IndexOutput: main, digest: crc32.NewIEEE()}
}

func (out *ChecksumIndexOutput) WriteByte(b byte) error {
	out.digest.Write([]byte{b})
	return out.IndexOutput.WriteByte(b)
}

func (out *ChecksumIndexOutput) WriteBytes(buf []byte) error {
	out.digest.Write(buf)
	return out.IndexOutput.WriteBytes(buf)
}

func (out *ChecksumIndexOutput) Seek(pos int64) {
	panic("not supported")
}
