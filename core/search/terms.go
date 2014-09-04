package search

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	. "github.com/balzaczyy/golucene/core/index/model"
	. "github.com/balzaczyy/golucene/core/search/model"
	"github.com/balzaczyy/golucene/core/util"
	"reflect"
)

type TermQuery struct {
	*AbstractQuery
	term               *index.Term
	docFreq            int
	perReaderTermState *index.TermContext
}

func NewTermQuery(t *index.Term) *TermQuery {
	return NewTermQueryWithDocFreq(t, -1)
}

func NewTermQueryWithDocFreq(t *index.Term, docFreq int) *TermQuery {
	ans := &TermQuery{}
	ans.AbstractQuery = NewAbstractQuery(ans)
	ans.term = t
	ans.docFreq = docFreq
	return ans
}

func (q *TermQuery) CreateWeight(ss *IndexSearcher) (w Weight, err error) {
	ctx := ss.TopReaderContext()
	var termState *index.TermContext
	if q.perReaderTermState == nil || q.perReaderTermState.TopReaderContext != ctx {
		// make TermQuery single-pass if we don't have a PRTS or if the context differs!
		termState, err = index.NewTermContextFromTerm(ctx, q.term)
		if err != nil {
			return nil, err
		}
	} else {
		// PRTS was pre-build for this IS
		termState = q.perReaderTermState
	}

	// we must not ignore the given docFreq - if set use the given value (lie)
	if q.docFreq != -1 {
		termState.DocFreq = q.docFreq
	}

	return NewTermWeight(q, ss, termState), nil
}

func (q *TermQuery) ToString(field string) string {
	var buf bytes.Buffer
	if q.term.Field != field {
		buf.WriteString(q.term.Field)
		buf.WriteRune(':')
	}
	buf.WriteString(string(q.term.Bytes))
	if q.boost != 1.0 {
		buf.WriteString(fmt.Sprintf("^%v", q.boost))
	}
	return buf.String()
}

type TermWeight struct {
	*WeightImpl
	*TermQuery
	similarity Similarity
	stats      SimWeight
	termStates *index.TermContext
}

func NewTermWeight(owner *TermQuery, ss *IndexSearcher, termStates *index.TermContext) *TermWeight {
	assert(termStates != nil)
	sim := ss.similarity
	ans := &TermWeight{
		TermQuery:  owner,
		similarity: sim,
		stats: sim.computeWeight(
			owner.boost,
			ss.CollectionStatistics(owner.term.Field),
			ss.TermStatistics(owner.term, termStates)),
		termStates: termStates,
	}
	ans.WeightImpl = newWeightImpl(ans)
	return ans
}

func (tw *TermWeight) String() string {
	return fmt.Sprintf("weight(%v)", tw.TermQuery)
}

func (tw *TermWeight) ValueForNormalization() float32 {
	return tw.stats.ValueForNormalization()
}

func (tw *TermWeight) Normalize(norm float32, topLevelBoost float32) {
	tw.stats.Normalize(norm, topLevelBoost)
}

func (tw *TermWeight) IsScoresDocsOutOfOrder() bool {
	return false
}

func (tw *TermWeight) Scorer(context *index.AtomicReaderContext,
	acceptDocs util.Bits) (Scorer, error) {

	assert2(tw.termStates.TopReaderContext == index.TopLevelContext(context),
		"The top-reader used to create Weight (%v) is not the same as the current reader's top-reader (%v)",
		tw.termStates.TopReaderContext, index.TopLevelContext(context))
	termsEnum, err := tw.termsEnum(context)
	if termsEnum == nil || err != nil {
		return nil, err
	}
	assert(termsEnum != nil)
	docs, err := termsEnum.Docs(acceptDocs, nil)
	if err != nil {
		return nil, err
	}
	assert(docs != nil)
	simScorer, err := tw.similarity.simScorer(tw.stats, context)
	if err != nil {
		return nil, err
	}
	return newTermScorer(tw, docs, simScorer), nil
}

func (tw *TermWeight) termsEnum(ctx *index.AtomicReaderContext) (TermsEnum, error) {
	state := tw.termStates.State(ctx.Ord)
	if state == nil { // term is not present in that reader
		assert2(tw.termNotInReader(ctx.Reader(), tw.term),
			"no termstate found but term exists in reader term=%v", tw.term)
		return nil, nil
	}
	terms := ctx.Reader().(index.AtomicReader).Terms(tw.term.Field)
	te := terms.Iterator(nil)
	err := te.SeekExactFromLast(tw.term.Bytes, state)
	return te, err
}

func (tw *TermWeight) termNotInReader(reader index.IndexReader, term *index.Term) bool {
	n, err := reader.DocFreq(term)
	assert(err == nil)
	return n == 0
}

func (tw *TermWeight) Explain(ctx *index.AtomicReaderContext, doc int) (Explanation, error) {
	scorer, err := tw.Scorer(ctx, ctx.Reader().(index.AtomicReader).LiveDocs())
	if err != nil {
		return nil, err
	}
	if scorer != nil {
		newDoc, err := scorer.Advance(doc)
		if err != nil {
			return nil, err
		}
		if newDoc == doc {
			freq, err := scorer.Freq()
			if err != nil {
				return nil, err
			}
			docScorer, err := tw.similarity.simScorer(tw.stats, ctx)
			if err != nil {
				return nil, err
			}
			scoreExplanation := docScorer.explain(doc,
				newExplanation(float32(freq), fmt.Sprintf("termFreq=%v", freq)))
			ans := newComplexExplanation(true,
				scoreExplanation.(*ExplanationImpl).value,
				fmt.Sprintf("weight(%v in %v) [%v], result of:",
					tw.TermQuery, doc, reflect.TypeOf(tw.similarity)))
			ans.details = []Explanation{scoreExplanation}
			return ans, nil
		}
	}
	return newComplexExplanation(false, 0, "no matching term"), nil
}

// search/TermScorer.java
/** Expert: A <code>Scorer</code> for documents matching a <code>Term</code>.
 */
type TermScorer struct {
	*abstractScorer
	docsEnum  DocsEnum
	docScorer SimScorer
}

func newTermScorer(w Weight, td DocsEnum, docScorer SimScorer) *TermScorer {
	ans := &TermScorer{docsEnum: td, docScorer: docScorer}
	ans.abstractScorer = newScorer(ans, w)
	return ans
}

func (ts *TermScorer) DocId() int {
	return ts.docsEnum.DocId()
}

func (ts *TermScorer) Freq() (int, error) {
	return ts.docsEnum.Freq()
}

/**
 * Advances to the next document matching the query. <br>
 *
 * @return the document matching the query or NO_MORE_DOCS if there are no more documents.
 */
func (ts *TermScorer) NextDoc() (d int, err error) {
	return ts.docsEnum.NextDoc()
}

func (ts *TermScorer) Score() (s float32, err error) {
	assert(ts.DocId() != NO_MORE_DOCS)
	freq, err := ts.docsEnum.Freq()
	if err != nil {
		return 0, err
	}
	return ts.docScorer.Score(ts.docsEnum.DocId(), float32(freq)), nil
}

/*
Advances to the first match beyond the current whose document number
is greater than or equal to a given target.
*/
func (ts *TermScorer) Advance(target int) (int, error) {
	return ts.docsEnum.Advance(target)
}

func (ts *TermScorer) String() string {
	return fmt.Sprintf("scorer(%v)", ts.weight)
}
