package analysis

import (
	"github.com/balzaczyy/golucene/core/util"
)

type myAttributeFactorySPI interface {
	Create() interface{}
}

/*
Expert: AttributeFactory returning an instance of the given clazz for
the attributes it implements. For all other attributes it calls the
given delegate factory as fallback. This clas can be used to prefer a
specific AttributeImpl which combines multiple attributes over
separate classes.
*/
type myAttributeFactory struct {
	delegate util.AttributeFactory
	clazz    map[string]bool
	ctor     func() util.AttributeImpl
}

func (f *myAttributeFactory) Create(name string) util.AttributeImpl {
	if _, ok := f.clazz[name]; ok {
		return f.ctor()
	}
	return f.delegate.Create(name)
}

/*
Returns an AttributeFactory returning an instance of the given clazz
for the attributes it implements. The given clazz must have a public
no-arg contructor. For all other attributes it calls the given
delegate factory as fallback. This method can be used to prefer a
specific AttributeImpl which combines multiple attributes over
separate classes.

Please save instances created by this method in a global field,
because on each call, this does
*/
func assembleAttributeFactory(factory util.AttributeFactory,
	clazz map[string]bool, ctor func() util.AttributeImpl) util.AttributeFactory {
	return &myAttributeFactory{factory, clazz, ctor}
}
