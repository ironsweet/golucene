package util

import (
	. "github.com/balzaczyy/golucene/core/analysis"
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
	spi FilteringTokenFilterSPI
}

func (f *FilteringTokenFilter) IncrementToken() (bool, error) {
	panic("not implemented yet")
}

func (f *FilteringTokenFilter) Reset() error {
	panic("not implemented yet")
}

func (f *FilteringTokenFilter) End() error {
	panic("not implemented yet")
}
