package automaton

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"sort"
	"unicode"
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
	isAccept      *util.OpenBitSet
	deterministic bool
}

func newEmptyAutomaton() *Automaton {
	return &Automaton{
		deterministic: true,
		curState:      -1,
		isAccept:      util.NewOpenBitSet(),
	}
}

func (a *Automaton) String() string {
	return fmt.Sprintf("{curState=%v,states=%v,transitions=%v,isAccept=%v,%v}",
		a.curState, a.states, a.transitions, a.isAccept, a.deterministic)
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
		a.isAccept.Set(int64(state))
	} else {
		a.isAccept.Clear(int64(state))
	}
}

/*
Sugar to get all transitions for all states. This is object-heavy;
it's better to iterate state by state instead.
*/
func (a *Automaton) sortedTransitions() [][]*Transition {
	numStates := a.numStates()
	transitions := make([][]*Transition, numStates)
	for s := 0; s < numStates; s++ {
		numTransitions := a.numTransitions(s)
		transitions[s] = make([]*Transition, numTransitions)
		for t := 0; t < numTransitions; t++ {
			transition := newTransition()
			a.transition(s, t, transition)
			transitions[s][t] = transition
		}
	}
	return transitions
}

/* Returns true if this state is an accept state. */
func (a *Automaton) IsAccept(state int) bool {
	return a.isAccept.Get(int64(state))
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

/*
Add a [virtual] epsilon transition between source and dest. Dest
state must already have all transitions added because this method
simply copies those same transitions over to source.
*/
func (a *Automaton) addEpsilon(source, dest int) {
	t := newTransition()
	count := a.initTransition(dest, t)
	for i := 0; i < count; i++ {
		a.nextTransition(t)
		a.addTransitionRange(source, t.dest, t.min, t.max)
	}
	if a.IsAccept(dest) {
		a.setAccept(source, true)
	}
}

/*
Copies over all state/transition from other. The state numbers are
sequentially assigned (appended).
*/
func (a *Automaton) copy(other *Automaton) {
	// bulk copy and then fixup the state pointers
	stateOffset := a.numStates()
	a.states = append(a.states, other.states...)
	for i := 0; i < len(other.states); i += 2 {
		if a.states[stateOffset*2+i] != -1 {
			a.states[stateOffset*2+i] += len(a.transitions)
		}
	}
	otherAcceptState := other.isAccept
	for state := otherAcceptState.NextSetBit(0); state != -1; state = otherAcceptState.NextSetBit(state + 1) {
		a.setAccept(stateOffset+int(state), true)
	}

	// bulk copy and then fixup dest for each transition
	transOffset := len(a.transitions)
	a.transitions = append(a.transitions, other.transitions...)
	for i := 0; i < len(other.transitions); i += 3 {
		a.transitions[transOffset+i] += stateOffset
	}

	if !other.deterministic {
		a.deterministic = false
	}
}

/* Freezes the last state, sorting and reducing the transitions. */
func (a *Automaton) finishCurrentState() {
	numTransitions := a.states[2*a.curState+1]
	assert(numTransitions > 0)

	offset := a.states[2*a.curState]
	start := offset / 3
	util.NewInPlaceMergeSorter(destMinMaxSorter(a.transitions)).Sort(start, start+numTransitions)

	// reduce any "adjacent" transitions:
	upto, min, max, dest := 0, -1, -1, -1

	for i := 0; i < numTransitions; i++ {
		tDest := a.transitions[offset+3*i]
		tMin := a.transitions[offset+3*i+1]
		tMax := a.transitions[offset+3*i+2]

		if dest == tDest {
			if tMin <= max+1 {
				if tMax > max {
					max = tMax
				}
			} else {
				if dest != -1 {
					a.transitions[offset+3*upto] = dest
					a.transitions[offset+3*upto+1] = min
					a.transitions[offset+3*upto+2] = max
					upto++
				}
				min, max = tMin, tMax
			}
		} else {
			if dest != -1 {
				a.transitions[offset+3*upto] = dest
				a.transitions[offset+3*upto+1] = min
				a.transitions[offset+3*upto+2] = max
				upto++
			}
			dest, min, max = tDest, tMin, tMax
		}
	}

	if dest != -1 {
		// last transition
		a.transitions[offset+3*upto] = dest
		a.transitions[offset+3*upto+1] = min
		a.transitions[offset+3*upto+2] = max
		upto++
	}

	a.transitions = a.transitions[:len(a.transitions)-(numTransitions-upto)*3]
	a.states[2*a.curState+1] = upto

	// sort transitions by min/max/dest:
	util.NewInPlaceMergeSorter(minMaxDestSorter(a.transitions)).Sort(start, start+upto)

	if a.deterministic && upto > 1 {
		lastMax := a.transitions[offset+2]
		for i := 1; i < upto; i++ {
			min = a.transitions[offset+3*i+1]
			if min <= lastMax {
				a.deterministic = false
				break
			}
			lastMax = a.transitions[offset+3*i+2]
		}
	}
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

/* How many states this automaton has. */
func (a *Automaton) numStates() int {
	return len(a.states) / 2
}

/* How many transitions this state has. */
func (a *Automaton) numTransitions(state int) int {
	if count := a.states[2*state+1]; count != -1 {
		return count
	}
	return 0
}

type destMinMaxSorter []int

func (s destMinMaxSorter) Len() int {
	panic("niy")
}

func (s destMinMaxSorter) Swap(i, j int) {
	iStart, jStart := 3*i, 3*j
	for n := 0; n < 3; n++ {
		s[iStart+n], s[jStart+n] = s[jStart+n], s[iStart+n]
	}
}

func (s destMinMaxSorter) Less(i, j int) bool {
	iStart := 3 * i
	jStart := 3 * j

	// first dest:
	iDest := s[iStart]
	jDest := s[jStart]
	if iDest < jDest {
		return true
	} else if iDest > jDest {
		return false
	}

	// then min:
	iMin := s[iStart+1]
	jMin := s[jStart+1]
	if iMin < jMin {
		return true
	} else if iMin > jMin {
		return false
	}

	// then max:
	iMax := s[iStart+2]
	jMax := s[jStart+2]
	return iMax < jMax
}

type minMaxDestSorter []int

func (s minMaxDestSorter) Len() int {
	panic("niy")
}

func (s minMaxDestSorter) Swap(i, j int) {
	iStart, jStart := 3*i, 3*j
	for n := 0; n < 3; n++ {
		s[iStart+n], s[jStart+n] = s[jStart+n], s[iStart+n]
	}
}

func (s minMaxDestSorter) Less(i, j int) bool {
	iStart, jStart := 3*i, 3*j

	iMin, jMin := s[iStart+1], s[jStart+1]
	if iMin < jMin {
		return true
	} else if iMin > jMin {
		return false
	}

	iMax, jMax := s[iStart+2], s[jStart+2]
	if iMax < jMax {
		return true
	} else if iMax > jMax {
		return false
	}

	iDest, jDest := s[iStart], s[jStart]
	return iDest < jDest
}

/*
Initialize the provided Transition to iterate through all transitions
leaving the specified state. You must call nextTransition() to get
each transition. Returns the number of transitions leaving this tate.
*/
func (a *Automaton) initTransition(state int, t *Transition) int {
	assert2(state < a.numStates(), "state=%v nextState=%v", state, a.numStates())
	t.source = state
	t.transitionUpto = a.states[2*state]
	return a.numTransitions(state)
}

/* Iterate to the next transition after the provided one */
func (a *Automaton) nextTransition(t *Transition) {
	// make sure there is still a transition left
	assert((t.transitionUpto + 3 - a.states[2*t.source]) <= 3*a.states[2*t.source+1])
	t.dest = a.transitions[t.transitionUpto]
	t.min = a.transitions[t.transitionUpto+1]
	t.max = a.transitions[t.transitionUpto+2]
	t.transitionUpto += 3
}

/*
Fill the provided Transition with the index'th transition leaving the
specified state.
*/
func (a *Automaton) transition(state, index int, t *Transition) {
	i := a.states[2*state] + 3*index
	t.source = state
	t.dest = a.transitions[i]
	t.min = a.transitions[i+1]
	t.max = a.transitions[i+2]
}

// L563
/* Returns sorted array of all interval start points. */
func (a *Automaton) startPoints() []int {
	pointset := make(map[int]bool)
	pointset[MIN_CODE_POINT] = true
	// fmt.Println("getStartPoints")
	for s := 0; s < len(a.states); s += 2 {
		trans := a.states[s]
		limit := trans + 3*a.states[s+1]
		// fmt.Printf("  state=%v trans=%v limit=%v\n", s/2, trans, limit)
		for trans < limit {
			min, max := a.transitions[trans+1], a.transitions[trans+2]
			// fmt.Printf("    min=%v\n", min)
			pointset[min] = true
			if max < unicode.MaxRune {
				pointset[max+1] = true
			}
			trans += 3
		}
	}
	var points []int
	for m, _ := range pointset {
		points = append(points, m)
	}
	sort.Ints(points)
	return points
}

/* Performs lookup in transitions, assuming determinism. */
func (a *Automaton) step(state, label int) int {
	assert(state >= 0)
	assert(label >= 0)
	if 2*state >= len(a.states) {
		return -1 // invalid state
	}
	trans := a.states[2*state]
	limit := trans + 3*a.states[2*state+1]
	// TODO binary search
	for trans < limit {
		dest, min, max := a.transitions[trans], a.transitions[trans+1], a.transitions[trans+2]
		if min <= label && label <= max {
			return dest
		}
		trans += 3
	}
	return -1
}

// Go doesn't have unicode.MinRune which should be 0
const MIN_CODE_POINT = 0

type AutomatonBuilder struct {
	transitions []int
	a           *Automaton
}

func newAutomatonBuilder() *AutomatonBuilder {
	return &AutomatonBuilder{
		a: newEmptyAutomaton(),
	}
}

func (b *AutomatonBuilder) addTransitionRange(source, dest, min, max int) {
	b.transitions = append(b.transitions, source, dest, min, max)
}

type srcMinMaxDestSorter []int

func (s srcMinMaxDestSorter) Len() int {
	panic("niy")
}

func (s srcMinMaxDestSorter) Swap(i, j int) {
	iStart, jStart := 4*i, 4*j
	for n := 0; n < 4; n++ {
		s[iStart+n], s[jStart+n] = s[jStart+n], s[iStart+n]
	}
}

func (s srcMinMaxDestSorter) Less(i, j int) bool {
	iStart, jStart := 4*i, 4*j

	iSrc, jSrc := s[iStart], s[jStart]
	if iSrc < jSrc {
		return true
	} else if iSrc > jSrc {
		return false
	}

	iMin, jMin := s[iStart+2], s[jStart+2]
	if iMin < jMin {
		return true
	} else if iMin > jMin {
		return false
	}

	iMax, jMax := s[iStart+3], s[jStart+3]
	if iMax < jMax {
		return true
	} else if iMax > jMax {
		return false
	}

	iDest, jDest := s[iStart+1], s[jStart+1]
	return iDest < jDest
}

/* Compiles all added states and transitions into a new Automaton and returns it. */
func (b *AutomatonBuilder) finish() *Automaton {
	// fmt.Printf("LA.Builder.finish: count=%v\n", len(b.transitions)/4)
	// fmt.Println("finish pending")
	util.NewInPlaceMergeSorter(srcMinMaxDestSorter(b.transitions)).Sort(0, len(b.transitions)/4)
	for upto := 0; upto < len(b.transitions); upto += 4 {
		b.a.addTransitionRange(
			b.transitions[upto],
			b.transitions[upto+1],
			b.transitions[upto+2],
			b.transitions[upto+3],
		)
	}

	b.a.finishState()
	return b.a
}

func (b *AutomatonBuilder) createState() int {
	return b.a.createState()
}

func (b *AutomatonBuilder) setAccept(state int, accept bool) {
	b.a.setAccept(state, accept)
}

func (b *AutomatonBuilder) isAccept(state int) bool {
	return b.a.IsAccept(state)
}

func (b *AutomatonBuilder) copy(other *Automaton) {
	offset := b.a.numStates()
	otherNumStates := other.numStates()
	for s := 0; s < otherNumStates; s++ {
		newState := b.createState()
		b.setAccept(newState, other.IsAccept(s))
	}
	t := newTransition()
	for s := 0; s < otherNumStates; s++ {
		count := other.initTransition(s, t)
		for i := 0; i < count; i++ {
			other.nextTransition(t)
			b.addTransitionRange(offset+s, offset+t.dest, t.min, t.max)
		}
	}
}
