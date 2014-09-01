package classic

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/search"
)

type QueryBuilder struct {
	analyzer                 analysis.Analyzer
	enablePositionIncrements bool
}

func newQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		enablePositionIncrements: true,
	}
}

// L193
func (qp *QueryBuilder) createFieldQuery(analyzer analysis.Analyzer,
	operator search.Occur, field, queryText string, quoted bool, phraseSlop int) search.Query {

	panic("not implemented yet")
}

// L379
func (qp *QueryBuilder) newBooleanQuery(disableCoord bool) *search.BooleanQuery {
	panic("not implemented yet")
}
