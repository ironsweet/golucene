package store

import (
	"github.com/balzaczyy/golucene/core/codec"
)

/*
Clones the provided input, reads all bytes from the file, and calls
CheckFooter().

Note that this method may be slow, as it must process the entire file.
If you just need to extract the checksum value, call retrieveChecksum().
*/
func ChecksumEntireFile(input IndexInput) (hash int64, err error) {
	clone := input.Clone()
	if err = clone.Seek(0); err != nil {
		return 0, err
	}
	in := newBufferedChecksumIndexInput(clone)
	assert(in.FilePointer() == 0)
	if err = in.Seek(in.Length() - codec.FOOTER_LENGTH); err != nil {
		return 0, err
	}
	return codec.CheckFooter(in)
}
