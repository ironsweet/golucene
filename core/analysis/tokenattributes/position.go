package tokenattributes

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

/*
Determines the position of this token relative to the previous Token
in a TokenStream, used in phrase searching.

The default value is one.

Some common uses for this are:

	- Set it to zero to put multiple terms in the same position. This
	is useful if, e.g., a word has multiple stems. Searches for phrases
	including either stem will match. In this case, all but the first
	stem's increment should be set to zero: the increment of the first
	instance should be one. Repeating a token with an increment of zero
	can also be used to boost the scores of matches on that token.

	-	Set it to values greater than one to inhibit exact phrase matches.
	If, for example, one does not want phrases to match across removed
	stop words, then one could build a stop word filter that removes
	stop words and also sets the incremeent to the number of stop words
	removed before each non-stop word. Then axact phrase queries will
	only match when the terms occur with no intervening stop words.
*/
type PositionIncrementAttribute interface {
	util.Attribute
	// Set the position increment. The deafult value is one.
	SetPositionIncrement(int)
	// Returns the position increment of this token.
	PositionIncrement() int
}

/* Default implementation of ositionIncrementAttribute */
type PositionIncrementAttributeImpl struct {
	positionIncrement int
}

func newPositionIncrementAttributeImpl() util.AttributeImpl {
	return &PositionIncrementAttributeImpl{
		positionIncrement: 1,
	}
}

func (a *PositionIncrementAttributeImpl) Interfaces() []string {
	return []string{"PositionIncrementAttribute"}
}

func (a *PositionIncrementAttributeImpl) SetPositionIncrement(positionIncrement int) {
	assert2(positionIncrement >= 0, "Increment must be zero or greater: got %v", positionIncrement)
	a.positionIncrement = positionIncrement
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (a *PositionIncrementAttributeImpl) PositionIncrement() int {
	return a.positionIncrement
}

func (a *PositionIncrementAttributeImpl) Clear() {
	a.positionIncrement = 1
}

func (a *PositionIncrementAttributeImpl) Clone() util.AttributeImpl {
	return &PositionIncrementAttributeImpl{
		positionIncrement: a.positionIncrement,
	}
}

func (a *PositionIncrementAttributeImpl) CopyTo(target util.AttributeImpl) {
	target.(PositionIncrementAttribute).SetPositionIncrement(a.positionIncrement)
}
