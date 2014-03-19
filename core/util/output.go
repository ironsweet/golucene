package util

import ()

/*
Abstract base class for performing write operations of Lucene's
low-level data types.

DataOutput may only be used from one thread, because it is not thread
safe (it keeps internal state like file position).
*/
type DataOutput interface {
	DataWriter
	CopyBytes(input DataInput, numBytes int64) error
}

type DataWriter interface {
	WriteByte(b byte) error
	WriteBytes(buf []byte) error
}

type DataOutputImpl struct {
	DataWriter
	copyBuffer []byte
}

func NewDataOutput(part DataWriter) *DataOutputImpl {
	return &DataOutputImpl{DataWriter: part}
}

const DATA_OUTPUT_COPY_BUFFER_SIZE = 16384

func (out *DataOutputImpl) CopyBytes(input DataInput, numBytes int64) error {
	assert(numBytes >= 0)
	left := numBytes
	if out.copyBuffer == nil {
		out.copyBuffer = make([]byte, DATA_OUTPUT_COPY_BUFFER_SIZE)
	}
	for left > 0 {
		var toCopy int32
		if left > DATA_OUTPUT_COPY_BUFFER_SIZE {
			toCopy = DATA_OUTPUT_COPY_BUFFER_SIZE
		} else {
			toCopy = int32(left)
		}
		err := input.ReadBytes(out.copyBuffer[0:toCopy])
		if err != nil {
			return err
		}
		err = out.WriteBytes(out.copyBuffer[0:toCopy])
		if err != nil {
			return err
		}
		left -= int64(toCopy)
	}
	return nil
}
