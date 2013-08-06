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

type Terms interface {
	Iterator(reuse TermsEnum) TermsEnum
	DocCount() int
	SumTotalTermFreq() int64
	SumDocFreq() int64
}

const (
	SEEK_STATUS_END       = 1
	SEEK_STATUS_FOUND     = 2
	SEEK_STATUS_NOT_FOUND = 3
)

var (
	TERMS_ENUM_EMPTY = &TermsEnumEmpty{}
)

type SeekStatus int

type TermsEnum interface {
	// BytesRefIterator
	Next() (buf []byte, err error)
	Comparator() sort.Interface

	Attributes() util.AttributeSource
	SeekExactUsingCache(text []byte, useCache bool) bool
	SeekCeilUsingCache(text []byte, useCache bool) SeekStatus
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

func (e *TermsEnumImpl) SeekExactUsingCache(text []byte, useCache bool) bool {
	return e.SeekCeilUsingCache(text, useCache) == SEEK_STATUS_FOUND
}

func (e *TermsEnumImpl) SeekCeil(text []byte) SeekStatus {
	return e.SeekCeilUsingCache(text, true)
}

func (e *TermsEnumImpl) SeekExactFromLast(text []byte, state TermState) error {
	if !e.SeekExactUsingCache(text, true) {
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

type TermContext struct {
	TopReaderContext IndexReaderContext
	states           []*TermState
	DocFreq          int
	TotalTermFreq    int64
}

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

func NewTermContextFromTerm(ctx IndexReaderContext, t Term, cache bool) *TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	perReaderTermState := NewTermContext(ctx)
	for _, v := range ctx.Leaves() {
		fields := v.reader.Fields()
		if fields != nil {
			terms := fields.Terms(t.Field)
			if terms != nil {
				termsEnum := terms.Iterator(nil)
				if termsEnum.SeekExactUsingCache(t.Bytes, cache) {
					termState := termsEnum.TermState()
					perReaderTermState.register(termState, v.Ord, termsEnum.DocFreq(), termsEnum.TotalTermFreq())
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
