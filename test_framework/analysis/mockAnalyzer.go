package analysis

import (
	ca "github.com/balzaczyy/golucene/core/analysis"
	auto "github.com/balzaczyy/golucene/core/util/automaton"
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
	runAutomaton         *auto.CharacterRunAutomaton
	lowerCase            bool
	filter               *auto.CharacterRunAutomaton
	positionIncrementgap int
	random               *rand.Rand
	previousMappings     map[string]int
	enableChecks         bool
	maxTokenLength       int
}

// Creates a new MockAnalyzer.
func NewMockAnalyzer(r *rand.Rand, runAutomaton *auto.CharacterRunAutomaton, lowerCase bool, filter *auto.CharacterRunAutomaton) *MockAnalyzer {
	return &MockAnalyzer{
		AnalyzerImpl: ca.NewAnalyzerWithStrategy(ca.PER_FIELD_REUSE_STRATEGY),
		// TODO: this should be solved in a different way; Random should not be shared (!)
		random:           rand.New(rand.NewSource(r.Int63())),
		runAutomaton:     runAutomaton,
		lowerCase:        lowerCase,
		filter:           filter,
		previousMappings: make(map[string]int),
		enableChecks:     true,
		maxTokenLength:   DEFAULT_MAX_TOKEN_LENGTH,
	}
}

func NewMockAnalyzer3(r *rand.Rand, runAutomation *auto.CharacterRunAutomaton, lowerCase bool) *MockAnalyzer {
	return NewMockAnalyzer(r, runAutomation, lowerCase, EMPTY_STOPSET)
}

// Creates a Whitespace-lowercasing analyzer with no stopwords removal.
func NewMockAnalyzerWithRandom(r *rand.Rand) *MockAnalyzer {
	return NewMockAnalyzer3(r, WHITESPACE, true)
}

// analysis/MockTokenFilter.java

var EMPTY_STOPSET = auto.NewCharacterRunAutomaton(auto.MakeEmpty())

// analysis/MockTokenizer.java

/*
Tokenizer for testing.

This tokenizer is a replacement for WHITESPACE, SIMPLE, and KEYWORD
tokenizers. If you are writing a component such as a TokenFilter,
it's a great idea to test it wrapping this tokenizer instead for
extra checks. This tokenizer has the following behavior:

1. An internal state-machine is used for checking consumer
consistency. These checks can be disabled with DisableChecks(bool).
2. For convenience, optionally lowercases terms that it outputs.
*/
type MockTokenizer struct {
}

// Acts Similar to WhitespaceTokenizer
var WHITESPACE = auto.NewCharacterRunAutomaton(auto.NewRegExp("[^ \t\r\n]+").ToAutomaton())
