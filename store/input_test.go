package store

import (
	"fmt"
	"math"
	"testing"
)

// func writeBytes(path string, size int64) error {
// 	fos.Open(name)
// }

const TEST_FILE_LENGTH = int64(100 * 1024)

func TestReadByte(t *testing.T) {
	input := newMyBufferedIndexInput(math.MaxInt64)
	for i := 0; i < BUFFER_SIZE; i++ {
		b, err := input.ReadByte()
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(t, b, byten(int64(i)))
	}
}

func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("Expected '%v', but '%v'", b, a)
	}
}

func byten(n int64) byte {
	return byte(n * n % 256)
}

type MyBufferedIndexInput struct {
	*BufferedIndexInput
	pos    int64
	length int64
}

func newMyBufferedIndexInput(length int64) MyBufferedIndexInput {
	ans := MyBufferedIndexInput{pos: 0, length: length}
	ans.BufferedIndexInput = newBufferedIndexInputBySize(fmt.Sprintf(
		"MyBufferedIndexInput(len=%v)", length), BUFFER_SIZE)
	ans.BufferedIndexInput.readInternal = func(buf []byte) error {
		for i, _ := range buf {
			buf[i] = byten(ans.pos)
			ans.pos++
		}
		return nil
	}
	ans.BufferedIndexInput.seekInternal = func(pos int64) {
		ans.pos = pos
	}
	ans.BufferedIndexInput.close = func() error {
		return nil
	}
	ans.BufferedIndexInput.length = func() int64 {
		return ans.length
	}
	return ans
}
