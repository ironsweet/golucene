package util

import (
	"sync/atomic"
)

// util/ByteBlockPool.java

/*
An array holding the offset into the LEVEL_SIZE_ARRAY to quickly
navigate to the next slice level.
*/
var NEXT_LEVEL_ARRAY = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 9}

/* An array holding the level sizes for byte slices. */
var LEVEL_SIZE_ARRAY = []int{5, 14, 20, 30, 40, 40, 80, 80, 120, 200}

/* THe first level size for new slices */
var FIRST_LEVEL_SIZE = LEVEL_SIZE_ARRAY[0]

/*
Class that Posting and PostingVector use to writ ebyte streams into
shared fixed-size []byte arrays. The idea is to allocate slices of
increasing lengths. For example, the first slice is 5 bytes, the next
slice is 14, etc. We start by writing out bytes into the first 5
bytes. When we hit the end of the slice, we allocate the next slice
and then write the address of the next slice into the last 4 bytes of
the previous slice (the "forwarding address").

Each slice is filled with 0's initially, and we mark the end with a
non-zero byte. This way the methods that are writing into the slice
don't need to record its length and instead allocate a new slice once
they hit a non-zero byte.
*/
type ByteBlockPool struct {
	Buffers    [][]byte
	bufferUpto int
	ByteUpto   int
	Buffer     []byte
	ByteOffset int
	allocator  ByteAllocator
}

func NewByteBlockPool(allocator ByteAllocator) *ByteBlockPool {
	return &ByteBlockPool{
		bufferUpto: -1,
		ByteUpto:   BYTE_BLOCK_SIZE,
		ByteOffset: -BYTE_BLOCK_SIZE,
		allocator:  allocator,
	}
}

/* Expert: Resets the pool to its initial state reusing the first buffer. */
func (pool *ByteBlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
	// TODO consolidate with IntBlockPool.Reset()
	if pool.bufferUpto != -1 {
		// We allocated at least one buffer
		if zeroFillBuffers {
			panic("not implemented yet")
		}

		if pool.bufferUpto > 0 || !reuseFirst {
			offset := 0
			if reuseFirst {
				offset = 1
			}
			// Recycle all but the first buffer
			pool.allocator.recycle(pool.Buffers[offset : 1+pool.bufferUpto])
			for i := offset; i <= pool.bufferUpto; i++ {
				pool.Buffers[i] = nil
			}
		}
		if reuseFirst {
			panic("not implemented yet")
		} else {
			pool.bufferUpto = -1
			pool.ByteUpto = BYTE_BLOCK_SIZE
			pool.ByteOffset = -BYTE_BLOCK_SIZE
			pool.Buffer = nil
		}
	}
}

/*
Advances the pool to its next buffer. This method should be called
once after the constructor to initialize the pool. In contrast to the
constructor, a ByteBlockPool.Reset() call will advance the pool to
its first buffer immediately.
*/
func (pool *ByteBlockPool) NextBuffer() {
	if 1+pool.bufferUpto == len(pool.Buffers) {
		newBuffers := make([][]byte, Oversize(len(pool.Buffers)+1, NUM_BYTES_OBJECT_REF))
		copy(newBuffers, pool.Buffers)
		pool.Buffers = newBuffers
	}
	pool.Buffer = pool.allocator.allocate()
	pool.Buffers[1+pool.bufferUpto] = pool.Buffer
	pool.bufferUpto++

	pool.ByteUpto = 0
	pool.ByteOffset += BYTE_BLOCK_SIZE
}

/* Allocates a new slice with the given size. */
func (pool *ByteBlockPool) NewSlice(size int) int {
	if pool.ByteUpto > BYTE_BLOCK_SIZE-size {
		pool.NextBuffer()
	}
	upto := pool.ByteUpto
	pool.ByteUpto += size
	pool.Buffer[pool.ByteUpto-1] = 16
	return upto
}

/*
Creates a new byte slice with the given starting size and returns the
slices offset in the pool.
*/
func (p *ByteBlockPool) AllocSlice(slice []byte, upto int) int {
	level := slice[upto] & 15
	newLevel := NEXT_LEVEL_ARRAY[level]
	newSize := LEVEL_SIZE_ARRAY[newLevel]

	// maybe allocate another block
	if p.ByteUpto > BYTE_BLOCK_SIZE-newSize {
		p.NextBuffer()
	}

	newUpto := p.ByteUpto
	offset := newUpto + p.ByteOffset
	p.ByteUpto += newSize

	// copy forward the post 3 bytes (which we are about to overwrite
	// with the forwarding address):
	p.Buffer[newUpto] = slice[upto-3]
	p.Buffer[newUpto+1] = slice[upto-2]
	p.Buffer[newUpto+2] = slice[upto-1]

	// write forwarding address at end of last slice:
	slice[upto-3] = byte(uint(offset) >> 24)
	slice[upto-2] = byte(uint(offset) >> 16)
	slice[upto-1] = byte(uint(offset) >> 8)
	slice[upto] = byte(offset)

	// write new level:
	p.Buffer[p.ByteUpto-1] = byte(16 | newLevel)

	return newUpto + 3
}

/* Fill in a BytesRef from term's length & bytes encoded in byte block */
func (p *ByteBlockPool) SetBytesRef(term *BytesRef, textStart int) {
	bytes := p.Buffers[textStart>>BYTE_BLOCK_SHIFT]
	term.Bytes = bytes
	pos := textStart & BYTE_BLOCK_MASK
	if (bytes[pos] & 0x80) == 0 {
		// length is 1 byte
		term.Length = int(bytes[pos])
		term.Offset = pos + 1
	} else {
		// length is 2 bytes
		term.Length = (int(bytes[pos]) & 0x7f) + ((int(bytes[pos+1]) & 0xff) << 7)
		term.Offset = pos + 2
	}
	assert(term.Length >= 0)
}

/* Abstract class for allocating and freeing byte blocks. */
type ByteAllocator interface {
	recycle(blocks [][]byte)
	allocate() []byte
}

type ByteAllocatorImpl struct {
	blockSize int
}

func newByteAllocator(blockSize int) *ByteAllocatorImpl {
	return &ByteAllocatorImpl{blockSize}
}

func (a *ByteAllocatorImpl) allocate() []byte {
	assert(a.blockSize <= 1000000) // should not allocate more than 1MB?
	return make([]byte, a.blockSize)
}

/* A simple Allocator that never recycles, but tracks how much total RAM is in use. */
type DirectTrackingAllocator struct {
	*ByteAllocatorImpl
	bytesUsed Counter
}

func NewDirectTrackingAllocator(bytesUsed Counter) *DirectTrackingAllocator {
	return &DirectTrackingAllocator{
		ByteAllocatorImpl: newByteAllocator(BYTE_BLOCK_SIZE),
		bytesUsed:         bytesUsed,
	}
}

func (alloc *DirectTrackingAllocator) recycle(blocks [][]byte) {
	alloc.bytesUsed.AddAndGet(int64(-len(blocks) * alloc.blockSize))
	for i, _ := range blocks {
		blocks[i] = nil
	}
}

// util/Counter.java

type Counter interface {
	AddAndGet(delta int64) int64
	Get() int64
}

func NewCounter() Counter {
	return &serialCounter{0}
}

func NewAtomicCounter() Counter {
	return &atomicCounter{0}
}

type serialCounter struct {
	count int64
}

func (sc *serialCounter) AddAndGet(delta int64) int64 {
	sc.count += delta
	return sc.count
}

func (sc *serialCounter) Get() int64 {
	return sc.count
}

type atomicCounter struct {
	count int64
}

func (ac *atomicCounter) AddAndGet(delta int64) int64 {
	return atomic.AddInt64(&ac.count, delta)
}

func (ac *atomicCounter) Get() int64 {
	return atomic.LoadInt64(&ac.count)
}
