package store

import (
	"hash"
)

/*
Wraps another Checksum with an internal buffer to speed up checksum
calculations.
*/
type BufferedChecksum struct {
	in     hash.Hash32
	buffer []byte
	upto   int
}

/* Create a new BufferedChecksum with DEFAULT_BUFFERSIZE */
func newBufferedChecksum(in hash.Hash32) *BufferedChecksum {
	return newBufferedChecksumWithBuffer(in, 256)
}

/* Create a new BufferedChecksum with the specified bufferSize */
func newBufferedChecksumWithBuffer(in hash.Hash32, bufferSize int) *BufferedChecksum {
	return &BufferedChecksum{in: in, buffer: make([]byte, bufferSize)}
}

func (bc *BufferedChecksum) BlockSize() int {
	return bc.in.BlockSize()
}

func (bc *BufferedChecksum) Reset() {
	bc.upto = 0
	bc.in.Reset()
}

func (bc *BufferedChecksum) Size() int {
	return bc.in.Size()
}

func (bc *BufferedChecksum) Sum(p []byte) []byte {
	return bc.in.Sum(p)
}

func (bc *BufferedChecksum) Sum32() uint32 {
	bc.flush()
	return bc.in.Sum32()
}

func (bc *BufferedChecksum) flush() {
	if bc.upto > 0 {
		bc.in.Write(bc.buffer[:bc.upto])
	}
	bc.upto = 0
}

func (bc *BufferedChecksum) Write(p []byte) (int, error) {
	if len(p) >= len(bc.buffer) {
		bc.flush()
		return bc.in.Write(p)
	}
	if bc.upto+len(p) > len(bc.buffer) {
		bc.flush()
	}
	copy(bc.buffer, p)
	bc.upto += len(p)
	return len(p), nil
}
