package automaton

import (
	"fmt"
	. "github.com/balzaczyy/golucene/test_framework/util"
	"testing"
)

func TestMinimizeCase1(t *testing.T) {
	// for i := 0; i < 20; i++ {
	s1 := string([]rune{46, 93, 42, 9794, 64126})
	s2 := string([]rune{46, 453, 46, 91, 64417, 65, 65533, 65533, 93, 46, 42, 93, 124, 124})
	a1 := complement(NewRegExpWithFlag(s1, NONE).ToAutomaton())
	a2 := complement(NewRegExpWithFlag(s2, NONE).ToAutomaton())
	a := minus(a1, a2)
	b := minimize(a)
	assert(sameLanguage(a, b))
	// }
}

func TestMinimizeCase2(t *testing.T) {
	// for i := 0; i < 20; i++ {
	s1 := ")]"
	s2 := "]"
	a1 := complement(NewRegExpWithFlag(s1, NONE).ToAutomaton())
	a2 := complement(NewRegExpWithFlag(s2, NONE).ToAutomaton())
	a := minus(a1, a2)
	b := minimize(a)
	assert(sameLanguage(a, b))
	// }
}

func TestMinimizeCase3(t *testing.T) {
	s := "*.?-"
	r := NewRegExpWithFlag(s, NONE)
	a := r.ToAutomaton()
	b := minimize(a)
	assert(sameLanguage(a, b))
}

func TestRemoveDeadStatesSimple(t *testing.T) {
	a := newEmptyAutomaton()
	a.createState()
	assert(a.numStates() == 1)
	a = removeDeadStates(a)
	assert(a.numStates() == 0)
}

// util/automaton/TestMinimize.java
// This test builds some randomish NFA/DFA and minimizes them.

// The minimal and non-minimal are compared to ensure they are the same.
func TestMinimize(t *testing.T) {
	num := AtLeast(200)
	for i := 0; i < num; i++ {
		a := randomAutomaton(Random())
		fmt.Println("DEBUG4", removeDeadStates(a))
		la := determinize(removeDeadStates(a))
		lb := minimize(a)
		assert(sameLanguage(la, lb))
	}
}

/*
Compare minimized against minimized with a slower, simple impl. We
check not only that they are the same, but that transitions are the
same.
*/
func TestAgainstBrzozowski(t *testing.T) {
	num := AtLeast(200)
	for i := 0; i < num; i++ {
		a := randomAutomaton(Random())
		minimizeSimple(a)
		b := minimize(a)
		assert(sameLanguage(a, b))
		assert(a.numStates() == b.numStates())

		sum1 := 0
		for s := 0; s < a.numStates(); s++ {
			sum1 += a.numTransitions(s)
		}
		sum2 := 0
		for s := 0; s < b.numStates(); s++ {
			sum2 += b.numTransitions(s)
		}

		assert(sum1 == sum2)
	}
}

// n^2 space usage in Hopcroft minimization?
func TestMinimizeHuge(t *testing.T) {
	NewRegExpWithFlag("+-*(A|.....|BC)*", NONE).ToAutomaton()
}
