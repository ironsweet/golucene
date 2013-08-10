package util

import (
	"testing"
)

func TestStripSegmentName(t *testing.T) {
	s := StripSegmentName("_0.fnm")
	if s != ".fnm" {
		t.Errorf("Expected '.fnm' but was '%v'", s)
	}
}
