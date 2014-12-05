package standard

import (
	. "github.com/balzaczyy/golucene/analysis/core"
	. "github.com/balzaczyy/golucene/analysis/util"
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

// standard/StandardAnalyzer.java

/* Default maximum allowed token length */
const DEFAULT_MAX_TOKEN_LENGTH = 255

/* An unmodifiable set containing some common English words that are usually not useful for searching */
var STOP_WORDS_SET = ENGLISH_STOP_WORDS_SET

/*
Filters StandardTokenizer with StandardFilter, LowerCaseFilter and
StopFilter, using a list of English stop words.

# Version
You must specify the required Version compatibility when creating
StandardAnalyzer:
	- GoLucene supports 4.5+ only.
*/
type StandardAnalyzer struct {
	*StopwordAnalyzerBase
	matchVersion   util.Version
	stopWordSet    map[string]bool
	maxTokenLength int
}

/* Builds an analyzer with the given stop words. */
func NewStandardAnalyzerWithStopWords(matchVersion util.Version, stopWords map[string]bool) *StandardAnalyzer {
	ans := &StandardAnalyzer{
		NewStopwordAnalyzerBaseWithStopWords(matchVersion, stopWords),
		matchVersion,
		stopWords,
		DEFAULT_MAX_TOKEN_LENGTH,
	}
	ans.Spi = ans
	return ans
}

/* Buils an analyzer with the default stop words (STOP_WORDS_SET). */
func NewStandardAnalyzer(matchVersion util.Version) *StandardAnalyzer {
	return NewStandardAnalyzerWithStopWords(matchVersion, STOP_WORDS_SET)
}

func (a *StandardAnalyzer) CreateComponents(fieldName string, reader io.RuneReader) *TokenStreamComponents {
	src := newStandardTokenizer(a.matchVersion, reader)
	src.maxTokenLength = a.maxTokenLength
	var tok TokenStream = newStandardFilter(a.matchVersion, src)
	tok = NewLowerCaseFilter(a.matchVersion, tok)
	tok = NewStopFilter(a.matchVersion, tok, a.stopWordSet)
	ans := NewTokenStreamComponents(src, tok)
	super := ans.SetReader
	ans.SetReader = func(reader io.RuneReader) error {
		src.maxTokenLength = a.maxTokenLength
		return super(reader)
	}
	return ans
}
