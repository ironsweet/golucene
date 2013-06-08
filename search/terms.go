package search

import (
	"lucene/index"
	"lucene/util"
)

type TermQuery struct {
	*AbstractQuery
	term               index.Term
	docFreq            int
	perReaderTermState index.TermContext
}

func (q *TermQuery) CreateWeight(ss IndexSearcher) Weight {
	ctx := ss.TopReaderContext()
	var termState index.TermContext
	if q.perReaderTermState.TopReaderContext != ctx {
		// make TermQuery single-pass if we don't have a PRTS or if the context differs!
		termState = index.NewTermContextFromTerm(ctx, q.term, true)
	} else {
		// PRTS was pre-build for this IS
		termState = q.perReaderTermState
	}

	// we must not ignore the given docFreq - if set use the given value (lie)
	if q.docFreq != -1 {
		termState.DocFreq = q.docFreq
	}

	return NewTermWeight(q, ss, termState)
}

type TermWeight struct {
	query      TermQuery
	similarity Similarity
	stats      SimWeight
	termStates index.TermContext
}

func NewTermWeight(q TermQuery, ss IndexSearcher, termStates index.TermContext) TermWeight {
	sim := ss.Similarity
	return TermWeight{q, sim, sim.computeWeight(
		q.AbstractQuery.boost,
		ss.CollectionStatistics(q.term.Field),
		ss.TermStatistics(q.term, termStates)), termStates}
}

func (tw TermWeight) ValueForNormalization() float32 {
	return tw.stats.ValueForNormalization()
}

func (tw TermWeight) Normalize(norm, topLevelBoost float32) float32 {
	return tw.stats.Normalize(norm, topLevelBoost)
}

func (tw TermWeight) IsScoresDocsOutOfOrder() bool {
	return false
}

func (tw TermWeight) Scorer(ctx index.AtomicReaderContext, inOrder bool, topScorer bool, acceptDocs util.Bits) (sc Scorer, ok bool) {
	// assert termStates.topReaderContext == ReaderUtil.getTopLevelContext(context) : "The top-reader used to create Weight (" + termStates.topReaderContext + ") is not the same as the current reader's top-reader (" + ReaderUtil.getTopLevelContext(context);
	termsEnum, ok := termsEnum(context)
	if !ok {
		return Scorer{}, false
	}
	docs := termsEnum.docs(acceptDocs, null)
	// assert docs != null;
	return NewTermScorer(tw, docs, tw.similarity.exactSimScorer(stats, context))
}

func (tw TermWeight) termsEnum(ctx index.AtomicReaderContext) (te index.TermsEnum, ok bool) {
	state := tw.termStates.State(ctx.Ord)
	if state == nil { // term is not present in that reader
		// assert termNotInReader(ctx.Reader(), tw.Term)
		// : "no termstate found but term exists in reader term=" + term;
		return TermsEnum{}, false
	}
	te = ctx.Reader().Terms(tw.query.term.Field).Iterator(nil)
	te.SeekExact(tw.query.term.Bytes, state)
	return te, true
}

type TermScorer struct {
	*Scorer
	docScroer ExactSimScorer
	docsEnum  DocsEnum
}

func NewTermScorer(w Weight, td DocsEnum, docScorer ExactSimScorer) *TermScorer {
	return TermScorer{Scorer{w}, td, docScorer}
}
