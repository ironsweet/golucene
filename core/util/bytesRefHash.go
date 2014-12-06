package util

import (
	"fmt"
)

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

	scratch1        *BytesRef
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
		scratch1:        NewEmptyBytesRef(),
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
Returns the ids array in arbitrary order. Valid ids start at offset
of 0 and end at a limit of size() - 1

Note: This is a destructive operation. clear() must be called in
order to reuse this BytesRefHash instance.
*/
func (h *BytesRefHash) compact() []int {
	assert2(h.bytesStart != nil, "bytesStart is nil - not initialized")
	upto := 0
	for i := 0; i < h.hashSize; i++ {
		if h.ids[i] != -1 {
			if upto < i {
				h.ids[upto] = h.ids[i]
				h.ids[i] = -1
			}
			upto++
		}
	}

	assert(upto == h.count)
	h.lastCount = h.count
	return h.ids
}

type bytesRefIntroSorter struct {
	*IntroSorter
	owner    *BytesRefHash
	compact  []int
	comp     func([]byte, []byte) bool
	pivot    *BytesRef
	scratch1 *BytesRef
	scratch2 *BytesRef
}

func newBytesRefIntroSorter(owner *BytesRefHash, v []int,
	comp func([]byte, []byte) bool) *bytesRefIntroSorter {
	ans := &bytesRefIntroSorter{
		owner:    owner,
		compact:  v,
		comp:     comp,
		pivot:    NewEmptyBytesRef(),
		scratch1: NewEmptyBytesRef(),
		scratch2: NewEmptyBytesRef(),
	}
	ans.IntroSorter = NewIntroSorter(ans, ans)
	return ans
}

func (a *bytesRefIntroSorter) Len() int      { return len(a.compact) }
func (a *bytesRefIntroSorter) Swap(i, j int) { a.compact[i], a.compact[j] = a.compact[j], a.compact[i] }
func (a *bytesRefIntroSorter) Less(i, j int) bool {
	id1, id2 := a.compact[i], a.compact[j]
	assert(len(a.owner.bytesStart) > id1 && len(a.owner.bytesStart) > id2)
	a.owner.pool.SetBytesRef(a.scratch1, a.owner.bytesStart[id1])
	a.owner.pool.SetBytesRef(a.scratch2, a.owner.bytesStart[id2])
	return a.comp(a.scratch1.ToBytes(), a.scratch2.ToBytes())
}

func (a *bytesRefIntroSorter) SetPivot(i int) {
	id := a.compact[i]
	assert(len(a.owner.bytesStart) > id)
	a.owner.pool.SetBytesRef(a.pivot, a.owner.bytesStart[id])
}

func (a *bytesRefIntroSorter) PivotLess(j int) bool {
	id := a.compact[j]
	assert(len(a.owner.bytesStart) > id)
	a.owner.pool.SetBytesRef(a.scratch2, a.owner.bytesStart[id])
	return a.comp(a.pivot.ToBytes(), a.scratch2.ToBytes())
}

/*
Returns the values array sorted by the referenced byte values.

Note: this is a destructive operation. clear() must be called in
order to reuse this BytesRefHash instance.
*/
func (h *BytesRefHash) Sort(comp func(a, b []byte) bool) []int {
	compact := h.compact()
	s := newBytesRefIntroSorter(h, compact, comp)
	s.Sort(0, h.count)
	// TODO remove this
	// for i, _ := range compact {
	// 	if compact[i+1] == -1 {
	// 		break
	// 	}
	// 	assert(!s.Less(i+1, i))
	// 	if ok := !s.Less(i+1, i); !ok {
	// 		fmt.Println("DEBUG1", compact)
	// 		assert(ok)
	// 	}
	// }
	return compact
}

func (h *BytesRefHash) equals(id int, b []byte) bool {
	h.pool.SetBytesRef(h.scratch1, h.bytesStart[id])
	return h.scratch1.bytesEquals(b)
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

type MaxBytesLengthExceededError string

func (e MaxBytesLengthExceededError) Error() string {
	return string(e)
}

/* Adds a new BytesRef. */
func (h *BytesRefHash) Add(bytes []byte) (int, error) {
	assert2(h.bytesStart != nil, "Bytesstart is null - not initialized")
	length := len(bytes)
	// final position
	hashPos := h.findHash(bytes)
	e := h.ids[hashPos]

	if e == -1 {
		// new entry
		if len2 := 2 + len(bytes); len2+h.pool.ByteUpto > BYTE_BLOCK_SIZE {
			if len2 > BYTE_BLOCK_SIZE {
				return 0, MaxBytesLengthExceededError(fmt.Sprintf(
					"bytes can be at most %v in length; got %v",
					BYTE_BLOCK_SIZE-2, len(bytes)))
			}
			h.pool.NextBuffer()
		}
		buffer := h.pool.Buffer
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
		return e, nil
	}
	return -(e + 1), nil
}

func (h *BytesRefHash) findHash(bytes []byte) int {
	assert2(h.bytesStart != nil, "bytesStart is null - not initialized")
	code := h.doHash(bytes)
	// final position
	hashPos := code & h.hashMask
	if e := h.ids[hashPos]; e != -1 && !h.equals(e, bytes) {
		// conflict; use linear probe to find an open slot
		// (see LUCENE-5604):
		for {
			code++
			hashPos = code & h.hashMask
			e = h.ids[hashPos]
			if e == -1 || h.equals(e, bytes) {
				break
			}
		}
	}
	return hashPos
}

/* Called when has is too small (> 50% occupied) or too large (< 20% occupied). */
func (h *BytesRefHash) rehash(newSize int, hashOnData bool) {
	newMask := newSize - 1
	h.bytesUsed.AddAndGet(NUM_BYTES_INT * int64(newSize))
	newHash := make([]int, newSize)
	for i, _ := range newHash {
		newHash[i] = -1
	}
	for i := 0; i < h.hashSize; i++ {
		if e0 := h.ids[i]; e0 != -1 {
			var code int
			if hashOnData {
				off := h.bytesStart[e0]
				start := off & BYTE_BLOCK_MASK
				bytes := h.pool.Buffers[off>>BYTE_BLOCK_SHIFT]
				var length int
				var pos int
				if bytes[start]&0x80 == 0 {
					// length is 1 byte
					length = int(bytes[start])
					pos = start + 1
				} else {
					length = int(bytes[start]&0x7f) + (int(bytes[start+1]&0xff) << 7)
					pos = start + 2
				}
				code = h.doHash(bytes[pos : pos+length])
			} else {
				code = h.bytesStart[e0]
			}

			hashPos := code & newMask
			assert(hashPos >= 0)
			if newHash[hashPos] != -1 {
				// conflict; use linear probe to find an open slot
				// (see LUCENE-5604)
				for {
					code++
					hashPos = code & newMask
					if newHash[hashPos] == -1 {
						break
					}
				}
			}
			assert(newHash[hashPos] == -1)
			newHash[hashPos] = e0
		}
	}

	h.hashMask = newMask
	h.bytesUsed.AddAndGet(NUM_BYTES_INT * int64(-len(h.ids)))
	h.ids = newHash
	h.hashSize = newSize
	h.hashHalfSize = newSize / 2
}

func (h *BytesRefHash) doHash(p []byte) int {
	return int(MurmurHash3_x86_32(p, GOOD_FAST_HASH_SEED))
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
