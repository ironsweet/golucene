package codec

import (
	"errors"
	"fmt"
)

type CompressionMode interface {
	NewCompressor() Compressor
	NewDecompressor() Decompressor
}

const (
	COMPRESSION_MODE_FAST = CompressionModeDefaults(1)
)

type CompressionModeDefaults int

func (m CompressionModeDefaults) NewCompressor() Compressor {
	panic("not supported yet")
	return nil
}

func (m CompressionModeDefaults) NewDecompressor() Decompressor {
	switch int(m) {
	case 1:
		return LZ4_DECOMPRESSOR
	default:
		panic("not implemented yet")
	}
}

type Compressor interface{}

// codec/compressing/Decompressor.java

// A decompressor
type Decompressor interface {
	/*
		Decompress 'bytes' that were stored between [offset:offset+length]
		in the original stream from the compressed stream 'in' to 'bytes'.
		The length of the returned bytes (len(buf)) must be equal to 'length'.
		Implementations of this method are free to resize 'bytes' depending
		on their needs.
	*/
	Decompress(in DataInput, originalLength, offset, length int, bytes []byte) (buf []byte, err error)
	Clone() Decompressor
}

var (
	LZ4_DECOMPRESSOR = LZ4Decompressor(1)
)

type LZ4Decompressor int

func (d LZ4Decompressor) Decompress(in DataInput, originalLength, offset, length int, bytes []byte) (res []byte, err error) {
	assert(offset+length <= originalLength)
	// add 7 padding bytes, this is not necessary but can help decompression run faster
	res = bytes
	if len(res) < originalLength+7 {
		// res = make([]byte, util.Oversize(originalLength+7, 1))
		// FIXME util cause dep circle
		res = make([]byte, originalLength+7)
	}
	decompressedLength, err := LZ4Decompress(in, offset+length, res)
	if err != nil {
		return nil, err
	}
	if decompressedLength > originalLength {
		return nil, errors.New(fmt.Sprintf("Corrupted: lengths mismatch: %v > %v (resource=%v)", decompressedLength, originalLength, in))
	}
	return res[offset : offset+length], nil
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (d LZ4Decompressor) Clone() Decompressor {
	return d
}
