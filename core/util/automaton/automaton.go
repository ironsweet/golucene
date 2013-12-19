package automaton

// util/automaton/Automaton.java

/*
Finite-state automaton with regular expression operations.

Class invariants:

1. An automaton is either represented explicitly (with State and
Transition object) or with a singleton string (see Singleton() and
expandSingleton()) in case the automaton is known to accept exactly
one string. (Implicitly, all states and transitions of an automaton
are reachable from its initial state.)
2. Automata are always reduced (see Reduce()) and have no transitions
to dead states (see RemoveDeadTransitions()).
3. If an automaton is nondeterministic, then IsDeterministic()
returns false (but the converse is not required).
4. Automata provided as input to operations are generally assumed to
be disjoint.

If the states or transitions are manipulated manually, the
RestoreInvariant() and SetDeterministic(bool) methods should be used
afterwards to restore representation invariants that are assumed by
the built-in automata operations.

Note: This class has internal mutable state and is not thread safe.
It is the caller's responsibility to ensure any necessary
synchronization if you wish to use the same Automaton from multiple
threads. In general it is instead recommended to use a RunAutomaton
for multithreaded matching: it is immutable, thread safe, and much
faster.
*/
type Automaton struct {
	// Initial state of this automation.
	initial *State
	// If true, then this automaton is definitely deterministic (i.e.,
	// there are no choices for any run, but a run may crash).
	deterministic bool
}

// Constructs a new automaton that accepts the empty language. Using
// this constructor, automata can be constructed manually from State
// and Transition objects.
func newAutomatonWithState(initial *State) *Automaton {
	return &Automaton{initial: initial, deterministic: true}
}

func newEmptyAutomaton() *Automaton {
	return newAutomatonWithState(newState())
}

// util/automaton/State.java

var next_id int

// Automaton state
type State struct {
	accept           bool
	transitionsArray []*Transition
	numTransitions   int

	number int

	id int
}

// Constructs a new state. Initially, the new state is a reject state.
func newState() *State {
	s := &State{}
	s.resetTransitions()
	s.id = next_id
	next_id++
	return s
}

// Resets transition set.
func (s *State) resetTransitions() {
	s.transitionsArray = make([]*Transition, 0)
	s.numTransitions = 0
}

// util/automaton/Transition.java

/*
Automaton transition.

A transition, which belongs to a source state, consists of a Unicode
codepoint interval and a destination state.
*/
type Transition struct {
}

// util/automaton/BasicAutomata.java

// Returns a new (deterministic) automaton with the empty language.
func MakeEmpty() *Automaton {
	a := newEmptyAutomaton()
	a.initial = newState()
	a.deterministic = true
	return a
}
