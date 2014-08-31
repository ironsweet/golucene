package classic

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/util"
	"strings"
)

type QueryParser struct {
	*QueryParserBase

	analyzer analysis.Analyzer

	token_source           *TokenManager
	token                  *Token // current token
	jj_nt                  *Token // next token
	jj_ntk                 int
	jj_scanpos, jj_lastpos *Token
	jj_la                  int
	jj_gen                 int
	jj_la1                 []int

	jj_2_rtns []*JJCalls
	jj_rescan bool
	jj_gc     int
}

func NewQueryParser(matchVersion util.Version, f string, a analysis.Analyzer) *QueryParser {
	qp := &QueryParser{
		analyzer:     a,
		token_source: newTokenManager(newFastCharStream(strings.NewReader(""))),
		jj_la1:       make([]int, 21),
		jj_2_rtns:    make([]*JJCalls, 1),
	}
	qp.QueryParserBase = newQueryParserBase(qp)
	qp.ReInit(newFastCharStream(strings.NewReader("")))
	// base
	qp.field = f
	qp.autoGeneratePhraseQueries = !matchVersion.OnOrAfter(util.VERSION_31)
	return qp
}

func (qp *QueryParser) TopLevelQuery(field string) (search.Query, error) {
	panic("not implemented yet")
}

func (qp *QueryParser) ReInit(stream CharStream) {
	qp.token_source.ReInit(stream)
	qp.token = new(Token)
	qp.jj_ntk = -1
	qp.jj_gen = 0
	for i, _ := range qp.jj_la1 {
		qp.jj_la1[i] = -1
	}
	for i, _ := range qp.jj_2_rtns {
		qp.jj_2_rtns[i] = new(JJCalls)
	}

}

type JJCalls struct {
	gen   int
	first *Token
	arg   int
	next  *JJCalls
}
