package automaton

// util/automaton/Transition.java

/*
Holds one transition from an {@link Automaton}.  This is typically
used temporarily when iterating through transitions by invoking
{@link Automaton#initTransition} and {@link Automaton#getNextTransition}.
*/
type Transition struct {
	source, dest   int
	min, max       int
	transitionUpto int
}

// Constructs a new singleton interval transition.
func newTransition() *Transition {
	return &Transition{
		transitionUpto: -1,
	}
}

func (t *Transition) String() string {
	panic("niy")
	// var b bytes.Buffer
	// appendCharString(t.min, &b)
	// if t.min != t.max {
	// 	b.WriteString("-")
	// 	appendCharString(t.max, &b)
	// }
	// fmt.Fprintf(&b, " -> %v", t.to.number)
	// return b.String()
}
