package util

import (
	"testing"
)

func TestMurmurHash3_x86_32(t *testing.T) {
	verifyMurmurHash3_x86_32(t, []byte{98, 97, 114}, 1553420910, 609023304)
	verifyMurmurHash3_x86_32(t, []byte{98, 97, 114}, 1553490497, 818114846)
	verifyMurmurHash3_x86_32(t, []byte{231, 175, 135}, 223189302, 1845636694)
}

func verifyMurmurHash3_x86_32(t *testing.T, data []byte, seed, expected uint32) {
	if hash := MurmurHash3_x86_32(data, seed); hash != expected {
		t.Error("Fail to do hash using MurmurHash3_x86_32")
	}
}
