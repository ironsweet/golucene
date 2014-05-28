package standard

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

/* Normalizes tokens extracted with StandardTokenizer */
type StandardFilter struct {
	*TokenFilter
}

func newStandardFilter(matchVersion util.Version, in TokenStream) *StandardFilter {
	panic("not implemented yet")
}

func (f *StandardFilter) IncrementToken() (bool, error) {
	panic("not implemented yet")
}
