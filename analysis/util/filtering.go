package util

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

type FilteringTokenFilterSPI interface {
	// Override this method and return if the current input token
	// should be returned by IncrementToken().
	Accept() bool
}

/*
Abstract base class for TokenFilters that may remove tokens. You have
to implement Accept() and return a boolean if the current token
should be preserved. IncrementToken() uses this method to decide if a
token should be passed to the caller.

As for Lucene 4.4, GoLucene would panic when trying to disable
position increments when filtering terms.
*/
type FilteringTokenFilter struct {
	*TokenFilter
	spi                      FilteringTokenFilterSPI
	version                  util.Version
	enablePositionIncrements bool
	first                    bool
	skippedPositions         int
}

/* Creates a new FilteringTokenFilter. */
func NewFilteringTokenFilter(version util.Version, in TokenStream) *FilteringTokenFilter {
	return &FilteringTokenFilter{
		TokenFilter:              NewTokenFilter(in),
		version:                  version,
		enablePositionIncrements: true,
		first: true,
	}
}

func (f *FilteringTokenFilter) IncrementToken() (bool, error) {
	panic("not implemented yet")
}

func (f *FilteringTokenFilter) Reset() error {
	err := f.TokenFilter.Reset()
	if err == nil {
		f.first = true
		f.skippedPositions = 0
	}
	return err
}

func (f *FilteringTokenFilter) End() error {
	panic("not implemented yet")
}
