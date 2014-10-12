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
	return &Sorter{
		Interface: arr,
	}
}

func (sorter *Sorter) checkRange(from, to int) {
	assert2(from <= to, fmt.Sprintf("'to' must be >= 'from', got from=%v, and to=%v", from, to))
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (s *Sorter) mergeInPlace(from, mid, to int) {
	if from == mid || mid == to || !s.Less(mid, mid-1) {
		return
	}
	if to-from == 2 {
		s.Swap(mid-1, mid)
		return
	}
	for !s.Less(mid, from) {
		from++
	}
	for !s.Less(to-1, mid-1) {
		to--
	}
	var first_cut, second_cut int
	var len11, len22 int
	if mid-from > to-mid {
		len11 = int(uint(mid-from) >> 1)
		first_cut = from + len11
		second_cut = s.lower(mid, to, first_cut)
		len22 = second_cut - mid
	} else {
		len22 = int(uint(to-mid) >> 1)
		second_cut = mid + len22
		first_cut = s.upper(from, mid, second_cut)
		// len11 = first_cut - from
	}
	s.rotate(first_cut, mid, second_cut)
	new_mid := first_cut + len22
	s.mergeInPlace(from, first_cut, new_mid)
	s.mergeInPlace(new_mid, second_cut, to)
}

func (s *Sorter) lower(from, to, val int) int {
	size := to - from
	for size > 0 {
		half := int(uint(size) >> 1)
		mid := from + half
		if s.Less(mid, val) {
			from = mid + 1
			size = size - half - 1
		} else {
			size = half
		}
	}
	return from
}

func (s *Sorter) upper(from, to, val int) int {
	size := to - from
	for size > 0 {
		half := int(uint(size) >> 1)
		mid := from + half
		if s.Less(val, mid) {
			size = half
		} else {
			from = mid + 1
			size = size - half - 1
		}
	}
	return from
}

func (s *Sorter) rotate(lo, mid, hi int) {
	assert(lo <= mid && mid <= hi)
	if lo == mid || mid == hi {
		return
	}
	s.doRotate(lo, mid, hi)
}

func (s *Sorter) doRotate(lo, mid, hi int) {
	if mid-lo == hi-mid {
		// happens rarely but saves n/2 swaps
		for mid < hi {
			s.Swap(lo, mid)
			lo++
			mid++
		}
	} else {
		s.reverse(lo, mid)
		s.reverse(mid, hi)
		s.reverse(lo, hi)
	}
}

func (sorter *Sorter) reverse(from, to int) {
	for to--; from < to; from, to = from+1, to-1 {
		sorter.Swap(from, to)
	}
}

func (sorter *Sorter) insertionSort(from, to int) {
	for i := from + 1; i < to; i++ {
		for j := i; j > from; j-- {
			if sorter.Less(j, j-1) {
				sorter.Swap(j-1, j)
			} else {
				break
			}
		}
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

func (s *Sorter) heapSort(from, to int) {
	if to-from <= 1 {
		return
	}
	s.heapify(from, to)
	for end := to - 1; end > from; end-- {
		s.Swap(from, end)
		s.siftDown(from, from, end)
	}
	// TODO remove this
	// for i := from; i < to-1; i++ {
	// 	assert(!s.Less(i+1, i))
	// }
}

func (s *Sorter) heapify(from, to int) {
	for i := s.heapParent(from, to-1); i >= from; i-- {
		s.siftDown(i, from, to)
	}
}

func (s *Sorter) siftDown(i, from, to int) {
	for leftChild := s.heapChild(from, i); leftChild < to; leftChild = s.heapChild(from, i) {
		rightChild := leftChild + 1
		if s.Less(i, leftChild) {
			if rightChild < to && s.Less(leftChild, rightChild) {
				s.Swap(i, rightChild)
				i = rightChild
			} else {
				s.Swap(i, leftChild)
				i = leftChild
			}
		} else if rightChild < to && s.Less(i, rightChild) {
			s.Swap(i, rightChild)
			i = rightChild
		} else {
			break
		}
	}
}

func (s *Sorter) heapParent(from, i int) int {
	return int(uint(i-1-from)>>1) + from
}

func (s *Sorter) heapChild(from, i int) int {
	return ((i - from) << 1) + 1 + from
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

type IntroSorterSPI interface {
	// Save the value at slot i so that it can later be used as a pivot.
	SetPivot(int)
	// Compare the pivot with the slot at j, similarly to Less(int,int).
	PivotLess(int) bool
}

/*
Sorter implementation based on a variant of the quicksort algorithm
called introsort: when the recursion level exceeds the log of the
length of the array to sort, it falls back to heapsort. This prevents
quicksort from running into its worst-case quadratic runtime. Small
arrays are sorted with insertion sort.
*/
type IntroSorter struct {
	spi IntroSorterSPI
	*Sorter
}

func NewIntroSorter(spi IntroSorterSPI, arr sort.Interface) *IntroSorter {
	return &IntroSorter{spi, newSorter(arr)}
}

// 32 - leadingZero(n-1)
func ceilLog2(n int) int {
	assert(n >= 1)
	if n == 1 {
		return 0
	}
	n--
	ans := 0
	for n > 0 {
		n >>= 1
		ans++
	}
	return ans
}

func (s *IntroSorter) Sort(from, to int) {
	s.checkRange(from, to)
	s.quicksort(from, to, ceilLog2(to-from))
}

func (s *IntroSorter) quicksort(from, to, maxDepth int) {
	if to-from < SORTER_THRESHOLD {
		s.insertionSort(from, to)
		// for i := from; i < to-1; i++ {
		// 	assert(!s.Less(i+1, i))
		// }
		return
	}
	if maxDepth--; maxDepth < 0 {
		s.heapSort(from, to)
		// for i := from; i < to-1; i++ {
		// 	assert(!s.Less(i+1, i))
		// }
		return
	}

	mid := (from + to) >> 1

	if s.Less(mid, from) {
		s.Swap(from, mid)
	}

	if s.Less(to-1, mid) {
		s.Swap(mid, to-1)
		if s.Less(mid, from) {
			s.Swap(from, mid)
		}
	}

	left := from + 1
	right := to - 2

	s.spi.SetPivot(mid)
	for {
		for s.spi.PivotLess(right) {
			right--
		}

		for left < right && !s.spi.PivotLess(left) {
			left++
		}

		if left < right {
			s.Swap(left, right)
			right--
		} else {
			break
		}
	}

	s.quicksort(from, left+1, maxDepth)
	s.quicksort(left+1, to, maxDepth)
	// for i := from; i < to-1; i++ {
	// 	assert(!s.Less(i+1, i))
	// }
}

// util/InPlaceMergeSorter.java

/*
Sorter implementation absed on the merge-sort algorithm that merges
in place (no extra memory will be allocated). Small arrays are sorter
with insertion sort.
*/
type InPlaceMergeSorter struct {
	*Sorter
}

func NewInPlaceMergeSorter(impl sort.Interface) *InPlaceMergeSorter {
	return &InPlaceMergeSorter{
		Sorter: newSorter(impl),
	}
}

func (s *InPlaceMergeSorter) Sort(from, to int) {
	s.checkRange(from, to)
	s.mergeSort(from, to)
}

func (s *InPlaceMergeSorter) mergeSort(from, to int) {
	if to-from < SORTER_THRESHOLD {
		s.insertionSort(from, to)
	} else {
		mid := int((uint(from) + uint(to)) >> 1)
		s.mergeSort(from, mid)
		s.mergeSort(mid, to)
		s.mergeInPlace(from, mid, to)
	}
}
