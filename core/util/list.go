package util

import (
	"sort"
)

// util/CollectionUtil.java

func IntroSort(data sort.Interface) {
	if data.Len() <= 1 {
		return
	}
	newListIntroSorter(data).Sort(0, data.Len())
}

type ListIntroSorter struct {
	*IntroSorter
}

func newListIntroSorter(data sort.Interface) *ListIntroSorter {
	ans := new(ListIntroSorter)
	ans.IntroSorter = NewIntroSorter(ans, data)
	return ans
}

func (s *ListIntroSorter) SetPivot(i int) {
	panic("not implemented yet")
}

func (s *ListIntroSorter) PivotLess(j int) bool {
	panic("not implemented yet")
}
