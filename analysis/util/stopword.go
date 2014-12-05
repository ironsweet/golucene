package util

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

/* Base class for Analyzers that need to make use of stopword sets. */
type StopwordAnalyzerBase struct {
	*AnalyzerImpl
	stopwords    map[string]bool
	matchVersion util.Version
}

func NewStopwordAnalyzerBaseWithStopWords(version util.Version, stopwords map[string]bool) *StopwordAnalyzerBase {
	ans := new(StopwordAnalyzerBase)
	ans.AnalyzerImpl = NewAnalyzer()
	ans.matchVersion = version
	ans.stopwords = make(map[string]bool)
	if stopwords != nil {
		for k, v := range stopwords {
			ans.stopwords[k] = v
		}
	}
	return ans
}
