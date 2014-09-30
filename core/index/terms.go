package index

import (
	"fmt"
	// "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
	// "sort"
)

// index/Term.java

/*
A Term represents a word from text. This is the unit of search. It is
composed of two elements, the text of the word, as a string, and the
name of the field that the text occurred in.

Note that terms may represents more than words from text fields, but
also things like dates, email addresses, urls, etc.
*/
type Term struct {
	Field string
	Bytes []byte
}

func NewTermFromBytes(fld string, bytes []byte) *Term {
	return &Term{fld, bytes}
}

func NewTerm(fld string, text string) *Term {
	return &Term{fld, []byte(text)}
}

/*
Constructs a Term with the given field and empty text. This serves
two purposes: 1) reuse of a Term with the same field. 2) pattern for
a query.
*/
func NewEmptyTerm(fld string) *Term {
	return &Term{fld, nil}
}

type TermSorter []*Term

func (s TermSorter) Len() int      { return len(s) }
func (s TermSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TermSorter) Less(i, j int) bool {
	if s[i].Field == s[j].Field {
		return util.UTF8SortedAsUnicodeLess(s[i].Bytes, s[j].Bytes)
	}
	return s[i].Field < s[j].Field
}

func (t *Term) String() string {
	return fmt.Sprintf("%v:%v", t.Field, utf8ToString(t.Bytes))
}

// TermContext.java

type TermContext struct {
	TopReaderContext IndexReaderContext
	states           []TermState
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
	return &TermContext{TopReaderContext: ctx, states: make([]TermState, n)}
}

/**
 * Creates a {@link TermContext} from a top-level {@link IndexReaderContext} and the
 * given {@link Term}. This method will lookup the given term in all context's leaf readers
 * and register each of the readers containing the term in the returned {@link TermContext}
 * using the leaf reader's ordinal.
 * <p>
 * Note: the given context must be a top-level context.
 */
func NewTermContextFromTerm(ctx IndexReaderContext, t *Term) (tc *TermContext, err error) {
	assert(ctx != nil && ctx.Parent() == nil)
	perReaderTermState := NewTermContext(ctx)
	// fmt.Printf("prts.build term=%v\n", t)
	for _, leaf := range ctx.Leaves() {
		// fmt.Printf("  r=%v\n", leaf.reader)
		if fields := leaf.reader.Fields(); fields != nil {
			if terms := fields.Terms(t.Field); terms != nil {
				termsEnum := terms.Iterator(nil)
				ok, err := termsEnum.SeekExact(t.Bytes)
				if err != nil {
					return nil, err
				}
				if ok {
					termState, err := termsEnum.TermState()
					if err != nil {
						return nil, err
					}
					// log.Println("    found")
					df, err := termsEnum.DocFreq()
					if err != nil {
						return nil, err
					}
					tf, err := termsEnum.TotalTermFreq()
					if err != nil {
						return nil, err
					}
					perReaderTermState.register(termState, leaf.Ord, df, tf)
				}
			}
		}
	}
	return perReaderTermState, nil
}

func (tc *TermContext) register(state TermState, ord, docFreq int, totalTermFreq int64) {
	assert2(state != nil, "state must not be nil")
	assert(ord >= 0 && ord < len(tc.states))
	assert2(tc.states[ord] == nil, "state for ord: %v already registered", ord)
	tc.DocFreq += docFreq
	if tc.TotalTermFreq >= 0 && totalTermFreq >= 0 {
		tc.TotalTermFreq += totalTermFreq
	} else {
		tc.TotalTermFreq = -1
	}
	tc.states[ord] = state
}

func (tc *TermContext) State(ord int) TermState {
	assert(ord >= 0 && ord < len(tc.states))
	return tc.states[ord]
}

type MultiTerms struct {
	subs      []Terms
	subSlices []ReaderSlice
}

func NewMultiTerms(subs []Terms, subSlices []ReaderSlice) *MultiTerms {
	// TODO support customized comparator
	return &MultiTerms{subs, subSlices}
}

func (mt *MultiTerms) Iterator(reuse TermsEnum) TermsEnum {
	panic("not implemented yet")
}

func (mt *MultiTerms) DocCount() int {
	sum := 0
	for _, terms := range mt.subs {
		if v := terms.DocCount(); v != -1 {
			sum += v
		} else {
			return -1
		}
	}
	return sum
}

func (mt *MultiTerms) SumTotalTermFreq() int64 {
	var sum int64
	for _, terms := range mt.subs {
		if v := terms.SumTotalTermFreq(); v != -1 {
			sum += v
		} else {
			return -1
		}
	}
	return sum
}

func (mt *MultiTerms) SumDocFreq() int64 {
	var sum int64
	for _, terms := range mt.subs {
		if v := terms.SumDocFreq(); v != -1 {
			sum += v
		} else {
			return -1
		}
	}
	return sum
}
