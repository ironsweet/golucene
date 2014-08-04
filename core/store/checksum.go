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

type ChecksumIndexInputImplSPI interface {
	util.DataReader
	FilePointer() int64
}

type ChecksumIndexInputImpl struct {
	*IndexInputImpl
	spi ChecksumIndexInputImplSPI
}

func NewChecksumIndexInput(desc string, spi ChecksumIndexInputImplSPI) *ChecksumIndexInputImpl {
	return &ChecksumIndexInputImpl{NewIndexInputImpl(desc, spi), spi}
}

func (in *ChecksumIndexInputImpl) Seek(pos int64) error {
	skip := pos - in.spi.FilePointer()
	assert2(skip >= 0, "ChecksumIndexInput cannot seek backwards")
	return in.SkipBytes(skip)
}
