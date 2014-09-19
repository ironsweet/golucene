package packed

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestByteCount(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	inters := 10 // >= 3
	for i := 0; i < inters; i++ {
		valueCount := rand.Int31n(math.MaxInt32-1) + 1 // [1, 1^32-1]
		for j := 0; j <= 1; j++ {
			format := PackedFormat(j)
			for bpv := uint32(1); bpv <= 64; bpv++ {
				byteCount := format.ByteCount(VERSION_CURRENT, valueCount, bpv)
				msg := fmt.Sprintf("format=%v, byteCount=%v, valueCount=%v, bpv=%v", format, byteCount, valueCount, bpv)
				if byteCount*8 < int64(valueCount)*int64(bpv) {
					t.Errorf(msg)
				}
				if format == PACKED {
					if (byteCount-1)*8 >= int64(valueCount)*int64(bpv) {
						t.Errorf(msg)
					}
				}
			}
		}
	}
}

func TestMaxValue(t *testing.T) {
	if MaxValue(0) != 0 {
		t.Error("0 bit -> 0")
	}
	if MaxValue(1) != 1 {
		t.Error("1 bit -> 1")
	}
	if MaxValue(2) != 3 {
		t.Error("2 bits -> 3")
	}
	if MaxValue(64) != 0x7fffffffffffffff {
		t.Error("64 bits -> 0x7fffffffffffffff")
	}
}

func TestGetPackedIntsDecoder(t *testing.T) {
	decoder := GetPackedIntsDecoder(PackedFormat(PACKED), int32(PACKED_VERSION_START), 1)
	if n := decoder.ByteValueCount(); n != 8 {
		t.Errorf("ByteValueCount() should be 8, instead of %v", n)
	}
}

func TestUnsignedBitsRequired(t *testing.T) {
	if n := UnsignedBitsRequired(-158146830731166066); n != 64 {
		t.Errorf("-158146830731166066 -> 64bit (got %v)", n)
	}
}
