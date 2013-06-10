package index

type Term struct {
	Field string
	Bytes []byte
}

type Terms interface {
	Iterator(reuse TermsEnum) TermsEnum
	DocCount() int
	SumTotalTermFreq() int
	SumDocFreq() int
}

const (
	SEEK_STATUS_END       = 1
	SEEK_STATUS_FOUND     = 2
	SEEK_STATUS_NOT_FOUND = 3
)

type TermsEnum interface {
	SeekExact(text []byte, useCache bool) bool
	SeekCeil(text []byte, useCache bool) int
	docFreq() int
	totalTermFreq() int64
	TermState() TermState
}

type AbstractTermsEnum struct {
	SeekCeil func(text []byte, useCache bool) int
}

func (iter *AbstractTermsEnum) SeekExact(text []byte, useCache bool) bool {
	return iter.SeekCeil(text, useCache) == SEEK_STATUS_FOUND
}

func (iter *AbstractTermsEnum) TermState() TermState {
	return TermState{copyFrom: func(other TermState) { panic("not implemented yet!") }}
}

type TermContext struct {
	TopReaderContext IndexReaderContext
	states           []*TermState
	DocFreq          int
	TotalTermFreq    int64
}

func NewTermContext(ctx IndexReaderContext) TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	var n int
	if ctx.Leaves() == nil {
		n = 1
	} else {
		n = len(ctx.Leaves())
	}
	return TermContext{TopReaderContext: ctx, states: make([]*TermState, n)}
}

func NewTermContextFromTerm(ctx IndexReaderContext, t Term, cache bool) TermContext {
	// assert ctx != nil && ctx.IsTopLevel
	perReaderTermState := NewTermContext(ctx)
	for _, v := range ctx.Leaves() {
		fields := v.reader.Fields()
		if fields != nil {
			terms := fields.terms(t.Field)
			if terms != nil {
				termsEnum := terms.Iterator(nil)
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

func (mt MultiTerms) SumTotalTermFreq() int {
	panic("not implemented yet")
}

func (mt MultiTerms) SumDocFreq() int {
	panic("not implemented yet")
}
