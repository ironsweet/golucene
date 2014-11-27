package util

import (
	"sort"
)

/*
Abstract base class for performing write operations of Lucene's
low-level data types.

DataOutput may only be used from one thread, because it is not thread
safe (it keeps internal state like file position).
*/
type DataOutput interface {
	DataWriter
	WriteInt(i int32) error
	WriteVInt(i int32) error
	WriteLong(i int64) error
	WriteVLong(i int64) error
	WriteString(s string) error
	CopyBytes(input DataInput, numBytes int64) error
	WriteStringStringMap(m map[string]string) error
	WriteStringSet(m map[string]bool) error
}

type DataWriter interface {
	WriteByte(b byte) error
	WriteBytes(buf []byte) error
}

type DataOutputImpl struct {
	Writer     DataWriter
	copyBuffer []byte
}

func NewDataOutput(part DataWriter) *DataOutputImpl {
	assert(part != nil)
	return &DataOutputImpl{Writer: part}
}

/*
Writes an int as four bytes.

32-bit unsigned integer written as four bytes, high-order bytes first.
*/
func (out *DataOutputImpl) WriteInt(i int32) error {
	assert(out.Writer != nil)
	err := out.Writer.WriteByte(byte(i >> 24))
	if err == nil {
		err = out.Writer.WriteByte(byte(i >> 16))
		if err == nil {
			err = out.Writer.WriteByte(byte(i >> 8))
			if err == nil {
				err = out.Writer.WriteByte(byte(i))
			}
		}
	}
	return err
}

/*
Writes an int in a variable-length format. Writes between one and
five bytes. Smaller values take fewer bytes. Negative numbers are
supported, by should be avoided.

VByte is a variable-length format. For positive integers, it is
defined where the high-order bit of each byte indicates whether more
bytes remain to be read. The low-order seven bits are appended as
increasingly more significant bits in the resulting integer value.
Thus values from zero to 127 may be stored in a single byte, values
from 128 to 16,383 may be stored in two bytes, and so on.

VByte Encoding Examle

	| Value		| Byte 1		| Byte 2		| Byte 3		|
	| 0				| 00000000	|
	| 1				| 00000001	|
	| 2				| 00000010	|
	| ...			|
	| 127			| 01111111	|
	| 128			| 10000000	| 00000001	|
	| 129			| 10000001	| 00000001	|
	| 130			| 10000010	| 00000001	|
	| ...			|
	| 16,383	| 11111111	| 01111111	|
	| 16,384	| 10000000	| 10000000	| 00000001	|
	| 16,385	| 10000001	| 10000000	| 00000001	|
	| ...			|

This provides compression while still being efficient to decode.
*/
func (out *DataOutputImpl) WriteVInt(i int32) error {
	for (i & ^0x7F) != 0 {
		err := out.Writer.WriteByte(byte(i&0x7F) | 0x80)
		if err != nil {
			return err
		}
		i = int32(uint32(i) >> 7)
	}
	return out.Writer.WriteByte(byte(i))
}

/*
Writes a long as eight bytes.

64-bit unsigned integer written as eight bytes, high-order bytes first.
*/
func (out *DataOutputImpl) WriteLong(i int64) error {
	err := out.WriteInt(int32(i >> 32))
	if err == nil {
		err = out.WriteInt(int32(i))
	}
	return err
}

/*
Writes an long in a variable-length format. Writes between one and
none bytes. Smaller values take fewer bytes. Negative number are not
supported.

The format is described further in WriteVInt().
*/
func (out *DataOutputImpl) WriteVLong(i int64) error {
	assert(i >= 0)
	return out.writeNegativeVLong(i)
}

/* write a potentially negative gLong */
func (out *DataOutputImpl) writeNegativeVLong(i int64) error {
	for (i & ^0x7F) != 0 {
		err := out.Writer.WriteByte(byte((i & 0x7F) | 0x80))
		if err != nil {
			return err
		}
		i = int64(uint64(i) >> 7)
	}
	return out.Writer.WriteByte(byte(i))
}

/*
Writes a string.

Writes strings as UTF-8 encoded bytes. First the length, in bytes, is
written as a VInt, followed by the bytes.
*/
func (out *DataOutputImpl) WriteString(s string) error {
	bytes := []byte(s)
	err := out.WriteVInt(int32(len(bytes)))
	if err == nil {
		err = out.Writer.WriteBytes(bytes)
	}
	return err
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
		err = out.Writer.WriteBytes(out.copyBuffer[0:toCopy])
		if err != nil {
			return err
		}
		left -= int64(toCopy)
	}
	return nil
}

/*
Writes a string map.

First the size is written as an int32, followed by each key-value
pair written as two consecutive strings.
*/
func (out *DataOutputImpl) WriteStringStringMap(m map[string]string) error {
	if m == nil {
		return out.WriteInt(0)
	}
	err := out.WriteInt(int32(len(m)))
	if err != nil {
		return err
	}
	// enforce key order during serialization
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		// for k, v := range m {
		err = out.WriteString(k)
		if err == nil {
			err = out.WriteString(v)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Writes a String set.

First the size is written as an int32, followed by each value written
as a string.
*/
func (out *DataOutputImpl) WriteStringSet(m map[string]bool) error {
	if m == nil {
		return out.WriteInt(0)
	}
	err := out.WriteInt(int32(len(m)))
	if err != nil {
		return err
	}
	for value, _ := range m {
		err = out.WriteString(value)
		if err != nil {
			return err
		}
	}
	return nil
}
