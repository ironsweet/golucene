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
	QuerySPI
	CreateWeight(ss *IndexSearcher) (w Weight, err error)
	Rewrite(r index.IndexReader) Query
}

type QuerySPI interface {
	ToString(string) string
}

type AbstractQuery struct {
	spi   QuerySPI
	value Query
	boost float32
}

func NewAbstractQuery(self interface{}) *AbstractQuery {
	return &AbstractQuery{
		spi:   self.(QuerySPI),
		value: self.(Query),
		boost: 1.0,
	}
}

func (q *AbstractQuery) SetBoost(b float32) { q.boost = b }
func (q *AbstractQuery) Boost() float32     { return q.boost }

func (q *AbstractQuery) String() string { return q.spi.ToString("") }

func (q *AbstractQuery) CreateWeight(ss *IndexSearcher) (w Weight, err error) {
	panic(fmt.Sprintf("Query %v does not implement createWeight", q))
}

func (q *AbstractQuery) Rewrite(r index.IndexReader) Query {
	return q.value
}
