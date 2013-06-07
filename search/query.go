package search

import (
	"fmt"
	"lucene/index"
)

type Query interface {
	CreateWeight(ss IndexSearcher) Weight
	Rewrite(r index.Reader) Query
}

type AbstractQuery struct {
	boost float32
}

func NewAbstractQuery() Query {
	return &AbstractQuery{1.0}
}

func (q *AbstractQuery) CreateWeight(ss IndexSearcher) Weight {
	panic(fmt.Sprintf("Query %v does not implement createWeight", q))
}

func (q *AbstractQuery) Rewrite(r index.Reader) Query {
	return q
}

type TermQuery struct {
	*AbstractQuery
	term               index.Term
	docFreq            int
	perReaderTermState index.TermContext
}

func (q *TermQuery) CreateWeight(ss IndexSearcher) Weight {
	ctx := ss.TopReaderContext()
	var termState index.TermContext
	// FIXME !=
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
	similarity Similarity
	stats      SimWeight
	termStates index.TermContext
}

func NewTermWeight(q *TermQuery, ss IndexSearcher, termStates index.TermContext) TermWeight {
	sim := ss.Similarity
	return TermWeight{sim, sim.computeWeight(
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
