package automaton

// util/automaton/DaciukMihovAutomatonBuilder.java

type DaciukMihovAutomatonBuilder struct {
}

// DFSA state with rune labels on transition
type dfsaState struct {
}

/*
Add another rune sequence to this automaton. The sequence must be
lexicographically larger or equal compared to any previous sequences
added to this automaton (the input must be sorted)
*/
func (builder *DaciukMihovAutomatonBuilder) add(current []rune) {
	panic("niy")
	// assert2(builder.stateRegistry != nil, "Automaton already builder.")
	// assert2(builder.previous == nil || builder.previous <= cur)
}

/*
Finalize the automaton and return the root state. No more strings can
be added to the builder after this call.
*/
func (builder *DaciukMihovAutomatonBuilder) complete() *dfsaState {
	panic("not implemented yet")
}

// Internal recursive traversal for conversion.
func convert(a *AutomatonBuilder, s *dfsaState, visited map[*dfsaState]int) int {
	panic("not implemented yet")
}

/*
Build a minimal, deterministic automaton from a sorted list of []byte
representing strings in UTF-8. These strings must be binary-sorted.
*/
func buildDaciukMihovAutomaton(input [][]byte) *Automaton {
	// builder := &DaciukMihovAutomatonBuilder{}
	// scratch := make([]rune, 0)
	// for _, b := range input {
	panic("not implemented yet")
	// 	builder.add(scratch)
	// }

	// a := newEmptyAutomaton()
	// a.initial = convert(
	// 	builder.complete(),
	// 	make(map[*dfsaState]*State))
	// a.deterministic = true
	// return a
}

// utils/CharsRef.java

func compareUTF16SortedAsUTF8(a, b []rune) int {
	// if a == b {
	// 	return 0
	// }

	for i, lenA, lenB := 0, len(a), len(b); i < lenA && i < lenB; i++ {
		aChar, bChar := a[i], b[i]
		if aChar != bChar {
			// http://icu-project.org/docs/papers/utf16_code_point_order.html

			// aChar != bChar, fix up each one if they're both in or above
			// the surrogate range, then compare them
			if aChar >= 0xd800 && bChar >= 0xd800 {
				if aChar >= 0xe000 {
					aChar -= 0x800
				} else {
					aChar += 0x2000
				}

				if bChar >= 0xe000 {
					bChar -= 0x800
				} else {
					bChar += 0x2000
				}
			}

			// now aChar and bChar are in code point order
			return int(aChar) - int(bChar)
		}
	}

	// One is a prefix of the other, or, they are equal:
	return len(a) - len(b)
}
