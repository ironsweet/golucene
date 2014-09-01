package classic

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/search"
	"strings"
)

const (
	CONJ_NONE = iota
	CONJ_AND
	CONJ_OR
)

const (
	MOD_NONE = 0
	MOD_NOT  = 10
	MOD_REQ  = 11
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

// L116
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

// L408
func (qp *QueryParserBase) addClause(clauses []search.BooleanClause, conj, mods int, q search.Query) {
	panic("not implemented yet")
}

// L827
func (qp *QueryParserBase) handleBareTokenQuery(qField string,
	term, fuzzySlop *Token, prefix, wildcard, fuzzy, regexp bool) (q search.Query, err error) {

	panic("not implemented yet")
}

// L876
func (qp *QueryParserBase) handleBoost(q search.Query, boost *Token) search.Query {
	panic("not implemented yet")
}
