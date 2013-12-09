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

func (ts *BlockTermState) internalCopyFrom(ots *BlockTermState) {
	ts.OrdTermState.CopyFrom(ots.OrdTermState)
	ts.docFreq = ots.docFreq
	ts.totalTermFreq = ots.totalTermFreq
	ts.termBlockOrd = ots.termBlockOrd
	ts.blockFilePointer = ots.blockFilePointer
}

func (ts *BlockTermState) CopyFrom(other TermState) {
	if ots, ok := other.(*BlockTermState); ok {
		if ts.Self != nil && ots.Self != nil {
			ts.Self.CopyFrom(ots.Self)
		} else {
			ts.internalCopyFrom(ots)
		}
	} else if ts.Self != nil {
		// try copy from sub class
		ts.Self.CopyFrom(other)
	} else {
		panic(fmt.Sprintf("Can not copy from %v", reflect.TypeOf(other).Name()))
	}
}

func (ts *BlockTermState) Clone() TermState {
	if ts.Self != nil {
		return ts.Self.Clone()
	}
	clone := NewBlockTermState()
	clone.CopyFrom(ts)
	return clone
}

func (ts *BlockTermState) toString() string {
	return fmt.Sprintf("docFreq=%v totalTermFreq=%v termBlockOrd=%v blockFP=%v",
		ts.docFreq, ts.totalTermFreq, ts.termBlockOrd, ts.blockFilePointer)
}

func (ts *BlockTermState) String() string {
	if ts.Self != nil {
		return fmt.Sprintf("%v", ts.Self) // sub-class's string conversion first
	}
	return ts.toString()
}
