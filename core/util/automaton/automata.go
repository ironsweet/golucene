package automaton

import (
	"unicode"
)

// util/automaton/Automata.java

// Returns a new (deterministic) automaton with the empty language.
func MakeEmpty() *Automaton {
	a := newEmptyAutomaton()
	a.finishState()
	return a
}

// Returns a new (deterministic) automaton that accepts only the empty string.
func makeEmptyString() *Automaton {
	panic("not implemented yet")
}

// Returns a new (deterministic) automaton that accepts any single codepoint.
func makeAnyChar() *Automaton {
	return makeCharRange(MIN_CODE_POINT, unicode.MaxRune)
}

// Returns a new (deterministic) automaton that accepts a single codepoint of the given value.
func makeChar(c int) *Automaton {
	return makeCharRange(c, c)
}

/*
Returns a new (deterministic) automaton that accepts a single rune
whose value is in the given interval (including both end points)
*/
func makeCharRange(min, max int) *Automaton {
	if min > max {
		return MakeEmpty()
	}

	a := newEmptyAutomaton()
	s1 := a.createState()
	s2 := a.createState()
	a.setAccept(s2, true)
	a.addTransitionRange(s1, s2, min, max)
	a.finishState()
	return a
}

// L237
// Returns a new (deterministic) automaton that accepts the single given string
func makeString(s string) *Automaton {
	a := newEmptyAutomaton()
	lastState := a.createState()
	for _, r := range s {
		state := a.createState()
		a.addTransitionRange(lastState, state, int(r), int(r))
		lastState = state
	}

	a.setAccept(lastState, true)
	a.finishState()

	assert(a.deterministic)
	assert(!hasDeadStates(a))

	return a
}

// L271
/*
Returns a new (deterministic and minimal) automaton that accepts the
union of the given collection of []byte representing UTF-8 encoded
strings.
*/
func makeStringUnion(utf8Strings [][]byte) *Automaton {
	if len(utf8Strings) == 0 {
		return MakeEmpty()
	}
	return buildDaciukMihovAutomaton(utf8Strings)
}
