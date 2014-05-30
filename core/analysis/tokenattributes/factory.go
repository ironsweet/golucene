package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

type DefaultAttributeFactory struct{}

func (fac *DefaultAttributeFactory) Create(name string) *util.AttributeImpl {
	switch name {
	case "PositionIncrementAttribute":
		return newPositionIncrementAttributeImpl()
	case "CharTermAttribute":
		return newCharTermAttributeImpl()
	}
	panic("not supported yet")
}

/*
This is the default factory that creates AttributeImpls using the
class name of the supplied Attribute interface class by appending
Impl to it.
*/
var DEFAULT_ATTRIBUTE_FACTORY = new(DefaultAttributeFactory)
