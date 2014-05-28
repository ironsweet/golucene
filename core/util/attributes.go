package util

import (
	"reflect"
)

// util/Attribute.java

/* Base interface for attributes. */
type Attribute interface{}

// util/AttributeImpl.java

/*
Base class for Attributes that can be added to a AttributeSource.

Attributes are used to add data in a dynamic, yet type-safe way to a
source of usually streamed ojects, e.g. a TokenStream.
*/
type AttributeImpl struct{}

// util/AttributeSource.java

/* An AttributeFactory creates instances of AttributeImpls. */
type AttributeFactory interface {
	Create(reflect.Type) *AttributeImpl
}

type DefaultAttributeFactory struct {
	attTypeImplMap map[reflect.Type]Attribute
}

func (fac *DefaultAttributeFactory) Create(t reflect.Type) *AttributeImpl {
	panic("not implemented yet")
}

var DEFAULT_ATTRIBUTE_FACTORY = &DefaultAttributeFactory{
	make(map[reflect.Type]Attribute),
}

/*
An AttributeSource contains a list of different AttributeImpls, and
methods to add and get them. There can only be a single instance of
an attribute in the same AttributeSource instance. This is ensured by
passing in the actual type of the Attribute (reflect.TypeOf(Attribute))
to the #AddAttribute(Type), which then checks if an instance of that
type is already present. If yes, it returns the instance, otherwise
it creates a new instance and returns it.
*/
type AttributeSource struct {
}

/* An AttributeSource using the default attribute factory */
func NewAttributeSource() *AttributeSource {
	return NewAttributeSourceWith(DEFAULT_ATTRIBUTE_FACTORY)
}

/* An AttributeSource using the supplied AttributeFactory for creating new Attribute instance. */
func NewAttributeSourceWith(factory AttributeFactory) *AttributeSource {
	panic("not implemented yet")
}
