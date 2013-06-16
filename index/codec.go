package index

import (
	"fmt"
	"lucene/store"
)

const (
	CODEC_MAGIC = 0x3fd76c17
)

func CheckHeader(in *store.DataInput, codec string, minVersion, maxVersion int) (v int, err error) {
	// Safety to guard against reading a bogus string:
	actualHeader, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	if actualHeader != CODEC_MAGIC {
		return 0, &CorruptIndexError{fmt.Sprintf(
			"codec header mismatch: actual header=%v vs expected header=%v (resource: %v)",
			actualHeader, CODEC_MAGIC, in)}
	}
	return CheckHeaderNoMagic(in, codec, minVersion, maxVersion)
}

func CheckHeaderNoMagic(in *store.DataInput, codec string, minVersion, maxVersion int) (v int, err error) {
	actualCodec, err := in.ReadString()
	if err != nil {
		return 0, err
	}
	if actualCodec != codec {
		return 0, &CorruptIndexError{fmt.Sprintf(
			"codec mismatch: actual codec=%v vs expected codec=%v (resource: %v)", actualCodec, codec, in)}
	}

	actualVersion, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	if actualVersion < minVersion {
		return 0, NewIndexFormatTooOldError(in, actualVersion, minVersion, maxVersion)
	}
	if actualVersion > maxVersion {
		return 0, NewIndexFormatTooNewError(in, actualVersion, minVersion, maxVersion)
	}

	return actualVersion, nil
}
