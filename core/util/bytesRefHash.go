package util

/*
BytesRefHash is a special purpose hash map like data structure
optimized for BytesRef instances. BytesRefHash maintains mappings of
byte arrays to ids (map[[]byte]int) sorting the hashed bytes
efficiently in continuous storage. The mapping to the id is
encapsulated inside BytesRefHash and is guaranteed to be increased
for each added BytesRef.

Note: The maximum capacity BytesRef instance passed to add() must not
be longer than BYTE_BLOCK_SIZE-2. The internal storage is limited to
2GB total byte storage.
*/
type BytesRefHash struct {
	pool       *ByteBlockPool
	bytesStart []int

	hashSize        int
	hashHalfSize    int
	hashMask        int
	count           int
	lastCount       int
	ids             []int
	bytesStartArray BytesStartArray
	bytesUsed       Counter
}

func NewBytesRefHash(pool *ByteBlockPool, capacity int,
	bytesStartArray BytesStartArray) *BytesRefHash {
	ids := make([]int, capacity)
	for i, _ := range ids {
		ids[i] = -1
	}
	counter := bytesStartArray.BytesUsed()
	if counter == nil {
		counter = NewCounter()
	}
	counter.AddAndGet(int64(capacity) * NUM_BYTES_INT)
	return &BytesRefHash{
		hashSize:        capacity,
		hashHalfSize:    capacity >> 1,
		hashMask:        capacity - 1,
		lastCount:       -1,
		pool:            pool,
		ids:             ids,
		bytesStartArray: bytesStartArray,
		bytesStart:      bytesStartArray.Init(),
		bytesUsed:       counter,
	}
}

/* Returns the number of values in this hash. */
func (h *BytesRefHash) Size() int {
	return h.count
}

/*
Returns the values array sorted by the referenced byte values.

Note: this is a destructive operation. clear() must be called in
order to reuse this BytesRefHash instance.
*/
func (h *BytesRefHash) Sort(comp func(a, b []byte) bool) []int {
	panic("not implemented yet")
}

func (h *BytesRefHash) equals(id int, b []byte) bool {
	panic("not implemented yet")
}

func (h *BytesRefHash) shrink(targetSize int) bool {
	// Cannot use util.Shrink because we require power of 2:
	newSize := h.hashSize
	for newSize >= 8 && newSize/4 > targetSize {
		newSize /= 2
	}
	if newSize != h.hashSize {
		h.bytesUsed.AddAndGet(NUM_BYTES_INT * -int64(h.hashSize-newSize))
		h.hashSize = newSize
		h.ids = make([]int, h.hashSize)
		for i, _ := range h.ids {
			h.ids[i] = -1
		}
		h.hashHalfSize = newSize / 2
		h.hashMask = newSize - 1
		return true
	}
	return false
}

/* Clears the BytesRef which maps to the given BytesRef */
func (h *BytesRefHash) Clear(resetPool bool) {
	h.lastCount = h.count
	h.count = 0
	if resetPool {
		h.pool.Reset(false, false) // we don't need to 0-fill the bufferes
	}
	h.bytesStart = h.bytesStartArray.Clear()
	if h.lastCount != -1 && h.shrink(h.lastCount) {
		// shurnk clears the hash entries
		return
	}
	for i, _ := range h.ids {
		h.ids[i] = -1
	}
}

/* Adds a new BytesRef with pre-calculated hash code. */
func (h *BytesRefHash) Add(bytes []byte, code int) (int, bool) {
	assert(bytes != nil)
	assert(len(bytes) > 0)
	assert2(h.bytesStart != nil, "Bytesstart is null - not initialized")
	length := len(bytes)
	// final position
	hashPos := h.findHash(bytes, code)
	e := h.ids[hashPos]

	if e == -1 {
		// new entry
		if len2 := 2 + len(bytes); len2+h.pool.ByteUpto > BYTE_BLOCK_SIZE {
			if len2 > BYTE_BLOCK_SIZE {
				return 0, false
			}
			h.pool.NextBuffer()
		}
		buffer := h.pool.buffer
		bufferUpto := h.pool.ByteUpto
		if h.count >= len(h.bytesStart) {
			h.bytesStart = h.bytesStartArray.Grow()
			assert2(h.count < len(h.bytesStart)+1, "count: %v len: %v", h.count, len(h.bytesStart))
		}
		e = h.count
		h.count++

		h.bytesStart[e] = bufferUpto + h.pool.ByteOffset

		// We first encode the length, followed by the bytes. Length is
		// encoded as vint, but will consume 1 or 2 bytes at most (we
		// reject too-long terms, above).
		if length < 128 {
			// 1 byte to store length
			buffer[bufferUpto] = byte(length)
			h.pool.ByteUpto += length + 1
			assert2(length >= 0, "Length must be positive: %v", length)
			copy(buffer[bufferUpto+1:], bytes)
		} else {
			// 2 bytes to store length
			buffer[bufferUpto] = byte(0x80 | (length & 0x7f))
			buffer[bufferUpto+1] = byte((length >> 7) & 0xff)
			h.pool.ByteUpto += length + 2
			copy(buffer[bufferUpto+2:], bytes)
		}
		assert(h.ids[hashPos] == -1)
		h.ids[hashPos] = e

		if h.count == h.hashHalfSize {
			h.rehash(2*h.hashSize, true)
		}
		return e, true
	}
	return -(e + 1), true
}

func (h *BytesRefHash) findHash(bytes []byte, code int) int {
	assert(bytes != nil)
	assert2(h.bytesStart != nil, "bytesStart is null - not initialized")
	// final position
	hashPos := code & h.hashMask
	if e := h.ids[hashPos]; e != -1 && !h.equals(e, bytes) {
		panic("not implemented yet")
	}
	return hashPos
}

/* Claled when has is too small (> 50% occupied) or too large (< 20% occupied). */
func (h *BytesRefHash) rehash(newSize int, hashOnData bool) {
	panic("not implemented yet")
}

/*
reinitializes the BytesRefHash after a previous clear() call. If
clear() has not been called previously this method has no effect.
*/
func (h *BytesRefHash) Reinit() {
	if h.bytesStart == nil {
		h.bytesStart = h.bytesStartArray.Init()
	}
	if h.ids == nil {
		h.ids = make([]int, h.hashSize)
		h.bytesUsed.AddAndGet(NUM_BYTES_INT * int64(h.hashSize))
	}
}

/*
Returns the bytesStart offset into the internally used ByteBlockPool
for the given bytesID.
*/
func (h *BytesRefHash) ByteStart(bytesId int) int {
	assert2(h.bytesStart != nil, "bytesStart is null - not initialized")
	assert2(bytesId >= 0 && bytesId <= h.count, "%v", bytesId)
	return h.bytesStart[bytesId]
}

/* Manages allocation of per-term addresses. */
type BytesStartArray interface {
	// Initializes the BytesStartArray. This call will allocate memory
	Init() []int
	// A Counter reference holding the number of bytes used by this
	// BytesStartArray. The BytesRefHash uses this reference to track
	// its memory usage
	BytesUsed() Counter
	// Grows the BytesStartArray
	Grow() []int
	// clears the BytesStartArray and returns the cleared instance.
	Clear() []int
}
