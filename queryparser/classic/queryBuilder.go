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

func (qp *QueryBuilder) newBooleanQuery(disableCoord bool) *search.BooleanQuery {
	panic("not implemented yet")
}
