package classic

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
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

	operator Operator

	field      string
	phraseSlop int

	autoGeneratePhraseQueries bool
}

func newQueryParserBase(spi QueryParserBaseSPI) *QueryParserBase {
	return &QueryParserBase{
		QueryBuilder: newQueryBuilder(),
		spi:          spi,
		operator:     OP_OR,
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
func (qp *QueryParserBase) addClause(clauses []*search.BooleanClause,
	conj, mods int, q search.Query) []*search.BooleanClause {
	var required, prohibited bool

	// If this term is introduced by AND, make the preceding term required,
	// unless it's already prohibited
	if len(clauses) > 0 && conj == CONJ_AND {
		panic("not implemented yet")
	}

	if len(clauses) > 0 && qp.operator == OP_AND && conj == CONJ_OR {
		panic("not implemented yet")
	}

	// We might have been passed an empty query; the term might have been
	// filtered away by the analyzer.
	if q == nil {
		return clauses
	}

	if qp.operator == OP_OR {
		// We set REQUIRED if we're introduced by AND or +; PROHIBITED if
		// introduced by NOT or -; make sure not to set both.
		prohibited = (mods == MOD_NOT)
		required = (mods == MOD_REQ)
		if conj == CONJ_AND && !prohibited {
			required = true
		}
	} else {
		panic("not implemented yet")
	}
	if required {
		panic("not implemented yet")
	} else if !prohibited {
		return append(clauses, qp.newBooleanClause(q, search.SHOULD))
	} else {
		panic("not implemented yet")
	}
}

// L461
func (qp *QueryParserBase) fieldQuery(field, queryText string, quoted bool) search.Query {
	return qp.newFieldQuery(qp.analyzer, field, queryText, quoted)
}

func (qp *QueryParserBase) newFieldQuery(analyzer analysis.Analyzer,
	field, queryText string, quoted bool) search.Query {

	var occur search.Occur
	if qp.operator == OP_AND {
		occur = search.MUST
	} else {
		occur = search.SHOULD
	}
	return qp.createFieldQuery(analyzer, occur, field, queryText,
		quoted || qp.autoGeneratePhraseQueries, qp.phraseSlop)
}

// L539

func (qp *QueryParserBase) newBooleanClause(q search.Query, occur search.Occur) *search.BooleanClause {
	return search.NewBooleanClause(q, occur)
}

// L676
/*
Factory method for generating query, given a set of clauses.
By default creates a boolean query composed of clauses passed in.

Can be overridden by extending classes, to modify query being
returned.
*/
func (qp *QueryParserBase) booleanQuery(clauses []*search.BooleanClause) (search.Query, error) {
	return qp.booleanQueryDisableCoord(clauses, false)
}

func (qp *QueryParserBase) booleanQueryDisableCoord(clauses []*search.BooleanClause, disableCoord bool) (search.Query, error) {
	if len(clauses) == 0 {
		return nil, nil // all clause words were filetered away by the analyzer.
	}
	query := qp.newBooleanQuery(disableCoord)
	for _, clause := range clauses {
		query.AddClause(clause)
	}
	return query, nil
}

// L827
func (qp *QueryParserBase) handleBareTokenQuery(qField string,
	term, fuzzySlop *Token, prefix, wildcard, fuzzy, regexp bool) (q search.Query, err error) {

	var termImage string
	if termImage, err = qp.discardEscapeChar(term.image); err != nil {
		return nil, err
	}
	if wildcard {
		panic("not implemented yet")
	} else if prefix {
		panic("not implemented yet")
	} else if regexp {
		panic("not implemented yet")
	} else if fuzzy {
		panic("not implemented yet")
	} else {
		return qp.fieldQuery(qField, termImage, false), nil
	}
}

// L876
func (qp *QueryParserBase) handleBoost(q search.Query, boost *Token) search.Query {
	if boost != nil {
		panic("not implemented yet")
	}
	return q
}

// L906
func (qp *QueryParserBase) discardEscapeChar(input string) (string, error) {
	output := make([]rune, len(input))

	length := 0

	lastCharWasEscapeChar := false

	codePointMultiplier := 0

	// codePoint := 0

	for _, curChar := range input {
		if codePointMultiplier > 0 {
			panic("not implemented yet")
		} else if lastCharWasEscapeChar {
			panic("not implemented yet")
		} else {
			if curChar == '\\' {
				lastCharWasEscapeChar = true
			} else {
				output[length] = curChar
				length++
			}
		}
	}

	if codePointMultiplier > 0 {
		return "", errors.New("Truncated unicode escape sequence.")
	}
	if lastCharWasEscapeChar {
		return "", errors.New("Term can not end with escape character.")
	}
	return string(output), nil
}
