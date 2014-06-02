package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

/* The start and end character offset of a Token. */
type OffsetAttribute interface {
	util.Attribute
	// Returns this Token's starting offset, the position of the first
	// character corresponding to this token in the source text.
	//
	// Note that the difference between endOffset() and startOffset()
	// may not be equal to the termText.Length(), as the term text may
	// have been altered by a stemmer or some other filter.
	// StartOffset() int
	// Set the starting and ending offset.
	SetOffset(int, int)
	// Returns this TOken's ending offset, one greater than the
	// position of the last character corresponding to this token in
	// the source text. The length of the token in the source text is
	// (endOffset() - startOffset()).
	// EndOffset() int
}

/* Default implementation of OffsetAttribute */
type OffsetAttributeImpl struct {
	startOffset, endOffset int
}

func newOffsetAttributeImpl() *util.AttributeImpl {
	return util.NewAttributeImpl(new(OffsetAttributeImpl))
}

func (a *OffsetAttributeImpl) Interfaces() []string {
	return []string{"OffsetAttribute"}
}

func (a *OffsetAttributeImpl) SetOffset(startOffset, endOffset int) {
	panic("not implemented yet")
}

func (a *OffsetAttributeImpl) Clear() {
	// TODO: we could use -1 as default here? Then we can assert in SetOffset...
	a.startOffset = 0
	a.endOffset = 0
}
