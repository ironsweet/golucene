package core

import (
	. "github.com/balzaczyy/golucene/analysis/util"
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
	charUtils *CharacterUtils
}

/* Create a new LowerCaseFilter, that normalizes token text to lower case. */
func NewLowerCaseFilter(matchVersion util.Version, in TokenStream) *LowerCaseFilter {
	return &LowerCaseFilter{
		TokenFilter: NewTokenFilter(in),
		charUtils:   GetCharacterUtils(matchVersion),
	}
}

func (f *LowerCaseFilter) IncrementToken() (bool, error) {
	panic("not implemented yet")
}
