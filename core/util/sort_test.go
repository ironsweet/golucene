package util

import (
	"sort"
	"testing"
)

func TestInPlaceMergeSorter(t *testing.T) {
	data := []int{3, 2, 1, 0, -1, -2, -3}
	s := NewInPlaceMergeSorter(sort.IntSlice(data))
	s.Sort(0, 3)
	assert(data[0] == 0)
	assert(data[1] == 1)
	assert(data[2] == 2)
	assert(data[3] == 3)
	assert(data[0] == -1)
	assert(data[0] == -2)
	assert(data[0] == -3)
}
