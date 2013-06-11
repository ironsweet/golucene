package index

import (
	"lucene/util"
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
	TERMS_ENUM_EMPTY = TermsEnum{}
)

type TermsEnum struct {
	SeekCeil        func(text []byte, useCache bool) int
	docFreq         func() int
	totalTermFreq   func() int64
	DocsForTermFlag func(liveDocs util.Bits, reuse DocsEnum, flags int) DocsEnum
}

func (iter *TermsEnum) SeekExact(text []byte, useCache bool) bool {
	return iter.SeekCeil(text, useCache) == SEEK_STATUS_FOUND
}

func (iter *TermsEnum) DocsForTerm(liveDocs util.Bits, reuse DocsEnum) DocsEnum {
	return iter.DocsForTermFlag(liveDocs, reuse, DOCS_ENUM_FLAG_FREQS)
}

func (iter *TermsEnum) TermState() TermState {
	return TermState{copyFrom: func(other TermState) { panic("not supported!") }}
}

type TermContext struct {
	TopReaderContext *IndexReaderContext
	states           []*TermState
	DocFreq          int
	TotalTermFreq    int64
}

func NewTermContext(ctx *IndexReaderContext) *TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	var n int
	if ctx.Leaves() == nil {
		n = 1
	} else {
		n = len(ctx.Leaves())
	}
	return &TermContext{TopReaderContext: ctx, states: make([]*TermState, n)}
}

func NewTermContextFromTerm(ctx *IndexReaderContext, t Term, cache bool) *TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	perReaderTermState := NewTermContext(ctx)
	for _, v := range ctx.Leaves() {
		fields := v.reader.Fields()
		if fields != nil {
			terms := fields.terms(t.Field)
			if terms != nil {
				termsEnum := terms.Iterator(TermsEnum{}) // empty TermsEnum
				if termsEnum.SeekExact(t.Bytes, cache) {
					termState := termsEnum.TermState()
					perReaderTermState.register(termState, v.Ord, termsEnum.docFreq(), termsEnum.totalTermFreq())
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
