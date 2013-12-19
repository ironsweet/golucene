package analysis

import ()

// analysis/Analyzer.java

/*
An Analyzer builds TokenStreams, which analyze text. It thus reprents a policy
for extracting index terms for text.

In order to define what analysis is done, subclass must define their
TokenStreamConents in CreateComponents(string, Reader). The components are
then reused in each call to TokenStream(string, Reader).
*/
type Analyzer interface {
}

type AnalyzerImpl struct {
	reuseStrategy ReuseStrategy
}

func NewAnalyzer() *AnalyzerImpl {
	return &AnalyzerImpl{}
}

func NewAnalyzerWithStrategy(reuseStrategy ReuseStrategy) *AnalyzerImpl {
	return &AnalyzerImpl{reuseStrategy: reuseStrategy}
}

// L329

// Strategy defining how TokenStreamComponents are reused per call to
// TokenStream(string, io.Reader)
type ReuseStrategy interface {
	// Gets the reusable TokenStreamComponents for the field with the
	// given name.
	// ReusableComponents(analyzer Analyzer, fieldName string)
	// Stores the given TokenStreamComponents as the reusable
	// components for the field with the given name.
	// SetReusableComponents(analyzer Analyzer, fieldName string, components TokenStreamComponents)
}
