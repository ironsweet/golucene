package util

import (
	"reflect"
)

// util/Attribute.java

/* Base interface for attributes. */
type Attribute interface {
}

// util/AttributeImpl.java

/*
Base class for Attributes that can be added to a AttributeSource.

Attributes are used to add data in a dynamic, yet type-safe way to a
source of usually streamed ojects, e.g. a TokenStream.
*/
type AttributeImpl interface {
	Interfaces() []string
	Clone() AttributeImpl
	// Clears the values in this AttributeImpl and resets it to its
	// default value. If this implementation implements more than one
	// Attribute interface, it clears all.
	Clear()
	CopyTo(target AttributeImpl)
}

// util/AttributeFactory.java

/* An AttributeFactory creates instances of AttributeImpls. */
type AttributeFactory interface {
	Create(string) AttributeImpl
}

// util/AttributeSource.java

/* This class holds the state of an AttributeSource */
type AttributeState struct {
	attribute AttributeImpl
	next      *AttributeState
}

func (s *AttributeState) Clone() *AttributeState {
	ans := new(AttributeState)
	ans.attribute = s.attribute.Clone()
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
	attributes     map[string]AttributeImpl
	attributeImpls map[reflect.Type]AttributeImpl
	_currentState  []*AttributeState
	factory        AttributeFactory
}

/* An AttributeSource using the default attribute factory */
// func NewAttributeSource() *AttributeSource {
// 	return NewAttributeSourceWith(DEFAULT_ATTRIBUTE_FACTORY)
// }

/* An AttributeSource that uses the same attributes as the supplied one. */
func NewAttributeSourceFrom(input *AttributeSource) *AttributeSource {
	assert2(input != nil, "input AttributeSource must not be null")
	return &AttributeSource{
		attributes:     input.attributes,
		attributeImpls: input.attributeImpls,
		_currentState:  input._currentState,
		factory:        input.factory,
	}
}

/* An AttributeSource using the supplied AttributeFactory for creating new Attribute instance. */
func NewAttributeSourceWith(factory AttributeFactory) *AttributeSource {
	// Note that Lucene Java use LinkedHashMap to keep insert order.
	// But it's used by Solr only and GoLucene doesn't have plan to
	// port GoSolr. So we use plain map here.
	return &AttributeSource{
		attributes:     make(map[string]AttributeImpl),
		attributeImpls: make(map[reflect.Type]AttributeImpl),
		_currentState:  make([]*AttributeState, 1),
		factory:        factory,
	}
}

/*
Expert: Adds a custom AttributeImpl instance with one or more
Attribute interfaces.

Please note: it is not guaranteed, that att is added to the
AttributeSource, because the provided attributes may already exist.
You should always retrieve the wanted attributes using Get() after
adding with this method and cast to you class.

The recommended way to use custom implementations is using an
AttributeFactory.
*/
func (as *AttributeSource) AddImpl(att AttributeImpl) {
	typ := reflect.TypeOf(att)
	if _, ok := as.attributeImpls[typ]; ok {
		return
	}

	// add all interfaces of this AttributeImpl to the maps
	for _, curInterface := range att.Interfaces() {
		// Attribute is a superclass of this interface
		if _, ok := as.attributes[curInterface]; !ok {
			// invalidate state to force recomputation in captureState()
			as._currentState[0] = nil
			as.attributes[curInterface] = att
			as.attributeImpls[typ] = att
		}
	}
}

/* Returns true, iff this AttributeSource has any attributes */
func (as *AttributeSource) hasAny() bool {
	return len(as.attributes) > 0
}

/* Returns true, iff this AttributeSource contains the passed-in Attribute. */
func (as *AttributeSource) Has(s string) bool {
	_, ok := as.attributes[s]
	return ok
}

/*
The caller must pass in a Attribute instance. This method first
checks if an instance of that type is already in this AttributeSource
and returns it. Otherwise a new instance is created, added to this
AttributeSource and returned.
*/
func (as *AttributeSource) Add(s string) Attribute {
	attImpl, ok := as.attributes[s]
	if !ok {
		attImpl = as.factory.Create(s)
		as.AddImpl(attImpl)
	}
	return attImpl
}

/* Returns the instance of the passe in Attribute contained in this AttributeSource */
func (as *AttributeSource) Get(s string) Attribute {
	return as.attributes[s]
}

func (as *AttributeSource) currentState() *AttributeState {
	s := as._currentState[0]
	if s != nil || !as.hasAny() {
		return s
	}
	s = new(AttributeState)
	var c *AttributeState
	for _, v := range as.attributeImpls {
		if c == nil {
			c = s
			c.attribute = v
		} else {
			c.next = &AttributeState{attribute: v}
			c = c.next
		}
	}
	return s
}

func (as *AttributeSource) CaptureState() (state *AttributeState) {
	if state = as.currentState(); state != nil {
		state = state.Clone()
	}
	return
}

func (as *AttributeSource) RestoreState(state *AttributeState) {
	if state == nil {
		return
	}

	for {
		targetImpl, ok := as.attributeImpls[reflect.TypeOf(state.attribute)]
		assert2(ok,
			"State contains AttributeImpl of type %v that is not in in this AttributeSource",
			reflect.TypeOf(state.attribute).Name())
		state.attribute.CopyTo(targetImpl)
		state = state.next
		if state == nil {
			break
		}
	}
}

/*
Resets all Attributes in this AttributeSource by calling
AttributeImpl.clear() on each Attribute implementation.
*/
func (as *AttributeSource) Clear() {
	for state := as.currentState(); state != nil; state = state.next {
		state.attribute.Clear()
	}
}

/*
Returns a string consisting of the class's simple name, the hex
representation of the identity hash code, and the current reflection
of all attributes.
*/
func (as *AttributeSource) String() string {
	panic("not implemented yet")
}
