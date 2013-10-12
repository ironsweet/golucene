package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/util"
	"sort"
)

type Term struct {
	Field string
	Bytes []byte
}

func NewTerm(fld string, text string) Term {
	return Term{fld, []byte(text)}
}

type Terms interface {
	Iterator(reuse TermsEnum) TermsEnum
	DocCount() int
	SumTotalTermFreq() int64
	SumDocFreq() int64
}

// TermsEnum.java

var (
	TERMS_ENUM_EMPTY = &TermsEnumEmpty{}
)

type SeekStatus int

const (
	SEEK_STATUS_END       = 1
	SEEK_STATUS_FOUND     = 2
	SEEK_STATUS_NOT_FOUND = 3
)

type TermsEnum interface {
	// BytesRefIterator
	Next() (buf []byte, err error)
	Comparator() sort.Interface

	Attributes() util.AttributeSource
	SeekExact(text []byte) bool
	SeekCeil(text []byte) SeekStatus
	SeekExactByPosition(ord int64) error
	SeekExactFromLast(text []byte, state TermState) error
	Term() []byte
	Ord() int64
	DocFreq() int
	TotalTermFreq() int64
	Docs(liveDocs util.Bits, reuse DocsEnum) DocsEnum
	DocsByFlags(liveDocs util.Bits, reuse DocsEnum, flags int) DocsEnum
	DocsAndPositions(liveDocs util.Bits, reuse DocsAndPositionsEnum) DocsAndPositionsEnum
	DocsAndPositionsByFlags(liveDocs util.Bits, reuse DocsAndPositionsEnum, flags int) DocsAndPositionsEnum
	TermState() TermState
}

type TermsEnumImpl struct {
	TermsEnum
	atts util.AttributeSource
}

func newTermsEnumImpl(self TermsEnum) *TermsEnumImpl {
	return &TermsEnumImpl{self, util.NewAttributeSource()}
}

func (e *TermsEnumImpl) Attributes() util.AttributeSource {
	return e.atts
}

func (e *TermsEnumImpl) SeekExact(text []byte) bool {
	return e.SeekCeil(text) == SEEK_STATUS_FOUND
}

func (e *TermsEnumImpl) SeekExactFromLast(text []byte, state TermState) error {
	if !e.SeekExact(text) {
		panic(fmt.Sprintf("term %v does not exist", text))
	}
	return nil
}

func (e *TermsEnumImpl) Docs(liveDocs util.Bits, reuse DocsEnum) DocsEnum {
	return e.DocsByFlags(liveDocs, reuse, DOCS_ENUM_FLAG_FREQS)
}

func (e *TermsEnumImpl) DocsAndPositions(liveDocs util.Bits, reuse DocsAndPositionsEnum) DocsAndPositionsEnum {
	return e.DocsAndPositionsByFlags(liveDocs, reuse, DOCS_POSITIONS_ENUM_FLAG_OFF_SETS|DOCS_POSITIONS_ENUM_FLAG_PAYLOADS)
}

func (e *TermsEnumImpl) TermState() TermState {
	return TermState{copyFrom: func(other TermState) { panic("not supported!") }}
}

type TermsEnumEmpty struct {
	*TermsEnumImpl
}

func (e *TermsEnumEmpty) SeekCeilUsingCache(term []byte, useCache bool) SeekStatus {
	return SEEK_STATUS_END
}

func (e *TermsEnumEmpty) SeekExactByPosition(ord int64) error {
	return nil
}

func (e *TermsEnumEmpty) Term() []byte {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) Comparator() sort.Interface {
	return nil
}

func (e *TermsEnumEmpty) DocFreq() int {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) TotalTermFreq() int64 {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) Ord() int64 {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) DocsByFlags(liveDocs util.Bits, reuse DocsEnum, flags int) DocsEnum {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) DocsAndPositionsByFlags(liveDocs util.Bits, reuse DocsAndPositionsEnum, flags int) DocsAndPositionsEnum {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) Next() (term []byte, err error) {
	return nil, nil
}

func (e *TermsEnumEmpty) TermState() TermState {
	panic("this method should never be called")
}

func (e *TermsEnumEmpty) SeekExactFromLast(term []byte, state TermState) error {
	panic("this method should never be called")
}

// TermContext.java

type TermContext struct {
	TopReaderContext IndexReaderContext
	states           []*TermState
	DocFreq          int
	TotalTermFreq    int64
}

/**
 * Creates an empty {@link TermContext} from a {@link IndexReaderContext}
 */
func NewTermContext(ctx IndexReaderContext) *TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	var n int
	if ctx.Leaves() == nil {
		n = 1
	} else {
		n = len(ctx.Leaves())
	}
	return &TermContext{TopReaderContext: ctx, states: make([]*TermState, n)}
}

/**
 * Creates a {@link TermContext} from a top-level {@link IndexReaderContext} and the
 * given {@link Term}. This method will lookup the given term in all context's leaf readers
 * and register each of the readers containing the term in the returned {@link TermContext}
 * using the leaf reader's ordinal.
 * <p>
 * Note: the given context must be a top-level context.
 */
func NewTermContextFromTerm(ctx IndexReaderContext, t Term) *TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	perReaderTermState := NewTermContext(ctx)
	for _, leaf := range ctx.Leaves() {
		if fields := leaf.reader.Fields(); fields != nil {
			if terms := fields.Terms(t.Field); terms != nil {
				if termsEnum := terms.Iterator(nil); termsEnum.SeekExact(t.Bytes) {
					termState := termsEnum.TermState()
					perReaderTermState.register(termState, leaf.Ord, termsEnum.DocFreq(), termsEnum.TotalTermFreq())
				}
			}
		}
	}
	return perReaderTermState
}

func (tc TermContext) register(state TermState, ord, docFreq int, totalTermFreq int64) {
	// assert ord >= 0 && ord < len(states)
	// assert states[ord] == null : "state for ord: " + ord + " already registered";
	tc.DocFreq += docFreq
	if tc.TotalTermFreq >= 0 && totalTermFreq >= 0 {
		tc.TotalTermFreq += totalTermFreq
	} else {
		tc.TotalTermFreq = -1
	}
	tc.states[ord] = &state
}

func (tc TermContext) State(ord int) *TermState {
	// asert ord >= 0 && ord < len(states)
	return tc.states[ord]
}

type TermState struct {
	copyFrom func(other TermState)
}

type MultiTerms struct {
	subs      []Terms
	subSlices []ReaderSlice
}

func NewMultiTerms(subs []Terms, subSlices []ReaderSlice) MultiTerms {
	// TODO support customized comparator
	return MultiTerms{subs, subSlices}
}

func (mt MultiTerms) Iterator(reuse TermsEnum) TermsEnum {
	panic("not implemented yet")
}

func (mt MultiTerms) DocCount() int {
	panic("not implemented yet")
}

func (mt MultiTerms) SumTotalTermFreq() int64 {
	panic("not implemented yet")
}

func (mt MultiTerms) SumDocFreq() int64 {
	panic("not implemented yet")
}
