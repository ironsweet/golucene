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
}
