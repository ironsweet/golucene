package standard

import (
	. "github.com/balzaczyy/golucene/core/analysis/core"
	. "github.com/balzaczyy/golucene/core/analysis/util"
	"github.com/balzaczyy/golucene/core/util"
)

// standard/StandardAnalyzer.java

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
}

/* Builds an analyzer with the given stop words. */
func NewStandardAnalyzerWithStopWords(matchVersion util.Version, stopWords map[string]bool) *StandardAnalyzer {
	return &StandardAnalyzer{NewStopwordAnalyzerBaseWithStopWords(matchVersion, stopWords)}
}

/* Buils an analyzer with the default stop words (STOP_WORDS_SET). */
func NewStandardAnalyzer(matchVersion util.Version) *StandardAnalyzer {
	return NewStandardAnalyzerWithStopWords(matchVersion, STOP_WORDS_SET)
}
