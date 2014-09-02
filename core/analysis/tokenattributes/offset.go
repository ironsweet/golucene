package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

/* The start and end character offset of a Token. */
type OffsetAttribute interface {
	util.Attribute
	// Returns this Token's starting offset, the position of the first
	// character corresponding to this token in the source text.
	StartOffset() int
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
	EndOffset() int
}

/* Default implementation of OffsetAttribute */
type OffsetAttributeImpl struct {
	startOffset, endOffset int
}

func newOffsetAttributeImpl() util.AttributeImpl {
	return new(OffsetAttributeImpl)
}

func (a *OffsetAttributeImpl) Interfaces() []string {
	return []string{"OffsetAttribute"}
}

func (a *OffsetAttributeImpl) StartOffset() int {
	return a.startOffset
}

func (a *OffsetAttributeImpl) SetOffset(startOffset, endOffset int) {
	// TODO: we could assert that this is set-once, ie, current value
	// are -1? Very few token filters should change offsets once set by
	// the tokenizer... and tokenizer should call clearAtts before
	// re-using OffsetAtt
	assert2(startOffset >= 0 && startOffset <= endOffset,
		"startOffset must be non-negative, and endOffset must be >= startOffset, startOffset=%v,endOffset=%v",
		startOffset, endOffset)
	a.startOffset = startOffset
	a.endOffset = a.endOffset
}

func (a *OffsetAttributeImpl) EndOffset() int {
	return a.endOffset
}

func (a *OffsetAttributeImpl) Clear() {
	// TODO: we could use -1 as default here? Then we can assert in SetOffset...
	a.startOffset = 0
	a.endOffset = 0
}

func (a *OffsetAttributeImpl) Clone() util.AttributeImpl {
	return &OffsetAttributeImpl{
		startOffset: a.startOffset,
		endOffset:   a.endOffset,
	}
}

func (a *OffsetAttributeImpl) CopyTo(target util.AttributeImpl) {
	target.(OffsetAttribute).SetOffset(a.startOffset, a.endOffset)
}
