package codec

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

// codecs/CodecUtil.java

/* Constant to identify the start of a codec header. */
const CODEC_MAGIC = 0x3fd76c17

/* Constant to identify the start of a codec footer. */
const FOOTER_MAGIC = ^CODEC_MAGIC

const FOOTER_LENGTH = 16

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

type IndexOutput interface {
	WriteInt(n int32) error
	WriteLong(n int64) error
	Checksum() int64
}

/*
Writes a codec footer, which records both a checksum algorithm ID and
a checksum. This footer can be parsed and validated with CheckFooter().

CodecFooter --> Magic,AlgorithmID,Checksum
	- Magic --> uint32. This identifies the start of the footer. It is
		always FOOTER_MAGIC.
	- AlgorithmID --> uing32. This indicates the checksum algorithm
		used. Currently this is always 0, for zlib-crc32.
	- Checksum --> uint64. The actual checksum value for all previous
		bytes in the stream, including the bytes from Magic and AlgorithmID.
*/
func WriteFooter(out IndexOutput) (err error) {
	if err = out.WriteInt(FOOTER_MAGIC); err == nil {
		if err = out.WriteInt(0); err == nil {
			err = out.WriteLong(out.Checksum())
		}
	}
	return
}

type ChecksumIndexInput interface {
	IndexInput
	Checksum() int64
}

/* Validates the codec footer previously written by WriteFooter(). */
func CheckFooter(in ChecksumIndexInput) (cs int64, err error) {
	if err = validateFooter(in); err == nil {
		cs = in.Checksum()
		var cs2 int64
		if cs2, err = in.ReadLong(); err == nil {
			if cs != cs2 {
				return 0, errors.New(fmt.Sprintf(
					"checksum failed (hardware problem?): expected=%v actual=%v (resource=%v)",
					util.ItoHex(cs2), util.ItoHex(cs), in))
			}
			if in.FilePointer() != in.Length() {
				return 0, errors.New(fmt.Sprintf(
					"did not read all bytes from file: read %v vs size %v (resource: %v)",
					in.FilePointer(), in.Length(), in))
			}
		}
	}
	return
}

/* Returns (but does not validate) the checksum previously written by CheckFooter. */
func RetrieveChecksum(in IndexInput) (int64, error) {
	var err error
	if err = in.Seek(in.Length() - FOOTER_LENGTH); err != nil {
		return 0, err
	}
	if err = validateFooter(in); err != nil {
		return 0, err
	}
	return in.ReadLong()
}

func validateFooter(in IndexInput) error {
	magic, err := in.ReadInt()
	if err != nil {
		return err
	}
	if magic != FOOTER_MAGIC {
		return errors.New(fmt.Sprintf(
			"codec footer mismatch: actual footer=%v vs expected footer=%v (resource: %v)",
			magic, FOOTER_MAGIC, in))
	}

	algorithmId, err := in.ReadInt()
	if err != nil {
		return err
	}
	if algorithmId != 0 {
		return errors.New(fmt.Sprintf(
			"codec footer mismatch: unknown algorithmID: %v",
			algorithmId))
	}
	return nil
}

type IndexInput interface {
	FilePointer() int64
	Seek(int64) error
	Length() int64
	ReadInt() (int32, error)
	ReadLong() (int64, error)
}

/* Checks that the stream is positioned at the end, and returns error if it is not. */
func CheckEOF(in IndexInput) error {
	if in.FilePointer() != in.Length() {
		return errors.New(fmt.Sprintf(
			"did not read all bytes from file: read %v vs size %v (resources: %v)",
			in.FilePointer(), in.Length(), in))
	}
	return nil
}
