package automaton

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math/big"
	"sort"
	"strconv"
	"unicode"
)

// util/automaton/Automaton.java

// Minimize always flag.
const MINIMIZE_ALWAYS = false

// Selects whether operations may modify the input automata (default: false)
const ALLOW_MUTATION = false

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
	// Singleton string. Empty if not applicable.
	singleton string
	// cached
	numberedStates []*State
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

// L185
func (a *Automaton) checkMinimizeAlways() {
	if MINIMIZE_ALWAYS {
		minimize(a)
	}
}

func (a *Automaton) isSingleton() bool {
	return a.singleton != ""
}

// L269
func (a *Automaton) NumberedStates() []*State {
	if a.numberedStates == nil {
		a.ExpandSingleton()
		visited := make(map[int]bool)
		worklist := list.New()
		states := make([]*State, 0, 4) // pre-allocate 4 slots
		// Lucence use its own append algorithm
		// here I use Go's default append, witout 'upto' as index
		upto := 0
		worklist.PushBack(a.initial)
		visited[a.initial.id] = true
		a.initial.number = upto
		states = append(a.numberedStates, a.initial)
		upto++
		for worklist.Len() > 0 {
			s := worklist.Front().Value.(*State)
			worklist.Remove(worklist.Front())
			for _, t := range s.transitionsArray {
				if _, ok := visited[t.to.id]; !ok {
					visited[t.to.id] = true
					worklist.PushBack(t.to)
					t.to.number = upto
					states = append(states, t.to)
					upto++
				}
			}
		}
		a.numberedStates = states[:upto]
	}
	return a.numberedStates
}

func (a *Automaton) setNumberedStates(states []*State) {
	a.numberedStates = states
}

func (a *Automaton) clearNumberedStates() {
	a.numberedStates = nil
}

// Returns the set of reachable accept states.
func (a *Automaton) acceptStates() map[int]*State {
	a.ExpandSingleton()
	accepts := make(map[int]*State)
	visited := make(map[int]*State)
	worklist := list.New()
	worklist.PushBack(a.initial)
	visited[a.initial.id] = a.initial
	for worklist.Len() > 0 {
		s := worklist.Front().Value.(*State)
		worklist.Remove(worklist.Front())
		if s.accept {
			accepts[s.id] = s
		}
		for _, t := range s.transitionsArray {
			if _, ok := visited[t.to.id]; !ok {
				visited[t.to.id] = t.to
				worklist.PushBack(t.to)
			}
		}
	}
	return accepts
}

// Adds transitions to explicit crash state to ensure that transition
// function is total
func (a *Automaton) totalize() {
	s := newState()
	s.addTransition(newTransitionRange(MIN_CODE_POINT, unicode.MaxRune, s))
	for _, p := range a.NumberedStates() {
		maxi := MIN_CODE_POINT
		p.sortTransitions(compareByMinMaxThenDest)
		for _, t := range p.transitionsArray {
			if t.min > maxi {
				p.addTransition(newTransitionRange(maxi, t.min-1, s))
			}
			if t.max+1 > maxi {
				maxi = t.max + 1
			}
		}
		if maxi <= unicode.MaxRune {
			p.addTransition(newTransitionRange(maxi, unicode.MaxRune, s))
		}
	}
	a.clearNumberedStates()
}

// Go doesn't have unicode.MinRune which should be 0
const MIN_CODE_POINT = 0

// L390
// Reduces this automaton. An automaton is "reduced" by combining
// overlapping and adjacent edge intervals with the same destination.
func (a *Automaton) reduce() {
	states := a.NumberedStates()
	if a.isSingleton() {
		return
	}
	for _, s := range states {
		s.reduce()
	}
}

// Returns sorted array of all interval start points.
func (a *Automaton) startPoints() []int {
	states := a.NumberedStates()
	pointset := make(map[int]bool)
	pointset[MIN_CODE_POINT] = true
	for _, s := range states {
		for _, t := range s.transitionsArray {
			pointset[t.min] = true
			if t.max < unicode.MaxRune {
				pointset[t.max+1] = true
			}
		}
	}
	points := make([]int, 0, len(pointset))
	for k, _ := range pointset {
		points = append(points, k)
	}
	sort.Ints(points)
	return points
}

// Returns the set of live states. A state is "live" if an accept
// state is reachable from it.
func (a *Automaton) liveStates() []*State {
	states := a.NumberedStates()
	live := make(map[int]*State)
	for _, q := range states {
		if q.accept {
			live[q.id] = q
		}
	}
	// map[int]map[int]*state
	stateMap := make([]map[int]*State, len(states))
	for i, _ := range stateMap {
		stateMap[i] = make(map[int]*State)
	}
	for _, s := range states {
		for _, t := range s.transitionsArray {
			stateMap[t.to.number][s.id] = s
		}
	}
	worklist := list.New()
	for _, s := range live {
		worklist.PushBack(s)
	}
	for worklist.Len() > 0 {
		s := worklist.Front().Value.(*State)
		worklist.Remove(worklist.Front())
		for _, p := range stateMap[s.number] {
			if _, ok := live[p.id]; !ok {
				live[p.id] = p
				worklist.PushBack(p)
			}
		}
	}

	ans := make([]*State, 0, len(live))
	for _, s := range live {
		ans = append(ans, s)
	}
	return ans
}

// Removes transitions to dead states and calls reduce().
// (A state is "dead" if no accept state is reachable from it.)
func (a *Automaton) removeDeadTransitions() {
	states := a.NumberedStates()
	if a.isSingleton() {
		return
	}
	live := a.liveStates()

	liveSet := big.NewInt(0)
	for _, s := range live {
		setBit(liveSet, s.number)
	}

	for _, s := range states {
		// filter out transitions to dead states:
		upto := 0
		for _, t := range s.transitionsArray {
			if liveSet.Bit(t.to.number) == 1 {
				s.transitionsArray[upto] = t
				upto++
			}
		}
		// release remaining references for GC
		for i, limit := upto, len(s.transitionsArray); i < limit; i++ {
			s.transitionsArray[i] = nil
		}
		s.transitionsArray = s.transitionsArray[:upto]
	}
	for i, s := range live {
		s.number = i
	}
	if len(live) > 0 {
		a.setNumberedStates(live)
	} else {
		// sneaky corner case -- if machine accepts no strings
		a.clearNumberedStates()
	}
	a.reduce()
}

// Returns a sorted array of transitions for each state (and sets
// state numbers).
func (a *Automaton) sortedTransitions() [][]*Transition {
	states := a.NumberedStates()
	transitions := make([][]*Transition, len(states))
	for _, s := range states {
		s.sortTransitions(compareByMinMaxThenDest)
		transitions[s.number] = s.transitionsArray
		assert(s.transitionsArray != nil)
	}
	return transitions
}

// Expands singleton representation to normal representation. Does
// nothing if not in singleton representation.
func (a *Automaton) ExpandSingleton() {
	if a.isSingleton() {
		p := newState()
		a.initial = p
		for _, cp := range a.singleton {
			q := newState()
			p.addTransition(newTransition(int(cp), q))
			p = q
		}
		p.accept = true
		a.deterministic = true
		a.singleton = ""
	}
}

// Returns the number of states in this automaton.
func (a *Automaton) NumberOfStates() int {
	if a.isSingleton() {
		return len([]rune(a.singleton))
	}
	return len(a.NumberedStates())
}

// Returns the number of transitions in this automaton. This number
// is counted as the total number of edges, where one edge may be a
// character interval.
func (a *Automaton) NumberOfTransitions() int {
	if a.isSingleton() {
		return len([]rune(a.singleton))
	}
	c := 0
	for _, s := range a.NumberedStates() {
		c += len(s.transitionsArray)
	}
	return c
}

// Returns a string representation of this automaton.
func (a *Automaton) String() string {
	var b bytes.Buffer
	if a.isSingleton() {
		b.WriteString("singleton: ")
		for _, c := range a.singleton {
			appendCharString(int(c), &b)
		}
		b.WriteRune('\n')
	} else {
		states := a.NumberedStates()
		fmt.Fprintf(&b, "initial state: %v\n", a.initial.number)
		for _, s := range states {
			b.WriteString(s.String())
		}
	}
	return b.String()
}

// L614
// Returns a clone of this automaton, expands if singleton.
func (a *Automaton) cloneExpanded() *Automaton {
	clone := a.Clone()
	clone.ExpandSingleton()
	return clone
}

// Returns a clone of this automaton unless ALLOW_MUTATION is set,
// expands if singleton.
func (a *Automaton) cloneExpandedIfRequired() *Automaton {
	if ALLOW_MUTATION {
		a.ExpandSingleton()
		return a
	}
	return a.cloneExpanded()
}

// Returns a clone of this automaton.
func (a *Automaton) Clone() *Automaton {
	clone := *a
	if !a.isSingleton() {
		m := make(map[int]*State)
		states := a.NumberedStates()
		for _, s := range states {
			m[s.id] = newState()
		}
		for _, s := range states {
			p := m[s.id]
			p.accept = s.accept
			if s == a.initial {
				clone.initial = p
			}
			for _, t := range s.transitionsArray {
				p.addTransition(newTransitionRange(t.min, t.max, m[t.to.id]))
			}
		}
	}
	clone.clearNumberedStates()
	return &clone
}

// Returns a clone of this automaton, or this automaton itself if
// ALLOW_MUTATION flag is set.
func (a *Automaton) cloneIfRequired() *Automaton {
	if ALLOW_MUTATION {
		return a
	}
	return a.Clone()
}

func (a *Automaton) optional() *Automaton {
	return optional(a)
}

func (a *Automaton) repeat() *Automaton {
	return repeat(a)
}

func (a *Automaton) repeatMin(min int) *Automaton {
	return repeatMin(a, min)
}

func (a *Automaton) complement() *Automaton {
	return complement(a)
}

func (a *Automaton) intersection(b *Automaton) *Automaton {
	return intersection(a, b)
}

func (a *Automaton) determinize() {
	determinize(a)
}

func (a *Automaton) isEmptyString() bool {
	return isEmptyString(a)
}

// util/automaton/State.java

var next_id int

// Automaton state
type State struct {
	accept           bool
	transitionsArray []*Transition
	// numTransitions   int
	number int
	id     int
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
}

// Adds ann outgoing transition.
func (s *State) addTransition(t *Transition) {
	s.transitionsArray = append(s.transitionsArray, t)
}

// Performs lookup in transitions, assuming determinisim.
func (s *State) step(c int) *State {
	assert(c >= 0)
	for _, t := range s.transitionsArray {
		if t.min <= c && c <= t.max {
			return t.to
		}
	}
	return nil
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

/*
Virtually adds an epsilon transition to the target state. This is
implemented by copying all transitions from given state to this
state, and if it is an accept state then set accept for this state.
*/
func (s *State) addEpsilon(to *State) {
	if to.accept {
		s.accept = true
	}
	for _, t := range to.transitionsArray {
		s.transitionsArray = append(s.transitionsArray, t)
	}
}

// Reduce this state. A state is "reduced" by combining overlapping
// and adjacent edge intervals with same destination
func (s *State) reduce() {
	if len(s.transitionsArray) <= 1 {
		return
	}
	s.sortTransitions(compareByDestThenMinMax)
	var p *State
	min, max, upto := -1, -1, 0
	for _, t := range s.transitionsArray {
		if p == t.to {
			if t.min <= max+1 {
				if t.max > max {
					max = t.max
				}
			} else {
				if p != nil {
					s.transitionsArray[upto] = newTransitionRange(min, max, p)
					upto++
				}
				min, max = t.min, t.max
			}
		} else {
			if p != nil {
				s.transitionsArray[upto] = newTransitionRange(min, max, p)
				upto++
			}
			p, min, max = t.to, t.min, t.max
		}
	}

	if p != nil {
		s.transitionsArray[upto] = newTransitionRange(min, max, p)
		upto++
	}
	for i, limit := upto, len(s.transitionsArray); i < limit; i++ {
		s.transitionsArray[i] = nil
	}
	s.transitionsArray = s.transitionsArray[:upto]
}

type TransitionArraySorter struct {
	arr []*Transition
	by  func(t1, t2 *Transition) bool
}

func (s TransitionArraySorter) Len() int           { return len(s.arr) }
func (s TransitionArraySorter) Less(i, j int) bool { return s.by(s.arr[i], s.arr[j]) }
func (s TransitionArraySorter) Swap(i, j int)      { s.arr[i], s.arr[j] = s.arr[j], s.arr[i] }

// Returns sorted list of outgoing transitions.
// Sorts transitions array in-place
func (s *State) sortTransitions(f func(t1, t2 *Transition) bool) {
	// merge-sort seems to perform better on already sorted arrays
	if len(s.transitionsArray) > 1 {
		util.TimSort(TransitionArraySorter{s.transitionsArray, f})
	}
}

// Returns string describing this state. Normally invoked via Automaton.
func (s *State) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "state %v", s.number)
	if s.accept {
		b.WriteString(" [accept]")
	} else {
		b.WriteString(" [reject]")
	}
	b.WriteString("\n")
	for _, t := range s.transitionsArray {
		fmt.Fprintf(&b, "  %v\n", t.String())
	}
	return b.String()
}

// util/automaton/Transition.java

/*
Automaton transition.

A transition, which belongs to a source state, consists of a Unicode
codepoint interval and a destination state.
*/
type Transition struct {
	// CLASS INVARIANT: min <= max
	min, max int
	to       *State
}

// Constructs a new singleton interval transition.
func newTransition(c int, to *State) *Transition {
	assert(c >= 0)
	return &Transition{c, c, to}
}

// Constructs a new transition. Both end points are included in the interval
func newTransitionRange(min, max int, to *State) *Transition {
	assert(min >= 0)
	assert(max >= 0)
	if max < min {
		max, min = min, max
	}
	return &Transition{min, max, to}
}

func appendCharString(c int, b *bytes.Buffer) {
	if c >= 0x21 && c <= 0x7e && c != '\\' && c != '"' {
		b.WriteRune(rune(c))
	} else {
		b.WriteString("\\\\U")
		s := fmt.Sprintf("%x", c)
		if c < 0x10 {
			fmt.Fprintf(b, "0000000%v", s)
		} else if c < 0x100 {
			fmt.Fprintf(b, "000000%v", s)
		} else if c < 0x1000 {
			fmt.Fprintf(b, "00000%v", s)
		} else if c < 0x10000 {
			fmt.Fprintf(b, "0000%v", s)
		} else if c < 0x100000 {
			fmt.Fprintf(b, "000%v", s)
		} else if c < 0x1000000 {
			fmt.Fprintf(b, "00%v", s)
		} else if c < 0x10000000 {
			fmt.Fprintf(b, "0%v", s)
		} else {
			b.WriteString(s)
		}
	}
}

// Returns a string describing this transition. Normally invoked via State.
func (t *Transition) String() string {
	var b bytes.Buffer
	appendCharString(t.min, &b)
	if t.min != t.max {
		b.WriteString("-")
		appendCharString(t.max, &b)
	}
	fmt.Fprintf(&b, " -> %v", t.to.number)
	return b.String()
}

var compareByDestThenMinMax = func(t1, t2 *Transition) bool {
	if t1.to != t2.to {
		return t1.to.number < t2.to.number
	}
	return t1.min < t2.min || t1.min == t2.min && t1.max > t2.max
}

var compareByMinMaxThenDest = func(t1, t2 *Transition) bool {
	if t1.min < t2.min {
		return true
	}
	if t1.min > t2.min {
		return false
	}
	if t1.max > t2.max {
		return true
	}
	if t1.max < t2.max {
		return false
	}
	return t1.to != t2.to && t1.to.number < t2.to.number
}

// util/automaton/BasicAutomata.java

// Returns a new (deterministic) automaton with the empty language.
func MakeEmpty() *Automaton {
	a := newEmptyAutomaton()
	a.initial = newState()
	a.deterministic = true
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
	var b bytes.Buffer
	b.WriteRune(rune(c))
	a := newEmptyAutomaton()
	a.singleton = b.String()
	a.deterministic = true
	return a
}

/*
Returns a new (deterministic) automaton that accepts a single
codepoint whose value is in the given interval (including both end
points)
*/
func makeCharRange(min, max int) *Automaton {
	if min == max {
		return makeChar(min)
	}
	a := newEmptyAutomaton()
	a.initial = newState()
	a.deterministic = true
	s := newState()
	s.accept = true
	if min <= max {
		a.initial.addTransition(newTransitionRange(min, max, s))
	}
	return a
}

// L237
// Returns a new (deterministic) automaton that accepts the single given string
func makeString(s string) *Automaton {
	a := newEmptyAutomaton()
	a.singleton = s
	a.deterministic = true
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

// util/automaton/BasicOperation.java
// Basic automata operations.

/*
Returns an automaton that accepts the concatenation of the
languages of the given automata.

Complexity: linear in number of states.
*/
func concatenate(a1, a2 *Automaton) *Automaton {
	if a1.isSingleton() && a2.isSingleton() {
		return makeString(a1.singleton + a2.singleton)
	}
	if isEmpty(a1) || isEmpty(a2) {
		return MakeEmpty()
	}
	// adding epsilon transitions with the NFA concatenation algorithm
	// in this case always produces a rsulting DFA, preventing expensive
	// redundant determinize() calls for this common case.
	deterministic := a1.isSingleton() && a2.deterministic
	if a1 == a2 {
		a1 = a1.cloneExpanded()
		a2 = a2.cloneExpanded()
	} else {
		a1 = a1.cloneExpandedIfRequired()
		a2 = a2.cloneExpandedIfRequired()
	}
	for _, s := range a1.acceptStates() {
		s.accept = false
		s.addEpsilon(a2.initial)
	}
	a1.deterministic = deterministic
	a1.clearNumberedStates()
	a1.checkMinimizeAlways()
	return a1
}

/*
Returns an automaton that accepts the concatenation of the
languages of the given automata.

Complexity: linear in total number of states.
*/
func concatenateN(l []*Automaton) *Automaton {
	if len(l) == 0 {
		return makeEmptyString()
	}
	allSingleton := true
	for _, a := range l {
		if !a.isSingleton() {
			allSingleton = false
			break
		}
	}
	if allSingleton {
		var b bytes.Buffer
		for _, a := range l {
			b.WriteString(a.singleton)
		}
		return makeString(b.String())
	}
	for _, a := range l {
		if isEmpty(a) {
			return MakeEmpty()
		}
	}
	all := make(map[*Automaton]bool)
	hasAliases := false
	for _, a := range l {
		if _, ok := all[a]; ok {
			hasAliases = true
			break
		} else {
			all[a] = true
		}
	}
	b := l[0]
	if hasAliases {
		b = b.cloneExpanded()
	} else {
		b = b.cloneExpandedIfRequired()
	}
	ac := b.acceptStates()
	first := true
	for _, a := range l {
		if first {
			first = false
		} else {
			if a.isEmptyString() {
				continue
			}
			aa := a
			if hasAliases {
				aa = aa.cloneExpanded()
			} else {
				aa = aa.cloneExpandedIfRequired()
			}
			ns := aa.acceptStates()
			for _, s := range ac {
				s.accept = false
				s.addEpsilon(aa.initial)
				if s.accept {
					ns[s.id] = s
				}
			}
			ac = ns
		}
	}
	b.deterministic = false
	b.clearNumberedStates()
	b.checkMinimizeAlways()
	return b
}

/*
Returns an automaton that accepts the union of the empty string and
the language of the given automaton.

Complexity: linear in number of states.
*/
func optional(a *Automaton) *Automaton {
	s := newState()
	s.addEpsilon(a.initial)
	s.accept = true
	a = a.cloneExpandedIfRequired()
	a.initial = s
	a.deterministic = false
	a.clearNumberedStates()
	a.checkMinimizeAlways()
	return a
}

/*
Returns an automaton that accepts the Kleene star (zero or more
concatenated repetitions) of the language of the given automaton.
Never modifies the input automaton language.

Complexity: linear in number of states.
*/
func repeat(a *Automaton) *Automaton {
	a = a.cloneExpandedIfRequired()
	s := newState()
	s.accept = true
	s.addEpsilon(a.initial)
	for _, p := range a.acceptStates() {
		p.addEpsilon(s)
	}
	a.initial = s
	a.deterministic = false
	a.clearNumberedStates()
	a.checkMinimizeAlways()
	return a
}

/*
Returns an automaton that accepts min or more concatenated
repetitions of the language of the given automaton.

Complexity: linear in number of states and in min.
*/
func repeatMin(a *Automaton, min int) *Automaton {
	if min == 0 {
		return repeat(a)
	}
	as := make([]*Automaton, 0, min+1)
	for min > 0 {
		as = append(as, a)
		min--
	}
	as = append(as, repeat(a))
	return concatenateN(as)
}

/*
Returns a (deterministic) automaton that accepts the complement of
the language of the given automaton.

Complexity: linear in number of states (if already deterministic).
*/
func complement(a *Automaton) *Automaton {
	a = a.cloneExpandedIfRequired()
	a.determinize()
	a.totalize()
	for _, p := range a.NumberedStates() {
		p.accept = !p.accept
	}
	a.removeDeadTransitions()
	return a
}

/*
Returns a (deterministic) automaton that accepts the intersection of
the language of a1 and the complement of the language of a2. As a
side-effect, the automata may be determinized, if not already
deterministic.

Complexity: quadratic in number of states (if already deterministic).
*/
func minus(a1, a2 *Automaton) *Automaton {
	if isEmpty(a1) || a1 == a2 {
		return MakeEmpty()
	}
	if isEmpty(a2) {
		return a1.cloneIfRequired()
	}
	if a1.isSingleton() {
		if run(a2, a1.singleton) {
			return MakeEmpty()
		} else {
			return a1.cloneIfRequired()
		}
	}
	return intersection(a1, a2.complement())
}

// Pair of states.
type StatePair struct{ s, s1, s2 *State }

/*
Returns an automaton that accepts the intersection of the languages
of the given automata. Never modifies the input automata languages.

Complexity: quadratic in number of states.
*/
func intersection(a1, a2 *Automaton) *Automaton {
	if a1.isSingleton() {
		if run(a2, a1.singleton) {
			return a1.cloneIfRequired()
		}
		return MakeEmpty()
	}
	if a2.isSingleton() {
		if run(a1, a2.singleton) {
			return a2.cloneIfRequired()
		}
		return MakeEmpty()
	}
	if a1 == a2 {
		return a1.cloneIfRequired()
	}
	transitions1 := a1.sortedTransitions()
	transitions2 := a2.sortedTransitions()
	c := newEmptyAutomaton()
	worklist := list.New()
	newstates := make(map[string]*StatePair)
	hash := func(p *StatePair) string {
		return fmt.Sprintf("%v/%v", p.s1.id, p.s2.id)
	}
	p := &StatePair{c.initial, a1.initial, a2.initial}
	worklist.PushBack(p)
	newstates[hash(p)] = p
	for worklist.Len() > 0 {
		p = worklist.Front().Value.(*StatePair)
		worklist.Remove(worklist.Front())
		p.s.accept = (p.s1.accept && p.s2.accept)
		t1 := transitions1[p.s1.number]
		t2 := transitions2[p.s2.number]
		for n1, b2, t1Len := 0, 0, len(t1); n1 < t1Len; n1++ {
			t2Len := len(t2)
			for b2 < t2Len && t2[b2].max < t1[n1].min {
				b2++
			}
			for n2 := b2; n2 < t2Len && t1[n1].max >= t2[n2].min; n2++ {
				if t2[n2].max >= t1[n1].min {
					q := &StatePair{nil, t1[n1].to, t2[n2].to}
					r, ok := newstates[hash(q)]
					if !ok {
						q.s = newState()
						worklist.PushBack(q)
						newstates[hash(q)] = q
						r = q
					}
					min := or(t1[n1].min > t2[n2].min, t1[n1].min, t2[n2].min).(int)
					max := or(t1[n1].max < t2[n2].max, t1[n1].max, t2[n2].max).(int)
					p.s.addTransition(newTransitionRange(min, max, r.s))
				}
			}
		}
	}
	c.deterministic = (a1.deterministic && a2.deterministic)
	c.removeDeadTransitions()
	c.checkMinimizeAlways()
	return c
}

/*
Returns true if these two automata accept exactly the same language.
This is a costly computation! Note also that a1 and a2 will be
determinized as a side effect.
*/
func sameLanguage(a1, a2 *Automaton) bool {
	if a1 == a2 {
		return true
	}
	if a1.isSingleton() && a2.isSingleton() {
		return a1.singleton == a2.singleton
	}
	if a1.isSingleton() {
		// subsetOf is faster if the first automaton is a singleton
		return subsetOf(a1, a2) && subsetOf(a2, a1)
	}
	return subsetOf(a2, a1) && subsetOf(a1, a2)
}

/*
Returns true if the language of a1 is a subset of the language of a2.
As a side-effect, a2 is determinized if not already marked as
deterministic.
*/
func subsetOf(a1, a2 *Automaton) bool {
	if a1 == a2 {
		return true
	}
	if a1.isSingleton() {
		if a2.isSingleton() {
			return a1.singleton == a2.singleton
		}
		return run(a2, a1.singleton)
	}
	a2.determinize()
	transitions1 := a1.sortedTransitions()
	transitions2 := a2.sortedTransitions()
	worklist := list.New()
	visited := make(map[string]*StatePair)
	hash := func(p *StatePair) string {
		return fmt.Sprintf("%v/%v", p.s1.id, p.s2.id)
	}
	p := &StatePair{nil, a1.initial, a2.initial}
	worklist.PushBack(p)
	visited[hash(p)] = p
	for worklist.Len() > 0 {
		p = worklist.Front().Value.(*StatePair)
		worklist.Remove(worklist.Front())
		if p.s1.accept && !p.s2.accept {
			return false
		}
		t1 := transitions1[p.s1.number]
		t2 := transitions2[p.s2.number]
		for n1, b2, t1Len := 0, 0, len(t1); n1 < t1Len; n1++ {
			t2Len := len(t2)
			for b2 < t2Len && t2[b2].max < t1[n1].min {
				b2++
			}
			min1, max1 := t1[n1].min, t1[n1].max

			for n2 := b2; n2 < t2Len && t1[n1].max >= t2[n2].min; n2++ {
				if t2[n2].min > min1 {
					return false
				}
				if t2[n2].max < unicode.MaxRune {
					min1 = t2[n2].max + 1
				} else {
					min1, max1 = unicode.MaxRune, MIN_CODE_POINT
				}
				q := &StatePair{nil, t1[n1].to, t2[n2].to}
				if _, ok := visited[hash(q)]; !ok {
					worklist.PushBack(q)
					visited[hash(q)] = q
				}
			}
			if min1 <= max1 {
				return false
			}
		}
	}
	return true
}

/*
Returns an automaton that accepts the union of the languages of the
given automta.

Complexity: linear in number of states.
*/
func union(a1, a2 *Automaton) *Automaton {
	if a1 == a2 ||
		(a1.isSingleton() && a2.isSingleton() && a1.singleton == a2.singleton) {
		a1.cloneIfRequired()
	}
	if a1 == a2 {
		a1 = a1.cloneExpanded()
		a2 = a2.cloneExpanded()
	} else {
		a1 = a1.cloneExpandedIfRequired()
		a2 = a2.cloneExpandedIfRequired()
	}
	s := newState()
	s.addEpsilon(a1.initial)
	s.addEpsilon(a2.initial)
	a1.initial = s
	a1.deterministic = false
	a1.clearNumberedStates()
	a1.checkMinimizeAlways()
	return a1
}

/*
Returns an automaton that accepts the union of the languages of the
given automata.

Complexity: linear in number of states.
*/
func unionN(l []*Automaton) *Automaton {
	// ids := make(map[int]bool)
	hasAliases := false
	s := newState()
	for _, b := range l {
		if isEmpty(b) {
			continue
		}
		bb := b
		if hasAliases {
			// bb = bb.cloneExpanded()
			panic("not implemented yet")
		} else {
			bb = bb.cloneExpandedIfRequired()
		}
		s.addEpsilon(bb.initial)
	}
	a := newEmptyAutomaton()
	a.initial = s
	a.deterministic = false
	a.clearNumberedStates()
	a.checkMinimizeAlways()
	return a
}

// Holds all transitions that start on this int point, or end at this
// point-1
type PointTransitions struct {
	point  int
	ends   []*Transition
	starts []*Transition
}

func newPointTransitions() *PointTransitions {
	return &PointTransitions{
		ends:   make([]*Transition, 0, 2),
		starts: make([]*Transition, 0, 2),
	}
}

func (pt *PointTransitions) reset(point int) {
	pt.point = point
	pt.ends = pt.ends[:0]     // reuse slice
	pt.starts = pt.starts[:0] // reuse slice
}

const HASHMAP_CUTOVER = 30

type PointTransitionSet struct {
	points  []*PointTransitions
	dict    map[int]*PointTransitions
	useHash bool
}

func newPointTransitionSet() *PointTransitionSet {
	return &PointTransitionSet{
		points:  make([]*PointTransitions, 0, 5),
		dict:    make(map[int]*PointTransitions),
		useHash: false,
	}
}

func (pts *PointTransitionSet) next(point int) *PointTransitions {
	// 1st time we are seeing this point
	p := newPointTransitions()
	pts.points = append(pts.points, p)
	p.reset(point)
	return p
}

func (pts *PointTransitionSet) find(point int) *PointTransitions {
	if pts.useHash {
		p, ok := pts.dict[point]
		if !ok {
			p = pts.next(point)
			pts.dict[point] = p
		}
		return p
	}

	for _, p := range pts.points {
		if p.point == point {
			return p
		}
	}

	p := pts.next(point)
	if len(pts.points) == HASHMAP_CUTOVER {
		// switch to hash map on the fly
		assert(len(pts.dict) == 0)
		for _, v := range pts.points {
			pts.dict[v.point] = v
		}
		pts.useHash = true
	}
	return p
}

func (pts *PointTransitionSet) reset() {
	if pts.useHash {
		pts.dict = make(map[int]*PointTransitions)
		pts.useHash = false
	}
	pts.points = pts.points[:0] // reuse slice
}

type PointTransitionsArray []*PointTransitions

func (a PointTransitionsArray) Len() int           { return len(a) }
func (a PointTransitionsArray) Less(i, j int) bool { return a[i].point < a[j].point }
func (a PointTransitionsArray) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (pts *PointTransitionSet) sort() {
	// Tim sort performs well on already sorted arrays:
	if len(pts.points) > 0 {
		util.TimSort(PointTransitionsArray(pts.points))
	}
}

func (pts *PointTransitionSet) add(t *Transition) {
	p1 := pts.find(t.min)
	p1.starts = append(p1.starts, t)

	p2 := pts.find(1 + t.max)
	p2.ends = append(p2.ends, t)
}

func (pts *PointTransitionSet) String() string {
	panic("not implemented yet")
}

/*
Determinizes the given automaton.

Split the code points in ranges, and merge overlapping states.

Worst case complexity: exponential in number of states.
*/
func determinize(a *Automaton) {
	if a.deterministic || a.isSingleton() {
		return
	}

	allStates := a.NumberedStates()

	// subset construction
	initAccept := a.initial.accept
	initNumber := a.initial.number
	a.initial = newState()
	initialset := newFrozenIntSet([]int{initNumber}, a.initial)

	worklist := list.New()
	newstate := make(map[string]*FrozenIntSet)

	worklist.PushBack(initialset)

	a.initial.accept = initAccept
	newstate[initialset.String()] = initialset

	newStatesArray := make([]*State, 0, 5)
	newStatesArray = append(newStatesArray, a.initial)
	a.initial.number = 0

	// like map[int]*PointTransitions
	points := newPointTransitionSet()

	// like sorted map[int]int
	statesSet := newSortedIntSet(5)

	for worklist.Len() > 0 {
		s := worklist.Front().Value.(*FrozenIntSet)
		worklist.Remove(worklist.Front())

		// Collate all outgoing transitions by min/1+max
		for _, v := range s.values {
			s0 := allStates[v]
			for _, t := range s0.transitionsArray {
				points.add(t)
			}
		}

		if len(points.points) == 0 {
			// No outgoing transitions -- skip it
			continue
		}

		points.sort()

		lastPoint := -1
		accCount := 0

		r := s.state
		for _, v := range points.points {
			point := v.point

			if len(statesSet.values) > 0 {
				assert(lastPoint != -1)

				hashKey := statesSet.computeHash().String()

				var q *State
				ss, ok := newstate[hashKey]
				if !ok {
					q = newState()
					p := statesSet.freeze(q)
					worklist.PushBack(p)
					newStatesArray = append(newStatesArray, q)
					q.number = len(newStatesArray) - 1
					q.accept = accCount > 0
					newstate[p.String()] = p
				} else {
					q = ss.state
					assert2(q.accept == (accCount > 0), fmt.Sprintf(
						"accCount=%v vs existing accept=%v states=%v",
						accCount, q.accept, statesSet))
				}

				r.addTransition(newTransitionRange(lastPoint, point-1, q))
			}

			// process transitions that end on this point
			// (closes an overlapping interval)
			for _, t := range v.ends {
				statesSet.decr(t.to.number)
				if t.to.accept {
					accCount--
				}
			}
			v.ends = v.ends[:0] // reuse slice

			// process transitions that start on this point
			// (opens a new interval)
			for _, t := range v.starts {
				statesSet.incr(t.to.number)
				if t.to.accept {
					accCount++
				}
			}
			v.starts = v.starts[:0] // reuse slice
			lastPoint = point
		}
		points.reset()
		assert2(len(statesSet.values) == 0, fmt.Sprintf("upto=%v", len(statesSet.values)))
	}
	a.deterministic = true
	a.setNumberedStates(newStatesArray)
}

// L775
// Returns true if the given automaton accepts the empty string and
// nothing else.
func isEmptyString(a *Automaton) bool {
	if a.isSingleton() {
		return len(a.singleton) == 0
	}
	return a.initial.accept && len(a.initial.transitionsArray) == 0
}

// Returns true if the given automaton accepts no strings.
func isEmpty(a *Automaton) bool {
	return !a.isSingleton() && !a.initial.accept && len(a.initial.transitionsArray) == 0
}

/*
Returns true if the given string is accepted by the autmaton.

Complexity: linear in the length of the string.

Note: for fll performance, use the RunAutomation class.
*/
func run(a *Automaton, s string) bool {
	if a.isSingleton() {
		return s == a.singleton
	}
	if a.deterministic {
		p := a.initial
		for _, ch := range s {
			q := p.step(int(ch))
			if q == nil {
				return false
			}
			p = q
		}
		return p.accept
	}
	// states := a.NumberedStates()
	panic("not implemented yet")
}

// util/automaton/SortedIntSet.java

// If we hold more than this many states, we swtich from O(N*2)
// linear ops to O(N log(N)) TreeMap
const TREE_MAP_CUTOVER = 30

/*
Just holds a set of []int states, plus a corresponding []int count
per state. Used by determinize().

I have to disable hashCode and use string key to mimic Lucene's
custom hashing function here.
*/
type SortedIntSet struct {
	values []int
	counts []int
	// hashCode   int
	dict       map[int]int // keys need sort
	useTreeMap bool
	state      *State
}

func newSortedIntSet(capacity int) *SortedIntSet {
	return &SortedIntSet{
		values: make([]int, 0, capacity),
		counts: make([]int, 0, capacity),
	}
}

// Adds this state ot the set
func (sis *SortedIntSet) incr(num int) {
	// log.Println("DEBUG incr", sis, num)
	if sis.useTreeMap {
		val, ok := sis.dict[num]
		if !ok {
			sis.dict[num] = 1
		} else {
			sis.dict[num] = 1 + val
		}
		return
	}

	for i, v := range sis.values {
		if v == num {
			sis.counts[i]++
			return
		} else if num < v {
			// insert here
			sis.values = append(sis.values[:i], append([]int{num}, sis.values[i:]...)...)
			sis.counts = append(sis.counts[:i], append([]int{1}, sis.counts[i:]...)...)
			return
		}
	}

	// append
	sis.values = append(sis.values, num)
	sis.counts = append(sis.counts, 1)

	if len(sis.values) == TREE_MAP_CUTOVER {
		sis.useTreeMap = true
		for i, v := range sis.values {
			sis.dict[v] = sis.counts[i]
		}
	}
}

// Removes the state from the set, if count decrs to 0
func (sis *SortedIntSet) decr(num int) {
	// log.Println("DEBUG decr", sis, num)
	if sis.useTreeMap {
		count, ok := sis.dict[num]
		assert(ok)
		if count == 1 {
			delete(sis.dict, num)
			// Fall back to simple arrays once we touch zero again
			if len(sis.dict) == 0 {
				sis.useTreeMap = false
				sis.values = sis.values[:0] // reuse slice
				sis.counts = sis.counts[:0] // reuse slice
			}
		} else {
			sis.dict[num] = count - 1
		}
		return
	}

	for i, v := range sis.values {
		if v == num {
			sis.counts[i]--
			if sis.counts[i] == 0 {
				limit := len(sis.values) - 1
				if i < limit {
					sis.values = append(sis.values[:i], sis.values[i+1:]...)
					sis.counts = append(sis.counts[:i], sis.counts[i+1:]...)
				} else {
					sis.values = sis.values[:i]
					sis.counts = sis.counts[:i]
				}
			}
			return
		}
	}

	panic("should not be here!")
}

func (sis *SortedIntSet) computeHash() *FrozenIntSet {
	// do nothing related to hash
	if sis.useTreeMap {
		if size := len(sis.dict); size > len(sis.values) {
			sis.values = make([]int, 0, size)
			sis.counts = make([]int, 0, size)
		}
		for state, _ := range sis.dict {
			sis.values = append(sis.values, state)
		}
	} else {
		// do nothing
	}
	return sis.freeze(sis.state)
}

func (sis *SortedIntSet) freeze(state *State) *FrozenIntSet {
	c := make([]int, len(sis.values))
	copy(c, sis.values)
	sort.Ints(c)
	return newFrozenIntSet(c, state)
}

func (sis *SortedIntSet) String() string {
	var b bytes.Buffer
	b.WriteRune('[')
	for i, limit := 0, len(sis.values); i < limit; i++ {
		if i > 0 {
			b.WriteRune(' ')
		}
		fmt.Fprintf(&b, "%v:%v", sis.values[i], sis.counts[i])
	}
	b.WriteRune(']')
	return b.String()
}

type FrozenIntSet struct {
	values []int
	state  *State
}

func newFrozenIntSet(values []int, state *State) *FrozenIntSet {
	return &FrozenIntSet{values, state}
}

func (fis *FrozenIntSet) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, v := range fis.values {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(strconv.Itoa(v))
	}
	buf.WriteString("]")
	return buf.String()
}

// util/automaton/SpecialOperations.java

// Finds the largest entry whose value is less than or equal to c, or
// 0 if there is no such entry.
func findIndex(c int, points []int) int {
	a, b := 0, len(points)
	for b-a > 1 {
		d := int(uint(a+b) >> 1)
		if points[d] > c {
			b = d
		} else if points[d] < c {
			a = d
		} else {
			return d
		}
	}
	return a
}

// Reverses the language of the given (non-singleton) automaton while
// returning the set of new initial states.
func reverse(a *Automaton) map[int]*State {
	a.ExpandSingleton()
	// reverse all edges
	m := make(map[int]map[string]*Transition)
	hash := func(t *Transition) string {
		return fmt.Sprintf("%v/%v/%v", t.min, t.max, t.to.id)
	}
	states := a.NumberedStates()
	accept := make(map[int]*State)
	for _, s := range states {
		if s.accept {
			accept[s.id] = s
		}
	}
	for _, r := range states {
		m[r.id] = make(map[string]*Transition)
		r.accept = false
	}
	for _, r := range states {
		for _, t := range r.transitionsArray {
			tt := newTransitionRange(t.min, t.max, r)
			m[t.to.id][hash(tt)] = tt
		}
	}
	for _, r := range states {
		tr := m[r.id]
		arr := make([]*Transition, 0, len(tr))
		for _, t := range tr {
			arr = append(arr, t)
		}
		r.transitionsArray = arr
	}
	// make new initial + final states
	a.initial.accept = true
	a.initial = newState()
	for _, r := range accept {
		a.initial.addEpsilon(r) // ensures that all initial states are reachable
	}
	a.deterministic = false
	a.clearNumberedStates()
	return accept
}

// util/automaton/MinimizationOperations.java

// Minimizes (and determinizes if not already deterministic) the
// given automaton
func minimize(a *Automaton) {
	if !a.isSingleton() {
		minimizeHopcroft(a)
	}
}

// Minimizes the given automaton using Hopcroft's alforithm.
func minimizeHopcroft(a *Automaton) {
	// log.Println("Minimizing using Hopcraft...")
	a.determinize()
	if len(a.initial.transitionsArray) == 1 {
		t := a.initial.transitionsArray[0]
		if t.to == a.initial && t.min == MIN_CODE_POINT && t.max == unicode.MaxRune {
			return
		}
	}
	a.totalize()

	// initialize data structure
	sigma := a.startPoints()
	states := a.NumberedStates()
	sigmaLen, statesLen := len(sigma), len(states)
	reverse := make([][][]*State, statesLen)
	for i, _ := range reverse {
		reverse[i] = make([][]*State, sigmaLen)
	}
	partition := make([]map[int]*State, statesLen)
	splitblock := make([][]*State, statesLen)
	block := make([]int, statesLen)
	active := make([][]*StateList, statesLen)
	for i, _ := range active {
		active[i] = make([]*StateList, sigmaLen)
	}
	active2 := make([][]*StateListNode, statesLen)
	for i, _ := range active2 {
		active2[i] = make([]*StateListNode, sigmaLen)
	}
	pending := list.New()
	pending2 := big.NewInt(0) // sigmaLen * statesLen bits
	split := big.NewInt(0)    // statesLen bits
	refine := big.NewInt(0)   // statesLen bits
	refine2 := big.NewInt(0)  // statesLen bits
	for q, _ := range splitblock {
		splitblock[q] = make([]*State, 0)
		partition[q] = make(map[int]*State)
		for x, _ := range active[q] {
			active[q][x] = &StateList{}
		}
	}
	// find initial partition and reverse edges
	for q, qq := range states {
		j := or(qq.accept, 0, 1).(int)
		partition[j][qq.id] = qq
		block[q] = j
		for x, v := range sigma {
			r := reverse[qq.step(v).number]
			if r[x] == nil {
				r[x] = make([]*State, 0)
			}
			r[x] = append(r[x], qq)
		}
	}
	// initialize active sets
	for j := 0; j <= 1; j++ {
		for x := 0; x < sigmaLen; x++ {
			for _, qq := range partition[j] {
				if reverse[qq.number][x] != nil {
					active2[qq.number][x] = active[j][x].add(qq)
				}
			}
		}
	}
	// initialize pending
	for x := 0; x < sigmaLen; x++ {
		j := or(active[0][x].size <= active[1][x].size, 0, 1).(int)
		pending.PushBack(&IntPair{j, x})
		setBit(pending2, x*statesLen+j)
	}
	// process pending until fixed point
	k := 2
	for pending.Len() > 0 {
		ip := pending.Front().Value.(*IntPair)
		pending.Remove(pending.Front())
		p, x := ip.n1, ip.n2
		clearBit(pending2, x*statesLen+p)
		// find states that need to be split off their blocks
		for m := active[p][x].first; m != nil; m = m.next {
			if r := reverse[m.q.number][x]; r != nil {
				for _, s := range r {
					if i := s.number; split.Bit(i) == 0 {
						setBit(split, i)
						j := block[i]
						splitblock[j] = append(splitblock[j], s)
						if refine2.Bit(j) == 0 {
							setBit(refine2, j)
							setBit(refine, j)
						}
					}
				}
			}
		}
		// refine blocks
		// TODO Go's Big int may not be efficient here.
		for j, bitsLen := 0, refine.BitLen(); j < bitsLen; j++ {
			if refine.Bit(j) == 0 {
				continue
			}

			sb := splitblock[j]
			if len(sb) < len(partition[j]) {
				b1, b2 := partition[j], partition[k]
				for _, s := range sb {
					delete(b1, s.id)
					b2[s.id] = s
					block[s.number] = k
					for c, sn := range active2[s.number] {
						if sn != nil && sn.sl == active[j][c] {
							sn.remove()
							active2[s.number][c] = active[k][c].add(s)
						}
					}
				}
				// update pending
				for c, _ := range active[j] {
					aj := active[j][c].size
					ak := active[k][c].size
					ofs := c * statesLen
					if pending2.Bit(ofs+j) == 0 && 0 < aj && aj <= ak {
						setBit(pending2, ofs+j)
						pending.PushBack(&IntPair{j, c})
					} else {
						setBit(pending2, ofs+k)
						pending.PushBack(&IntPair{k, c})
					}
				}
				k++
			}
			clearBit(refine2, j)
			for _, s := range sb {
				clearBit(split, s.number)
			}
			splitblock[j] = splitblock[j][:0] // clear sb
		}
		refine = big.NewInt(0) // not quite efficient
	}
	// make a new state for each equivalence class, set initial state
	newstates := make([]*State, k)
	for n, _ := range newstates {
		s := newState()
		newstates[n] = s
		for _, q := range partition[n] {
			if q == a.initial {
				a.initial = s
			}
			s.accept = q.accept
			s.number = q.number // select representative
			q.number = n
		}
	}
	// build transitions and set acceptance
	for _, s := range newstates {
		s.accept = states[s.number].accept
		for _, t := range states[s.number].transitionsArray {
			s.addTransition(newTransitionRange(t.min, t.max, newstates[t.to.number]))
		}
	}
	a.clearNumberedStates()
	a.removeDeadTransitions()
}

func setBit(n *big.Int, bitIndex int) {
	n.SetBit(n, bitIndex, 1)
}

func clearBit(n *big.Int, bitIndex int) {
	n.SetBit(n, bitIndex, 0)
}

func bitSetToString(n *big.Int) string {
	var b bytes.Buffer
	b.WriteRune('{')
	for i, bitsLen := 0, n.BitLen(); i < bitsLen; i++ {
		if n.Bit(i) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%v, ", i)
	}
	b.WriteRune('}')
	return b.String()
}

func or(cond bool, v1, v2 interface{}) interface{} {
	if cond {
		return v1
	}
	return v2
}

type IntPair struct{ n1, n2 int }

type StateList struct {
	size        int
	first, last *StateListNode
}

func (sl *StateList) add(q *State) *StateListNode {
	return newStateListNode(q, sl)
}

type StateListNode struct {
	q          *State
	next, prev *StateListNode
	sl         *StateList
}

func newStateListNode(q *State, sl *StateList) *StateListNode {
	ans := &StateListNode{q: q, sl: sl}
	sl.size++
	if sl.size == 1 {
		sl.first, sl.last = ans, ans
	} else {
		sl.last.next = ans
		ans.prev = sl.last
		sl.last = ans
	}
	return ans
}

func (node *StateListNode) remove() {
	node.sl.size--
	if node.sl.first == node {
		node.sl.first = node.next
	} else {
		node.prev.next = node.next
	}
	if node.sl.last == node {
		node.sl.last = node.prev
	} else {
		node.next.prev = node.prev
	}
}
