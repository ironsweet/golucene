package automaton

// util/automaton/Automaton.java

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
}

// util/automaton/BasicAutomata.java

func MakeEmpty() *Automaton {
	panic("not implemented yet")
}
