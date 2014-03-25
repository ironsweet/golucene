package util

import (
	"sync/atomic"
)

type Allocator interface {
	// Allocate() interface{}
	// Recycle([]interface{})
}

type blockAllocator struct {
	blockSize int
}

func newBlockAllocator(blockSize int) *blockAllocator {
	return &blockAllocator{blockSize}
}

/* A simple Allocator that never recycles, but tracks how much total RAM is in use. */
type DirectTrackingAllocator struct {
}

func NewDirectTrackingAllocator(bytesUsed Counter) *DirectTrackingAllocator {
	panic("not implemented yet")
}

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
