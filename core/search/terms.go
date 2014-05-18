package search

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/util"
)

type TermQuery struct {
	*AbstractQuery
	term               index.Term
	docFreq            int
	perReaderTermState *index.TermContext
}

func NewTermQuery(t index.Term) *TermQuery {
	return NewTermQueryWithDocFreq(t, -1)
}

func NewTermQueryWithDocFreq(t index.Term, docFreq int) *TermQuery {
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

	return NewTermWeight(q, ss, *termState), nil
}

func (q *TermQuery) String() string {
	boost := ""
	if q.boost != 1.0 {
		boost = fmt.Sprintf("^%v", q.boost)
	}
	return fmt.Sprintf("%v:%v%v", q.term.Field, string(q.term.Bytes), boost)
}

type TermWeight struct {
	*TermQuery
	similarity Similarity
	stats      SimWeight
	termStates index.TermContext
}

func NewTermWeight(owner *TermQuery, ss *IndexSearcher, termStates index.TermContext) TermWeight {
	// assert(termStates != nil)
	sim := ss.similarity
	return TermWeight{owner, sim, sim.computeWeight(
		owner.boost,
		ss.CollectionStatistics(owner.term.Field),
		ss.TermStatistics(owner.term, termStates)), termStates}
}

func (tw TermWeight) String() string {
	return fmt.Sprintf("weight(%v)", tw.TermQuery)
}

func (tw TermWeight) ValueForNormalization() float32 {
	return tw.stats.ValueForNormalization()
}

func (tw TermWeight) Normalize(norm float32, topLevelBoost float32) {
	tw.stats.Normalize(norm, topLevelBoost)
}

func (tw TermWeight) IsScoresDocsOutOfOrder() bool {
	return false
}

func (tw TermWeight) Scorer(context index.AtomicReaderContext,
	inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, err error) {
	// assert termStates.topReaderContext == ReaderUtil.getTopLevelContext(context) : "The top-reader used to create Weight (" + termStates.topReaderContext + ") is not the same as the current reader's top-reader (" + ReaderUtil.getTopLevelContext(context);
	termsEnum, err := tw.termsEnum(context)
	if err != nil {
		return nil, err
	}
	docs, err := termsEnum.Docs(acceptDocs, index.DOCS_ENUM_EMPTY)
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

func (tw TermWeight) termsEnum(ctx index.AtomicReaderContext) (te index.TermsEnum, err error) {
	state := tw.termStates.State(ctx.Ord)
	if state == nil { // term is not present in that reader
		// assert termNotInReader(ctx.Reader(), tw.Term)
		// : "no termstate found but term exists in reader term=" + term;
		return index.EMPTY_TERMS_ENUM, nil
	}
	te = ctx.Reader().(index.AtomicReader).Terms(tw.term.Field).Iterator(index.EMPTY_TERMS_ENUM)
	err = te.SeekExactFromLast(tw.term.Bytes, *state)
	return
}

// search/TermScorer.java
/** Expert: A <code>Scorer</code> for documents matching a <code>Term</code>.
 */
type TermScorer struct {
	*abstractScorer
	docsEnum  index.DocsEnum
	docScorer SimScorer
}

func newTermScorer(w Weight, td index.DocsEnum, docScorer SimScorer) *TermScorer {
	ans := &TermScorer{docsEnum: td, docScorer: docScorer}
	ans.abstractScorer = newScorer(ans, w)
	return ans
}

func (ts *TermScorer) DocId() int {
	return ts.docsEnum.DocId()
}

/**
 * Advances to the next document matching the query. <br>
 *
 * @return the document matching the query or NO_MORE_DOCS if there are no more documents.
 */
func (ts *TermScorer) NextDoc() (d int, err error) {
	return ts.docsEnum.NextDoc()
}

func (ts *TermScorer) Score() (s float64, err error) {
	assert(ts.DocId() != index.NO_MORE_DOCS)
	freq, err := ts.docsEnum.Freq()
	if err != nil {
		return 0, err
	}
	return float64(ts.docScorer.Score(ts.docsEnum.DocId(), float32(freq))), nil
}

func (ts *TermScorer) String() string {
	return fmt.Sprintf("scorer(%v)", ts.weight)
}
