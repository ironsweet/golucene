package analysis

import (
	ca "github.com/balzaczyy/golucene/core/analysis"
	"math/rand"
)

type MockAnalyzer struct {
	*ca.AnalyzerImpl
}

func NewMockAnalyzerWithRandom(r *rand.Rand) *MockAnalyzer {
	panic("not implemented yet")
}
