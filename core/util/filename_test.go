package util

import (
	"testing"
)

func TestStripSegmentName(t *testing.T) {
	s := StripSegmentName("_0.fnm")
	if s != ".fnm" {
		t.Errorf("Expected '.fnm' but was '%v'", s)
	}

	s = StripSegmentName("_0_Lucene41_0.doc")
	if s != "_Lucene41_0.doc" {
		t.Errorf("Expected '_Lucene41_0.doc', but was '%v'", s)
	}
}
