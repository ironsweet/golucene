package util

import (
	. "github.com/balzaczyy/golucene/core/analysis"
)

// util/StopwordAnalyzerBase.java

/* Base class for Analyzers that need to make use of stopword sets. */
type StopwordAnalyzerBase struct {
	*AnalyzerImpl
	stopwords map[string]bool
}

func NewStopwordAnalyzerBaseWithStopWords(stopwords map[string]bool) *StopwordAnalyzerBase {
	ans := new(StopwordAnalyzerBase)
	ans.AnalyzerImpl = NewAnalyzer()
	ans.stopwords = make(map[string]bool)
	if stopwords != nil {
		for k, v := range stopwords {
			ans.stopwords[k] = v
		}
	}
	return ans
}
