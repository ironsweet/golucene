package util

import (
	"errors"
)

// store/DataInput.java

/*
Abstract base class for performing read operations of Lucene's low-level
data types.

DataInput may only be used from one thread, because it is not thread safe
(it keeps internal state like file position). To allow multithreaded use,
every DataInput instance must be cloned before  used in another thread.
Subclases must therefore implement Clone(), returning a new DataInput which
operates on the same underlying resource, but positioned independently.
*/
type DataInput interface {
	ReadByte() (b byte, err error)
	ReadBytes(buf []byte) error
	ReadShort() (n int16, err error)
	ReadInt() (n int32, err error)
	ReadVInt() (n int32, err error)
	ReadLong() (n int64, err error)
	ReadVLong() (n int64, err error)
	ReadString() (s string, err error)
	ReadStringStringMap() (m map[string]string, err error)
	ReadStringSet() (m map[string]bool, err error)
}

type DataReader interface {
	/* Reads and returns a single byte.	*/
	ReadByte() (b byte, err error)
	/* Reads a specified number of bytes into an array */
	ReadBytes(buf []byte) error
	/** Reads a specified number of bytes into an array at the
	 * specified offset with control over whether the read
	 * should be buffered (callers who have their own buffer
	 * should pass in "false" for useBuffer).  Currently only
	 * {@link BufferedIndexInput} respects this parameter.
	 */
	// ReadBytesBuffered(buf []byte, useBuffer bool) error
}

const SKIP_BUFFER_SIZE = 1024

type DataInputImpl struct {
	Reader DataReader
	// This buffer is used to skip over bytes with the default
	// implementation of skipBytes. The reason why we need to use an
	// instance member instead of sharing a single instance across
	// routines is that some delegating implementations of DataInput
	// might want to reuse the provided buffer in order to e.g. update
	// the checksum. If we shared the same buffer across routines, then
	// another routine might update the buffer while the checksum is
	// being computed, making it invalid. See LUCENE-5583 for more
	// information.
	skipBuffer []byte
}

func NewDataInput(spi DataReader) *DataInputImpl {
	return &DataInputImpl{Reader: spi}
}

func (in *DataInputImpl) ReadBytesBuffered(buf []byte, useBuffer bool) error {
	return in.Reader.ReadBytes(buf)
}

func (in *DataInputImpl) ReadShort() (n int16, err error) {
	if b1, err := in.Reader.ReadByte(); err == nil {
		if b2, err := in.Reader.ReadByte(); err == nil {
			return (int16(b1) << 8) | int16(b2), nil
		}
	}
	return 0, err
}

func (in *DataInputImpl) ReadInt() (n int32, err error) {
	if b1, err := in.Reader.ReadByte(); err == nil {
		if b2, err := in.Reader.ReadByte(); err == nil {
			if b3, err := in.Reader.ReadByte(); err == nil {
				if b4, err := in.Reader.ReadByte(); err == nil {
					return (int32(b1) << 24) | (int32(b2) << 16) | (int32(b3) << 8) | int32(b4), nil
				}
			}
		}
	}
	return 0, err
}

func (in *DataInputImpl) ReadVInt() (n int32, err error) {
	if b, err := in.Reader.ReadByte(); err == nil {
		n = int32(b) & 0x7F
		if b < 128 {
			return n, nil
		}
		if b, err = in.Reader.ReadByte(); err == nil {
			n |= (int32(b) & 0x7F) << 7
			if b < 128 {
				return n, nil
			}
			if b, err = in.Reader.ReadByte(); err == nil {
				n |= (int32(b) & 0x7F) << 14
				if b < 128 {
					return n, nil
				}
				if b, err = in.Reader.ReadByte(); err == nil {
					n |= (int32(b) & 0x7F) << 21
					if b < 128 {
						return n, nil
					}
					if b, err = in.Reader.ReadByte(); err == nil {
						// Warning: the next ands use 0x0F / 0xF0 - beware copy/paste errors:
						n |= (int32(b) & 0x0F) << 28

						if int32(b)&0xF0 == 0 {
							return n, nil
						}
						return 0, errors.New("Invalid vInt detected (too many bits)")
					}
				}
			}
		}
	}
	return 0, err
}

func (in *DataInputImpl) ReadLong() (n int64, err error) {
	d1, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	d2, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	return (int64(d1) << 32) | int64(d2)&0xFFFFFFFF, nil
}

func (in *DataInputImpl) ReadVLong() (int64, error) {
	return in.readVLong(false)
}

func (in *DataInputImpl) readVLong(allowNegative bool) (n int64, err error) {
	if b, err := in.Reader.ReadByte(); err == nil {
		n = int64(b & 0x7F)
		if b < 128 {
			return n, nil
		}
		if b, err = in.Reader.ReadByte(); err == nil {
			n |= (int64(b&0x7F) << 7)
			if b < 128 {
				return n, nil
			}
			if b, err = in.Reader.ReadByte(); err == nil {
				n |= (int64(b&0x7F) << 14)
				if b < 128 {
					return n, nil
				}
				if b, err = in.Reader.ReadByte(); err == nil {
					n |= (int64(b&0x7F) << 21)
					if b < 128 {
						return n, nil
					}
					if b, err = in.Reader.ReadByte(); err == nil {
						n |= (int64(b&0x7F) << 28)
						if b < 128 {
							return n, nil
						}
						if b, err = in.Reader.ReadByte(); err == nil {
							n |= (int64(b&0x7F) << 35)
							if b < 128 {
								return n, nil
							}
							if b, err = in.Reader.ReadByte(); err == nil {
								n |= (int64(b&0x7F) << 42)
								if b < 128 {
									return n, nil
								}
								if b, err = in.Reader.ReadByte(); err == nil {
									n |= (int64(b&0x7F) << 49)
									if b < 128 {
										return n, nil
									}
									if b, err = in.Reader.ReadByte(); err == nil {
										n |= (int64(b&0x7F) << 56)
										if b < 128 {
											return n, nil
										}
										if allowNegative {
											panic("niy")
										}
										return 0, errors.New("Invalid vLong detected (negative values disallowed)")
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return 0, err
}

func (in *DataInputImpl) ReadString() (s string, err error) {
	length, err := in.ReadVInt()
	if err != nil {
		return "", err
	}
	bytes := make([]byte, length)
	err = in.Reader.ReadBytes(bytes)
	return string(bytes), nil
}

func (in *DataInputImpl) ReadStringStringMap() (m map[string]string, err error) {
	count, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	m = make(map[string]string)
	for i := int32(0); i < count; i++ {
		key, err := in.ReadString()
		if err != nil {
			return nil, err
		}
		value, err := in.ReadString()
		if err != nil {
			return nil, err
		}
		m[key] = value
	}
	return m, nil
}

func (in *DataInputImpl) ReadStringSet() (s map[string]bool, err error) {
	count, err := in.ReadInt()
	if err != nil {
		return nil, err
	}
	s = make(map[string]bool)
	for i := int32(0); i < count; i++ {
		key, err := in.ReadString()
		if err != nil {
			return nil, err
		}
		s[key] = true
	}
	return s, nil
}

/*
Skip over numBytes bytes. The contract on this method is that it
should have the same behavior as reading the same number of bytes
into a buffer and discarding its content. Negative values of numBytes
are not supported.
*/
func (in *DataInputImpl) SkipBytes(numBytes int64) (err error) {
	assert2(numBytes >= 0, "numBytes must be >= 0, got %v", numBytes)
	if in.skipBuffer == nil {
		in.skipBuffer = make([]byte, SKIP_BUFFER_SIZE)
	}
	assert(len(in.skipBuffer) == SKIP_BUFFER_SIZE)
	var step int
	for skipped := int64(0); skipped < numBytes; {
		step = int(numBytes - skipped)
		if SKIP_BUFFER_SIZE < step {
			step = SKIP_BUFFER_SIZE
		}
		if err = in.ReadBytesBuffered(in.skipBuffer[:step], false); err != nil {
			return
		}
		skipped += int64(step)
	}
	return nil
}
