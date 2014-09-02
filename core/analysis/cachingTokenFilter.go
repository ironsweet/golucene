package analysis

import (
	"github.com/balzaczyy/golucene/core/util"
)

type CachingTokenFilter struct {
	*TokenFilter
	cache      []*util.AttributeState
	cacheIdx   int
	iterator   func() (*util.AttributeState, bool)
	finalState *util.AttributeState
}

func NewCachingTokenFilter(input TokenStream) *CachingTokenFilter {
	return &CachingTokenFilter{
		TokenFilter: NewTokenFilter(input),
	}
}

func (f *CachingTokenFilter) IncrementToken() (bool, error) {
	if f.cache == nil { // fill cache lazily
		if err := f.fillCache(); err != nil {
			return false, err
		}
		f.Reset()
	}

	if f.cacheIdx >= len(f.cache) {
		// the cache is exhausted, return false
		return false, nil
	}
	// Since the TokenFilter can be reset, the tokens need to be preserved as immutable.
	f.Attributes().RestoreState(f.cache[f.cacheIdx])
	f.cacheIdx++
	return true, nil
}

func (f *CachingTokenFilter) Reset() {
	if f.cache != nil {
		f.cacheIdx = 0
	}
}

func (f *CachingTokenFilter) fillCache() error {
	ok, err := f.input.IncrementToken()
	for ok && err == nil {
		f.cache = append(f.cache, f.Attributes().CaptureState())
		ok, err = f.input.IncrementToken()
	}
	if err != nil {
		return err
	}
	// capture final state
	if err = f.input.End(); err != nil {
		return err
	}
	f.finalState = f.Attributes().CaptureState()
	return nil
}
