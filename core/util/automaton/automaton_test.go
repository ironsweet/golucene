package automaton

import (
	"github.com/balzaczyy/golucene/core/util"
	. "github.com/balzaczyy/golucene/test_framework/util"
	"log"
	"math/rand"
	"testing"
)

func TestRegExpToAutomaton(t *testing.T) {
	a := NewRegExp("[^ \t\r\n]+").ToAutomaton()
	// fmt.Println(a)
	assert(a.deterministic)
	assert(1 == a.initial.number)
	assert(2 == len(a.numberedStates))
}

func TestMinusSimple(t *testing.T) {
	assert(sameLanguage(makeChar('b'), minus(makeCharRange('a', 'b'), makeChar('a'))))
	assert(sameLanguage(MakeEmpty(), minus(makeChar('a'), makeChar('a'))))
}

func TestComplementSimple(t *testing.T) {
	a := makeChar('a')
	assert(sameLanguage(a, complement(complement(a))))
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
					log.Println("DEBUG", r)
					ok = false
				}
			}()
			l := make([]rune, 0, len(regexp))
			for _, ch := range regexp {
				l = append(l, ch)
			}
			log.Println("DEBUG", l)
			// log.Println("Trying", regexp)
			NewRegExpWithFlag(regexp, NONE)
			return
		}(regexp); ok {
			log.Println("Valid regexp found:", regexp)
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
		// log.Println("DEBUG way 0")
		return concatenate(a1, a2)
	case 1:
		// log.Println("DEBUG way 1")
		return union(a1, a2)
	case 2:
		// log.Println("DEBUG way 2")
		return intersection(a1, a2)
	default:
		// log.Println("DEBUG way 3")
		return minus(a1, a2)
	}
}
