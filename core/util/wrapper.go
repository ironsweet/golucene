package util

import (
	"fmt"
	"sync"
)

// util/SetOnce.java

/*
A convenient class which offers a semi-immutable object wrapper
implementation which allows one to set the value of an object exactly
once, and retrieve it many times. If Set() is called more than once,
error is returned and the operation will fail.
*/
type SetOnce struct {
	*sync.Once
	obj interface{} // volatile
}

func NewSetOnce() *SetOnce {
	return &SetOnce{Once: &sync.Once{}}
}

func NewSetOnceOf(obj interface{}) *SetOnce {
	return &SetOnce{&sync.Once{}, obj}
}

// Sets the given object. If the object has already been set, an exception is thrown.
func (so *SetOnce) Set(obj interface{}) {
	so.Do(func() { so.obj = obj })
	assert2(so.obj == obj, "The object cannot be set twice!")
}

// Returns the object set by Set().
func (so *SetOnce) Get() interface{} {
	return so.obj
}

func (so *SetOnce) Clone() *SetOnce {
	if so.obj == nil {
		return NewSetOnce()
	}
	return NewSetOnceOf(so.obj)
}

func (so *SetOnce) String() string {
	if so.obj == nil {
		return "undefined"
	}
	return fmt.Sprintf("%v", so.obj)
}
