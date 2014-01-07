package util

import (
	"testing"
)

// util/TestSetOnce.java

func TestSetOnce(t *testing.T) {
	set := NewSetOnce()
	set.Set(5)
	assert(set.Get() == 5)
	defer func() {
		assert(recover() != nil)
	}()
	set.Set(7)
	panic("should not be here")
}
