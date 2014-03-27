package util

// util/IntBlockPool.java

/* A pool for int blocks similar to ByteBlockPool */
type IntBlockPool struct {
	allocator IntAllocator
}

func NewIntBlockPool(allocator IntAllocator) *IntBlockPool {
	return &IntBlockPool{allocator}
}

type IntAllocator interface{}
