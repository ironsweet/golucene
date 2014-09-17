package packed

import (
	"math/rand"
	"testing"
)

func TestAppendDeltaPackedLongBuffer(t *testing.T) {
	buf := NewAppendingDeltaPackedLongBuffer(16, 1024, 0)
	assert(buf.Size() == 0)
	buf.Add(0)
	assert(buf.Size() == 1)
	assert(buf.Get(0) == 0)
	buf.Add(1)
	assert(buf.Size() == 2)
	assert(buf.Get(1) == 1)

	buf = NewAppendingDeltaPackedLongBuffer(4, 64, 0)
	for i := 0; i <= 65; i++ {
		buf.Add(int64(i))
	}
	assert(buf.Size() == 66)
	for i := 0; i <= 65; i++ {
		v := buf.Get(int64(i))
		if v != int64(i) {
			t.Errorf("buf[%v] should be %v (got %v)", i, i, v)
		}
	}

	next := buf.Iterator()
	for i := 0; i <= 65; i++ {
		v, ok := next()
		assert(ok)
		if v != int64(i) {
			t.Fatalf("buf[%v] should be %v (got %v)", i, i, v)
		}
	}
	_, ok := next()
	assert(!ok)
}

func TestAppendingLongBuffer(t *testing.T) {
	arr := make([]int64, rand.Intn(100)+1)
	radioOptions := []float32{PackedInts.DEFAULT, PackedInts.COMPACT, PackedInts.FAST}
	for _, bpv := range []int{0, 1, 63, 64, rand.Intn(61) + 2} {
		t.Logf("BPV: %v", bpv)
		pageSize := 1 << uint(rand.Intn(15)+6)
		initialPageCount := rand.Intn(17)
		acceptableOverheadRatio := radioOptions[rand.Intn(len(radioOptions))]
		t.Logf("Init AppendingDeltaPackedLongBuffer(%v,%v,%v)", initialPageCount, pageSize, acceptableOverheadRatio)

		// TODO support more data type
		var buf AppendingLongBuffer
		buf = NewAppendingDeltaPackedLongBuffer(initialPageCount, pageSize, acceptableOverheadRatio)
		inc := 0

		switch bpv {
		case 0:
			arr[0] = rand.Int63()
			for i := 1; i < len(arr); i++ {
				arr[i] = arr[i-1] + int64(inc)
			}
		case 64:
			panic("niy")
		default:
			minValue := rand.Int63n(MaxValue(bpv))
			for i := 0; i < len(arr); i++ {
				arr[i] = minValue + int64(inc)*int64(i) + rand.Int63()&MaxValue(bpv)
			}
		}

		for _, v := range arr {
			buf.Add(v)
		}
		assert(int64(len(arr)) == buf.Size())
		if rand.Intn(2) == 0 {
			buf.freeze()
			if rand.Intn(2) == 0 {
				// make sure double freeze doesn't break anything
				buf.freeze()
			}
		}
		assert(int64(len(arr)) == buf.Size())

		for i, v := range arr {
			assert2(v == buf.Get(int64(i)), "[%v]%v vs %v", i, v, buf.Get(int64(i)))
		}

		next := buf.Iterator()
		for i, v := range arr {
			v2, ok := next()
			if rand.Intn(2) == 0 {
				assert(ok)
			}
			if v != v2 {
				t.Fatalf("buf[%v] should be %v (got %v)", i, v, v2)
			}
		}
		_, ok := next()
		assert(!ok)

		target := make([]int64, len(arr)+1024)
		for i := 0; i < len(arr); i += rand.Intn(10000) + 1 {
			lenToRead := rand.Intn(buf.pageSize()*2) + 1
			if len(target)-i < lenToRead {
				lenToRead = len(target) - i
			}
			lenToCheck := lenToRead
			if len(arr)-i < lenToCheck {
				lenToCheck = len(arr) - i
			}
			off := i
			for off < len(arr) && lenToRead > 0 {
				read := buf.GetBulk(int64(off), target[off:off+lenToRead])
				assert(read > 0)
				assert(read <= lenToRead)
				lenToRead -= read
				off += read
			}

			for j := 0; j < lenToCheck; j++ {
				assert(arr[j+i] == target[j+i])
			}
		}

		// TODO test ramBytesUsed()
	}
}
