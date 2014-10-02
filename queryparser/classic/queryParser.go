package classic

import (
	"errors"
	// "fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/util"
	"strings"
)

type Operator int

var (
	OP_OR  = Operator(1)
	OP_AND = Operator(2)
)

type QueryParser struct {
	*QueryParserBase

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
		token_source: newTokenManager(newFastCharStream(strings.NewReader(""))),
		jj_la1:       make([]int, 21),
		jj_2_rtns:    make([]*JJCalls, 1),
	}
	qp.QueryParserBase = newQueryParserBase(qp)
	qp.ReInit(newFastCharStream(strings.NewReader("")))
	// base
	qp.analyzer = a
	qp.field = f
	qp.autoGeneratePhraseQueries = !matchVersion.OnOrAfter(util.VERSION_31)
	return qp
}

func (qp *QueryParser) conjunction() (int, error) {
	ret := CONJ_NONE
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	switch qp.jj_ntk {
	case AND, OR:
		panic("niy")
	default:
		qp.jj_la1[1] = qp.jj_gen
	}
	return ret, nil
}

func (qp *QueryParser) modifiers() (ret int, err error) {
	ret = MOD_NONE
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	switch qp.jj_ntk {
	case NOT, PLUS, MINUS:
		panic("not implemented yet")
	default:
		qp.jj_la1[3] = qp.jj_gen
	}
	return
}

func (qp *QueryParser) TopLevelQuery(field string) (q search.Query, err error) {
	if q, err = qp.Query(field); err != nil {
		return nil, err
	}
	_, err = qp.jj_consume_token(0)
	return q, err
}

func (qp *QueryParser) Query(field string) (q search.Query, err error) {
	var clauses []*search.BooleanClause
	var conj, mods int
	if mods, err = qp.modifiers(); err != nil {
		return nil, err
	}
	if q, err = qp.clause(field); err != nil {
		return nil, err
	}
	clauses = qp.addClause(clauses, CONJ_NONE, mods, q)
	var firstQuery search.Query
	if mods == MOD_NONE {
		firstQuery = q
	}
	for {
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		var found = false
		switch qp.jj_ntk {
		case AND, OR, NOT, PLUS, MINUS, BAREOPER, LPAREN, STAR, QUOTED,
			TERM, PREFIXTERM, WILDTERM, REGEXPTERM, RANGEIN_START,
			RANGEEX_START, NUMBER:
		default:
			qp.jj_la1[4] = qp.jj_gen
			found = true
		}
		if found {
			break
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
		clauses = qp.addClause(clauses, conj, mods, q)
	}
	if len(clauses) == 1 && firstQuery != nil {
		return firstQuery, nil
	} else {
		return qp.booleanQuery(clauses)
	}
}

func (qp *QueryParser) clause(field string) (q search.Query, err error) {
	if qp.jj_2_1(2) {
		panic("not implemented yet")
	}
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	var boost *Token
	switch qp.jj_ntk {
	case BAREOPER, STAR, QUOTED, TERM, PREFIXTERM, WILDTERM,
		REGEXPTERM, RANGEIN_START, RANGEEX_START, NUMBER:
		if q, err = qp.term(field); err != nil {
			return nil, err
		}
	case LPAREN:
		panic("not implemented yet")
	default:
		qp.jj_la1[7] = qp.jj_gen
		if _, err = qp.jj_consume_token(-1); err != nil {
			return nil, err
		}
		return nil, errors.New("parse error")
	}
	return qp.handleBoost(q, boost), nil
}

func (qp *QueryParser) term(field string) (q search.Query, err error) {
	var term, boost, fuzzySlop /*, goop1, goop2*/ *Token
	var prefix, wildcard, fuzzy, regexp /*, startInc, endInc*/ bool
	if qp.jj_ntk == -1 {
		qp.get_jj_ntk()
	}
	switch qp.jj_ntk {
	case BAREOPER, STAR, TERM, PREFIXTERM, WILDTERM, REGEXPTERM, NUMBER:
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case TERM:
			if term, err = qp.jj_consume_token(TERM); err != nil {
				return nil, err
			}
		case STAR:
			panic("not implemented yet")
		case PREFIXTERM:
			panic("not implemented yet")
		case WILDTERM:
			panic("not implemented yet")
		case REGEXPTERM:
			panic("not implemented yet")
		case NUMBER:
			panic("not implemented yet")
		case BAREOPER:
			panic("not implemented yet")
		default:
			panic("not implemented yet")
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case FUZZY_SLOP:
			panic("not implemented yet")
		default:
			qp.jj_la1[9] = qp.jj_gen
		}
		if qp.jj_ntk == -1 {
			qp.get_jj_ntk()
		}
		switch qp.jj_ntk {
		case CARAT:
			panic("not implemented yet")
		default:
			qp.jj_la1[11] = qp.jj_gen
		}
		if q, err = qp.handleBareTokenQuery(field, term, fuzzySlop, prefix, wildcard, fuzzy, regexp); err != nil {
			return nil, err
		}

	case RANGEIN_START, RANGEEX_START:
		panic("not implemented yet")
	case QUOTED:
		panic("not implemented yet")
	default:
		panic("not implemented yet")
	}
	return qp.handleBoost(q, boost), nil
}

// L473
func (qp *QueryParser) jj_2_1(xla int) (ok bool) {
	qp.jj_la = xla
	qp.jj_lastpos = qp.token
	qp.jj_scanpos = qp.token
	defer func() {
		// Disable following recover() to deal with dev panic
		if err := recover(); err == lookAheadSuccess {
			ok = true
		}
		qp.jj_save(0, xla)
	}()
	return !qp.jj_3_1()
}

func (qp *QueryParser) jj_3R_2() bool {
	return qp.jj_scan_token(TERM) ||
		qp.jj_scan_token(COLON)
}

func (qp *QueryParser) jj_3_1() bool {
	xsp := qp.jj_scanpos
	if qp.jj_3R_2() {
		qp.jj_scanpos = xsp
		if qp.jj_3R_3() {
			return true
		}
	}
	return false
}

func (qp *QueryParser) jj_3R_3() bool {
	return qp.jj_scan_token(STAR) ||
		qp.jj_scan_token(COLON)
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

// L569
func (qp *QueryParser) jj_consume_token(kind int) (*Token, error) {
	oldToken := qp.token
	if qp.token.next != nil {
		qp.token = qp.token.next
	} else {
		qp.token.next = qp.token_source.nextToken()
		qp.token = qp.token.next
	}
	qp.jj_ntk = -1
	if qp.token.kind == kind {
		qp.jj_gen++
		if qp.jj_gc++; qp.jj_gc > 100 {
			qp.jj_gc = 0
			panic("not implemented yet")
		}
		return qp.token, nil
	}
	qp.token = oldToken
	panic("not implemented yet")
}

type LookAheadSuccess bool

var lookAheadSuccess = LookAheadSuccess(true)

func (qp *QueryParser) jj_scan_token(kind int) bool {
	if qp.jj_scanpos == qp.jj_lastpos {
		qp.jj_la--
		if qp.jj_scanpos.next == nil {
			nextToken := qp.token_source.nextToken()
			qp.jj_scanpos.next = nextToken
			qp.jj_scanpos = nextToken
			qp.jj_lastpos = nextToken
		} else {
			qp.jj_scanpos = qp.jj_scanpos.next
			qp.jj_lastpos = qp.jj_scanpos.next
		}
	} else {
		qp.jj_scanpos = qp.jj_scanpos.next
	}
	if qp.jj_rescan {
		panic("niy")
	}
	if qp.jj_scanpos.kind != kind {
		return true
	}
	if qp.jj_la == 0 && qp.jj_scanpos == qp.jj_lastpos {
		panic(lookAheadSuccess)
	}
	return false
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

// L738

func (qp *QueryParser) jj_save(index, xla int) {
	p := qp.jj_2_rtns[index]
	for p.gen > qp.jj_gen {
		if p.next == nil {
			p = new(JJCalls)
			p.next = p
			break
		}
		p = p.next
	}
	p.gen = qp.jj_gen + xla - qp.jj_la
	p.first = qp.token
	p.arg = xla
}

type JJCalls struct {
	gen   int
	first *Token
	arg   int
	next  *JJCalls
}
