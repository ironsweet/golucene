package automaton

import (
	"github.com/balzaczyy/golucene/core/util"
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
	panic("niy")
	// if len(l) == 0 {
	// 	return makeEmptyString()
	// }
	// allSingleton := true
	// for _, a := range l {
	// 	if !a.isSingleton() {
	// 		allSingleton = false
	// 		break
	// 	}
	// }
	// if allSingleton {
	// 	var b bytes.Buffer
	// 	for _, a := range l {
	// 		b.WriteString(a.singleton)
	// 	}
	// 	return makeString(b.String())
	// }
	// for _, a := range l {
	// 	if isEmpty(a) {
	// 		return MakeEmpty()
	// 	}
	// }
	// all := make(map[*Automaton]bool)
	// hasAliases := false
	// for _, a := range l {
	// 	if _, ok := all[a]; ok {
	// 		hasAliases = true
	// 		break
	// 	} else {
	// 		all[a] = true
	// 	}
	// }
	// b := l[0]
	// if hasAliases {
	// 	b = b.cloneExpanded()
	// } else {
	// 	b = b.cloneExpandedIfRequired()
	// }
	// ac := b.acceptStates()
	// first := true
	// for _, a := range l {
	// 	if first {
	// 		first = false
	// 	} else {
	// 		if a.isEmptyString() {
	// 			continue
	// 		}
	// 		aa := a
	// 		if hasAliases {
	// 			aa = aa.cloneExpanded()
	// 		} else {
	// 			aa = aa.cloneExpandedIfRequired()
	// 		}
	// 		ns := aa.acceptStates()
	// 		for _, s := range ac {
	// 			s.accept = false
	// 			s.addEpsilon(aa.initial)
	// 			if s.accept {
	// 				ns[s.id] = s
	// 			}
	// 		}
	// 		ac = ns
	// 	}
	// }
	// b.deterministic = false
	// b.clearNumberedStates()
	// b.checkMinimizeAlways()
	// return b
}

/*
Returns an automaton that accepts the union of the empty string and
the language of the given automaton.

Complexity: linear in number of states.
*/
func optional(a *Automaton) *Automaton {
	panic("niy")
	// s := newState()
	// s.addEpsilon(a.initial)
	// s.accept = true
	// a = a.cloneExpandedIfRequired()
	// a.initial = s
	// a.deterministic = false
	// a.clearNumberedStates()
	// a.checkMinimizeAlways()
	// return a
}

/*
Returns an automaton that accepts the Kleene star (zero or more
concatenated repetitions) of the language of the given automaton.
Never modifies the input automaton language.

Complexity: linear in number of states.
*/
func repeat(a *Automaton) *Automaton {
	panic("niy")
	// a = a.cloneExpandedIfRequired()
	// s := newState()
	// s.accept = true
	// s.addEpsilon(a.initial)
	// for _, p := range a.acceptStates() {
	// 	p.addEpsilon(s)
	// }
	// a.initial = s
	// a.deterministic = false
	// a.clearNumberedStates()
	// a.checkMinimizeAlways()
	// return a
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
	panic("niy")
	// a = a.cloneExpandedIfRequired()
	// a.determinize()
	// a.totalize()
	// for _, p := range a.NumberedStates() {
	// 	p.accept = !p.accept
	// }
	// a.removeDeadTransitions()
	// return a
}

/*
Returns a (deterministic) automaton that accepts the intersection of
the language of a1 and the complement of the language of a2. As a
side-effect, the automata may be determinized, if not already
deterministic.

Complexity: quadratic in number of states (if already deterministic).
*/
func minus(a1, a2 *Automaton) *Automaton {
	panic("niy")
	// if isEmpty(a1) || a1 == a2 {
	// 	return MakeEmpty()
	// }
	// if isEmpty(a2) {
	// 	return a1.cloneIfRequired()
	// }
	// if a1.isSingleton() {
	// 	if run(a2, a1.singleton) {
	// 		return MakeEmpty()
	// 	} else {
	// 		return a1.cloneIfRequired()
	// 	}
	// }
	// return intersection(a1, a2.complement())
}

// Pair of states.
type StatePair struct{ s, s1, s2 int }

/*
Returns an automaton that accepts the intersection of the languages
of the given automata. Never modifies the input automata languages.

Complexity: quadratic in number of states.
*/
func intersection(a1, a2 *Automaton) *Automaton {
	panic("niy")
	// if a1.isSingleton() {
	// 	if run(a2, a1.singleton) {
	// 		return a1.cloneIfRequired()
	// 	}
	// 	return MakeEmpty()
	// }
	// if a2.isSingleton() {
	// 	if run(a1, a2.singleton) {
	// 		return a2.cloneIfRequired()
	// 	}
	// 	return MakeEmpty()
	// }
	// if a1 == a2 {
	// 	return a1.cloneIfRequired()
	// }
	// transitions1 := a1.sortedTransitions()
	// transitions2 := a2.sortedTransitions()
	// c := newEmptyAutomaton()
	// worklist := list.New()
	// newstates := make(map[string]*StatePair)
	// hash := func(p *StatePair) string {
	// 	return fmt.Sprintf("%v/%v", p.s1.id, p.s2.id)
	// }
	// p := &StatePair{c.initial, a1.initial, a2.initial}
	// worklist.PushBack(p)
	// newstates[hash(p)] = p
	// for worklist.Len() > 0 {
	// 	p = worklist.Front().Value.(*StatePair)
	// 	worklist.Remove(worklist.Front())
	// 	p.s.accept = (p.s1.accept && p.s2.accept)
	// 	t1 := transitions1[p.s1.number]
	// 	t2 := transitions2[p.s2.number]
	// 	for n1, b2, t1Len := 0, 0, len(t1); n1 < t1Len; n1++ {
	// 		t2Len := len(t2)
	// 		for b2 < t2Len && t2[b2].max < t1[n1].min {
	// 			b2++
	// 		}
	// 		for n2 := b2; n2 < t2Len && t1[n1].max >= t2[n2].min; n2++ {
	// 			if t2[n2].max >= t1[n1].min {
	// 				q := &StatePair{nil, t1[n1].to, t2[n2].to}
	// 				r, ok := newstates[hash(q)]
	// 				if !ok {
	// 					q.s = newState()
	// 					worklist.PushBack(q)
	// 					newstates[hash(q)] = q
	// 					r = q
	// 				}
	// 				min := or(t1[n1].min > t2[n2].min, t1[n1].min, t2[n2].min).(int)
	// 				max := or(t1[n1].max < t2[n2].max, t1[n1].max, t2[n2].max).(int)
	// 				p.s.addTransition(newTransitionRange(min, max, r.s))
	// 			}
	// 		}
	// 	}
	// }
	// c.deterministic = (a1.deterministic && a2.deterministic)
	// c.removeDeadTransitions()
	// c.checkMinimizeAlways()
	// return c
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
Returns true if the language of a1 is a subset of the language of a2.
As a side-effect, a2 is determinized if not already marked as
deterministic.
*/
func subsetOf(a1, a2 *Automaton) bool {
	panic("niy")
	// if a1 == a2 {
	// 	return true
	// }
	// if a1.isSingleton() {
	// 	if a2.isSingleton() {
	// 		return a1.singleton == a2.singleton
	// 	}
	// 	return run(a2, a1.singleton)
	// }
	// a2.determinize()
	// transitions1 := a1.sortedTransitions()
	// transitions2 := a2.sortedTransitions()
	// worklist := list.New()
	// visited := make(map[string]*StatePair)
	// hash := func(p *StatePair) string {
	// 	return fmt.Sprintf("%v/%v", p.s1.id, p.s2.id)
	// }
	// p := &StatePair{nil, a1.initial, a2.initial}
	// worklist.PushBack(p)
	// visited[hash(p)] = p
	// for worklist.Len() > 0 {
	// 	p = worklist.Front().Value.(*StatePair)
	// 	worklist.Remove(worklist.Front())
	// 	if p.s1.accept && !p.s2.accept {
	// 		return false
	// 	}
	// 	t1 := transitions1[p.s1.number]
	// 	t2 := transitions2[p.s2.number]
	// 	for n1, b2, t1Len := 0, 0, len(t1); n1 < t1Len; n1++ {
	// 		t2Len := len(t2)
	// 		for b2 < t2Len && t2[b2].max < t1[n1].min {
	// 			b2++
	// 		}
	// 		min1, max1 := t1[n1].min, t1[n1].max

	// 		for n2 := b2; n2 < t2Len && t1[n1].max >= t2[n2].min; n2++ {
	// 			if t2[n2].min > min1 {
	// 				return false
	// 			}
	// 			if t2[n2].max < unicode.MaxRune {
	// 				min1 = t2[n2].max + 1
	// 			} else {
	// 				min1, max1 = unicode.MaxRune, MIN_CODE_POINT
	// 			}
	// 			q := &StatePair{nil, t1[n1].to, t2[n2].to}
	// 			if _, ok := visited[hash(q)]; !ok {
	// 				worklist.PushBack(q)
	// 				visited[hash(q)] = q
	// 			}
	// 		}
	// 		if min1 <= max1 {
	// 			return false
	// 		}
	// 	}
	// }
	// return true
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
func determinize(a *Automaton) *Automaton {
	panic("niy")
	// if a.deterministic || a.isSingleton() {
	// 	return
	// }

	// allStates := a.NumberedStates()

	// // subset construction
	// initAccept := a.initial.accept
	// initNumber := a.initial.number
	// a.initial = newState()
	// initialset := newFrozenIntSet([]int{initNumber}, a.initial)

	// worklist := list.New()
	// newstate := make(map[string]*FrozenIntSet)

	// worklist.PushBack(initialset)

	// a.initial.accept = initAccept
	// newstate[initialset.String()] = initialset

	// newStatesArray := make([]*State, 0, 5)
	// newStatesArray = append(newStatesArray, a.initial)
	// a.initial.number = 0

	// // like map[int]*PointTransitions
	// points := newPointTransitionSet()

	// // like sorted map[int]int
	// statesSet := newSortedIntSet(5)

	// for worklist.Len() > 0 {
	// 	s := worklist.Front().Value.(*FrozenIntSet)
	// 	worklist.Remove(worklist.Front())

	// 	// Collate all outgoing transitions by min/1+max
	// 	for _, v := range s.values {
	// 		s0 := allStates[v]
	// 		for _, t := range s0.transitionsArray {
	// 			points.add(t)
	// 		}
	// 	}

	// 	if len(points.points) == 0 {
	// 		// No outgoing transitions -- skip it
	// 		continue
	// 	}

	// 	points.sort()

	// 	lastPoint := -1
	// 	accCount := 0

	// 	r := s.state
	// 	for _, v := range points.points {
	// 		point := v.point

	// 		if len(statesSet.values) > 0 {
	// 			assert(lastPoint != -1)

	// 			hashKey := statesSet.computeHash().String()

	// 			var q *State
	// 			ss, ok := newstate[hashKey]
	// 			if !ok {
	// 				q = newState()
	// 				p := statesSet.freeze(q)
	// 				worklist.PushBack(p)
	// 				newStatesArray = append(newStatesArray, q)
	// 				q.number = len(newStatesArray) - 1
	// 				q.accept = accCount > 0
	// 				newstate[p.String()] = p
	// 			} else {
	// 				q = ss.state
	// 				assert2(q.accept == (accCount > 0), fmt.Sprintf(
	// 					"accCount=%v vs existing accept=%v states=%v",
	// 					accCount, q.accept, statesSet))
	// 			}

	// 			r.addTransition(newTransitionRange(lastPoint, point-1, q))
	// 		}

	// 		// process transitions that end on this point
	// 		// (closes an overlapping interval)
	// 		for _, t := range v.ends {
	// 			statesSet.decr(t.to.number)
	// 			if t.to.accept {
	// 				accCount--
	// 			}
	// 		}
	// 		v.ends = v.ends[:0] // reuse slice

	// 		// process transitions that start on this point
	// 		// (opens a new interval)
	// 		for _, t := range v.starts {
	// 			statesSet.incr(t.to.number)
	// 			if t.to.accept {
	// 				accCount++
	// 			}
	// 		}
	// 		v.starts = v.starts[:0] // reuse slice
	// 		lastPoint = point
	// 	}
	// 	points.reset()
	// 	assert2(len(statesSet.values) == 0, fmt.Sprintf("upto=%v", len(statesSet.values)))
	// }
	// a.deterministic = true
	// a.setNumberedStates(newStatesArray)
}

// // L775
// // Returns true if the given automaton accepts the empty string and
// // nothing else.
// func isEmptyString(a *Automaton) bool {
// 	if a.isSingleton() {
// 		return len(a.singleton) == 0
// 	}
// 	return a.initial.accept && len(a.initial.transitionsArray) == 0
// }

// // Returns true if the given automaton accepts no strings.
// func isEmpty(a *Automaton) bool {
// 	return !a.isSingleton() && !a.initial.accept && len(a.initial.transitionsArray) == 0
// }

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
Removes transitions to dead states (a state is "dead" if it is not
reachable from the initial state or no accept state is reachable from
it.)
*/

func removeDeadStates(a *Automaton) *Automaton {
	panic("niy")
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
func reverse(a *Automaton) *Automaton {
	panic("niy")
	// a.ExpandSingleton()
	// // reverse all edges
	// m := make(map[int]map[string]*Transition)
	// hash := func(t *Transition) string {
	// 	return fmt.Sprintf("%v/%v/%v", t.min, t.max, t.to.id)
	// }
	// states := a.NumberedStates()
	// accept := make(map[int]*State)
	// for _, s := range states {
	// 	if s.accept {
	// 		accept[s.id] = s
	// 	}
	// }
	// for _, r := range states {
	// 	m[r.id] = make(map[string]*Transition)
	// 	r.accept = false
	// }
	// for _, r := range states {
	// 	for _, t := range r.transitionsArray {
	// 		tt := newTransitionRange(t.min, t.max, r)
	// 		m[t.to.id][hash(tt)] = tt
	// 	}
	// }
	// for _, r := range states {
	// 	tr := m[r.id]
	// 	arr := make([]*Transition, 0, len(tr))
	// 	for _, t := range tr {
	// 		arr = append(arr, t)
	// 	}
	// 	r.transitionsArray = arr
	// }
	// // make new initial + final states
	// a.initial.accept = true
	// a.initial = newState()
	// for _, r := range accept {
	// 	a.initial.addEpsilon(r) // ensures that all initial states are reachable
	// }
	// a.deterministic = false
	// a.clearNumberedStates()
	// return accept
}
