package codec

import (
	"errors"
	"fmt"
)

// codecs/CodecUtil.java

/* Constant to identify the start of a codec header */
const CODEC_MAGIC = 0x3fd76c17

type DataOutput interface {
	WriteInt(n int32) error
	WriteString(s string) error
}

/*
Writes a codc header, which records both a string to identify the
file and a version number. This header can be parsed and validated
with CheckHeader().

CodecHeader --> Magic,CodecName,Version
	Magic --> uint32. This identifies the start of the header. It is
	always CODEC_MAGIC.
	CodecName --> string. This is a string to identify this file.
	Version --> uint32. Records the version of the file.

Note that the length of a codec header depends only upon the name of
the codec, so this length can be computed at any time with
HeaderLength().
*/
func WriteHeader(out DataOutput, codec string, version int) error {
	assert(out != nil)
	bytes := []byte(codec)
	assert2(len(bytes) == len(codec) && len(bytes) < 128,
		"codec must be simple ASCII, less than 128 characters in length [got %v]", codec)
	err := out.WriteInt(CODEC_MAGIC)
	if err == nil {
		err = out.WriteString(codec)
		if err == nil {
			err = out.WriteInt(int32(version))
		}
	}
	return err
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

/* Computes the length of a codec header */
func HeaderLength(codec string) int {
	return 9 + len(codec)
}

type DataInput interface {
	ReadInt() (int32, error)
	ReadString() (string, error)
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
