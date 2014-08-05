package spi

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/index/model"
	"reflect"
)

// index/OrdTermState.java

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
	return fmt.Sprintf("TermState ord=%v", ts.ord)
}

// BlockTermState.java
/* Holds all state required for PostingsReaderBase
to produce a DocsEnum without re-seeking the
 terms dict. */
type BlockTermState struct {
	*OrdTermState
	// Allow sub-class to be converted
	Self TermState

	// how many docs have this term
	DocFreq int
	// total number of occurrences of this term
	TotalTermFreq int64

	// the term's ord in the current block
	TermBlockOrd int
	// fp into the terms dict primary file (_X.tim) that holds this term
	blockFilePointer int64
}

func NewBlockTermState() *BlockTermState {
	return &BlockTermState{OrdTermState: &OrdTermState{}}
}

func (ts *BlockTermState) CopyFrom(other TermState) {
	if ts.Self != nil {
		ts.Self.CopyFrom(other)
		return
	}
	ts.CopyFrom_(other)
}

func (ts *BlockTermState) CopyFrom_(other TermState) {
	if ots, ok := other.(*BlockTermState); ok {
		ts.OrdTermState.CopyFrom(ots.OrdTermState)
		ts.DocFreq = ots.DocFreq
		ts.TotalTermFreq = ots.TotalTermFreq
		ts.TermBlockOrd = ots.TermBlockOrd
		ts.blockFilePointer = ots.blockFilePointer
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

func (ts *BlockTermState) String() string {
	return fmt.Sprintf("docFreq=%v totalTermFreq=%v termBlockOrd=%v blockFP=%v",
		ts.DocFreq, ts.TotalTermFreq, ts.TermBlockOrd, ts.blockFilePointer)
}
