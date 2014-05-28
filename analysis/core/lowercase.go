package core

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

/*
Normalizes token text to lower case.

Version

You mus tspecify the required Version compatibility when creating
LowerCaseFilter:

	- As of 3.1, supplementary characters are properly lowercased.
*/
type LowerCaseFilter struct {
	*TokenFilter
}

func NewLowerCaseFilter(matchVersion util.Version, in TokenStream) *LowerCaseFilter {
	panic("not implemented yet")
}

func (f *LowerCaseFilter) IncrementToken() (bool, error) {
	panic("not implemented yet")
}
