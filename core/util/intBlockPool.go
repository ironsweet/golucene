package util

// util/IntBlockPool.java

/* A pool for int blocks similar to ByteBlockPool */
type IntBlockPool struct {
	allocator IntAllocator
}

func NewIntBlockPool(allocator IntAllocator) *IntBlockPool {
	return &IntBlockPool{allocator}
}

/* Expert: Resets the pool to its initial state reusing the first buffer. */
func (pool *IntBlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
	panic("not implemented yet")
}

type IntAllocator interface{}
