package search

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
)

type Query interface {
	CreateWeight(ss IndexSearcher) (w Weight, err error)
	Rewrite(r index.IndexReader) Query
}

type AbstractQuery struct {
	Query
	boost float32
}

func NewAbstractQuery(self Query) *AbstractQuery {
	return &AbstractQuery{self, 1.0}
}

func (q *AbstractQuery) CreateWeight(ss IndexSearcher) (w Weight, err error) {
	panic(fmt.Sprintf("Query %v does not implement createWeight", q))
}

func (q *AbstractQuery) Rewrite(r index.IndexReader) Query {
	return q.Query
}
