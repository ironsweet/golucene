package standard

import (
	"github.com/balzaczyy/golucene/core/util"
)

// standard/StandardAnalyzer.java

/*
Filters StandardTokenizer with StandardFilter, LowerCaseFilter and
StopFilter, using a list of English stop words.

# Version
You must specify the required Version compatibility when creating
StandardAnalyzer:
	- GoLucene supports 4.5+ only.
*/
type StandardAnalyzer struct {
}

/* Buils an analyzer with the default stop words (STOP_WORDS_SET). */
func NewStandardAnalyzer(matchVersion util.Version) *StandardAnalyzer {
	panic("not implemented yet")
}
