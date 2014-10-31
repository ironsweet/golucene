package util

import (
	"github.com/balzaczyy/golucene/core/util"
	"testing"
)

func TestStopwordAnalyzerBaseCtor(t *testing.T) {
	NewStopwordAnalyzerBaseWithStopWords(util.VERSION_4_10_1, nil)
}
