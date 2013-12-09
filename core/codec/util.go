package codec

import (
	"errors"
	"fmt"
)

const (
	CODEC_MAGIC = 0x3fd76c17
)

func HeaderLength(codec string) int {
	return 9 + len(codec)
}

type DataInput interface {
	ReadInt() (n int32, err error)
	ReadString() (s string, err error)
	ReadByte() (b byte, err error)
	ReadBytes(buf []byte) error
}

func CheckHeader(in DataInput, codec string, minVersion, maxVersion int32) (v int32, err error) {
	// Safety to guard against reading a bogus string:
	actualHeader, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	if actualHeader != CODEC_MAGIC {
		return 0, errors.New(fmt.Sprintf(
			"codec header mismatch: actual header=%v vs expected header=%v (resource: %v)",
			actualHeader, CODEC_MAGIC, in))
	}
	return CheckHeaderNoMagic(in, codec, minVersion, maxVersion)
}

func CheckHeaderNoMagic(in DataInput, codec string, minVersion, maxVersion int32) (v int32, err error) {
	actualCodec, err := in.ReadString()
	if err != nil {
		return 0, err
	}
	if actualCodec != codec {
		return 0, errors.New(fmt.Sprintf(
			"codec mismatch: actual codec=%v vs expected codec=%v (resource: %v)", actualCodec, codec, in))
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

func NewIndexFormatTooNewError(in DataInput, version, minVersion, maxVersion int32) error {
	return errors.New(fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v)",
		in, version, minVersion, maxVersion))
}

func NewIndexFormatTooOldError(in DataInput, version, minVersion, maxVersion int32) error {
	return errors.New(fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v). This version of Lucene only supports indexes created with release 3.0 and later.",
		in, version, minVersion, maxVersion))
}
