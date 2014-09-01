package classic

import (
	"github.com/balzaczyy/golucene/core/analysis"
	ta "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/util"
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

	assert(operator == search.SHOULD || operator == search.MUST)
	assert(analyzer != nil)
	// use the analyzer to get all the tokens, and then build a TermQuery,
	// PraseQuery, or nothing based on the term count
	var buffer *analysis.CachingTokenFilter
	var termAtt ta.TermToBytesRefAttribute
	var posIncrAtt ta.PositionIncrementAttribute
	var numTokens int
	if err := func() (err error) {
		var source analysis.TokenStream
		defer func() {
			util.CloseWhileSuppressingError(source)
		}()

		if source, err = analyzer.TokenStreamForString(field, queryText); err != nil {
			return
		}
		if err = source.Reset(); err != nil {
			return
		}
		buffer = analysis.NewCachingTokenFilter(source)
		buffer.Reset()

		termAtt = buffer.Attributes().Get("TermToBytesRefAttribute").(ta.TermToBytesRefAttribute)
		posIncrAtt = buffer.Attributes().Get("PositionIncrementAttribute").(ta.PositionIncrementAttribute)

		if termAtt != nil {
			panic("not implemented yet")
		}
		return nil
	}(); err != nil {
		panic(err)
	}

	// rewind the buffer stream
	buffer.Reset()

	// var bytes *util.BytesRef
	// if termAtt != nil {
	// 	bytes = termAtt.BytesRef()
	// }

	if numTokens == 0 {
		return nil
	} else if numTokens == 1 {
		panic("not implemented yet")
	} else {
		panic("not implemented yet")
	}
}

// L379
func (qp *QueryBuilder) newBooleanQuery(disableCoord bool) *search.BooleanQuery {
	panic("not implemented yet")
}
