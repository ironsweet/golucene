package classic

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/search"
	"strings"
)

type QueryParserBaseSPI interface {
	ReInit(CharStream)
	TopLevelQuery(string) (search.Query, error)
}

type QueryParserBase struct {
	*QueryBuilder

	spi QueryParserBaseSPI

	field string

	autoGeneratePhraseQueries bool
}

func newQueryParserBase(spi QueryParserBaseSPI) *QueryParserBase {
	return &QueryParserBase{
		QueryBuilder: newQueryBuilder(),
		spi:          spi,
	}
}

func (qp *QueryParserBase) Parse(query string) (res search.Query, err error) {
	qp.spi.ReInit(newFastCharStream(strings.NewReader(query)))
	if res, err = qp.spi.TopLevelQuery(qp.field); err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot parse '%v': %v", query, err))
	}
	if res != nil {
		return res, nil
	}
	return qp.newBooleanQuery(false), nil
}
