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

/*
This class encapsulates the outer components of a token stream. It
provides access to the source Tokenizer and the outer end (sink), an
instance of TokenFilter which also serves as the TokenStream returned
by Analyzer.tokenStream(string, Reader).
*/
type TokenStreamComponents struct {
}

// L329

// Strategy defining how TokenStreamComponents are reused per call to
// TokenStream(string, io.Reader)
type ReuseStrategy interface {
	// Gets the reusable TokenStreamComponents for the field with the
	// given name.
	ReusableComponents(analyzer Analyzer, fieldName string) *TokenStreamComponents
	// Stores the given TokenStreamComponents as the reusable
	// components for the field with the given name.
	SetReusableComponents(analyzer Analyzer, fieldName string, components *TokenStreamComponents)
}

// L423
// A predefined ReuseStrategy that reuses components per-field by
// maintaining a Map of TokenStreamComponent per field name.
var PER_FIELD_REUSE_STRATEGY = &PerFieldReuseStrategy{}

// Implementation of ReuseStrategy that reuses components per-field by
// maintianing a Map of TokenStreamComponent per field name.
type PerFieldReuseStrategy struct {
}

func (rs *PerFieldReuseStrategy) ReusableComponents(a Analyzer, fieldName string) *TokenStreamComponents {
	panic("not implemented yet")
}

func (rs *PerFieldReuseStrategy) SetReusableComponents(a Analyzer, fieldName string, components *TokenStreamComponents) {
	panic("not implemneted yet")
}
