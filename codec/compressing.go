package codec

import (
	"errors"
	"fmt"
)

type CompressionMode interface {
	newCompressor() Compressor
	newDecompressor() Decompressor
}

const (
	COMPRESSION_MODE_FAST = CompressionModeDefaults(1)
)

type CompressionModeDefaults int

func (m CompressionModeDefaults) newCompressor() Compressor {
	panic("not supported yet")
	return nil
}

func (m CompressionModeDefaults) newDecompressor() Decompressor {
	switch int(m) {
	case 1:
		return LZ4_DECOMPRESSOR
	default:
		panic("not implemented yet")
	}
}

type Compressor interface{}

type Decompressor interface {
	decompress(in DataInput, originalLength int, buf []byte) (res []byte, err error)
}

var (
	LZ4_DECOMPRESSOR = LZ4Decompressor(1)
)

type LZ4Decompressor int

func (d LZ4Decompressor) decompress(in DataInput, originalLength int, buf []byte) (res []byte, err error) {
	// assert offset + length <= originalLength
	// add 7 padding bytes, this is not necessary but can help decompression run faster
	res = buf
	if len(buf) < originalLength+7 {
		res = make([]byte, originalLength+7)
	}
	decompressedLength, err := LZ4Decompress(in, len(res), res)
	if decompressedLength > originalLength {
		return nil, errors.New(fmt.Sprintf("Corrupted: lengths mismatch: %v > %v (resource=%v)", decompressedLength, originalLength, in))
	}
	return res, nil
}
