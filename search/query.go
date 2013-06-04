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
	boost float
}

func NewAbstractQuery() {
	return QueryWithBoost{1.0}
}

func (q *AbstractQuery) CreateWeight(ss IndexSearcher) Weight {
	panic(fmt.Sprintf("Query %v does not implement createWeight", q))
}

func (q *AbstractQuery) Rewrite(r index.Reader) Query {
	return q
}

type TermQuery struct {
	*AbstractQuery
	term               Term
	perReaderTermState *TermContext
}

func (q *TermQuery) CreateWeight(ss IndexSearcher) Weight {
	ctx := ss.TopReaderContext()
	var termState index.TermContext
	if perReaderTermState == nil || perReaderTermState.TopReaderContext != ctx {
		// make TermQuery single-pass if we don't have a PRTS or if the context differs!
		termState = NewTermContextFromTerm(ctx, term, true)
	} else {
		// PRTS was pre-build for this IS
		termState = q.perReaderTermState
	}

	// we must not ignore the given docFreq - if set use the given value (lie)
	if docFreq != -1 {
		termState.docFreq = docFreq
	}

	return NewTermWeight(searcher, termState)
}

type TermWeight struct {
	similarity Similarty
	stats      SimWeight
	termStates TermContext
}

func NewTermWeight(q TermQuery, ss IndexSearcher, termStates TermContext) {
	sim := ss.Similarity()
	return TermWeight{sim, sim.computeWeight(
		q.AbstractQuery.boost,
		ss.CollectionStatustics(q.term.Field),
		ss.TermStatistics(q.term, termStates)), termStates}
}

func (tw TermWeight) ValueForNormalization() float {
	return tw.stats.ValueForNormalization()
}

func (tw TermWeight) Normalize(norm, topLevelBoost float) float {
	return tw.stats.Normalize(norm, topLevelBoost)
}
