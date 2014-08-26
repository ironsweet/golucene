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

func (a *IntsRef) At(i int) int {
	return a.Ints[a.Offset+i]
}

func (a *IntsRef) Value() []int {
	return a.Ints[a.Offset : a.Offset+a.Length]
}

/* Signed int order comparison */
func (a *IntsRef) Less(other *IntsRef) bool {
	if a == other {
		return false
	}

	aInts := a.Ints
	aUpto := a.Offset
	bInts := other.Ints
	bUpto := other.Offset

	var aStop int
	if a.Length < other.Length {
		aStop = aUpto + a.Length
	} else {
		aStop = aUpto + other.Length
	}

	for aUpto < aStop {
		aInt := aInts[aUpto]
		aUpto++
		bInt := bInts[bUpto]
		bUpto++
		if aInt > bInt {
			return false
		} else if aInt < bInt {
			return true
		}
	}

	// one is a prefix of the other, or, they are equal:
	return a.Length < other.Length
}

func (a *IntsRef) CopyInts(other *IntsRef) {
	if len(a.Ints)-a.Offset < other.Length {
		a.Ints = make([]int, other.Length)
		a.Offset = 0
	}
	copy(a.Ints, other.Ints[other.Offset:other.Offset+other.Length])
	a.Length = other.Length
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
