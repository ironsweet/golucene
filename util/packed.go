package util

import (
	"fmt"
	"math"
)

const (
	PACKED_VERSION_START        = 0
	PACKED_VERSION_BYTE_ALIGNED = 1
	PACKED_VERSION_CURRENT      = PACKED_VERSION_BYTE_ALIGNED
)

func CheckVersion(version int) {
	if version < PACKED_VERSION_START {
		panic(fmt.Sprintf("Version is too old, should be at least %v (got %v)", PACKED_VERSION_START, version))
	} else if version > PACKED_VERSION_CURRENT {
		panic(fmt.Sprintf("Version is too new, should be at most %v (got %v)", PACKED_VERSION_CURRENT, version))
	}
}

type PackedFormat int

const (
	PACKED              = 0
	PACKED_SINGLE_BLOCK = 1
)

func (f PackedFormat) ByteCount(packedIntsVersion, valueCount, bitsPerValue int) int64 {
	switch int(f) {
	case PACKED:
		if packedIntsVersion < PACKED_VERSION_BYTE_ALIGNED {
			return 8 * int64(math.Ceil(float64(valueCount)*float64(bitsPerValue)/64))
		}
		return int64(math.Ceil(float64(valueCount) * float64(bitsPerValue) / 8))
	}
	// assert bitsPerValue >= 0 && bitsPerValue <= 64
	// assume long-aligned
	return 8 * int64(f.longCount(packedIntsVersion, valueCount, bitsPerValue))
}

func (f PackedFormat) longCount(packedIntsVersion, valueCount, bitsPerValue int) int {
	switch int(f) {
	case PACKED_SINGLE_BLOCK:
		valuesPerBlock := 64 / bitsPerValue
		return int(math.Ceil(float64(valueCount) / float64(valuesPerBlock)))
	}
	// assert bitsPerValue >= 0 && bitsPerValue <= 64
	ans := f.ByteCount(packedIntsVersion, valueCount, bitsPerValue)
	// assert ans < 8 * math.MaxInt32()
	if ans%8 == 0 {
		return int(ans / 8)
	}
	return int(ans/8) + 1
}

type PackedIntsEncoder interface {
}

type PackedIntsDecoder interface {
	ByteValueCount() uint32
}

func GetPackedIntsEncoder(format PackedFormat, version, bitsPerValue int) PackedIntsEncoder {
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}

func GetPackedIntsDecoder(format PackedFormat, version, bitsPerValue int) PackedIntsDecoder {
	CheckVersion(version)
	return newBulkOperation(format, bitsPerValue)
}
