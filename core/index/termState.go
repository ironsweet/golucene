package index

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/index/model"
	"reflect"
)

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
