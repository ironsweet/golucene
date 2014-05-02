package util

import (
	"fmt"
	"sort"
)

// util/Sorter.java

const SORTER_THRESHOLD = 20

// Base class for sorting algorithms implementations.
type Sorter struct {
	sort.Interface
}

func newSorter(arr sort.Interface) *Sorter {
	return &Sorter{arr}
}

func (sorter *Sorter) checkRange(from, to int) {
	assert2(from <= to, fmt.Sprintf("'to' must be >= 'from', got from=%v, and to=%v", from, to))
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (sorter *Sorter) reverse(from, to int) {
	for to--; from < to; from, to = from+1, to-1 {
		sorter.Swap(from, to)
	}
}

func (sorter *Sorter) binarySort(from, to, i int) {
	// log.Printf("Binary sort [%v,%v] at %v", from, to, i)
	for ; i < to; i++ {
		l, h := from, i-1
		for l <= h {
			mid := int(uint(l+h) >> 1)
			if sorter.Less(i, mid) {
				h = mid - 1
			} else {
				l = mid + 1
			}
		}
		switch i - l {
		case 2:
			sorter.Swap(l+1, l+2)
			sorter.Swap(l, l+1)
		case 1:
			sorter.Swap(l, l+1)
		case 0:
		default:
			for j := i; j > l; j-- {
				sorter.Swap(j-1, j)
			}
		}
	}
}

// util/TimSorter.java

const (
	MINRUN        = 32
	RUN_THRESHOLD = 64
	STACKSIZE     = 40 // depends on MINRUN
	MIN_GALLOP    = 7
)

/*
Sorter implementation based on [TimSorter](http://svn.python.org/projects/python/trunk/Objects/listsort.txt) algorithm.

This implementation is especially good at sorting partially-sorted
arrays and sorts small arrays with binary sort.

NOTE: There are a few differences with the original implementation:

1. The extra amount of memory to perform merges is configurable. This
allows small merges to be very fast while large merges will be
performed in-place (slightly slower). You can make sure that the fast
merge routine will always be used by having maxTempSlots equal to
half of the length of the slice of data to sort.

2. Only the fast merge routine can gallop (the one that doesn't
in-place) and it only gallops on the longest slice.
*/
type TimSorter struct {
	*Sorter
	maxTempSlots int
	minRun       int
	to           int
	stackSize    int
	runEnds      []int
}

// Create a new TimSorter
func newTimSorter(arr sort.Interface, maxTempSlots int) *TimSorter {
	return &TimSorter{
		Sorter:       newSorter(arr),
		runEnds:      make([]int, 1+STACKSIZE),
		maxTempSlots: maxTempSlots,
	}
}

// Minimum run length for an array of given length.
func minRun(length int) int {
	assert2(length >= MINRUN, fmt.Sprintf("length=%v", length))
	n := length
	r := 0
	for n >= 64 {
		r = (r | (n & 1))
		n = int(uint(n) >> 1)
	}
	minRun := n + r
	assert(minRun >= MINRUN && minRun <= RUN_THRESHOLD)
	return minRun
}

func (sorter *TimSorter) runEnd(i int) int {
	return sorter.runEnds[sorter.stackSize-i]
}

func (sorter *TimSorter) pushRunLen(length int) {
	sorter.runEnds[sorter.stackSize+1] = sorter.runEnds[sorter.stackSize] + length
	sorter.stackSize++
}

// Compute the length of the next run, make the run sorted and return its length
func (sorter *TimSorter) nextRun() int {
	runBase := sorter.runEnd(0)
	assert2(runBase < sorter.to, fmt.Sprintf("runBase=%v to=%v", runBase, sorter.to))
	if runBase == sorter.to-1 {
		return 1
	}
	o := runBase + 2
	if sorter.Less(runBase+1, runBase) {
		// run must be strictly descending
		for o < sorter.to && sorter.Less(o, o-1) {
			o++
		}
		sorter.reverse(runBase, o)
	} else {
		// run must be non-descending
		for o < sorter.to && !sorter.Less(o, o-1) {
			o++
		}
	}
	runHi := runBase + sorter.minRun
	if sorter.to < runHi {
		runHi = sorter.to
	}
	if o > runHi {
		runHi = o
	}
	sorter.binarySort(runBase, runHi, o)
	for i := runBase; i < runHi-1; i++ {
		assert(!sorter.Less(i+1, i))
	}
	return runHi - runBase
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func (sorter *TimSorter) ensureInvariants() {
	for sorter.stackSize > 1 {
		panic("not implemented yet")
	}
}

func (sorter *TimSorter) exhaustStack() {
	for sorter.stackSize > 1 {
		panic("not implemented yet")
	}
}

func (sorter *TimSorter) reset(from, to int) {
	sorter.stackSize = 0
	for i, _ := range sorter.runEnds {
		sorter.runEnds[i] = 0
	}
	sorter.runEnds[0] = from
	sorter.to = to
	if length := to - from; length <= RUN_THRESHOLD {
		sorter.minRun = length
	} else {
		sorter.minRun = minRun(length)
	}
}

func (sorter *TimSorter) sort(from, to int) {
	sorter.checkRange(from, to)
	if to-from <= 1 {
		return
	}
	sorter.reset(from, to)
	for {
		sorter.ensureInvariants()
		sorter.pushRunLen(sorter.nextRun())
		if sorter.runEnd(0) >= to {
			break
		}
	}
	sorter.exhaustStack()
	assert(sorter.runEnd(0) == to)
}

// util/IntroSorter.java

/*
Sorter implementation based on a variant of the quicksort algorithm
called introsort: when the recursion level exceeds the log of the
length of the array to sort, it falls back to heapsort. This revents
quicksort from running into its worst-case quadratic runtime. Small
arrays are sorted with insertion sort.
*/
type IntroSorter struct {
}

func (s *IntroSorter) Sort(from, to int) {
	panic("not implemented yet")
}
