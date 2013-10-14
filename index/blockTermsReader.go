package index

import (
	"fmt"
	"reflect"
)

// BlockTermState.java
/*
Holds all state required for PostingsReaderBase
to produce a DocsEnum without re-seeking the
terms dict.
*/
type BlockTermState struct {
	*OrdTermState
	other TermState

	// how many docs have this term
	docFreq int
	// total number of occurrences of this term
	totalTermFreq int64

	// the term's ord in the current block
	termBlockOrd int
	// fp into the terms dict primary file (_X.tim) that holds this term
	blockFilePointer int64
}

func NewBlockTermState() *BlockTermState {
	return &BlockTermState{OrdTermState: &OrdTermState{}}
}

func (ts *BlockTermState) CopyFrom(other TermState) {
	if ots, ok := other.(*BlockTermState); ok {
		ts.OrdTermState.CopyFrom(ots.OrdTermState)
		if ts.other != nil {
			ts.other.CopyFrom(ots.other)
		}
		ts.docFreq = ots.docFreq
		ts.totalTermFreq = ots.totalTermFreq
		ts.termBlockOrd = ots.termBlockOrd
		ts.blockFilePointer = ots.blockFilePointer
	} else {
		panic(fmt.Sprintf("Can not copy from %v", reflect.TypeOf(other).Name()))
	}
}

func (ts *BlockTermState) Clone() TermState {
	clone := NewBlockTermState()
	clone.CopyFrom(ts)
	return clone
}

func (ts *BlockTermState) String() string {
	return fmt.Sprintf("%v docFreq=%v totalTermFreq=%v termBlockOrd=%v blockFP=%v other=%v",
		ts.OrdTermState, ts.docFreq, ts.totalTermFreq, ts.termBlockOrd, ts.blockFilePointer, ts.other)
}
