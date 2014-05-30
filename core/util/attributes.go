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

func (v *AttributeImpl) Clone() *AttributeImpl {
	panic("not implemented yet")
}

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

/* This class holds the state of an AttributeSource */
type AttributeState struct {
	value *AttributeImpl
	next  *AttributeState
}

func (s *AttributeState) Clone() *AttributeState {
	ans := new(AttributeState)
	ans.value = s.value.Clone()
	if s.next != nil {
		ans.next = s.next.Clone()
	}
	return ans
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
	attributes     map[reflect.Type]Attribute
	attributeImpls map[reflect.Type]*AttributeImpl
	currentState   []*AttributeState
	factory        AttributeFactory
}

/* An AttributeSource using the default attribute factory */
func NewAttributeSource() *AttributeSource {
	return NewAttributeSourceWith(DEFAULT_ATTRIBUTE_FACTORY)
}

/* An AttributeSource that uses the same attributes as the supplied one. */
func NewAttributeSourceFrom(input *AttributeSource) *AttributeSource {
	assert2(input != nil, "input AttributeSource must not be null")
	return &AttributeSource{
		attributes:     input.attributes,
		attributeImpls: input.attributeImpls,
		currentState:   input.currentState,
		factory:        input.factory,
	}
}

/* An AttributeSource using the supplied AttributeFactory for creating new Attribute instance. */
func NewAttributeSourceWith(factory AttributeFactory) *AttributeSource {
	// Note that Lucene Java use LinkedHashMap to keep insert order.
	// But it's used by Solr only and GoLucene doesn't have plan to
	// port GoSolr. So we use plain map here.
	return &AttributeSource{
		attributes:     make(map[reflect.Type]Attribute),
		attributeImpls: make(map[reflect.Type]*AttributeImpl),
		currentState:   make([]*AttributeState, 1),
		factory:        factory,
	}
}

func (as *AttributeSource) Add(v Attribute) Attribute {
	panic("not implemented yet")
}
