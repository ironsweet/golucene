package search

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
)

// search/Query.java

/*
The abstract base class for queries.
*/
type Query interface {
	SetBoost(b float32)
	Boost() float32
	CreateWeight(ss *IndexSearcher) (w Weight, err error)
	Rewrite(r index.IndexReader) Query
}

type AbstractQuery struct {
	Query
	boost float32
}

func NewAbstractQuery(self Query) *AbstractQuery {
	return &AbstractQuery{self, 1.0}
}

func (q *AbstractQuery) SetBoost(b float32) { q.boost = b }
func (q *AbstractQuery) Boost() float32     { return q.boost }

func (q *AbstractQuery) CreateWeight(ss *IndexSearcher) (w Weight, err error) {
	panic(fmt.Sprintf("Query %v does not implement createWeight", q))
}

func (q *AbstractQuery) Rewrite(r index.IndexReader) Query {
	return q.Query
}
