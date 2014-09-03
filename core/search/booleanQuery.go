package search

import (
	"bytes"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/util"
)

const maxClauseCount = 1024

type BooleanQuery struct {
	*AbstractQuery
	clauses          []*BooleanClause
	disableCoord     bool
	minNrShouldMatch int
}

func NewBooleanQuery() *BooleanQuery {
	return NewBooleanQueryDisableCoord(false)
}

func NewBooleanQueryDisableCoord(disableCoord bool) *BooleanQuery {
	ans := &BooleanQuery{
		disableCoord: disableCoord,
	}
	ans.AbstractQuery = NewAbstractQuery(ans)
	return ans
}

func (q *BooleanQuery) Add(query Query, occur Occur) {
	q.AddClause(NewBooleanClause(query, occur))
}

func (q *BooleanQuery) AddClause(clause *BooleanClause) {
	assert(len(q.clauses) < maxClauseCount)
	q.clauses = append(q.clauses, clause)
}

type BooleanWeight struct {
	owner        *BooleanQuery
	similarity   Similarity
	weights      []Weight
	maxCoord     int // num optional +num required
	disableCoord bool
}

func newBooleanWeight(owner *BooleanQuery,
	searcher *IndexSearcher, disableCoord bool) (w *BooleanWeight, err error) {

	w = &BooleanWeight{
		owner:        owner,
		similarity:   searcher.similarity,
		disableCoord: disableCoord,
	}
	var subWeight Weight
	for _, c := range owner.clauses {
		if subWeight, err = c.query.CreateWeight(searcher); err != nil {
			return nil, err
		}
		w.weights = append(w.weights, subWeight)
		if !c.IsProhibited() {
			w.maxCoord++
		}
	}
	return w, nil
}

func (w *BooleanWeight) ValueForNormalization() float32 {
	panic("not implemented yet")
}

func (w *BooleanWeight) Normalize(norm, topLevelBoost float32) {
	panic("not implemented yet")
}

func (w *BooleanWeight) Explain(context *index.AtomicReaderContext, doc int) (Explanation, error) {
	panic("not implemented yet")
}

func (w *BooleanWeight) BulkScorer(context *index.AtomicReaderContext,
	scoreDocsInOrder bool, acceptDocs util.Bits) (BulkScorer, error) {
	panic("not implemented yet")
}

func (w *BooleanWeight) IsScoresDocsOutOfOrder() bool {
	panic("not implemented yet")
}

func (q *BooleanQuery) CreateWeight(searcher *IndexSearcher) (Weight, error) {
	return newBooleanWeight(q, searcher, q.disableCoord)
}

func (q *BooleanQuery) Rewrite(reader index.IndexReader) Query {
	if q.minNrShouldMatch == 0 && len(q.clauses) == 1 {
		panic("not implemented yet")
	}

	var clone *BooleanQuery // recursively rewrite
	for _, c := range q.clauses {
		if query := c.query.Rewrite(reader); query != c.query {
			// clause rewrote: must clone
			if clone == nil {
				// The BooleanQuery clone is lazily initialized so only
				// initialize it if a rewritten clause differs from the
				// original clause (and hasn't been initialized already). If
				// nothing difers, the clone isn't needlessly created
				panic("not implemented yet")
			}
			panic("not implemented yet")
		}
	}
	if clone != nil {
		return clone // some clauses rewrote
	}
	return q
}

func (q *BooleanQuery) ToString(field string) string {
	var buf bytes.Buffer
	needParens := q.Boost() != 1 || q.minNrShouldMatch > 0
	if needParens {
		buf.WriteRune('(')
	}

	for i, c := range q.clauses {
		if c.IsProhibited() {
			buf.WriteRune('-')
		} else if c.IsRequired() {
			buf.WriteRune('+')
		}

		if subQuery := c.query; subQuery != nil {
			if _, ok := subQuery.(*BooleanQuery); ok { // wrap sub-bools in parens
				buf.WriteRune('(')
				buf.WriteString(subQuery.ToString(field))
				buf.WriteRune(')')
			} else {
				buf.WriteString(subQuery.ToString(field))
			}
		} else {
			buf.WriteString("nil")
		}

		if i != len(q.clauses)-1 {
			buf.WriteRune(' ')
		}
	}

	if needParens {
		buf.WriteRune(')')
	}

	if q.minNrShouldMatch > 0 {
		panic("not implemented yet")
	}

	if q.Boost() != 1 {
		panic("not implemented yet")
	}

	return buf.String()
}
