package tokenattributes

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

type DefaultAttributeFactory struct{}

func (fac *DefaultAttributeFactory) Create(name string) util.AttributeImpl {
	switch name {
	case "PositionIncrementAttribute":
		return newPositionIncrementAttributeImpl()
	case "CharTermAttribute":
		return newCharTermAttributeImpl()
	case "OffsetAttribute":
		return newOffsetAttributeImpl()
	case "TypeAttribute":
		return newTypeAttributeImpl()
	case "PayloadAttribute":
		return newPayloadAttributeImpl()
	}
	panic(fmt.Sprintf("not supported yet: %v", name))
}

/*
This is the default factory that creates AttributeImpls using the
class name of the supplied Attribute interface class by appending
Impl to it.
*/
var DEFAULT_ATTRIBUTE_FACTORY = new(DefaultAttributeFactory)
