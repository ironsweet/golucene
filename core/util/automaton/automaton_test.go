package automaton

import (
	"fmt"
	// . "github.com/balzaczyy/golucene/test_framework"
	// "sort"
	"testing"
)

func TestRegExpToAutomaton(t *testing.T) {
	a := NewRegExp("[^ \t\r\n]+").ToAutomaton()
	fmt.Println(a)
	assert(a.deterministic)
	assert(1 == a.initial.number)
	assert(2 == len(a.numberedStates))
}

// func TestStringUnion(t testing.T) {
// strings := make([]string, 0, 500)
// for i := NextInt(Random(), 0, 1000); i >= 0; i-- {
// 	strings = append(strings, RandomUnicodeString(Random()))
// }

// sort.Strings(strings)
// union := makeStringUnion(strings)
// assert(union.isDeterministic())
// assert(sameLanguage(union, naiveUnion(strings)))
// }
