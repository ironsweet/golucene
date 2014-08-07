package blocktree

import (
	"testing"
)

func TestBrToString(t *testing.T) {
	s := brToString(make([]byte, 0, 3))
	if len(s) != 3 {
		t.Error("Reserved space should be hidden.")
	}
}
