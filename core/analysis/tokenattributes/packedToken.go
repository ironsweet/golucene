package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

/*
Default implementation of the common attributes used by Lucene:
- CharTermAttribute
- TypeAttribute
- PositionIncrementAttribute
- PositionLengthAttribute
- OffsetAttribute
*/
type PackedTokenAttributeImpl struct {
}

func NewPackedTokenAttribute() *util.AttributeImpl {
	panic("not implemented yet")
}

func (a *PackedTokenAttributeImpl) Interfaces() []string {
	return []string{"CharTermAttribute", "TermToBytesRefAttribute",
		"TypeAttribute", "PositionIncrementAttribute",
		"PositionLengthAttribute", "OffsetAttribute"}
}

func (a *PackedTokenAttributeImpl) Clear() {
	panic("not implemented yet")
}
