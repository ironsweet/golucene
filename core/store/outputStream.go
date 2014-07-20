package store

import (
	"hash"
	"hash/crc32"
	"io"
)

// store/OutputStreamIndexOutput.java

/* Implementation class for buffered IndexOutput that writes to a WriterCloser. */
type OutputStreamIndexOutput struct {
	*IndexOutputImpl

	crc hash.Hash32
	os  io.WriteCloser

	bytesWritten int64
}

/* Creates a new OutputStreamIndexOutput with the given buffer size. */
func newOutputStreamIndexOutput(out io.WriteCloser, bufferSize int) *OutputStreamIndexOutput {
	ans := &OutputStreamIndexOutput{
		crc: crc32.NewIEEE(),
		os:  out,
	}
	ans.IndexOutputImpl = NewIndexOutput(ans)
	return ans
}

func (out *OutputStreamIndexOutput) WriteByte(b byte) error {
	out.crc.Write([]byte{b})
	if _, err := out.os.Write([]byte{b}); err != nil {
		return err
	}
	out.bytesWritten++
	return nil
}

func (out *OutputStreamIndexOutput) WriteBytes(p []byte) error {
	out.crc.Write(p)
	if _, err := out.os.Write(p); err != nil {
		return err
	}
	out.bytesWritten += int64(len(p))
	return nil
}

func (out *OutputStreamIndexOutput) Close() error {
	return out.os.Close()
}

func (out *OutputStreamIndexOutput) FilePointer() int64 {
	return out.bytesWritten
}

func (out *OutputStreamIndexOutput) Checksum() int64 {
	return int64(out.crc.Sum32())
}
