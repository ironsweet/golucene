package packed

import (
	"math/rand"
	// "fmt"
	"testing"
)

func TestAppendDeltaPackedLongBuffer1(t *testing.T) {
	buf := NewAppendingDeltaPackedLongBuffer(16, 1024, 0)
	assert(buf.Size() == 0)
	buf.Add(0)
	assert(buf.Size() == 1)
	assert(buf.Get(0) == 0)
	buf.Add(1)
	assert(buf.Size() == 2)
	assert(buf.Get(1) == 1)
}

func TestAppendDeltaPackedLongBuffer2(t *testing.T) {
	buf := NewAppendingDeltaPackedLongBuffer(4, 64, 0)
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

func TestAppendDeltaPackedLongBuffer3(t *testing.T) {
	buf := NewAppendingDeltaPackedLongBuffer(15, 128, 0)
	for _, v := range data {
		buf.Add(v)
	}
	next := buf.Iterator()
	for i, v := range data {
		v2, _ := next()
		if v2 != v {
			t.Fatalf("data[%v] should be %v (got %v)", i, v, v2)
		}
	}
}

var data = []int64{
	37159, 44666, 66150, 39156, 62910, 64700, 48509, 38309, 26583, 71682, 79900, 45908, 55380, 27865, 68578, 17602, 78550, 32543, 57900, 69956, 39162, 27937, 22125, 48992,
	79219, 41539, 18913, 24267, 61562, 24663, 53774, 58561, 22711, 18066, 77601, 27890, 33709, 79719, 76109, 27506, 26870, 23239, 40915, 20810, 40940, 41870, 45618, 29171, 76142,
	52051, 68272, 15313, 21244, 73731, 44526, 46098, 48452, 16142, 70296, 71232, 15191, 49664, 32567, 18864, 22958, 77735, 36760, 54198, 62200, 39413, 66646, 35346, 26964, 52708,
	73925, 54891, 26898, 66198, 74870, 54319, 75070, 22836, 65778, 62113, 70980, 24768, 27612, 46285, 54323, 50379, 57590, 78098, 19581, 77612, 62732, 18388, 18355, 62640, 68108,
	62156, 24707, 18373, 19003, 29426, 76082, 34032, 73548, 75005, 70351, 64601, 24904, 49544, 60419, 23496, 19232, 57691, 45808, 39036, 31510, 80476, 62186, 80546, 55021, 43474,
	29132, 51890, 24406, 71850, 55325, 76336, 58501, 76055, 44598, 55435, 68998, 38310, 52454, 73919, 21036, 79548, 38839, 32724, 77431, 68547, 36949, 63634, 60572, 49139, 29584,
	// 40556, 54411, 76689, 74509, 34163, 42154, 60224, 68327, 67244, 36882, 72817, 65023, 58443, 33602, 35579, 17116, 52379, 64929, 22796, 56112, 31701, 50842, 70785, 60829, 17858,
	// 46573, 41076, 48462, 68978, 50553, 63799, 65819, 66301, 71182, 44363, 38840, 49181, 51924, 67945, 43065, 78632, 67497, 54531, 79684, 21411, 30145, 40229, 58849, 26445, 20621,
}

func TestAppendDeltaPackedLongBuffer4(t *testing.T) {
	buf := NewAppendingDeltaPackedLongBuffer(16, 256, 0)
	for _, v := range data2 {
		buf.Add(v)
	}
	for i, v := range data2 {
		v2 := buf.Get(int64(i))
		if v2 != v {
			t.Errorf("buf[%v] should be %v (got %v)", i, v, v2)
		}
	}
	buf.freeze()
	next := buf.Iterator()
	for i, v := range data2 {
		v2, _ := next()
		if v2 != v {
			t.Fatalf("data2[%v] should be %v (got %v)", i, v, v2)
		}
	}
}

var data2 = []int64{
	6940219372182137078, 7778460041880943468, 7229210673732284536, 9110246765029446769, -7359921598399986212, -4607415284896715009,
	8311060135433658894, -8600976473745925460, -9178350477948938781, -7787173865322201423, 8945354000021812288, 9038212199459751411,
	-5848468738621273506, 8350620838402520651, 8208892880873945593, -8783468014022234340, -6848387412466609385, -4156839942976868659,
	9038326451371033698, 9074927733891398429, -5170658257362732224, 8240213471742998970, -4130612587730422708, -5685819278272085325,
	-7280529786002095952, 6595874591853951763, 7795145334146953943, -6510994955572395837, -3116903072584742493, 7067655052602529059,
	-6625778362009622413, 6880116486678206227, -5729119121351725364,
}

func testAppendingLongBuffer(t *testing.T) {
	arr := make([]int64, rand.Intn(1000000)+1)
	radioOptions := []float32{PackedInts.DEFAULT, PackedInts.COMPACT, PackedInts.FAST}
	for _, bpv := range []int{63, 0, 1, 63, 64, rand.Intn(61) + 2} {
		t.Logf("BPV: %v, len(arr): %v", bpv, len(arr))
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
			for i := 0; i < len(arr); i++ {
				arr[i] = rand.Int63()
			}
		default:
			minValue := rand.Int63n(MaxValue(bpv))
			for i := 0; i < len(arr); i++ {
				arr[i] = minValue + int64(inc)*int64(i) + rand.Int63()&MaxValue(bpv)
			}
		}

		for _, v := range arr {
			buf.Add(v)
		}
		if buf.Size() != int64(len(arr)) {
			t.Errorf("Size should be %v (got %v)", len(arr), buf.Size())
		}
		if rand.Intn(2) == 0 {
			buf.freeze()
			if rand.Intn(2) == 0 {
				// make sure double freeze doesn't break anything
				buf.freeze()
			}
		}
		if buf.Size() != int64(len(arr)) {
			t.Errorf("Size should be %v (got %v)", len(arr), buf.Size())
		}

		for i, v := range arr {
			v2 := buf.Get(int64(i))
			if v2 != v {
				t.Fatalf("[%v] should be %v (got %v)", i, v, v2)
			}
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
