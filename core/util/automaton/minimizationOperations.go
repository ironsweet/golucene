package automaton

import (
	"container/list"
	// "fmt"
	"github.com/balzaczyy/golucene/core/util"
	"unicode"
)

// util/automaton/MinimizationOperations.java

// Minimizes (and determinizes if not already deterministic) the
// given automaton
func minimize(a *Automaton) *Automaton {
	return minimizeHopcroft(a)
}

// Minimizes the given automaton using Hopcroft's alforithm.
func minimizeHopcroft(a *Automaton) *Automaton {
	if a.numStates() == 0 || !a.IsAccept(0) && a.numTransitions(0) == 0 {
		// fastmatch for common case
		return newEmptyAutomaton()
	}
	a = determinize(a)
	if a.numTransitions(0) == 1 {
		t := newTransition()
		a.transition(0, 0, t)
		if t.dest == 0 && t.min == MIN_CODE_POINT &&
			t.max == unicode.MaxRune {
			// accepts all strings
			return a
		}
	}
	a = totalize(a)

	// initialize data structure
	sigma := a.startPoints()
	sigmaLen, statesLen := len(sigma), a.numStates()

	reverse := make([][][]int, statesLen)
	for i, _ := range reverse {
		reverse[i] = make([][]int, sigmaLen)
	}
	partition := make([]map[int]bool, statesLen)
	splitblock := make([][]int, statesLen)
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
	pending2 := util.NewOpenBitSet() // sigmaLen * statesLen bits
	split := util.NewOpenBitSet()    // statesLen bits
	refine := util.NewOpenBitSet()   // statesLen bits
	refine2 := util.NewOpenBitSet()  // statesLen bits
	for q, _ := range splitblock {
		partition[q] = make(map[int]bool)
		for x, _ := range active[q] {
			active[q][x] = new(StateList)
		}
	}
	// find initial partition and reverse edges
	for q := 0; q < statesLen; q++ {
		j := or(a.IsAccept(q), 0, 1).(int)
		partition[j][q] = true
		block[q] = j
		for x, v := range sigma {
			n := a.step(q, v)
			assert2(n >= 0 && n < len(reverse), "%v", n)
			r := reverse[a.step(q, v)]
			r[x] = append(r[x], q)
		}
	}
	// initialize active sets
	for j := 0; j <= 1; j++ {
		for x := 0; x < sigmaLen; x++ {
			for q, _ := range partition[j] {
				if reverse[q][x] != nil {
					active2[q][x] = active[j][x].add(q)
				}
			}
		}
	}
	// initialize pending
	for x := 0; x < sigmaLen; x++ {
		j := or(active[0][x].size <= active[1][x].size, 0, 1).(int)
		pending.PushBack(&IntPair{j, x})
		pending2.Set(int64(x*statesLen + j))
	}
	// process pending until fixed point
	k := 2
	// fmt.Println("start min")
	for pending.Len() > 0 {
		// fmt.Println("  cycle pending")
		ip := pending.Remove(pending.Front()).(*IntPair)
		p, x := ip.n1, ip.n2
		// fmt.Printf("    pop n1=%v n2=%v\n", ip.n1, ip.n2)
		pending2.Clear(int64(x*statesLen + p))
		// find states that need to be split off their blocks
		for m := active[p][x].first; m != nil; m = m.next {
			if r := reverse[m.q][x]; r != nil {
				for _, i := range r {
					if !split.Get(int64(i)) {
						split.Set(int64(i))
						j := block[i]
						splitblock[j] = append(splitblock[j], i)
						if !refine2.Get(int64(j)) {
							refine2.Set(int64(j))
							refine.Set(int64(j))
						}
					}
				}
			}
		}
		// refine blocks
		for j := int(refine.NextSetBit(0)); j >= 0; j = int(refine.NextSetBit(int64(j) + 1)) {
			sb := splitblock[j]
			if len(sb) < len(partition[j]) {
				b1, b2 := partition[j], partition[k]
				for _, s := range sb {
					delete(b1, s)
					b2[s] = true
					block[s] = k
					for c, sn := range active2[s] {
						if sn != nil && sn.sl == active[j][c] {
							sn.remove()
							active2[s][c] = active[k][c].add(s)
						}
					}
				}
				// update pending
				for c, _ := range active[j] {
					aj := active[j][c].size
					ak := active[k][c].size
					ofs := int64(c * statesLen)
					if !pending2.Get(ofs+int64(j)) && 0 < aj && aj <= ak {
						pending2.Set(ofs + int64(j))
						pending.PushBack(&IntPair{j, c})
					} else {
						pending2.Set(ofs + int64(k))
						pending.PushBack(&IntPair{k, c})
					}
				}
				k++
			}
			refine2.Clear(int64(j))
			for _, s := range sb {
				split.Clear(int64(s))
			}
			splitblock[j] = nil // clear sb
		}
		refine = util.NewOpenBitSet() // not quite efficient
	}

	ans := newEmptyAutomaton()
	t := newTransition()
	// fmt.Printf("  k=%v\n", k)

	// make a new state for each equivalence class, set initial state
	stateMap := make([]int, statesLen)
	stateRep := make([]int, k)

	ans.createState()

	// fmt.Printf("min: k=%v\n", k)
	for n := 0; n < k; n++ {
		// fmt.Printf("    n=%v\n", n)

		isInitial := false
		for q, _ := range partition[n] {
			if q == 0 {
				isInitial = true
				// fmt.Println("    isInitial!")
				break
			}
		}

		newState := 0
		if !isInitial {
			newState = ans.createState()
		}

		// fmt.Printf("  newState=%v\n", newState)

		for q, _ := range partition[n] {
			stateMap[q] = newState
			// fmt.Printf("      q=%v isAccept?=%v\n", q, a.IsAccept(q))
			ans.setAccept(newState, a.IsAccept(q))
			stateRep[newState] = q // select representative
		}
	}

	// build transitions and set acceptance
	for n := 0; n < k; n++ {
		numTransitions := a.initTransition(stateRep[n], t)
		for i := 0; i < numTransitions; i++ {
			a.nextTransition(t)
			// fmt.Println("  add trans")
			ans.addTransitionRange(n, stateMap[t.dest], t.min, t.max)
		}
	}
	ans.finishState()
	// fmt.Printf("%v states\n", ans.numStates())

	return removeDeadStates(ans)
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

func (sl *StateList) add(q int) *StateListNode {
	return newStateListNode(q, sl)
}

type StateListNode struct {
	q          int
	next, prev *StateListNode
	sl         *StateList
}

func newStateListNode(q int, sl *StateList) *StateListNode {
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
