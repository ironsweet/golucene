package automaton

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"log"
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
		a.numberedStates = make([]*State, 0, 4) // pre-allocate 4 slots
		// Lucence use its own append algorithm
		// here I use Go's default append, witout 'upto' as index
		upto := 0
		worklist.PushBack(a.initial)
		visited[a.initial.id] = true
		a.initial.number = upto
		a.numberedStates = append(a.numberedStates, a.initial)
		upto++
		for worklist.Len() > 0 {
			s := worklist.Front().Value.(*State)
			worklist.Remove(worklist.Front())
			for _, t := range s.transitionsArray {
				if _, ok := visited[t.to.id]; !ok {
					visited[t.to.id] = true
					worklist.PushBack(t.to)
					t.to.number = upto
					a.numberedStates = append(a.numberedStates, t.to)
					upto++
				}
			}
		}
	}
	return a.numberedStates
}

func (a *Automaton) setNumberedStates(states []*State) {
	a.numberedStates = states
}

func (a *Automaton) clearNumberedStates() {
	a.numberedStates = nil
}

// 357
// Adds transitions to explicit crash state to ensure that transition
// function is total
func (a *Automaton) totalize() {
	panic("not implemented yet")
}

// Go doesn't have unicode.MinRune which should be 0
const MIN_CODE_POINT = 0

// L400
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

// L512
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

// L566
// Returns a string representation of this automaton.
func (a *Automaton) String() string {
	var b bytes.Buffer
	if a.isSingleton() {
		panic("not implemented yet")
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
	if !clone.isSingleton() {
		panic("not implemented yet")
	}
	clone.clearNumberedStates()
	return &clone
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

// util/automaton/BasicAutomata.java

// Returns a new (deterministic) automaton with the empty language.
func MakeEmpty() *Automaton {
	a := newEmptyAutomaton()
	a.initial = newState()
	a.deterministic = true
	return a
}

// Returns a new (deterministic) automaton that accepts any single codepoint.
func makeAnyChar() *Automaton {
	return makeCharRange(MIN_CODE_POINT, unicode.MaxRune)
}

// Returns a new (deterministic) automaton that accepts a single codepoint of the given value.
func makeChar(c int) *Automaton {
	return &Automaton{
		singleton:     strconv.QuoteRune(rune(c)),
		deterministic: true,
	}
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
	a := &Automaton{
		initial:       newState(),
		deterministic: true,
	}
	s := newState()
	s.accept = true
	if min <= max {
		a.initial.addTransition(newTransitionRange(min, max, s))
	}
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

// Returns an automaton that accepts the concatenation of the
// languages of the given automata.
func concatenate(l []*Automaton) *Automaton {
	panic("not implemented yet")
}

/*
Returns an automaton that accepts the Kleene star (zero or more
concatenated repetitions) of the language of the given automaton.
Never modifies the input automaton language.

Complexity: linear in number of states.
*/
func repeat(a *Automaton) *Automaton {
	panic("not implemented yet")
}

/*
Returns an automaton that accepts min or more concatenated
repetitions of the language of the given automaton.

Complexity: linear in number of states and in min.
*/
func repeatMin(a *Automaton, min int) *Automaton {
	panic("not implemented yet")
}

/*
Returns a (deterministic) automaton that accepts the complement of
the language of the given automaton.

Complexity: linear in number of states (if already deterministic).
*/
func complement(a *Automaton) *Automaton {
	panic("not implemented yet")
}

/*
Returns an automaton that accepts the intersection of the languages
of the given automata. Never modifies the input automata languages.

Complexity: quadratic in number of states.
*/
func intersection(a1, a2 *Automaton) *Automaton {
	panic("not implemented yet")
}

/*
Returns an automaton that accepts the union of the languages of the
given automata.

Complexity: linear in number of states.
*/
func union(l []*Automaton) *Automaton {
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
	a := &Automaton{
		initial:       s,
		deterministic: false,
	}
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
	log.Println("DEBUG", a)
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

// L780
// Returns true if the given automaton accepts no strings
func isEmpty(a *Automaton) bool {
	return !a.isSingleton() && !a.initial.accept && len(a.initial.transitionsArray) == 0
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
			t := append(sis.values[:i], num)
			sis.values = append(t, sis.values[i:]...)
			t = append(sis.counts[:i], 1)
			sis.counts = append(t, sis.counts[i:]...)
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
	panic("not implemented yet")
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
	log.Println("Minimizing using Hopcraft...")
	a.determinize()
	if len(a.initial.transitionsArray) == 1 {
		t := a.initial.transitionsArray[0]
		if t.to == a.initial && t.min == MIN_CODE_POINT && t.max == unicode.MaxRune {
			return
		}
	}
	a.totalize()

	panic("not implemented yet")
}
