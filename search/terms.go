package search

import (
	"fmt"
	"github.com/balzaczyy/golucene/index"
	"github.com/balzaczyy/golucene/util"
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

func (q *TermQuery) CreateWeight(ss IndexSearcher) Weight {
	ctx := ss.TopReaderContext()
	var termState *index.TermContext
	if q.perReaderTermState == nil || q.perReaderTermState.TopReaderContext != ctx {
		// make TermQuery single-pass if we don't have a PRTS or if the context differs!
		termState = index.NewTermContextFromTerm(ctx, q.term)
	} else {
		// PRTS was pre-build for this IS
		termState = q.perReaderTermState
	}

	// we must not ignore the given docFreq - if set use the given value (lie)
	if q.docFreq != -1 {
		termState.DocFreq = q.docFreq
	}

	return NewTermWeight(q, ss, *termState)
}

func (q *TermQuery) String() string {
	boost := ""
	if q.boost != 1.0 {
		boost = fmt.Sprintf("^%v", q.boost)
	}
	return fmt.Sprintf("%v:%v%v", q.term.Field, string(q.term.Bytes), boost)
}

type TermWeight struct {
	query      *TermQuery
	similarity Similarity
	stats      SimWeight
	termStates index.TermContext
}

func NewTermWeight(q *TermQuery, ss IndexSearcher, termStates index.TermContext) TermWeight {
	sim := ss.similarity
	return TermWeight{q, sim, sim.computeWeight(
		q.AbstractQuery.boost,
		ss.CollectionStatistics(q.term.Field),
		ss.TermStatistics(q.term, termStates)), termStates}
}

func (tw TermWeight) ValueForNormalization() float32 {
	return tw.stats.ValueForNormalization()
}

func (tw TermWeight) Normalize(norm float64, topLevelBoost float32) float32 {
	return tw.stats.Normalize(norm, topLevelBoost)
}

func (tw TermWeight) IsScoresDocsOutOfOrder() bool {
	return false
}

func (tw TermWeight) Scorer(context index.AtomicReaderContext,
	inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, ok bool) {
	// assert termStates.topReaderContext == ReaderUtil.getTopLevelContext(context) : "The top-reader used to create Weight (" + termStates.topReaderContext + ") is not the same as the current reader's top-reader (" + ReaderUtil.getTopLevelContext(context);
	termsEnum, ok := tw.termsEnum(context)
	if !ok {
		return Scorer{}, false
	}
	docs := termsEnum.Docs(acceptDocs, index.DOCS_ENUM_EMPTY)
	// assert docs != null;
	return *(newTermScorer(tw, docs, tw.similarity.exactSimScorer(tw.stats, context)).Scorer), true
}

func (tw TermWeight) termsEnum(ctx index.AtomicReaderContext) (te index.TermsEnum, ok bool) {
	state := tw.termStates.State(ctx.Ord)
	if state == nil { // term is not present in that reader
		// assert termNotInReader(ctx.Reader(), tw.Term)
		// : "no termstate found but term exists in reader term=" + term;
		return index.EMPTY_TERMS_ENUM, false
	}
	te = ctx.Reader().(index.AtomicReader).Terms(tw.query.term.Field).Iterator(index.EMPTY_TERMS_ENUM)
	te.SeekExactFromLast(tw.query.term.Bytes, *state)
	return te, true
}

type TermScorer struct {
	*Scorer
	docScorer ExactSimScorer
	docsEnum  index.DocsEnum
}

func newTermScorer(w Weight, td index.DocsEnum, docScorer ExactSimScorer) TermScorer {
	ans := &TermScorer{}
	scorer := newScorer(ans, w, func() float64 {
		// assert docID() != NO_MORE_DOCS
		return ans.docScorer.Score(ans.docsEnum.DocId(), ans.docsEnum.Freq())
	})
	ans.Scorer = &scorer
	ans.docScorer = docScorer
	ans.docsEnum = td
	return *ans
}
