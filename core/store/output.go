package store

import (
	"github.com/balzaczyy/golucene/core/util"
	"hash"
	"hash/crc64"
	"io"
)

// store/IndexOutput.java

type IndexOutput interface {
	io.Closer
	util.DataOutput
	// The numer of bytes in the file.
	Length() int64
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
	digest hash.Hash64
}

func NewChecksumIndexOutput(main IndexOutput) *ChecksumIndexOutput {
	return &ChecksumIndexOutput{IndexOutput: main, digest: crc64.New(crc64.MakeTable(crc64.ISO))}
}

func (out *ChecksumIndexOutput) WriteByte(b byte) error {
	out.digest.Write([]byte{b})
	return out.IndexOutput.WriteByte(b)
}

func (out *ChecksumIndexOutput) WriteBytes(buf []byte) error {
	out.digest.Write(buf)
	return out.IndexOutput.WriteBytes(buf)
}

func (out *ChecksumIndexOutput) Close() error {
	err := out.IndexOutput.WriteLong(int64(out.digest.Sum64()))
	if err == nil {
		err = out.IndexOutput.Close()
	}
	return err
}

// func (out *ChecksumIndexOutput) Seek(pos int64) {
// 	panic("not supported")
// }
