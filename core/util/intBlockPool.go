package util

// util/IntBlockPool.java

const INT_BLOCK_SHIFT = 13
const INT_BLOCK_SIZE = 1 << INT_BLOCK_SHIFT
const INT_BLOCK_MASK = INT_BLOCK_SIZE - 1

/* A pool for int blocks similar to ByteBlockPool */
type IntBlockPool struct {
	Buffers    [][]int
	bufferUpto int
	IntUpto    int
	Buffer     []int
	IntOffset  int
	allocator  IntAllocator
}

func NewIntBlockPool(allocator IntAllocator) *IntBlockPool {
	return &IntBlockPool{
		Buffers:    make([][]int, 10),
		bufferUpto: -1,
		IntUpto:    INT_BLOCK_SIZE,
		IntOffset:  -INT_BLOCK_SIZE,
		allocator:  allocator,
	}
}

/* Expert: Resets the pool to its initial state reusing the first buffer. */
func (pool *IntBlockPool) Reset(zeroFillBuffers, reuseFirst bool) {
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
			pool.allocator.Recycle(pool.Buffers[offset : 1+pool.bufferUpto])
			for i := offset; i <= pool.bufferUpto; i++ {
				pool.Buffers[i] = nil
			}
		}
		if reuseFirst {
			panic("not implemented yet")
		} else {
			pool.bufferUpto = -1
			pool.IntUpto = INT_BLOCK_SIZE
			pool.IntOffset = -INT_BLOCK_SIZE
			pool.Buffer = nil
		}
	}
}

/*
Advances the pool to its next buffer. This mthod should be called
once after the constructor to initialize the pool. In contrast to the
constructor a IntBlockPool.reset() call will advance the pool to its
first buffer immediately.
*/
func (p *IntBlockPool) NextBuffer() {
	if 1+p.bufferUpto == len(p.Buffers) {
		newBuffers := make([][]int, len(p.Buffers)+len(p.Buffers)/2)
		copy(newBuffers, p.Buffers)
		p.Buffers = newBuffers
	}
	p.Buffer = p.allocator.allocate()
	p.Buffers[1+p.bufferUpto] = p.Buffer
	p.bufferUpto++

	p.IntUpto = 0
	p.IntOffset += INT_BLOCK_SIZE
}

type IntAllocator interface {
	Recycle(blocks [][]int)
	allocate() []int
}

type IntAllocatorImpl struct {
	blockSize int
}

func NewIntAllocator(blockSize int) *IntAllocatorImpl {
	return &IntAllocatorImpl{blockSize}
}

func (a *IntAllocatorImpl) allocate() []int {
	return make([]int, a.blockSize)
}
