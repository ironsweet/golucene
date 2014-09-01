package analysis

import (
	"github.com/balzaczyy/golucene/core/util"
)

type CachingTokenFilter struct {
	*TokenFilter
	cache      []*util.AttributeState
	iterator   func() (*util.AttributeState, bool)
	finalState *util.AttributeState
}

func NewCachingTokenFilter(input TokenStream) *CachingTokenFilter {
	return &CachingTokenFilter{
		TokenFilter: NewTokenFilter(input),
	}
}

func (f *CachingTokenFilter) Reset() {
	panic("not implemented yet")
}
