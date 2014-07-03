package store

import ()

// store/OutputStreamIndexOutput.java

/* Implementation class for buffered IndexOutput that writes to a WriterCloser. */
type OutputStreamIndexOutput struct {
	*IndexOutputImpl
}

func (out *OutputStreamIndexOutput) WriteByte(b byte) error {
	panic("not implemented yet")
}

func (out *OutputStreamIndexOutput) WriteBytes(p []byte) error {
	panic("not implemented yet")
}

func (out *OutputStreamIndexOutput) FilePointer() int64 {
	panic("not implemented yet")
}
