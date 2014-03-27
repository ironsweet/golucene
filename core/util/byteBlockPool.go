package util

import (
	"sync/atomic"
)

// util/ByteBlockPool.java

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
	allocator ByteAllocator
}

func NewByteBlockPool(allocator ByteAllocator) *ByteBlockPool {
	return &ByteBlockPool{allocator}
}

type ByteAllocator interface {
	// Allocate() interface{}
	// Recycle([]interface{})
}

type ByteAllocatorImpl struct {
	blockSize int
}

func newByteAllocator(blockSize int) *ByteAllocatorImpl {
	return &ByteAllocatorImpl{blockSize}
}

/* A simple Allocator that never recycles, but tracks how much total RAM is in use. */
type DirectTrackingAllocator struct {
	bytesUsed Counter
}

func NewDirectTrackingAllocator(bytesUsed Counter) *DirectTrackingAllocator {
	return &DirectTrackingAllocator{bytesUsed}
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
