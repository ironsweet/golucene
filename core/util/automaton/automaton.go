package automaton

import (
	"math/big"
)

// util/automaton/Automaton.java

/*
Represents an automaton and all its states and transitions. States
are integers and must be created using {@link #createState}.  Mark a
state as an accept state using {@link #setAccept}.  Add transitions
using {@link #addTransition}.  Each state must have all of its
transitions added at once; if this is too restrictive then use
{@link Automaton.Builder} instead.  State 0 is always the
initial state.  Once a state is finished, either
because you've starting adding transitions to another state or you
call {@link #finishState}, then that states transitions are sorted
(first by min, then max, then dest) and reduced (transitions with
adjacent labels going to the same dest are combined).
*/
type Automaton struct {
	curState      int
	states        []int // 2x
	transitions   []int // 3x
	isAccept      *big.Int
	deterministic bool
}

func newEmptyAutomaton() *Automaton {
	return &Automaton{
		deterministic: true,
		curState:      -1,
		isAccept:      big.NewInt(0),
	}
}

/* Create a new state. */
func (a *Automaton) createState() int {
	state := len(a.states) / 2
	a.states = append(a.states, -1, 0)
	return state
}

/* Set or clear this state as an accept state. */
func (a *Automaton) setAccept(state int, accept bool) {
	assert2(state < a.numStates(), "state=%v is out of bounds (numStates=%v)", state, a.numStates())
	if accept {
		a.isAccept.SetBit(a.isAccept, state, 1)
	} else {
		a.isAccept.SetBit(a.isAccept, state, 0)
	}
}

/* Add a new transition with min = max = label. */
func (a *Automaton) addTransition(source, dest, label int) {
	a.addTransitionRange(source, dest, label, label)
}

/* Add a new transition with the specified source, dest, min, max. */
func (a *Automaton) addTransitionRange(source, dest, min, max int) {
	assert(len(a.transitions)%3 == 0)
	assert2(source < a.numStates(), "source=%v is out of bounds (maxState is %v)", source, a.numStates()-1)
	assert2(dest < a.numStates(), "dest=%v is out of bounds (maxState is %v)", dest, a.numStates()-1)

	if a.curState != source {
		if a.curState != -1 {
			a.finishCurrentState()
		}

		// move to next source:
		a.curState = source
		assert2(a.states[2*a.curState] == -1, "from state (%v) already had transitions added", source)
		assert(a.states[2*a.curState+1] == 0)
		a.states[2*a.curState] = len(a.transitions)
	}

	a.transitions = append(a.transitions, dest, min, max)

	// increment transition count for this state
	a.states[2*a.curState+1]++
}

/* Freezes the last state, sorting and reducing the transitions. */
func (a *Automaton) finishCurrentState() {
	panic("niy")
}

/*
Finishes the current state; call this once you are done adding
transitions for a state. This is automatically called if you start
adding transitions to a new source state, but for the last state you
add, you need to call this method yourself.
*/
func (a *Automaton) finishState() {
	if a.curState != -1 {
		a.finishCurrentState()
		a.curState = -1
	}
}

func (a *Automaton) numStates() int {
	return len(a.states) / 2
}

// Go doesn't have unicode.MinRune which should be 0
const MIN_CODE_POINT = 0

type AutomatonBuilder struct {
}
