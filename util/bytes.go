package util

import ()

type BytesStore struct {
	*DataOutput
	blocks    [][]byte
	blockSize uint32
	blockBits uint32
	blockMask uint32
	current   []byte
	nextWrite uint32
}

func newBytesStore() *BytesStore {
	self := &BytesStore{}
	self.DataOutput = &DataOutput{
		WriteByte: func(b byte) error {
			if self.nextWrite == self.blockSize {
				self.current = make([]byte, self.blockSize)
				self.blocks = append(self.blocks, self.current)
				self.nextWrite = 0
			}
			self.current[self.nextWrite] = b
			self.nextWrite++
			return nil
		},
		WriteBytes: func(buf []byte) error {
			var offset uint32 = 0
			length := uint32(len(buf))
			for length > 0 {
				chunk := self.blockSize - self.nextWrite
				if length <= chunk {
					copy(self.current[self.nextWrite:], buf[offset:offset+length])
					self.nextWrite += length
					break
				} else {
					if chunk > 0 {
						copy(self.current[self.nextWrite:], buf[offset:offset+chunk])
						offset += chunk
						length -= chunk
					}
					self.current = make([]byte, self.blockSize)
					self.blocks = append(self.blocks, self.current)
					self.nextWrite = 0
				}
			}
			return nil
		}}
	return self
}

func newBytesStoreFromBits(blockBits uint32) *BytesStore {
	blockSize := uint32(1) << blockBits
	self := newBytesStore()
	self.blockBits = blockBits
	self.blockSize = blockSize
	self.blockMask = blockSize - 1
	self.nextWrite = blockSize
	return self
}

func newBytesStoreFromInput(in DataInput, numBytes int64, maxBlockSize uint32) (bs *BytesStore, err error) {
	var blockSize uint32 = 2
	var blockBits uint32 = 1
	for int64(blockSize) < numBytes && blockSize < maxBlockSize {
		blockSize *= 2
		blockBits++
	}
	self := newBytesStore()
	self.blockBits = blockBits
	self.blockSize = blockSize
	self.blockMask = blockSize - 1
	left := numBytes
	for left > 0 {
		chunk := blockSize
		if left < int64(chunk) {
			chunk = uint32(left)
		}
		block := make([]byte, chunk)
		err = in.ReadBytes(block)
		if err != nil {
			return nil, err
		}
		self.blocks = append(self.blocks, block)
		left -= int64(chunk)
	}
	// So .getPosition still works
	self.nextWrite = uint32(len(self.blocks[len(self.blocks)-1]))
	return self, nil
}

type BytesStoreForwardReader struct {
	*BytesReader
	owner      *BytesStore
	current    []byte
	nextBuffer uint32
	nextRead   uint32
}

func (r *BytesStoreForwardReader) ReadByte() (b byte, err error) {
	if r.nextRead == r.owner.blockSize {
		r.current = r.owner.blocks[r.nextBuffer]
		r.nextBuffer++
		r.nextRead = 0
	}
	b = r.current[r.nextRead]
	r.nextRead++
	return b, nil
}

func (r *BytesStoreForwardReader) ReadBytes(buf []byte) error {
	var offset uint32 = 0
	length := uint32(len(buf))
	for length > 0 {
		chunkLeft := r.owner.blockSize - r.nextRead
		if length <= chunkLeft {
			copy(buf[offset:], r.current[r.nextRead:r.nextRead+length])
			r.nextRead += length
			break
		} else {
			if chunkLeft > 0 {
				copy(buf[offset:], r.current[r.nextRead:r.nextRead+chunkLeft])
				offset += chunkLeft
				length -= chunkLeft
			}
			r.current = r.owner.blocks[r.nextBuffer]
			r.nextBuffer++
			r.nextRead = 0
		}
	}
	return nil
}

func (bs *BytesStore) forwardReader() *BytesReader {
	if len(bs.blocks) == 1 {
		return newForwardBytesReader(bs.blocks[0])
	}
	self := &BytesStoreForwardReader{owner: bs, nextRead: bs.blockSize}
	self.BytesReader = &BytesReader{
		skipBytes: func(count int32) {
			self.setPosition(self.getPosition() + int64(count))
		}, getPosition: func() int64 {
			return (int64(self.nextBuffer)-1)*int64(bs.blockSize) + int64(self.nextRead)
		}, setPosition: func(pos int64) {
			bufferIndex := pos >> bs.blockBits
			self.nextBuffer = uint32(bufferIndex + 1)
			self.current = bs.blocks[bufferIndex]
			self.nextRead = uint32(pos) & bs.blockMask
			// assert self.getPosition() == pos
		}, reversed: func() bool {
			return false
		}}
	self.DataInputImpl = &DataInputImpl{self.BytesReader}
	return self.BytesReader
}

func (bs *BytesStore) reverseReader() *BytesReader {
	return bs.reverseReaderAllowSingle(true)
}

type BytesStoreReverseReader struct {
	*BytesReader
	owner      *BytesStore
	current    []byte
	nextBuffer int32
	nextRead   int32
}

func (r *BytesStoreReverseReader) ReadByte() (b byte, err error) {
	if r.nextRead == -1 {
		r.current = r.owner.blocks[r.nextBuffer]
		r.nextBuffer--
		r.nextRead = int32(r.owner.blockSize - 1)
	}
	r.nextRead--
	return r.current[r.nextRead+1], nil
}

func (r *BytesStoreReverseReader) ReadBytes(buf []byte) error {
	var err error
	for i, _ := range buf {
		buf[i], err = r.ReadByte()
		if err != nil {
			return err
		}
	}
	return err
}

func (bs *BytesStore) reverseReaderAllowSingle(allowSingle bool) *BytesReader {
	if allowSingle && len(bs.blocks) == 1 {
		return newReverseBytesReader(bs.blocks[0])
	}
	var current []byte = nil
	if len(bs.blocks) > 0 {
		current = bs.blocks[0]
	}
	self := &BytesStoreReverseReader{current: current, nextBuffer: -1, nextRead: 0}
	self.BytesReader = &BytesReader{
		skipBytes: func(count int32) {
			self.setPosition(self.getPosition() - int64(count))
		}, getPosition: func() int64 {
			return (int64(self.nextBuffer)+1)*int64(bs.blockSize) + int64(self.nextRead)
		}, setPosition: func(pos int64) {
			// NOTE: a little weird because if you
			// setPosition(0), the next byte you read is
			// bytes[0] ... but I would expect bytes[-1] (ie,
			// EOF)...?
			bufferIndex := int32(pos >> bs.blockSize)
			self.nextBuffer = bufferIndex - 1
			self.current = bs.blocks[bufferIndex]
			self.nextRead = int32(uint32(pos) & bs.blockMask)
			// assert getPosition() == pos
		}, reversed: func() bool {
			return true
		}}
	self.DataInputImpl = &DataInputImpl{self.BytesReader}
	return self.BytesReader
}

type ForwardBytesReader struct {
	*BytesReader
	bytes []byte
	pos   int32
}

func (r *ForwardBytesReader) ReadByte() (b byte, err error) {
	r.pos++
	return r.bytes[r.pos-1], nil
}

func (r *ForwardBytesReader) ReadBytes(buf []byte) error {
	copy(buf, r.bytes[r.pos:r.pos+int32(len(buf))])
	r.pos += int32(len(buf))
	return nil
}

func newForwardBytesReader(bytes []byte) *BytesReader {
	self := &ForwardBytesReader{bytes: bytes}
	self.BytesReader = &BytesReader{
		skipBytes: func(count int32) {
			self.pos += count
		}, getPosition: func() int64 {
			return int64(self.pos)
		}, setPosition: func(pos int64) {
			self.pos = int32(pos)
		}, reversed: func() bool {
			return false
		}}
	self.DataInputImpl = &DataInputImpl{self.BytesReader}
	return self.BytesReader
}

type ReverseBytesReader struct {
	*BytesReader
	bytes []byte
	pos   int32
}

func (r *ReverseBytesReader) ReadByte() (b byte, err error) {
	r.pos--
	return r.bytes[r.pos+1], nil
}

func (r *ReverseBytesReader) ReadBytes(buf []byte) error {
	for i, _ := range buf {
		buf[i] = r.bytes[r.pos]
		r.pos--
	}
	return nil
}

func newReverseBytesReader(bytes []byte) *BytesReader {
	self := &ReverseBytesReader{bytes: bytes}
	self.BytesReader = &BytesReader{
		skipBytes: func(count int32) {
			self.pos -= count
		}, getPosition: func() int64 {
			return int64(self.pos)
		}, setPosition: func(pos int64) {
			self.pos = int32(pos)
		}, reversed: func() bool {
			return true
		}}
	self.DataInputImpl = &DataInputImpl{self.BytesReader}
	return self.BytesReader
}
