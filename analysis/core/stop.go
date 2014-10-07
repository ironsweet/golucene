package core

import (
	. "github.com/balzaczyy/golucene/analysis/util"
	. "github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/util"
)

// core/StopAnalyzer.java

/* An unmodifiable set containing some common English words that are not usually useful for searching. */
var ENGLISH_STOP_WORDS_SET = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "but": true, "by": true,
	"for": true, "if": true, "in": true, "into": true, "is": true, "it": true,
	"no": true, "not": true, "of": true, "on": true, "or": true, "such": true,
	"that": true, "the": true, "their": true, "then": true, "there": true, "these": true,
	"they": true, "this": true, "to": true, "was": true, "will": true, "with": true,
}

// core/StopFilter.java

/*
Removes stop words from a token stream.

You may specify the Version
compatibility when creating StopFilter:

	- As of 3.1, StopFilter correctly handles Unicode 4.0 supplementary
	characters in stopwords and position increments are preserved
*/
type StopFilter struct {
	*FilteringTokenFilter
	stopWords map[string]bool
	termAtt   CharTermAttribute
}

/*
Constructs a filter which removes words from the input TokenStream
that are named in the Set.
*/
func NewStopFilter(matchVersion util.Version,
	in TokenStream, stopWords map[string]bool) *StopFilter {

	ans := &StopFilter{stopWords: stopWords}
	ans.FilteringTokenFilter = NewFilteringTokenFilter(ans, matchVersion, in)
	ans.termAtt = ans.Attributes().Add("CharTermAttribute").(CharTermAttribute)
	return ans
}

func (f *StopFilter) Accept() bool {
	term := string(f.termAtt.Buffer()[:f.termAtt.Length()])
	_, ok := f.stopWords[term]
	return !ok
}
