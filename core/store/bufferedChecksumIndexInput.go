package store

import (
	"fmt"
	"hash"
	"hash/crc32"
)

/*
Simple implementation of ChecksumIndexInput that wraps another input
and delegates calls.
*/
type BufferedChecksumIndexInput struct {
	*ChecksumIndexInputImpl
	main   IndexInput
	digest hash.Hash32
}

func newBufferedChecksumIndexInput(main IndexInput) *BufferedChecksumIndexInput {
	ans := &BufferedChecksumIndexInput{
		main:   main,
		digest: crc32.NewIEEE(),
	}
	ans.ChecksumIndexInputImpl = NewChecksumIndexInput(
		fmt.Sprintf("BufferedChecksumIndexInput(%v)", main), ans)
	return ans
}

func (in *BufferedChecksumIndexInput) ReadByte() (b byte, err error) {
	if b, err = in.main.ReadByte(); err == nil {
		in.digest.Write([]byte{b})
	}
	return
}

func (in *BufferedChecksumIndexInput) ReadBytes(p []byte) (err error) {
	if err = in.main.ReadBytes(p); err == nil {
		in.digest.Write(p)
	}
	return
}

func (in *BufferedChecksumIndexInput) Checksum() int64 {
	return int64(in.digest.Sum32())
}

func (in *BufferedChecksumIndexInput) Close() error {
	return in.main.Close()
}

func (in *BufferedChecksumIndexInput) FilePointer() int64 {
	return in.main.FilePointer()
}

func (in *BufferedChecksumIndexInput) Length() int64 {
	return in.main.Length()
}

func (in *BufferedChecksumIndexInput) Clone() IndexInput {
	panic("not supported")
}

func (in *BufferedChecksumIndexInput) Slice(desc string, offset, length int64) (IndexInput, error) {
	panic("not supported")
}
