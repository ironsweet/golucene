package standard

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

/* Normalizes tokens extracted with StandardTokenizer */
type StandardFilter struct {
	*TokenFilter
	matchVersion util.Version
}

func newStandardFilter(matchVersion util.Version, in TokenStream) *StandardFilter {
	return &StandardFilter{
		TokenFilter:  NewTokenFilter(in),
		matchVersion: matchVersion,
	}
}

func (f *StandardFilter) IncrementToken() (bool, error) {
	panic("not implemented yet")
}
