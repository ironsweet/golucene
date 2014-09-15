package fst

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

type BytesStore struct {
	*util.DataOutputImpl
	blocks    [][]byte
	blockSize uint32
	blockBits uint32
	blockMask uint32
	current   []byte
	nextWrite uint32
}

func newBytesStore() *BytesStore {
	bs := &BytesStore{}
	bs.DataOutputImpl = util.NewDataOutput(bs)
	return bs
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

func newBytesStoreFromInput(in util.DataInput, numBytes int64, maxBlockSize uint32) (bs *BytesStore, err error) {
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

func (bs *BytesStore) WriteByte(b byte) error {
	if bs.nextWrite == bs.blockSize {
		bs.current = make([]byte, bs.blockSize)
		bs.blocks = append(bs.blocks, bs.current)
		bs.nextWrite = 0
	}
	bs.current[bs.nextWrite] = b
	bs.nextWrite++
	return nil
}

func (bs *BytesStore) WriteBytes(buf []byte) error {
	var offset uint32 = 0
	length := uint32(len(buf))
	for length > 0 {
		chunk := bs.blockSize - bs.nextWrite
		if length <= chunk {
			copy(bs.current[bs.nextWrite:], buf[offset:offset+length])
			bs.nextWrite += length
			break
		} else {
			if chunk > 0 {
				copy(bs.current[bs.nextWrite:], buf[offset:offset+chunk])
				offset += chunk
				length -= chunk
			}
			bs.current = make([]byte, bs.blockSize)
			bs.blocks = append(bs.blocks, bs.current)
			bs.nextWrite = 0
		}
	}
	return nil
}

func (s *BytesStore) writeBytesAt(dest int64, b []byte) {
	length := len(b)
	assert2(dest+int64(length) <= s.position(),
		"dest=%v pos=%v len=%v", dest, s.position(), length)

	end := dest + int64(length)
	blockIndex := int(end >> s.blockBits)
	downTo := int(end & int64(s.blockMask))
	if downTo == 0 {
		blockIndex--
		downTo = int(s.blockSize)
	}
	block := s.blocks[blockIndex]

	for length > 0 {
		if length <= downTo {
			copy(block[downTo-length:], b[:length])
			break
		}
		length -= downTo
		copy(block, b[length:length+downTo])
		blockIndex--
		block = s.blocks[blockIndex]
		downTo = int(s.blockSize)
	}
}

func (s *BytesStore) copyBytesInside(src, dest int64, length int) {
	assert(src < dest)

	end := src + int64(length)

	blockIndex := int(end >> s.blockBits)
	downTo := int(end & int64(s.blockMask))
	if downTo == 0 {
		blockIndex--
		downTo = int(s.blockSize)
	}
	block := s.blocks[blockIndex]

	for length > 0 {
		if length <= downTo {
			s.writeBytesAt(dest, block[downTo-length:downTo])
			break
		}
		length -= downTo
		s.writeBytesAt(dest+int64(length), block[:downTo])
		blockIndex--
		block = s.blocks[blockIndex]
		downTo = int(s.blockSize)
	}
}

/* Reverse from srcPos, inclusive, to destPos, inclusive. */
func (s *BytesStore) reverse(srcPos, destPos int64) {
	assert(srcPos < destPos)
	assert(destPos < s.position())
	// fmt.Printf("reverse src=%v dest=%v\n", srcPos, destPos)

	srcBlockIndex := int(srcPos >> s.blockBits)
	src := int(srcPos & int64(s.blockMask))
	srcBlock := s.blocks[srcBlockIndex]

	destBlockIndex := int(destPos >> s.blockBits)
	dest := int(destPos & int64(s.blockMask))
	destBlock := s.blocks[destBlockIndex]

	// fmt.Printf("  srcBlock=%v destBlock=%v\n", srcBlockIndex, destBlockIndex)

	limit := int((destPos - srcPos + 1) / 2)
	for i := 0; i < limit; i++ {
		// fmt.Printf("  cycle src=%v dest=%v\n", src, dest)
		srcBlock[src], destBlock[dest] = destBlock[dest], srcBlock[src]
		if src++; src == int(s.blockSize) {
			srcBlockIndex++
			srcBlock = s.blocks[srcBlockIndex]
			fmt.Printf("  set destBlock=%v srcBlock=%v\n", destBlock, srcBlock)
			src = 0
		}

		if dest--; dest == -1 {
			destBlockIndex--
			destBlock = s.blocks[destBlockIndex]
			fmt.Printf("  set destBlock=%v srcBlock=%v\n", destBlock, srcBlock)
			dest = int(s.blockSize - 1)
		}
	}
}

func (s *BytesStore) skipBytes(length int) {
	for length > 0 {
		chunk := int(s.blockSize) - int(s.nextWrite)
		if length <= chunk {
			s.nextWrite += uint32(length)
			break
		}
		length -= chunk
		s.current = make([]byte, s.blockSize)
		s.blocks = append(s.blocks, s.current)
		s.nextWrite = 0
	}
}

func (s *BytesStore) position() int64 {
	return int64(len(s.blocks)-1)*int64(s.blockSize) + int64(s.nextWrite)
}

func (s *BytesStore) finish() {
	if s.current != nil {
		lastBuffer := make([]byte, s.nextWrite)
		copy(lastBuffer, s.current[:s.nextWrite])
		s.blocks[len(s.blocks)-1] = lastBuffer
		s.current = nil
	}
}

/* Writes all of our bytes to the target DataOutput. */
func (s *BytesStore) writeTo(out util.DataOutput) error {
	for _, block := range s.blocks {
		err := out.WriteBytes(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *BytesStore) String() string {
	return fmt.Sprintf("%v-bits x%v bytes store", s.blockBits, len(s.blocks))
}

type BytesStoreForwardReader struct {
	*util.DataInputImpl
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

func (r *BytesStoreForwardReader) skipBytes(count int64) {
	r.setPosition(r.getPosition() + count)
}

func (r *BytesStoreForwardReader) getPosition() int64 {
	return (int64(r.nextBuffer)-1)*int64(r.owner.blockSize) + int64(r.nextRead)
}

func (r *BytesStoreForwardReader) setPosition(pos int64) {
	bufferIndex := pos >> r.owner.blockBits
	r.nextBuffer = uint32(bufferIndex + 1)
	r.current = r.owner.blocks[bufferIndex]
	r.nextRead = uint32(pos) & r.owner.blockMask
	// assert self.getPosition() == pos
}

func (r *BytesStoreForwardReader) reversed() bool {
	return false
}

func (bs *BytesStore) forwardReader() BytesReader {
	if len(bs.blocks) == 1 {
		return newForwardBytesReader(bs.blocks[0])
	}
	ans := &BytesStoreForwardReader{owner: bs, nextRead: bs.blockSize}
	ans.DataInputImpl = util.NewDataInput(ans)
	return ans
}

func (bs *BytesStore) reverseReader() BytesReader {
	return bs.reverseReaderAllowSingle(true)
}

type BytesStoreReverseReader struct {
	*util.DataInputImpl
	owner      *BytesStore
	current    []byte
	nextBuffer int32
	nextRead   int32
}

func newBytesStoreReverseReader(owner *BytesStore, current []byte) *BytesStoreReverseReader {
	ans := &BytesStoreReverseReader{owner: owner, current: current, nextBuffer: -1}
	ans.DataInputImpl = util.NewDataInput(ans)
	return ans
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

func (r *BytesStoreReverseReader) skipBytes(count int64) {
	r.setPosition(r.getPosition() - count)
}

func (r *BytesStoreReverseReader) getPosition() int64 {
	return (int64(r.nextBuffer)+1)*int64(r.owner.blockSize) + int64(r.nextRead)
}

func (r *BytesStoreReverseReader) setPosition(pos int64) {
	// NOTE: a little weird because if you
	// setPosition(0), the next byte you read is
	// bytes[0] ... but I would expect bytes[-1] (ie,
	// EOF)...?
	bufferIndex := int32(pos >> r.owner.blockSize)
	r.nextBuffer = bufferIndex - 1
	r.current = r.owner.blocks[bufferIndex]
	r.nextRead = int32(uint32(pos) & r.owner.blockMask)
	// assert getPosition() == pos
}

func (r *BytesStoreReverseReader) reversed() bool {
	return true
}

func (bs *BytesStore) reverseReaderAllowSingle(allowSingle bool) BytesReader {
	if allowSingle && len(bs.blocks) == 1 {
		return newReverseBytesReader(bs.blocks[0])
	}
	var current []byte = nil
	if len(bs.blocks) > 0 {
		current = bs.blocks[0]
	}
	return newBytesStoreReverseReader(bs, current)
}

type ForwardBytesReader struct {
	*util.DataInputImpl
	bytes []byte
	pos   int
}

func (r *ForwardBytesReader) ReadByte() (b byte, err error) {
	r.pos++
	return r.bytes[r.pos-1], nil
}

func (r *ForwardBytesReader) ReadBytes(buf []byte) error {
	copy(buf, r.bytes[r.pos:r.pos+len(buf)])
	r.pos += len(buf)
	return nil
}

func (r *ForwardBytesReader) skipBytes(count int64) {
	r.pos += int(count)
}

func (r *ForwardBytesReader) getPosition() int64 {
	return int64(r.pos)
}

func (r *ForwardBytesReader) setPosition(pos int64) {
	r.pos = int(pos)
}

func (r *ForwardBytesReader) reversed() bool {
	return false
}

func newForwardBytesReader(bytes []byte) BytesReader {
	ans := &ForwardBytesReader{bytes: bytes}
	ans.DataInputImpl = util.NewDataInput(ans)
	return ans
}

type ReverseBytesReader struct {
	*util.DataInputImpl
	bytes []byte
	pos   int
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

func newReverseBytesReader(bytes []byte) BytesReader {
	ans := &ReverseBytesReader{bytes: bytes}
	ans.DataInputImpl = util.NewDataInput(ans)
	return ans
}

func (r *ReverseBytesReader) skipBytes(count int64) {
	r.pos -= int(count)
}

func (r *ReverseBytesReader) getPosition() int64 {
	return int64(r.pos)
}

func (r *ReverseBytesReader) setPosition(pos int64) {
	r.pos = int(pos)
}

func (r *ReverseBytesReader) reversed() bool {
	return true
}

func (r *ReverseBytesReader) String() string {
	return fmt.Sprintf("BytesReader(reversed, [%v,%v])", r.pos, len(r.bytes))
}
