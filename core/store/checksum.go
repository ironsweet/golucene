package store

import (
	// "hash"
	// "hash/crc32"
	"github.com/balzaczyy/golucene/core/util"
)

// store/ChecksumIndexInput.java

/*
Extension of IndexInput, computing checksum as it goes.
Callers can retrieve the checksum via Checksum().
*/
type ChecksumIndexInput interface {
	IndexInput
	Checksum() int64
}

type ChecksumIndexInputImpl struct {
	*IndexInputImpl
}

func NewChecksumIndexInput(desc string, spi util.DataReader) *ChecksumIndexInputImpl {
	return &ChecksumIndexInputImpl{
		NewIndexInputImpl(desc, spi),
	}
}

func (in *ChecksumIndexInputImpl) Seek(pos int64) error {
	panic("not implemented yet")
}

// store/ChecksumIndexOutput.java

/*
Writes bytes through to  a primary IndexOutput, computing checksum.
Note that you cannot use seek().
*/
// type ChecksumIndexOutput struct {
// 	*IndexOutputImpl
// 	main   IndexOutput
// 	digest hash.Hash32
// }

// func NewChecksumIndexOutput(main IndexOutput) *ChecksumIndexOutput {
// 	ans := &ChecksumIndexOutput{main: main, digest: crc32.NewIEEE()}
// 	ans.IndexOutputImpl = NewIndexOutput(ans)
// 	return ans
// }

// func (out *ChecksumIndexOutput) WriteByte(b byte) error {
// 	out.digest.Write([]byte{b})
// 	return out.main.WriteByte(b)
// }

// func (out *ChecksumIndexOutput) WriteBytes(buf []byte) error {
// 	out.digest.Write(buf)
// 	return out.main.WriteBytes(buf)
// }

// func (out *ChecksumIndexOutput) Flush() error {
// 	return out.main.Flush()
// }

// func (out *ChecksumIndexOutput) Close() error {
// 	return out.main.Close()
// }

// func (out *ChecksumIndexOutput) FilePointer() int64 {
// 	return out.main.FilePointer()
// }

// func (out *ChecksumIndexOutput) FinishCommit() error {
// 	return out.main.WriteLong(int64(out.digest.Sum32()))
// }

// func (out *ChecksumIndexOutput) Length() (int64, error) {
// 	return out.main.Length()
// }
