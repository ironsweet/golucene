package util

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/util"
)

// util/FilteringTokenFilter.java

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
	input                    TokenStream
	version                  util.Version
	posIncrAtt               PositionIncrementAttribute
	enablePositionIncrements bool
	first                    bool
	skippedPositions         int
}

/* Creates a new FilteringTokenFilter. */
func NewFilteringTokenFilter(spi FilteringTokenFilterSPI,
	version util.Version, in TokenStream) *FilteringTokenFilter {
	ans := &FilteringTokenFilter{
		spi:         spi,
		input:       in,
		TokenFilter: NewTokenFilter(in),
		version:     version,
		first:       true,
	}
	ans.posIncrAtt = ans.Attributes().Add("PositionIncrementAttribute").(PositionIncrementAttribute)
	ans.enablePositionIncrements = true
	return ans
}

func (f *FilteringTokenFilter) IncrementToken() (bool, error) {
	if f.enablePositionIncrements {
		f.skippedPositions = 0
		ok, err := f.input.IncrementToken()
		if err != nil {
			return false, err
		}
		for ok {
			if f.spi.Accept() {
				if f.skippedPositions != 0 {
					f.posIncrAtt.SetPositionIncrement(f.posIncrAtt.PositionIncrement() + f.skippedPositions)
				}
				return true, nil
			}
			f.skippedPositions += f.posIncrAtt.PositionIncrement()

			ok, err = f.input.IncrementToken()
			if err != nil {
				return false, err
			}
		}
	} else {
		ok, err := f.input.IncrementToken()
		if err != nil {
			return false, err
		}
		for ok {
			if f.spi.Accept() {
				if f.first {
					// first token having posinc=0 is illegal.
					if f.posIncrAtt.PositionIncrement() == 0 {
						f.posIncrAtt.SetPositionIncrement(1)
					}
					f.first = false
				}
				return true, nil
			}

			ok, err = f.input.IncrementToken()
			if err != nil {
				return false, err
			}
		}
	}
	// reached EOS -- return false
	return false, nil
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
	err := f.TokenFilter.End()
	if err == nil && f.enablePositionIncrements {
		f.posIncrAtt.SetPositionIncrement(f.posIncrAtt.PositionIncrement() + f.skippedPositions)
	}
	return err
}
