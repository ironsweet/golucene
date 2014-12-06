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
	Bytes  []byte
	Offset int
	Length int
}

func NewEmptyBytesRef() *BytesRef {
	return NewBytesRefFrom(EMPTY_BYTES)
}

func NewBytesRef(bytes []byte, offset, length int) *BytesRef {
	return &BytesRef{
		Bytes:  bytes,
		Offset: offset,
		Length: length,
	}
}

func NewBytesRefFrom(bytes []byte) *BytesRef {
	return NewBytesRef(bytes, 0, len(bytes))
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
Expert: compares the byte against another BytesRef, returning true if
the bytes are equal.
*/
func (br *BytesRef) bytesEquals(other []byte) bool {
	if br.Length != len(other) {
		return false
	}
	for i, v := range br.ToBytes() {
		if v != other[i] {
			return false
		}
	}
	return true
}

func (br *BytesRef) String() string {
	panic("not implemented yet")
}

func (br *BytesRef) ToBytes() []byte {
	return br.Bytes[br.Offset : br.Offset+br.Length]
}

/*
Copies the bytes from the given BytesRef

NOTE: if this would exceed the slice size, this method creates a new
reference array.
*/
func (a *BytesRef) copyBytes(other *BytesRef) {
	if len(a.Bytes)-a.Offset < other.Length {
		a.Bytes = make([]byte, other.Length)
		a.Offset = 0
	}
	copy(a.Bytes, other.Bytes[other.Offset:other.Offset+other.Length])
	a.Length = other.Length
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
	ref *BytesRef
}

func NewBytesRefBuilder() *BytesRefBuilder {
	return &BytesRefBuilder{
		ref: NewEmptyBytesRef(),
	}
}

/* Return a reference to the bytes of this build. */
func (b *BytesRefBuilder) Bytes() []byte {
	return b.ref.Bytes
}

/* Return the number of bytes in this buffer. */
func (b *BytesRefBuilder) Length() int {
	return b.ref.Length
}

/* Set the length. */
func (b *BytesRefBuilder) SetLength(length int) {
	b.ref.Length = length
}

/* Return the byte at the given offset. */
func (b *BytesRefBuilder) At(offset int) byte {
	return b.ref.Bytes[offset]
}

/* Set a byte. */
func (b *BytesRefBuilder) Set(offset int, v byte) {
	b.ref.Bytes[offset] = v
}

/* Ensure that this builder can hold at least capacity bytes without resizing. */
func (b *BytesRefBuilder) Grow(capacity int) {
	b.ref.Bytes = GrowByteSlice(b.ref.Bytes, capacity)
}

func (b *BytesRefBuilder) append(bytes []byte) {
	b.Grow(b.ref.Length + len(bytes))
	copy(b.ref.Bytes[b.ref.Length:], bytes)
	b.ref.Length += len(bytes)
}

func (b *BytesRefBuilder) clear() {
	b.SetLength(0)
}

func (b *BytesRefBuilder) Copy(ref []byte) {
	b.clear()
	b.append(ref)
}

func (b *BytesRefBuilder) Get() *BytesRef {
	assert2(b.ref.Offset == 0, "Modifying the offset of the returned ref is illegal")
	return b.ref
}
