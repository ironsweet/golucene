package automaton

import (
	"container/list"
	"github.com/balzaczyy/golucene/core/util"
	. "github.com/balzaczyy/golucene/test_framework/util"
	// "fmt"
	"math/rand"
	"testing"
	"unicode"
)

func TestRegExpToAutomaton(t *testing.T) {
	a := NewRegExp("[^ \t\r\n]+").ToAutomaton()
	assert(a.deterministic)
	assert(-1 == a.curState)
	assert(2 == a.numStates())
}

func TestMinusSimple(t *testing.T) {
	assert(sameLanguage(makeChar('b'), minus(makeCharRange('a', 'b'), makeChar('a'))))
	assert(sameLanguage(MakeEmpty(), minus(makeChar('a'), makeChar('a'))))
}

func TestComplementSimple(t *testing.T) {
	a := makeChar('a')
	assert(sameLanguage(a, complement(complement(a))))
}

func TestDeterminizeSimple(t *testing.T) {
	a1 := complement(NewRegExpWithFlag("-", NONE).ToAutomaton())
	a2 := NewRegExpWithFlag("ݖ|+", NONE).ToAutomaton()
	a := concatenate(a1, a2)
	a = removeDeadStates(a)
	a = determinize(a)
	assert(a.numStates() == 4)
}

// func TestStringUnion(t testing.T) {
// strings := make([]string, 0, 500)
// for i := NextInt(Random(), 0, 1000); i >= 0; i-- {
// 	strings = append(strings, RandomUnicodeString(Random()))
// }

// sort.Strings(strings)
// union := makeStringUnion(strings)
// assert(union.isDeterministic())
// assert(sameLanguage(union, naiveUnion(strings)))
// }

// util/automaton/AutomatonTestUtil.java
/*
Utilities for testing automata.

Capable of generating random regular expressions, and automata, and
also provides a number of very basic unoptimized implementations
(*slow) for testing.
*/

// Returns random string, including full unicode range.
func randomRegexp(r *rand.Rand) string {
	for i := 0; i < 500; i++ {
		regexp := randomRegexpString(r)
		// we will also generate some undefined unicode queries
		if !util.IsValidUTF16String([]rune(regexp)) {
			continue
		}
		if ok := func(regexp string) (ok bool) {
			ok = true
			defer func() {
				if r := recover(); r != nil {
					// log.Println("Recovered:", r)
					ok = false
				}
			}()
			// log.Println("Trying", regexp)
			NewRegExpWithFlag(regexp, NONE)
			return
		}(regexp); ok {
			// fmt.Println("Valid regexp found:", regexp)
			return regexp
		}
	}
	panic("should not be here")
}

func randomRegexpString(r *rand.Rand) string {
	end := r.Intn(20)
	if end == 0 {
		// allow 0 length
		return ""
	}
	buffer := make([]rune, 0, end)
	for i := 0; i < end; i++ {
		t := r.Intn(15)
		if 0 == t && i < end-1 {
			// Make a surrogate pair
			// High surrogate
			buffer = append(buffer, rune(NextInt(r, 0xd800, 0xdbff)))
			i++
			// Low surrogate
			buffer = append(buffer, rune(NextInt(r, 0xdc00, 0xdfff)))
		} else if t <= 1 {
			buffer = append(buffer, rune(r.Intn(0x80)))
		} else {
			switch t {
			case 2:
				buffer = append(buffer, rune(NextInt(r, 0x80, 0x800)))
			case 3:
				buffer = append(buffer, rune(NextInt(r, 0x800, 0xd7ff)))
			case 4:
				buffer = append(buffer, rune(NextInt(r, 0xe000, 0xffff)))
			case 5:
				buffer = append(buffer, '.')
			case 6:
				buffer = append(buffer, '?')
			case 7:
				buffer = append(buffer, '*')
			case 8:
				buffer = append(buffer, '+')
			case 9:
				buffer = append(buffer, '(')
			case 10:
				buffer = append(buffer, ')')
			case 11:
				buffer = append(buffer, '-')
			case 12:
				buffer = append(buffer, '[')
			case 13:
				buffer = append(buffer, ']')
			case 14:
				buffer = append(buffer, '|')
			}
		}
	}
	return string(buffer)
}

// L267
// Return a random NFA/DFA for testing
func randomAutomaton(r *rand.Rand) *Automaton {
	// get two random Automata from regexps
	a1 := NewRegExpWithFlag(randomRegexp(r), NONE).ToAutomaton()
	if r.Intn(2) == 0 {
		a1 = complement(a1)
	}

	a2 := NewRegExpWithFlag(randomRegexp(r), NONE).ToAutomaton()
	if r.Intn(2) == 0 {
		a2 = complement(a2)
	}

	// combine them in random ways
	switch r.Intn(4) {
	case 0:
		// fmt.Println("DEBUG way 0")
		return concatenate(a1, a2)
	case 1:
		// fmt.Println("DEBUG way 1")
		return union(a1, a2)
	case 2:
		// fmt.Println("DEBUG way 2")
		return intersection(a1, a2)
	default:
		// fmt.Println("DEBUG way 3")
		return minus(a1, a2)
	}
}

/**
 * below are original, unoptimized implementations of DFA operations for testing.
 * These are from brics automaton, full license (BSD) below:
 */

/*
 * dk.brics.automaton
 *
 * Copyright (c) 2001-2009 Anders Moeller
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. The name of the author may not be used to endorse or promote products
 *    derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
 * IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
 * OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 * IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
 * INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 * NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 * THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

/**
 * Simple, original brics implementation of Brzozowski minimize()
 */
func minimizeSimple(a *Automaton) *Automaton {
	var initialSet map[int]bool
	a, initialSet = reverse(a)
	a = determinizeSimple(a, initialSet)
	a, initialSet = reverse(a)
	a = determinizeSimple(a, initialSet)
	return a
}

/*
Simple original brics implementation of determinize()
Determinizes the given automaton using the given set of initial states.
*/
func determinizeSimple(a *Automaton, initialset map[int]bool) *Automaton {
	if a.numStates() == 0 {
		return a
	}
	points := a.startPoints()
	// subset construction
	sets := make(map[string]bool)
	hash := func(sets map[int]bool) string {
		n := util.NewOpenBitSet()
		for k, _ := range sets {
			n.Set(int64(k))
		}
		return n.String()
	}
	worklist := list.New()
	newstate := make(map[string]int)
	sets[hash(initialset)] = true
	worklist.PushBack(initialset)
	b := newAutomatonBuilder()
	b.createState()
	newstate[hash(initialset)] = 0
	t := newTransition()
	for worklist.Len() > 0 {
		s := worklist.Remove(worklist.Front()).(map[int]bool)
		r := newstate[hash(s)]
		for q, _ := range s {
			if a.IsAccept(q) {
				b.setAccept(r, true)
				break
			}
		}
		for n, point := range points {
			p := make(map[int]bool)
			for q, _ := range s {
				count := a.initTransition(q, t)
				for i := 0; i < count; i++ {
					a.nextTransition(t)
					if t.min <= point && point <= t.max {
						p[t.dest] = true
					}
				}
			}

			hashKey := hash(p)
			if _, ok := sets[hashKey]; !ok {
				sets[hashKey] = true
				worklist.PushBack(p)
				newstate[hashKey] = b.createState()
			}
			q := newstate[hashKey]
			min := point
			var max int
			if n+1 < len(points) {
				max = points[n+1] - 1
			} else {
				max = unicode.MaxRune
			}
			b.addTransitionRange(r, q, min, max)
		}
	}

	return removeDeadStates(b.finish())
}
