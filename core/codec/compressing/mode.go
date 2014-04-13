package compressing

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
	switch int(m) {
	case 1:
		var ht = new(LZ4HashTable)
		return Compressor(func(bytes []byte, out DataOutput) error {
			return LZ4Compress(bytes, out, ht)
		})
	default:
		panic("not implemented yet")
	}
}

func (m CompressionModeDefaults) NewDecompressor() Decompressor {
	switch int(m) {
	case 1:
		return LZ4Decompressor
	default:
		panic("not implemented yet")
	}
}

// codec/compressing/Compressor.java

/*
Compress bytes into out. It is the responsibility of the compressor
to add all necessary information so that a Decompressor will know
when to stop decompressing bytes from the stream.
*/
type Compressor func(bytes []byte, out DataOutput) error

// codec/compressing/Decompressor.java

/*
Decompress 'bytes' that were stored between [offset:offset+length]
in the original stream from the compressed stream 'in' to 'bytes'.
The length of the returned bytes (len(buf)) must be equal to 'length'.
Implementations of this method are free to resize 'bytes' depending
on their needs.
*/
type Decompressor func(in DataInput, originalLength, offset, length int, bytes []byte) (buf []byte, err error)

func LZ4Decompressor(in DataInput, originalLength, offset, length int, bytes []byte) (res []byte, err error) {
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
