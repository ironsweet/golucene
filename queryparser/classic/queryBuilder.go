package classic

import (
	"github.com/balzaczyy/golucene/core/analysis"
	ta "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/index"
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
	var positionCount int
	var severalTokensAtSamePosition bool
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
			hasMoreTokens, err := buffer.IncrementToken()
			for hasMoreTokens && err == nil {
				numTokens++
				positionIncrement := 1
				if posIncrAtt != nil {
					positionIncrement = posIncrAtt.PositionIncrement()
				}
				if positionIncrement != 0 {
					positionCount += positionIncrement
				} else {
					severalTokensAtSamePosition = true
				}
				hasMoreTokens, err = buffer.IncrementToken()
			} // ignore error
		}
		return nil
	}(); err != nil {
		panic(err)
	}

	// rewind the buffer stream
	buffer.Reset()

	var bytes *util.BytesRef
	if termAtt != nil {
		bytes = termAtt.BytesRef()
	}

	if numTokens == 0 {
		return nil
	} else if numTokens == 1 {
		if hasNext, err := buffer.IncrementToken(); err == nil {
			assert(hasNext)
			termAtt.FillBytesRef()
		} // safe to ignore error, because we know the number of tokens
		return qp.newTermQuery(index.NewTermFromBytes(field, util.DeepCopyOf(bytes).ToBytes()))
	} else {
		if severalTokensAtSamePosition || !quoted {
			if positionCount == 1 || !quoted {
				// no phrase query:

				if positionCount == 1 {
					panic("not implemented yet")
				} else {
					// multiple positions
					q := qp.newBooleanQuery(false)
					var currentQuery search.Query
					for i := 0; i < numTokens; i++ {
						hasNext, err := buffer.IncrementToken()
						if err != nil {
							continue // safe to ignore error, because we know the number of tokens
						}
						assert(hasNext)
						termAtt.FillBytesRef()

						if posIncrAtt != nil && posIncrAtt.PositionIncrement() == 0 {
							panic("not implemented yet")
						} else {
							if currentQuery != nil {
								q.Add(currentQuery, operator)
							}
							currentQuery = qp.newTermQuery(index.NewTermFromBytes(field, util.DeepCopyOf(bytes).ToBytes()))
						}
					}
					q.Add(currentQuery, operator)
					return q
				}
			} else {
				panic("not implemented yet")
			}
		} else {
			panic("not implemented yet")
		}
		panic("should not be here")
	}
}

// L379
func (qp *QueryBuilder) newBooleanQuery(disableCoord bool) *search.BooleanQuery {
	return search.NewBooleanQueryDisableCoord(disableCoord)
}

func (qp *QueryBuilder) newTermQuery(term *index.Term) search.Query {
	return search.NewTermQuery(term)
}
