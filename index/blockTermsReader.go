package index

import (
	"fmt"
	"reflect"
)

// BlockTermState.java
/* Holds all state required for PostingsReaderBase
to produce a DocsEnum without re-seeking the
 terms dict. */
type BlockTermState struct {
	*OrdTermState
	// Allow sub-class to be converted
	Self TermState

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
		if ts.Self != nil {
			ts.Self.CopyFrom(ots.Self)
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
	return fmt.Sprintf("%v docFreq=%v totalTermFreq=%v termBlockOrd=%v blockFP=%v",
		ts.OrdTermState, ts.docFreq, ts.totalTermFreq, ts.termBlockOrd, ts.blockFilePointer)
}
