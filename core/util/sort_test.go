package util

import (
	"sort"
	"testing"
)

func TestInPlaceMergeSorter(t *testing.T) {
	data := make([]int, 50)
	for i := 0; i < 50; i++ {
		data[i] = 25 - i
	}
	s := NewInPlaceMergeSorter(sort.IntSlice(data))
	s.Sort(0, 26)
	t.Log(data)
	for i := 0; i < 26; i++ {
		assert(data[i] == i)
	}
	for i := 26; i < 50; i++ {
		assert(data[i] == 25-i)
	}
}
