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
	if f.cache != nil {
		i := 0
		f.iterator = func() (*util.AttributeState, bool) {
			if i >= len(f.cache) {
				return nil, false
			}
			i++
			return f.cache[i-1], i == len(f.cache)
		}
	}
}
