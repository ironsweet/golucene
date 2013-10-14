package index

import (
	"fmt"
	"reflect"
)

// TermState.java
// Encapsulates all requried internal state to postiion the associated
// termsEnum without re-seeking
type TermState interface {
	CopyFrom(other TermState)
	Clone() TermState
}

var EMPTY_TERM_STATE = &EmptyTermState{}

type EmptyTermState struct{}

func (ts *EmptyTermState) CopyFrom(other TermState) {
	panic("not supported!")
}

func (ts *EmptyTermState) Clone() TermState {
	return ts
}

// An ordinal based TermState
type OrdTermState struct {
	// Term ordinal, i.e. it's position in the full list of
	// sorted terms
	ord int64
}

func (ts *OrdTermState) CopyFrom(other TermState) {
	if ots, ok := other.(*OrdTermState); ok {
		ts.ord = ots.ord
	} else {
		panic(fmt.Sprintf("Can not copy from %v", reflect.TypeOf(other).Name()))
	}
}

func (ts *OrdTermState) Clone() TermState {
	return &OrdTermState{ord: ts.ord}
}

func (ts *OrdTermState) String() string {
	return fmt.Sprintf("OrdtermState ord=%v", ts.ord)
}
