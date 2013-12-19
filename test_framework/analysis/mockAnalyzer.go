package analysis

import (
	ca "github.com/balzaczyy/golucene/core/analysis"
	"math/rand"
)

// analysis/MockAnalyzer.java

/*
Analyzer for testing

This analyzer is a replacement for Whitespace/Simple/KeywordAnalyzers
for unit tests. If you are testing a custom component such as a
queryparser or analyzer-wrapper that consumes analysis streams, it's
a great idea to test it with the anlyzer instead. MockAnalyzer as the
following behavior:

1. By default, the assertions in MockTokenizer are tured on for extra
checks that the consumer is consuming properly. These checks can be
disabled with SetEnableChecks(bool).
2. Payload data is randomly injected into the streams for more
thorough testing of payloads.
*/
type MockAnalyzer struct {
	*ca.AnalyzerImpl
}

// Creates a new MockAnalyzer.
func NewMockAnalyzerWithRandom(r *rand.Rand) *MockAnalyzer {
	panic("not implemented yet")
}
