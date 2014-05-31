package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

/* The term text of a Token. */
type CharTermAttribute interface {
	// Returns the internal termBuffer rune slice which you can then
	// directly alter. If the slice is too small for your token, use
	// ResizeBuffer(int) to increase it. After altering the buffer, be
	// sure to call SetLength() to record the number of valid runes
	// that were placed into the termBuffer.
	//
	// NOTE: the returned buffer may be larger than the valid Length().
	Buffer() []rune
	Length() int
}

const MIN_BUFFER_SIZE = 10

/* Default implementation of CharTermAttribute. */
type CharTermAttributeImpl struct {
	termBuffer []rune
	termLength int
}

func newCharTermAttributeImpl() *util.AttributeImpl {
	ans := &CharTermAttributeImpl{
		termBuffer: make([]rune, util.Oversize(MIN_BUFFER_SIZE, util.NUM_BYTES_CHAR)),
	}
	return util.NewAttributeImpl(ans)
}

func (a *CharTermAttributeImpl) Interfaces() []string {
	return []string{"CharTermAttribute"}
}

func (a *CharTermAttributeImpl) Buffer() []rune {
	return a.termBuffer
}

func (a *CharTermAttributeImpl) Length() int {
	return a.termLength
}

func (a *CharTermAttributeImpl) Clear() {
	a.termLength = 0
}

func (a *CharTermAttributeImpl) String() string {
	panic("not implemented yet")
}
