package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

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

func newTypeAttributeImpl() *util.AttributeImpl {
	panic("not implemented yet")
}

func (a *TypeAttributeImpl) Interfaces() []string {
	return []string{"TypeAttribute"}
}

func (a *TypeAttributeImpl) SetType(typ string) {
	a.typ = typ
}
