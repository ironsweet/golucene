package core

import (
	. "github.com/balzaczyy/golucene/analysis/util"
	. "github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/util"
)

// core/LowerCaseFilter.java

/*
Normalizes token text to lower case.

You may specify the Version
compatibility when creating LowerCaseFilter:

	- As of 3.1, supplementary characters are properly lowercased.
*/
type LowerCaseFilter struct {
	*TokenFilter
	input     TokenStream
	charUtils *CharacterUtils
	termAtt   CharTermAttribute
}

/* Create a new LowerCaseFilter, that normalizes token text to lower case. */
func NewLowerCaseFilter(matchVersion util.Version, in TokenStream) *LowerCaseFilter {
	ans := &LowerCaseFilter{
		TokenFilter: NewTokenFilter(in),
		input:       in,
		charUtils:   GetCharacterUtils(matchVersion),
	}
	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	return ans
}

func (f *LowerCaseFilter) IncrementToken() (bool, error) {
	ok, err := f.input.IncrementToken()
	if err != nil {
		return false, err
	}
	if ok {
		f.charUtils.ToLowerCase(f.termAtt.Buffer()[:f.termAtt.Length()])
		return true, nil
	}
	return false, nil
}
