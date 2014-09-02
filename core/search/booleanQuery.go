package search

import (
	"bytes"
	"fmt"
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
