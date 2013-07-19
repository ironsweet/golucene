package util

import (
	"errors"
)

type DataInput struct {
	/* Reads and returns a single byte.	*/
	ReadByte func() (b byte, err error)
	/* Reads a specified number of bytes into an array */
	ReadBytes func(buf []byte) error
	/** Reads a specified number of bytes into an array at the
	 * specified offset with control over whether the read
	 * should be buffered (callers who have their own buffer
	 * should pass in "false" for useBuffer).  Currently only
	 * {@link BufferedIndexInput} respects this parameter.
	 */
	ReadBytesBufferd func(buf []byte, useBuffer bool) error
}

func NewDataInput() DataInput {
	return DataInput{}
}

func (in *DataInput) ReadShort() (n int16, err error) {
	if b1, err := in.ReadByte(); err == nil {
		if b2, err := in.ReadByte(); err == nil {
			return (int16(b1) << 8) | int16(b2), nil
		}
	}
	return 0, err
}

func (in *DataInput) ReadInt() (n int32, err error) {
	if b1, err := in.ReadByte(); err == nil {
		if b2, err := in.ReadByte(); err == nil {
			if b3, err := in.ReadByte(); err == nil {
				if b4, err := in.ReadByte(); err == nil {
					return (int32(b1) << 24) | (int32(b2) << 16) | (int32(b3) << 8) | int32(b4), nil
				}
			}
		}
	}
	return 0, err
}

func (in *DataInput) ReadVInt() (n int32, err error) {
	if b, err := in.ReadByte(); err == nil {
		n = int32(b) & 0x7F
		if b >= 0 {
			return n, nil
		}
		if b, err = in.ReadByte(); err == nil {
			n |= (int32(b) & 0x7F) << 7
			if b >= 0 {
				return n, nil
			}
			if b, err = in.ReadByte(); err == nil {
				n |= (int32(b) & 0x7F) << 14
				if b >= 0 {
					return n, nil
				}
				if b, err = in.ReadByte(); err == nil {
					n |= (int32(b) & 0x7F) << 21
					if b >= 0 {
						return n, nil
					}
					if b, err = in.ReadByte(); err == nil {
						// Warning: the next ands use 0x0F / 0xF0 - beware copy/paste errors:
						n |= (int32(b) & 0x0F) << 28
						if int32(b)&0x0F == 0 {
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

func (in *DataInput) ReadLong() (n int64, err error) {
	d1, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	d2, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	return (int64(d1) << 32) | (int64(d2) & 0xFFFFFFFF), nil
}

func (in *DataInput) ReadVLong() (n int64, err error) {
	if b, err := in.ReadByte(); err == nil {
		n = int64(b) & 0x7F
		if b >= 0 {
			return n, nil
		}
		if b, err = in.ReadByte(); err == nil {
			n |= (int64(b) & 0x7F) << 7
			if b >= 0 {
				return n, nil
			}
			if b, err = in.ReadByte(); err == nil {
				n |= (int64(b) & 0x7F) << 14
				if b >= 0 {
					return n, nil
				}
				if b, err = in.ReadByte(); err == nil {
					n |= (int64(b) & 0x7F) << 21
					if b >= 0 {
						return n, nil
					}
					if b, err = in.ReadByte(); err == nil {
						n |= (int64(b) & 0x7F) << 28
						if b >= 0 {
							return n, nil
						}
						if b, err = in.ReadByte(); err == nil {
							n |= (int64(b) & 0x7F) << 35
							if b >= 0 {
								return n, nil
							}
							if b, err = in.ReadByte(); err == nil {
								n |= (int64(b) & 0x7F) << 42
								if b >= 0 {
									return n, nil
								}
								if b, err = in.ReadByte(); err == nil {
									n |= (int64(b) & 0x7F) << 49
									if b >= 0 {
										return n, nil
									}
									if b, err = in.ReadByte(); err == nil {
										n |= (int64(b) & 0x7F) << 56
										if b >= 0 {
											return n, nil
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

func (in *DataInput) ReadString() (s string, err error) {
	length, err := in.ReadVInt()
	if err != nil {
		return "", err
	}
	bytes := make([]byte, length)
	in.ReadBytes(bytes)
	return string(bytes), nil
}

func (in *DataInput) ReadStringStringMap() (m map[string]string, err error) {
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

func (in *DataInput) ReadStringSet() (s map[string]bool, err error) {
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
