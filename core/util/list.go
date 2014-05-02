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
	arr sort.Interface
}

func newListIntroSorter(data sort.Interface) *ListIntroSorter {
	return &ListIntroSorter{new(IntroSorter), data}
}
