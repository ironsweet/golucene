package analysis

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

// analysis/Analyzer.java

/*
An Analyzer builds TokenStreams, which analyze text. It thus reprents a policy
for extracting index terms for text.

In order to define what analysis is done, subclass must define their
TokenStreamConents in CreateComponents(string, Reader). The components are
then reused in each call to TokenStream(string, Reader).

Also note that one should Clone() Analyzer for each Go routine if
default ReuseStrategy is used.
*/
type Analyzer interface {
	TokenStreamForReader(string, io.RuneReader) (TokenStream, error)
	// Returns a TokenStream suitable for fieldName, tokenizing the
	// contents of text.
	//
	// This method uses createComponents(string, Reader) to obtain an
	// instance of TokenStreamComponents. It returns the sink of the
	// components and stores the components internally. Subsequent
	// calls to this method will reuse the previously stored components
	// after resetting them through TokenStreamComponents.SetReader(Reader).
	//
	// NOTE: After calling this method, the consumer must follow the
	// workflow described in TokenStream to propperly consume its
	// contents. See the Analysis package documentation for some
	// examples demonstrating this.
	TokenStreamForString(fieldName, text string) (TokenStream, error)
	PositionIncrementGap(string) int
	OffsetGap(string) int
}

type AnalyzerSPI interface {
	// Creates a new TokenStreamComponents instance for this analyzer.
	CreateComponents(fieldName string, reader io.RuneReader) *TokenStreamComponents
	// Override this if you want to add a CharFilter chain.
	//
	// The default implementation returns reader unchanged.
	InitReader(fieldName string, reader io.RuneReader) io.RuneReader
}

type container struct {
	value interface{}
}

type AnalyzerImpl struct {
	Spi           AnalyzerSPI
	reuseStrategy ReuseStrategy
	version       util.Version
	// Since Go doesn't have ThreadLocal alternatives, to share
	// Analyzer, one must Clone() the Analyzer for each Go routine. It
	// also means the performance may not be competitive compared to
	// Lucene Java Analyzer.
	storedValue *container
}

/*
Create a new Analyzer, reusing the same set of components per-thread
across calls to TokenStream(string, Reader).
*/
func NewAnalyzer() *AnalyzerImpl {
	return NewAnalyzerWithStrategy(GLOBAL_REUSE_STRATEGY)
}

func NewAnalyzerWithStrategy(reuseStrategy ReuseStrategy) *AnalyzerImpl {
	ans := &AnalyzerImpl{
		reuseStrategy: reuseStrategy,
		version:       util.VERSION_LATEST,
		storedValue:   &container{nil},
	}
	ans.Spi = ans
	return ans
}

func (a *AnalyzerImpl) CreateComponents(fieldName string, reader io.RuneReader) *TokenStreamComponents {
	panic("must be inherited and implemented")
}

func (a *AnalyzerImpl) TokenStreamForReader(fieldName string, reader io.RuneReader) (TokenStream, error) {
	components := a.reuseStrategy.ReusableComponents(a, fieldName)
	r := a.InitReader(fieldName, reader)
	if components == nil {
		panic("not implemented yet")
	} else {
		if err := components.SetReader(r); err != nil {
			return nil, err
		}
	}
	return components.TokenStream(), nil
}

func (a *AnalyzerImpl) TokenStreamForString(fieldName, text string) (TokenStream, error) {
	components := a.reuseStrategy.ReusableComponents(a, fieldName)
	var strReader *ReusableStringReader
	if components == nil || components.reusableStringReader == nil {
		strReader = new(ReusableStringReader)
	} else {
		strReader = components.reusableStringReader
	}
	strReader.setValue(text)
	r := a.InitReader(fieldName, strReader)
	if components == nil {
		components = a.Spi.CreateComponents(fieldName, r)
		a.reuseStrategy.SetReusableComponents(a, fieldName, components)
	} else {
		err := components.SetReader(r)
		if err != nil {
			return nil, err
		}
	}
	components.reusableStringReader = strReader
	return components.TokenStream(), nil
}

func (a *AnalyzerImpl) InitReader(fieldName string, reader io.RuneReader) io.RuneReader {
	return reader
}

func (a *AnalyzerImpl) PositionIncrementGap(fieldName string) int {
	return 0
}

func (a *AnalyzerImpl) OffsetGap(fieldName string) int {
	return 1
}

func (a *AnalyzerImpl) SetVersion(v util.Version) {
	a.version = v
}

func (a *AnalyzerImpl) Version() util.Version {
	return a.version
}

type myTokenizer interface {
	SetReader(io.RuneReader) error
}

/*
This class encapsulates the outer components of a token stream. It
provides access to the source Tokenizer and the outer end (sink), an
instance of TokenFilter which also serves as the TokenStream returned
by Analyzer.tokenStream(string, Reader).
*/
type TokenStreamComponents struct {
	// Original source of tokens.
	source myTokenizer
	// Sink tokenStream, such as the outer tokenFilter decorating the
	// chain. This can be the source if there are no filters.
	sink TokenStream
	// Internal cache only used by Analyzer.TokenStreamForString().
	reusableStringReader *ReusableStringReader
	// Resets the encapculated components with the given reader. If the
	// components canno be reset, an error should be returned.
	SetReader func(io.RuneReader) error
}

func NewTokenStreamComponents(source myTokenizer, result TokenStream) *TokenStreamComponents {
	ans := &TokenStreamComponents{source: source, sink: result}
	ans.SetReader = func(reader io.RuneReader) error {
		return ans.source.SetReader(reader)
	}
	return ans
}

/* Returns the sink TokenStream */
func (cp *TokenStreamComponents) TokenStream() TokenStream {
	return cp.sink
}

// L329

// Strategy defining how TokenStreamComponents are reused per call to
// TokenStream(string, io.Reader)
type ReuseStrategy interface {
	// Gets the reusable TokenStreamComponents for the field with the
	// given name.
	ReusableComponents(*AnalyzerImpl, string) *TokenStreamComponents
	// Stores the given TokenStreamComponents as the reusable
	// components for the field with the given name.
	SetReusableComponents(*AnalyzerImpl, string, *TokenStreamComponents)
}

type ReuseStrategyImpl struct {
}

/* Returns the currently stored value */
func (rs *ReuseStrategyImpl) storedValue(a *AnalyzerImpl) interface{} {
	assert2(a.storedValue != nil, "this Analyzer is closed")
	return a.storedValue.value
}

/* Set the stored value. */
func (rs *ReuseStrategyImpl) setStoredValue(a *AnalyzerImpl, v interface{}) {
	assert2(a.storedValue != nil, "this Analyzer is closed")
	a.storedValue.value = v
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

/* A predefined ReuseStrategy that reuses the same components for every field */
var GLOBAL_REUSE_STRATEGY = new(GlobalReuseStrategy)

type GlobalReuseStrategy struct {
	*ReuseStrategyImpl
}

func (rs *GlobalReuseStrategy) ReusableComponents(a *AnalyzerImpl, fieldName string) *TokenStreamComponents {
	if ans := rs.storedValue(a); ans != nil {
		return ans.(*TokenStreamComponents)
	}
	return nil
}

func (rs *GlobalReuseStrategy) SetReusableComponents(a *AnalyzerImpl, fieldName string, components *TokenStreamComponents) {
	rs.setStoredValue(a, components)
}

// L423
// A predefined ReuseStrategy that reuses components per-field by
// maintaining a Map of TokenStreamComponent per field name.
var PER_FIELD_REUSE_STRATEGY = &PerFieldReuseStrategy{}

// Implementation of ReuseStrategy that reuses components per-field by
// maintianing a Map of TokenStreamComponent per field name.
type PerFieldReuseStrategy struct {
}

func (rs *PerFieldReuseStrategy) ReusableComponents(a *AnalyzerImpl, fieldName string) *TokenStreamComponents {
	panic("not implemented yet")
}

func (rs *PerFieldReuseStrategy) SetReusableComponents(a *AnalyzerImpl, fieldName string, components *TokenStreamComponents) {
	panic("not implemneted yet")
}

// analysis/ReusableStringReader.java

/* Internal class to enale reuse of the string reader by Analyzer.TokenStreamForString() */
type ReusableStringReader struct {
	s *bytes.Buffer
}

func (r *ReusableStringReader) setValue(s string) {
	r.s = bytes.NewBufferString(s)
}

func (r *ReusableStringReader) Read(p []byte) (int, error) {
	return r.s.Read(p)
}

func (r *ReusableStringReader) ReadRune() (rune, int, error) {
	return r.s.ReadRune()
}

func (r *ReusableStringReader) Close() error {
	r.s = nil
	return nil
}
