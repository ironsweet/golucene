package automaton

import (
	"bytes"
	"fmt"
	"math/big"
)

// util/automaton/MinimizationOperations.java

// Minimizes (and determinizes if not already deterministic) the
// given automaton
func minimize(a *Automaton) *Automaton {
	return minimizeHopcroft(a)
}

// Minimizes the given automaton using Hopcroft's alforithm.
func minimizeHopcroft(a *Automaton) *Automaton {
	panic("niy")
	// // log.Println("Minimizing using Hopcraft...")
	// a.determinize()
	// if len(a.initial.transitionsArray) == 1 {
	// 	t := a.initial.transitionsArray[0]
	// 	if t.to == a.initial && t.min == MIN_CODE_POINT && t.max == unicode.MaxRune {
	// 		return
	// 	}
	// }
	// a.totalize()

	// // initialize data structure
	// sigma := a.startPoints()
	// states := a.NumberedStates()
	// sigmaLen, statesLen := len(sigma), len(states)
	// reverse := make([][][]*State, statesLen)
	// for i, _ := range reverse {
	// 	reverse[i] = make([][]*State, sigmaLen)
	// }
	// partition := make([]map[int]*State, statesLen)
	// splitblock := make([][]*State, statesLen)
	// block := make([]int, statesLen)
	// active := make([][]*StateList, statesLen)
	// for i, _ := range active {
	// 	active[i] = make([]*StateList, sigmaLen)
	// }
	// active2 := make([][]*StateListNode, statesLen)
	// for i, _ := range active2 {
	// 	active2[i] = make([]*StateListNode, sigmaLen)
	// }
	// pending := list.New()
	// pending2 := big.NewInt(0) // sigmaLen * statesLen bits
	// split := big.NewInt(0)    // statesLen bits
	// refine := big.NewInt(0)   // statesLen bits
	// refine2 := big.NewInt(0)  // statesLen bits
	// for q, _ := range splitblock {
	// 	splitblock[q] = make([]*State, 0)
	// 	partition[q] = make(map[int]*State)
	// 	for x, _ := range active[q] {
	// 		active[q][x] = &StateList{}
	// 	}
	// }
	// // find initial partition and reverse edges
	// for q, qq := range states {
	// 	j := or(qq.accept, 0, 1).(int)
	// 	partition[j][qq.id] = qq
	// 	block[q] = j
	// 	for x, v := range sigma {
	// 		r := reverse[qq.step(v).number]
	// 		if r[x] == nil {
	// 			r[x] = make([]*State, 0)
	// 		}
	// 		r[x] = append(r[x], qq)
	// 	}
	// }
	// // initialize active sets
	// for j := 0; j <= 1; j++ {
	// 	for x := 0; x < sigmaLen; x++ {
	// 		for _, qq := range partition[j] {
	// 			if reverse[qq.number][x] != nil {
	// 				active2[qq.number][x] = active[j][x].add(qq)
	// 			}
	// 		}
	// 	}
	// }
	// // initialize pending
	// for x := 0; x < sigmaLen; x++ {
	// 	j := or(active[0][x].size <= active[1][x].size, 0, 1).(int)
	// 	pending.PushBack(&IntPair{j, x})
	// 	setBit(pending2, x*statesLen+j)
	// }
	// // process pending until fixed point
	// k := 2
	// for pending.Len() > 0 {
	// 	ip := pending.Front().Value.(*IntPair)
	// 	pending.Remove(pending.Front())
	// 	p, x := ip.n1, ip.n2
	// 	clearBit(pending2, x*statesLen+p)
	// 	// find states that need to be split off their blocks
	// 	for m := active[p][x].first; m != nil; m = m.next {
	// 		if r := reverse[m.q.number][x]; r != nil {
	// 			for _, s := range r {
	// 				if i := s.number; split.Bit(i) == 0 {
	// 					setBit(split, i)
	// 					j := block[i]
	// 					splitblock[j] = append(splitblock[j], s)
	// 					if refine2.Bit(j) == 0 {
	// 						setBit(refine2, j)
	// 						setBit(refine, j)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// 	// refine blocks
	// 	// TODO Go's Big int may not be efficient here.
	// 	for j, bitsLen := 0, refine.BitLen(); j < bitsLen; j++ {
	// 		if refine.Bit(j) == 0 {
	// 			continue
	// 		}

	// 		sb := splitblock[j]
	// 		if len(sb) < len(partition[j]) {
	// 			b1, b2 := partition[j], partition[k]
	// 			for _, s := range sb {
	// 				delete(b1, s.id)
	// 				b2[s.id] = s
	// 				block[s.number] = k
	// 				for c, sn := range active2[s.number] {
	// 					if sn != nil && sn.sl == active[j][c] {
	// 						sn.remove()
	// 						active2[s.number][c] = active[k][c].add(s)
	// 					}
	// 				}
	// 			}
	// 			// update pending
	// 			for c, _ := range active[j] {
	// 				aj := active[j][c].size
	// 				ak := active[k][c].size
	// 				ofs := c * statesLen
	// 				if pending2.Bit(ofs+j) == 0 && 0 < aj && aj <= ak {
	// 					setBit(pending2, ofs+j)
	// 					pending.PushBack(&IntPair{j, c})
	// 				} else {
	// 					setBit(pending2, ofs+k)
	// 					pending.PushBack(&IntPair{k, c})
	// 				}
	// 			}
	// 			k++
	// 		}
	// 		clearBit(refine2, j)
	// 		for _, s := range sb {
	// 			clearBit(split, s.number)
	// 		}
	// 		splitblock[j] = splitblock[j][:0] // clear sb
	// 	}
	// 	refine = big.NewInt(0) // not quite efficient
	// }
	// // make a new state for each equivalence class, set initial state
	// newstates := make([]*State, k)
	// for n, _ := range newstates {
	// 	s := newState()
	// 	newstates[n] = s
	// 	for _, q := range partition[n] {
	// 		if q == a.initial {
	// 			a.initial = s
	// 		}
	// 		s.accept = q.accept
	// 		s.number = q.number // select representative
	// 		q.number = n
	// 	}
	// }
	// // build transitions and set acceptance
	// for _, s := range newstates {
	// 	s.accept = states[s.number].accept
	// 	for _, t := range states[s.number].transitionsArray {
	// 		s.addTransition(newTransitionRange(t.min, t.max, newstates[t.to.number]))
	// 	}
	// }
	// a.clearNumberedStates()
	// a.removeDeadTransitions()
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
