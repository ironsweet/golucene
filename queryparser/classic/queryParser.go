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

func (qp *QueryParser) conjunction() (int, error) {
	panic("not implemented yet")
}

func (qp *QueryParser) modifiers() (int, error) {
	panic("not implemented yet")
}

func (qp *QueryParser) TopLevelQuery(field string) (q search.Query, err error) {
	if q, err = qp.Query(field); err != nil {
		return nil, err
	}
	_, err = qp.jj_consume_token(0)
	return q, err
}

func (qp *QueryParser) Query(field string) (q search.Query, err error) {
	var clauses []search.BooleanClause
	var conj, mods int
	if mods, err = qp.modifiers(); err != nil {
		return nil, err
	}
	if q, err = qp.clause(field); err != nil {
		return nil, err
	}
	qp.addClause(clauses, CONJ_NONE, mods, q)
	var firstQuery search.Query
	if mods == MOD_NONE {
		firstQuery = q
	}
	var found = false
	for !found {
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case AND, OR, NOT, PLUS, MINUS, BAREOPER, LPAREN, STAR, QUOTED,
			TERM, PREFIXTERM, WILDTERM, REGEXPTERM, RANGEIN_START,
			RANGEEX_START, NUMBER:
		default:
			found = true
			qp.jj_la1[4] = qp.jj_gen
		}
		if conj, err = qp.conjunction(); err != nil {
			return nil, err
		}
		if mods, err = qp.modifiers(); err != nil {
			return nil, err
		}
		if q, err = qp.clause(field); err != nil {
			return nil, err
		}
		qp.addClause(clauses, conj, mods, q)
	}
	if len(clauses) == 1 && firstQuery != nil {
		return firstQuery, nil
	} else {
		panic("not implemented yet")
		// return qp.booleanQuery(clauses)
	}
}

func (qp *QueryParser) clause(field string) (search.Query, error) {
	panic("not implemented yet")
}

// L540
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

func (qp *QueryParser) jj_consume_token(kind int) (*Token, error) {
	panic("not implemented yet")
}

// L636
func (qp *QueryParser) get_jj_ntk() int {
	if qp.jj_nt = qp.token.next; qp.jj_nt == nil {
		qp.token.next = qp.token_source.nextToken()
		qp.jj_ntk = qp.token.next.kind
	} else {
		qp.jj_ntk = qp.jj_nt.kind
	}
	return qp.jj_ntk
}

type JJCalls struct {
	gen   int
	first *Token
	arg   int
	next  *JJCalls
}
