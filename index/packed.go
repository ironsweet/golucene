package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
)

func newPackedHeaderNoHeader(in *store.DataInput, format util.PackedFormat, version, valueCount int32, bitsPerValue uint32) (r util.PackedIntsReader, err error) {
	util.CheckVersion(version)
	switch format {
	case util.PACKED_SINGLE_BLOCK:
		return newPacked64SingleBlock(in, valueCount, bitsPerValue)
	case util.PACKED:
		switch bitsPerValue {
		case 8:
			return newDirect8(version, in, valueCount)
		case 16:
			return newDirect16(version, in, valueCount)
		case 32:
			return newDirect32(version, in, valueCount)
		case 64:
			return newDirect64(version, in, valueCount)
		case 24:
			if valueCount <= PACKED8_THREE_BLOCKS_MAX_SIZE {
				return newPacked8ThreeBlocks(version, in, valueCount)
			}
		case 48:
			if valueCount <= PACKED16_THREE_BLOCKS_MAX_SIZE {
				return newPacked16ThreeBlocks(version, in, valueCount)
			}
		}
		return newPacked64(version, in, valueCount, bitsPerValue)
	default:
		panic(fmt.Sprintf("Unknown Writer foramt: %v", format))
	}
}

func newPackedReader(in *store.DataInput) (r util.PackedIntsReader, err error) {
	version, err := store.CheckHeader(in, util.PACKED_CODEC_NAME, util.PACKED_VERSION_START, util.PACKED_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	bitsPerValue, err := in.ReadVInt()
	if err != nil {
		return nil, err
	}
	// assert bitsPerValue > 0 && bitsPerValue <= 64
	valueCount, err := in.ReadVInt()
	if err != nil {
		return nil, err
	}
	id, err := in.ReadVInt()
	if err != nil {
		return nil, err
	}
	format := util.PackedFormat(id)
	return newPackedReaderNoHeader(in, format, version, valueCount, bitsPerValue)
}
