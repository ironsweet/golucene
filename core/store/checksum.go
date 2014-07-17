package store

import (
	// "hash"
	// "hash/crc32"
	"github.com/balzaczyy/golucene/core/util"
)

// store/ChecksumIndexInput.java

/*
Extension of IndexInput, computing checksum as it goes.
Callers can retrieve the checksum via Checksum().
*/
type ChecksumIndexInput interface {
	IndexInput
	Checksum() int64
}

type ChecksumIndexInputImpl struct {
	*IndexInputImpl
}

func NewChecksumIndexInput(desc string, spi util.DataReader) *ChecksumIndexInputImpl {
	return &ChecksumIndexInputImpl{
		NewIndexInputImpl(desc, spi),
	}
}

func (in *ChecksumIndexInputImpl) Seek(pos int64) error {
	panic("not implemented yet")
}
