package search

const maxClauseCount = 1024

type BooleanQuery struct {
	*AbstractQuery
	clauses      []*BooleanClause
	disableCoord bool
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
