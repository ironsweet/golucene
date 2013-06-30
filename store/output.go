package store

type DataOutput struct {
	WriteByte  func(b byte) error
	WriteBytes func(buf []byte) error

	copyBuffer []byte
}

const DATA_OUTPUT_COPY_BUFFER_SIZE = 16384

func (out *DataOutput) CopyBytes(input *DataInput, numBytes int64) error {
	// assert numBytes >= 0
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
		_, err := input.ReadBytes(out.copyBuffer[0:toCopy])
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
