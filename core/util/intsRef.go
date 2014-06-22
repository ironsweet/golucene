package util

// util/IntsRef.java

/* An empty integer array for convenience */
var EMPTY_INTS = []int{}

/*
Represents []int, as a slice (offset + length) into an existing []int.
The ints member should never be nil; use EMPTY_INTS if necessary.

Go's native slice is always preferrable unless the reference pointer
need to remain unchanged, in which case, this class is more useful.
*/
type IntsRef struct {
	// The contents of teh IntsRef. Should never be nil.
	Ints []int
	// Offset of first valid integer.
	Offset int
	// Length of used ints.
	Length int
}

func NewEmptyIntsRef() *IntsRef {
	return &IntsRef{Ints: EMPTY_INTS}
}

func (a *IntsRef) Value() []int {
	return a.Ints[a.Offset : a.Offset+a.Length]
}

/* Signed int order comparison */
func (a *IntsRef) CompareTo(other *IntsRef) bool {
	panic("not implemented yet")
}

func (a *IntsRef) CopyInts(other *IntsRef) {
	panic("not implemented yet")
}

/*
Used to grow the reference slice.

In general this should not be used as it does not take the offset into account.
*/
func (a *IntsRef) Grow(newLength int) {
	assert(a.Offset == 0)
	if len(a.Ints) < newLength {
		a.Ints = GrowIntSlice(a.Ints, newLength)
	}
}
