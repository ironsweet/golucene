package automaton

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// util/automaton/SortedIntSet.java

// If we hold more than this many states, we swtich from O(N*2)
// linear ops to O(N log(N)) TreeMap
const TREE_MAP_CUTOVER = 30

/*
Just holds a set of []int states, plus a corresponding []int count
per state. Used by determinize().

I have to disable hashCode and use string key to mimic Lucene's
custom hashing function here.
*/
type SortedIntSet struct {
	values []int
	counts []int
	// hashCode   int
	dict       map[int]int // keys need sort
	useTreeMap bool
	state      int
}

func newSortedIntSet(capacity int) *SortedIntSet {
	return &SortedIntSet{
		values: make([]int, 0, capacity),
		counts: make([]int, 0, capacity),
	}
}

// Adds this state ot the set
func (sis *SortedIntSet) incr(num int) {
	if sis.useTreeMap {
		val, ok := sis.dict[num]
		if !ok {
			sis.dict[num] = 1
		} else {
			sis.dict[num] = 1 + val
		}
		return
	}

	for i, v := range sis.values {
		if v == num {
			sis.counts[i]++
			return
		} else if num < v {
			// insert here
			sis.values = append(sis.values[:i], append([]int{num}, sis.values[i:]...)...)
			sis.counts = append(sis.counts[:i], append([]int{1}, sis.counts[i:]...)...)
			return
		}
	}

	// append
	sis.values = append(sis.values, num)
	sis.counts = append(sis.counts, 1)

	if len(sis.values) == TREE_MAP_CUTOVER {
		sis.useTreeMap = true
		for i, v := range sis.values {
			sis.dict[v] = sis.counts[i]
		}
	}
}

// Removes the state from the set, if count decrs to 0
func (sis *SortedIntSet) decr(num int) {
	if sis.useTreeMap {
		count, ok := sis.dict[num]
		assert(ok)
		if count == 1 {
			delete(sis.dict, num)
			// Fall back to simple arrays once we touch zero again
			if len(sis.dict) == 0 {
				sis.useTreeMap = false
				sis.values = sis.values[:0] // reuse slice
				sis.counts = sis.counts[:0] // reuse slice
			}
		} else {
			sis.dict[num] = count - 1
		}
		return
	}

	for i, v := range sis.values {
		if v == num {
			sis.counts[i]--
			if sis.counts[i] == 0 {
				limit := len(sis.values) - 1
				if i < limit {
					sis.values = append(sis.values[:i], sis.values[i+1:]...)
					sis.counts = append(sis.counts[:i], sis.counts[i+1:]...)
				} else {
					sis.values = sis.values[:i]
					sis.counts = sis.counts[:i]
				}
			}
			return
		}
	}

	panic("should not be here!")
}

func (sis *SortedIntSet) computeHash() *FrozenIntSet {
	// do nothing related to hash
	if sis.useTreeMap {
		if size := len(sis.dict); size > len(sis.values) {
			sis.values = make([]int, 0, size)
			sis.counts = make([]int, 0, size)
		}
		for state, _ := range sis.dict {
			sis.values = append(sis.values, state)
		}
		sort.Ints(sis.values) // keys in map are not sorted
	} else {
		// do nothing
	}
	return sis.freeze(sis.state)
}

func (sis *SortedIntSet) freeze(state int) *FrozenIntSet {
	c := make([]int, len(sis.values))
	copy(c, sis.values)
	return newFrozenIntSet(c, state)
}

func (sis *SortedIntSet) String() string {
	var b bytes.Buffer
	b.WriteRune('[')
	for i, v := range sis.values {
		if i > 0 {
			b.WriteRune(' ')
		}
		fmt.Fprintf(&b, "%v:%v", v, sis.counts[i])
	}
	b.WriteRune(']')
	return b.String()
}

type FrozenIntSet struct {
	values []int
	state  int
}

func newFrozenIntSet(values []int, state int) *FrozenIntSet {
	return &FrozenIntSet{values, state}
}

func newFrozenIntSetOf(num, state int) *FrozenIntSet {
	return &FrozenIntSet{
		values: []int{num},
		state:  state,
	}
}

func (fis *FrozenIntSet) String() string {
	var b bytes.Buffer
	b.WriteRune('[')
	for i, v := range fis.values {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteString(strconv.Itoa(v))
	}
	b.WriteRune(']')
	return b.String()
}
