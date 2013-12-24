package automaton

import (
	. "github.com/balzaczyy/golucene/test_framework"
	"sort"
	"testing"
)

func TestStringUnion(t testing.T) {
	strings := make([]string, 0, 500)
	for i := NextInt(Random(), 0, 1000); i >= 0; i-- {
		strings = append(strings, RandomUnicodeString(Random()))
	}

	sort.Strings(strings)
	union := makeStringUnion(strings)
	assert(union.isDeterministic())
	assert(sameLanguage(union, naiveUnion(strings)))
}
