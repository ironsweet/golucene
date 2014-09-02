package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

const DEFAULT_TYPE = "word"

/* A Token's lexical type. The default value is "word". */
type TypeAttribute interface {
	util.Attribute
	// Set the lexical type.
	SetType(string)
}

/* Default implementation of TypeAttribute */
type TypeAttributeImpl struct {
	typ string
}

func newTypeAttributeImpl() util.AttributeImpl {
	return &TypeAttributeImpl{
		typ: DEFAULT_TYPE,
	}
}

func (a *TypeAttributeImpl) Interfaces() []string {
	return []string{"TypeAttribute"}
}

func (a *TypeAttributeImpl) SetType(typ string) {
	a.typ = typ
}

func (a *TypeAttributeImpl) Clear() {
	a.typ = DEFAULT_TYPE
}

func (a *TypeAttributeImpl) Clone() util.AttributeImpl {
	return &TypeAttributeImpl{
		typ: a.typ,
	}
}

func (a *TypeAttributeImpl) CopyTo(target util.AttributeImpl) {
	target.(TypeAttribute).SetType(a.typ)
}
