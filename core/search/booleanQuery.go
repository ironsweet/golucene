package search

type BooleanQuery struct {
	*AbstractQuery
}

func NewBooleanQuery() *BooleanQuery {
	ans := new(BooleanQuery)
	ans.AbstractQuery = NewAbstractQuery(ans)
	return ans
}

func NewBooleanQueryDisableCoord(disableCoord bool) *BooleanQuery {
	panic("not implemented yet")
}

func (q *BooleanQuery) Add(query Query, occur Occur) {
	panic("not imlemented yet")
}
