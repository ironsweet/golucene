package automaton

// util/automaton/RunAutomaton.java

// Finite-state automaton with fast run operation.
type RunAutomaton struct {
}

func (ra *RunAutomaton) String() string {
	panic("not implemented yet")
}

// Automaton representation for matching []char
type CharacterRunAutomaton struct {
	*RunAutomaton
}

func NewCharacterRunAutomaton(a *Automaton) *CharacterRunAutomaton {
	panic("not implemented yet")
}
