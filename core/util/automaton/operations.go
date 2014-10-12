package automaton

import (
	"container/list"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"unicode"
)

// Basic automata operations.

/*
Returns an automaton that accepts the concatenation of the
languages of the given automata.

Complexity: linear in total number of states.
*/
func concatenate(a1, a2 *Automaton) *Automaton {
	return concatenateN([]*Automaton{a1, a2})
}

/*
Returns an automaton that accepts the concatenation of the
languages of the given automata.

Complexity: linear in total number of states.
*/
func concatenateN(l []*Automaton) *Automaton {
	ans := newEmptyAutomaton()

	// first pass: create all states
	for _, a := range l {
		if a.numStates() == 0 {
			ans.finishState()
			return ans
		}
		numStates := a.numStates()
		for s := 0; s < numStates; s++ {
			ans.createState()
		}
	}

	// second pass: add transitions, carefully linking accept
	// states of A to init state of next A:
	stateOffset := 0
	t := newTransition()
	for i, a := range l {
		numStates := a.numStates()

		var nextA *Automaton
		if i < len(l)-1 {
			nextA = l[i+1]
		}

		for s := 0; s < numStates; s++ {
			numTransitions := a.initTransition(s, t)
			for j := 0; j < numTransitions; j++ {
				a.nextTransition(t)
				ans.addTransitionRange(stateOffset+s, stateOffset+t.dest, t.min, t.max)
			}

			if a.IsAccept(s) {
				followA := nextA
				followOffset := stateOffset
				upto := i + 1
				for {
					if followA != nil {
						// adds a "virtual" epsilon transition:
						numTransitions = followA.initTransition(0, t)
						for j := 0; j < numTransitions; j++ {
							followA.nextTransition(t)
							ans.addTransitionRange(stateOffset+s, followOffset+numStates+t.dest, t.min, t.max)
						}
						if followA.IsAccept(0) {
							// keep chaning if followA accepts empty string
							followOffset += followA.numStates()
							if upto < len(l)-1 {
								followA = l[upto+1]
							} else {
								followA = nil
							}
							upto++
						} else {
							break
						}
					} else {
						ans.setAccept(stateOffset+s, true)
						break
					}
				}
			}
		}

		stateOffset += numStates
	}

	if ans.numStates() == 0 {
		ans.createState()
	}

	ans.finishState()
	return ans
}

/*
Returns an automaton that accepts the union of the empty string and
the language of the given automaton.

Complexity: linear in number of states.
*/
func optional(a *Automaton) *Automaton {
	ans := newEmptyAutomaton()
	ans.createState()
	ans.setAccept(0, true)
	if a.numStates() > 0 {
		ans.copy(a)
		ans.addEpsilon(0, 1)
	}
	ans.finishState()
	return ans
}

/*
Returns an automaton that accepts the Kleene star (zero or more
concatenated repetitions) of the language of the given automaton.
Never modifies the input automaton language.

Complexity: linear in number of states.
*/
func repeat(a *Automaton) *Automaton {
	if isEmpty(a) {
		return a
	}

	b := newAutomatonBuilder()
	b.createState()
	b.setAccept(0, true)
	b.copy(a)

	t := newTransition()
	count := a.initTransition(0, t)
	for i := 0; i < count; i++ {
		a.nextTransition(t)
		b.addTransitionRange(0, t.dest+1, t.min, t.max)
	}

	numStates := a.numStates()
	for s := 0; s < numStates; s++ {
		if a.IsAccept(s) {
			count = a.initTransition(0, t)
			for i := 0; i < count; i++ {
				a.nextTransition(t)
				b.addTransitionRange(s+1, t.dest+1, t.min, t.max)
			}
		}
	}

	return b.finish()
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
	a = totalize(determinize(a))
	numStates := a.numStates()
	for p := 0; p < numStates; p++ {
		a.setAccept(p, !a.IsAccept(p))
	}
	return removeDeadStates(a)
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
		return a1
	}
	return intersection(a1, complement(a2))
}

// Pair of states.
type StatePair struct{ s, s1, s2 int }

/*
Returns an automaton that accepts the intersection of the languages
of the given automata. Never modifies the input automata languages.

Complexity: quadratic in number of states.
*/
func intersection(a1, a2 *Automaton) *Automaton {
	if a1 == a2 || a1.numStates() == 0 {
		return a1
	}
	if a2.numStates() == 0 {
		return a2
	}

	transitions1 := a1.sortedTransitions()
	transitions2 := a2.sortedTransitions()
	c := newEmptyAutomaton()
	c.createState()
	worklist := list.New()
	newstates := make(map[string]*StatePair)
	hash := func(p *StatePair) string {
		return fmt.Sprintf("%v/%v", p.s1, p.s2)
	}
	p := &StatePair{0, 0, 0}
	worklist.PushBack(p)
	newstates[hash(p)] = p
	for worklist.Len() > 0 {
		p = worklist.Remove(worklist.Front()).(*StatePair)
		c.setAccept(p.s, a1.IsAccept(p.s1) && a2.IsAccept(p.s2))
		t1 := transitions1[p.s1]
		t2 := transitions2[p.s2]
		for n1, b2 := 0, 0; n1 < len(t1); n1++ {
			for b2 < len(t2) && t2[b2].max < t1[n1].min {
				b2++
			}
			for n2 := b2; n2 < len(t2) && t1[n1].max >= t2[n2].min; n2++ {
				if t2[n2].max >= t1[n1].min {
					q := &StatePair{-1, t1[n1].dest, t2[n2].dest}
					r, ok := newstates[hash(q)]
					if !ok {
						q.s = c.createState()
						worklist.PushBack(q)
						newstates[hash(q)] = q
						r = q
					}
					min := or(t1[n1].min > t2[n2].min, t1[n1].min, t2[n2].min).(int)
					max := or(t1[n1].max < t2[n2].max, t1[n1].max, t2[n2].max).(int)
					c.addTransitionRange(p.s, r.s, min, max)
				}
			}
		}
	}
	c.finishState()
	return removeDeadStates(c)
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
	return subsetOf(a2, a1) && subsetOf(a1, a2)
}

/*
Returns true if the automaton has any states that cannot be reached
from the initial state or cannot reach an accept state.
Cost is O(numTransitions+numStates).
*/
func hasDeadStates(a *Automaton) bool {
	liveStates := liveStates(a)
	numLive := liveStates.Cardinality()
	numStates := a.numStates()
	assert2(numLive <= int64(numStates), "numLive=%v numStates=%v %v", numLive, numStates, liveStates)
	return numLive < int64(numStates)
}

func hasDeadStatesFromInitial(a *Automaton) bool {
	r1 := liveStatesFromInitial(a)
	r2 := liveStatesToAccept(a)
	r1.AndNot(r2)
	return !r1.IsEmpty()
}

/*
Returns true if the language of a1 is a subset of the language of a2.
As a side-effect, a2 is determinized if not already marked as
deterministic.
*/
func subsetOf(a1, a2 *Automaton) bool {
	assert2(a1.deterministic, "a1 must be deterministic")
	assert2(a2.deterministic, "a2 must be deterministic")
	assert(!hasDeadStatesFromInitial(a1))
	assert2(!hasDeadStatesFromInitial(a2), "%v", a2)
	if a1.numStates() == 0 {
		// empty language is always a subset of any other language
		return true
	} else if a2.numStates() == 0 {
		return isEmpty(a1)
	}

	transitions1 := a1.sortedTransitions()
	transitions2 := a2.sortedTransitions()
	worklist := list.New()
	visited := make(map[string]*StatePair)
	hash := func(p *StatePair) string {
		return fmt.Sprintf("%v/%v", p.s1, p.s2)
	}
	p := &StatePair{-1, 0, 0}
	worklist.PushBack(p)
	visited[hash(p)] = p
	for worklist.Len() > 0 {
		p = worklist.Remove(worklist.Front()).(*StatePair)
		if a1.IsAccept(p.s1) && !a2.IsAccept(p.s2) {
			return false
		}
		t1 := transitions1[p.s1]
		t2 := transitions2[p.s2]
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
				q := &StatePair{-1, t1[n1].dest, t2[n2].dest}
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
	return unionN([]*Automaton{a1, a2})
}

/*
Returns an automaton that accepts the union of the languages of the
given automata.

Complexity: linear in number of states.
*/
func unionN(l []*Automaton) *Automaton {
	ans := newEmptyAutomaton()
	// create initial state
	ans.createState()
	// copy over all automata
	for _, a := range l {
		ans.copy(a)
	}
	// add epsilon transition from new initial state
	stateOffset := 1
	for _, a := range l {
		if a.numStates() == 0 {
			continue
		}
		ans.addEpsilon(0, stateOffset)
		stateOffset += a.numStates()
	}
	ans.finishState()
	return removeDeadStates(ans)
}

/* Simple custom []*Transition */
type TransitionList struct {
	transitions []int // dest,min,max
}

func (l *TransitionList) add(t *Transition) {
	l.transitions = append(l.transitions, t.dest, t.min, t.max)
}

// Holds all transitions that start on this int point, or end at this
// point-1
type PointTransitions struct {
	point  int
	ends   *TransitionList
	starts *TransitionList
}

func newPointTransitions() *PointTransitions {
	return &PointTransitions{
		ends:   new(TransitionList),
		starts: new(TransitionList),
	}
}

func (pt *PointTransitions) reset(point int) {
	pt.point = point
	pt.ends.transitions = pt.ends.transitions[:0]
	pt.starts.transitions = pt.starts.transitions[:0]
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
	pts.find(t.min).starts.add(t)
	pts.find(1 + t.max).ends.add(t)
}

func (pts *PointTransitionSet) String() string {
	panic("not implemented yet")
}

/*
Determinizes the given automaton.

Split the code points in ranges, and merge overlapping states.

Worst case complexity: exponential in number of states.
*/
func determinize(a *Automaton) *Automaton {
	if a.deterministic || a.numStates() <= 1 {
		return a
	}

	// subset construction
	b := newAutomatonBuilder()

	// fmt.Println("DET:")

	initialset := newFrozenIntSetOf(0, 0)

	// craete state 0:
	b.createState()

	worklist := list.New()
	newstate := make(map[string]int)
	hash := func(s *FrozenIntSet) string {
		return s.String()
	}

	worklist.PushBack(initialset)

	b.setAccept(0, a.IsAccept(0))
	newstate[hash(initialset)] = 0

	// like map[int]*PointTransitions
	points := newPointTransitionSet()

	// like sorted map[int]int
	statesSet := newSortedIntSet(5)

	t := newTransition()

	for worklist.Len() > 0 {
		s := worklist.Remove(worklist.Front()).(*FrozenIntSet)
		// fmt.Printf("det: pop set=%v\n", s)

		// Collate all outgoing transitions by min/1+max
		for _, s0 := range s.values {
			numTransitions := a.numTransitions(s0)
			a.initTransition(s0, t)
			for j := 0; j < numTransitions; j++ {
				a.nextTransition(t)
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

				q, ok := newstate[hashKey]
				if !ok {
					q = b.createState()
					p := statesSet.freeze(q)
					// fmt.Printf("  make new state=%v -> %v accCount=%v\n", q, p, accCount)
					worklist.PushBack(p)
					b.setAccept(q, accCount > 0)
					newstate[hash(p)] = q
				} else {
					assert2(b.isAccept(q) == (accCount > 0),
						"accCount=%v vs existing accept=%v states=%v",
						accCount, b.isAccept(q), statesSet)
				}

				// fmt.Printf("  add trans src=%v dest=%v min=%v max=%v\n",
				// 	r, q, lastPoint, point-1)
				b.addTransitionRange(r, q, lastPoint, point-1)
			}

			// process transitions that end on this point
			// (closes an overlapping interval)
			for j, limit := 0, len(v.ends.transitions); j < limit; j += 3 {
				dest := v.ends.transitions[j]
				statesSet.decr(dest)
				if a.IsAccept(dest) {
					accCount--
				}
			}
			v.ends.transitions = v.ends.transitions[:0] // reuse slice

			// process transitions that start on this point
			// (opens a new interval)
			for j, limit := 0, len(v.starts.transitions); j < limit; j += 3 {
				dest := v.starts.transitions[j]
				statesSet.incr(dest)
				if a.IsAccept(dest) {
					accCount++
				}
			}
			v.starts.transitions = v.starts.transitions[:0] // reuse slice

			lastPoint = point
		}
		points.reset()
		assert2(len(statesSet.values) == 0, "upto=%v", len(statesSet.values))
	}

	ans := b.finish()
	assert(ans.deterministic)
	return ans
}

// // L779
// Returns true if the given automaton accepts no strings.
func isEmpty(a *Automaton) bool {
	if a.numStates() == 0 {
		// common case: no states
		return true
	}
	if !a.IsAccept(0) && a.numTransitions(0) == 0 {
		// common case: just one initial state
		return true
	}
	if a.IsAccept(0) {
		// apparently common case: it accepts the empty string
		return false
	}

	workList := list.New()
	seen := util.NewOpenBitSet()
	workList.PushBack(0)
	seen.Set(0)

	t := newTransition()
	for workList.Len() > 0 {
		state := workList.Remove(workList.Front()).(int)
		if a.IsAccept(state) {
			return false
		}
		count := a.initTransition(state, t)
		for i := 0; i < count; i++ {
			a.nextTransition(t)
			if !seen.Get(int64(t.dest)) {
				workList.PushBack(t.dest)
				seen.Set(int64(t.dest))
			}
		}
	}

	return true
}

// /*
// Returns true if the given string is accepted by the autmaton.

// Complexity: linear in the length of the string.

// Note: for fll performance, use the RunAutomation class.
// */
// func run(a *Automaton, s string) bool {
// 	if a.isSingleton() {
// 		return s == a.singleton
// 	}
// 	if a.deterministic {
// 		p := a.initial
// 		for _, ch := range s {
// 			q := p.step(int(ch))
// 			if q == nil {
// 				return false
// 			}
// 			p = q
// 		}
// 		return p.accept
// 	}
// 	// states := a.NumberedStates()
// 	panic("not implemented yet")
// }

/*
Returns the set of live states. A state is "live" if an accept state
is reachable from it and if it is reachable from the initial state.
*/
func liveStates(a *Automaton) *util.OpenBitSet {
	live := liveStatesFromInitial(a)
	live.And(liveStatesToAccept(a))
	return live
}

/* Returns BitSet marking states reachable from the initial state. */
func liveStatesFromInitial(a *Automaton) *util.OpenBitSet {
	numStates := a.numStates()
	live := util.NewOpenBitSet()
	if numStates == 0 {
		return live
	}
	workList := list.New()
	live.Set(0)
	workList.PushBack(0)

	t := newTransition()
	for workList.Len() > 0 {
		s := workList.Remove(workList.Front()).(int)
		count := a.initTransition(s, t)
		for i := 0; i < count; i++ {
			a.nextTransition(t)
			if !live.Get(int64(t.dest)) {
				live.Set(int64(t.dest))
				workList.PushBack(t.dest)
			}
		}
	}

	return live
}

/* Returns BitSet marking states that can reach an accept state. */
func liveStatesToAccept(a *Automaton) *util.OpenBitSet {
	builder := newAutomatonBuilder()

	// NOTE: not quite the same thing as what SpecialOperations.reverse does:
	t := newTransition()
	numStates := a.numStates()
	for s := 0; s < numStates; s++ {
		builder.createState()
	}
	for s := 0; s < numStates; s++ {
		count := a.initTransition(s, t)
		for i := 0; i < count; i++ {
			a.nextTransition(t)
			builder.addTransitionRange(t.dest, s, t.min, t.max)
		}
	}
	a2 := builder.finish()

	workList := list.New()
	live := util.NewOpenBitSet()
	acceptBits := a.isAccept
	s := 0
	for s < numStates {
		s = int(acceptBits.NextSetBit(int64(s)))
		if s == -1 {
			break
		}
		live.Set(int64(s))
		workList.PushBack(s)
		s++
	}

	for workList.Len() > 0 {
		s = workList.Remove(workList.Front()).(int)
		count := a2.initTransition(s, t)
		for i := 0; i < count; i++ {
			a2.nextTransition(t)
			if !live.Get(int64(t.dest)) {
				live.Set(int64(t.dest))
				workList.PushBack(t.dest)
			}
		}
	}

	return live
}

/*
Removes transitions to dead states (a state is "dead" if it is not
reachable from the initial state or no accept state is reachable from
it.)
*/
func removeDeadStates(a *Automaton) *Automaton {
	numStates := a.numStates()
	liveSet := liveStates(a)

	m := make([]int, numStates)

	ans := newEmptyAutomaton()
	// fmt.Printf("liveSet: %v numStates=%v\n", liveSet, numStates)
	for i := 0; i < numStates; i++ {
		if liveSet.Get(int64(i)) {
			m[i] = ans.createState()
			ans.setAccept(m[i], a.IsAccept(i))
		}
	}

	t := newTransition()

	for i := 0; i < numStates; i++ {
		if liveSet.Get(int64(i)) {
			numTransitions := a.initTransition(i, t)
			// filter out transitions to dead states:
			for j := 0; j < numTransitions; j++ {
				a.nextTransition(t)
				if liveSet.Get(int64(t.dest)) {
					ans.addTransitionRange(m[i], m[t.dest], t.min, t.max)
				}
			}
		}
	}

	ans.finishState()
	assert(!hasDeadStates(ans))
	return ans
}

/*
Finds the largest entry whose value is less than or equal to c, or
0 if there is no such entry.
*/
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

/* Returns an automaton accepting the reverse language. */
func reverse(a *Automaton) (*Automaton, map[int]bool) {
	if isEmpty(a) {
		return newEmptyAutomaton(), nil
	}

	numStates := a.numStates()

	// build a new automaton with all edges reversed
	b := newAutomatonBuilder()

	// initial node; we'll add epsilon transitions in the end:
	b.createState()
	for s := 0; s < numStates; s++ {
		b.createState()
	}

	// old initial state becomes new accept state:
	b.setAccept(1, true)

	t := newTransition()
	for s := 0; s < numStates; s++ {
		numTransitions := a.numTransitions(s)
		a.initTransition(s, t)
		for i := 0; i < numTransitions; i++ {
			a.nextTransition(t)
			b.addTransitionRange(t.dest+1, s+1, t.min, t.max)
		}
	}

	ans := b.finish()
	initialStates := make(map[int]bool)

	acceptStates := a.isAccept
	for s := acceptStates.NextSetBit(0); s != -1; s = acceptStates.NextSetBit(s + 1) {
		ans.addEpsilon(0, int(s+1))
		initialStates[int(s+1)] = true
	}

	ans.finishState()
	return ans, initialStates
}

/*
Returns a new automaton accepting the same language with added
transitions to a dead state so that from every state and every label
there is a transition.
*/
func totalize(a *Automaton) *Automaton {
	ans := newEmptyAutomaton()
	numStates := a.numStates()
	for i := 0; i < numStates; i++ {
		ans.createState()
		ans.setAccept(i, a.IsAccept(i))
	}

	deadState := ans.createState()
	ans.addTransitionRange(deadState, deadState, MIN_CODE_POINT, unicode.MaxRune)

	t := newTransition()
	for i := 0; i < numStates; i++ {
		maxi := MIN_CODE_POINT
		count := a.initTransition(i, t)
		for j := 0; j < count; j++ {
			a.nextTransition(t)
			ans.addTransitionRange(i, t.dest, t.min, t.max)
			if t.min > maxi {
				ans.addTransitionRange(i, deadState, maxi, t.min-1)
			}
			if t.max+1 > maxi {
				maxi = t.max + 1
			}
		}

		if maxi <= unicode.MaxRune {
			ans.addTransitionRange(i, deadState, maxi, unicode.MaxRune)
		}
	}

	ans.finishState()
	return ans
}
