package search

type BooleanQuery struct {
	*AbstractQuery
}

func NewBooleanQuery() *BooleanQuery {
	ans := new(BooleanQuery)
	ans.AbstractQuery = NewAbstractQuery(ans)
	return ans
}
