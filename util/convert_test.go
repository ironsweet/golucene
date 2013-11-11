package util

import (
	"testing"
)

func TestByte315ToFloat(t *testing.T) {
	data := [...]float32{0, 5.820766E-10, 6.9849193E-10, 8.1490725E-10}
	for i, target := range data {
		v := Byte315ToFloat(byte(i))
		if v-target > 1E-10 {
			t.Error("Bits to float conversion fail.")
		}
	}
}
