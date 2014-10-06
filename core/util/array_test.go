package util

import (
	"testing"
)

/* Ensure nextSize() gives linear amortized cost of realloc/copy */
func TestGrowth(t *testing.T) {
	var currentSize = 0
	var copyCost int64 = 0

	// Make sure it hits math.MaxInt32, if we insist:
	for currentSize != MAX_ARRAY_LENGTH {
		nextSize := Oversize(1+currentSize, NUM_BYTES_OBJECT_REF)
		assert2(nextSize > currentSize, "%v -> %v", currentSize, nextSize)
		if currentSize > 0 {
			copyCost += int64(currentSize)
			copyCostPerElement := float64(copyCost) / float64(currentSize)
			assert2(copyCostPerElement < 10, "cost %v", copyCostPerElement)
		}
		currentSize = nextSize
	}
}
