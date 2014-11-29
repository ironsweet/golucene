package util

// util/BytesRef.java

/* An empty byte slice for convenience */
var EMPTY_BYTES = []byte{}

/*
Represents []byte, as a slice (offset + length) into an existing
[]byte, similar to Go's byte slice.

Important note: Unless otherwise noted, GoLucene uses []byte directly
to represent terms that are encoded as UTF8 bytes in the index. It
uses this class in cases when caller needs to hold a reference, while
allowing underlying []byte to change.
*/
type BytesRef struct {
	// The contents of the BytesRef.
	Value []byte
}

func NewEmptyBytesRef() *BytesRef {
	return NewBytesRef(EMPTY_BYTES)
}

func NewBytesRef(bytes []byte) *BytesRef {
	return &BytesRef{bytes}
}

/*
Expert: compares the byte against another BytesRef, returning true if
the bytes are equal.
*/
func (br *BytesRef) bytesEquals(other []byte) bool {
	assert(other != nil)
	if len(br.Value) == len(other) {
		for i, v := range br.Value {
			if v != other[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (br *BytesRef) String() string {
	panic("not implemented yet")
}

/*
Creates a new BytesRef that points to a copy of the bytes from
other.

The returned BytesRef will have a length of other.length and an
offset of zero.
*/
func DeepCopyOf(other *BytesRef) *BytesRef {
	copy := NewEmptyBytesRef()
	copy.copyBytes(other)
	return copy
}

/*
Copies the bytes from the given BytesRef

NOTE: if this would exceed the slice size, this method creates a new
reference array.
*/
func (a *BytesRef) copyBytes(other *BytesRef) {
	if len(a.Value) < len(other.Value) {
		a.Value = make([]byte, len(other.Value))
	}
	copy(a.Value, other.Value)
}

func UTF8SortedAsUnicodeLess(aBytes, bBytes []byte) bool {
	aLen, bLen := len(aBytes), len(bBytes)

	for i, v := range aBytes {
		if i >= bLen {
			break
		}
		if v < bBytes[i] {
			return true
		} else if v > bBytes[i] {
			return false
		}
	}

	// One is a prefix of the other, or, they are equal:
	return aLen < bLen
}

type BytesRefs [][]byte

func (br BytesRefs) Len() int {
	return len(br)
}

func (br BytesRefs) Less(i, j int) bool {
	aBytes, bBytes := br[i], br[j]
	return UTF8SortedAsUnicodeLess(aBytes, bBytes)
}

func (br BytesRefs) Swap(i, j int) {
	br[i], br[j] = br[j], br[i]
}

// util/BytesRefBuilder.java

type BytesRefBuilder struct {
	ref []byte
}

func NewBytesRefBuilder() *BytesRefBuilder {
	return new(BytesRefBuilder)
}

/* Return a reference to the bytes of this build. */
func (b *BytesRefBuilder) Bytes() []byte {
	return b.ref
}

/* Return the number of bytes in this buffer. */
func (b *BytesRefBuilder) Length() int {
	return len(b.ref)
}

/* Set the length. */
func (b *BytesRefBuilder) SetLength(length int) {
	assert(length <= len(b.ref))
	b.ref = b.ref[:length]
}

/* Return the byte at the given offset. */
func (b *BytesRefBuilder) At(offset int) byte {
	return b.ref[offset]
}

/* Set a byte. */
func (b *BytesRefBuilder) Set(offset int, v byte) {
	b.ref[offset] = v
}

/* Ensure that this builder can hold at least capacity bytes without resizing. */
func (b *BytesRefBuilder) Grow(capacity int) {
	b.ref = GrowByteSlice(b.ref, capacity)
}

func (b *BytesRefBuilder) Get() *BytesRef {
	return NewBytesRef(b.ref)
}

func (b *BytesRefBuilder) Copy(ref []byte) {
	b.ref = GrowByteSlice(b.ref, len(ref))
	copy(b.ref, ref)
}
