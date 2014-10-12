package util

import (
	. "github.com/balzaczyy/gounit"
	"testing"
)

func TestNextSetBit(t *testing.T) {
	n := NewOpenBitSet()
	n.Set(0)
	n.Set(64)
	i := n.NextSetBit(0)
	It(t).Should("First set bit is 0 (got %v)", i).Verify(i == 0)
	i = n.NextSetBit(i + 1)
	It(t).Should("Second set bit is 64 (got %v)", i).Verify(i == 64)
}
