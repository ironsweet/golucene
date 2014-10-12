package store

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"
)

func writeBytes(aFile *os.File, size int64) (err error) {
	buf := make([]byte, 0, 4096)
	for i := int64(0); i < size; i++ {
		if i%4096 == 0 && len(buf) > 0 {
			_, err = aFile.Write(buf)
			if err != nil {
				return err
			}
			buf = make([]byte, 0, 4096)
		}
		buf = append(buf, byten(i))
	}
	if len(buf) > 0 {
		_, err = aFile.Write(buf)
	}
	return err
}

const TEST_FILE_LENGTH = int64(100 * 1024)
const TEMP_DIR = ""

// Call readByte() repeatedly, past the buffer boundary, and see that it
// is working as expected.
// Our input comes from a dynamically generated/ "file" - see
// MyBufferedIndexInput below.
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

// Call readBytes() repeatedly, with various chunk sizes (from 1 byte to
// larger than the buffer size), and see that it returns the bytes we expect.
// Our input comes from a dynamically generated "file" -
// see MyBufferedIndexInput below.
func TestReadBytes(t *testing.T) {
	input := newMyBufferedIndexInput(math.MaxInt64)
	err := runReadBytes(input, BUFFER_SIZE, random(), t)
	if err != nil {
		t.Error(err)
	}

	// This tests the workaround code for LUCENE-1566 where readBytesInternal
	// provides a workaround for a JVM Bug that incorrectly raises a OOM Error
	// when a large byte buffer is passed to a file read.
	// NOTE: this does only test the chunked reads and NOT if the Bug is triggered.
	//final int tmpFileSize = 1024 * 1024 * 5;
	inputBufferSize := 128
	tmpInputFile, err := ioutil.TempFile(TEMP_DIR, "IndexInput")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		tmpInputFile.Close()
		os.Remove(tmpInputFile.Name())
	}()
	err = writeBytes(tmpInputFile, TEST_FILE_LENGTH)
	if err != nil {
		t.Fatal(err)
	}

	// run test with chunk size of 10 bytes
	in, err := newSimpleFSIndexInput(fmt.Sprintf("SimpleFSIndexInput(path='%v')", tmpInputFile), tmpInputFile.Name(),
		newTestIOContext(random()))
	if err != nil {
		t.Error(err)
	}
	err = runReadBytesAndClose(in, inputBufferSize, random(), t)
	if err != nil {
		t.Error(err)
	}
}

func random() *rand.Rand {
	seed := time.Now().Unix()
	fmt.Println("Seed: ", seed)
	return rand.New(rand.NewSource(seed))
}

func runReadBytesAndClose(input IndexInput, bufferSize int, r *rand.Rand, t *testing.T) (err error) {
	defer func() {
		err = input.Close()
	}()

	return runReadBytes(input, bufferSize, r, t)
}

func runReadBytes(input IndexInput, bufferSize int, r *rand.Rand, t *testing.T) (err error) {
	pos := 0
	// gradually increasing size:
	for size := 1; size < bufferSize*10; size += size/200 + 1 {
		err = checkReadBytes(input, size, pos, t)
		if err != nil {
			return err
		}
		if pos += size; int64(pos) >= TEST_FILE_LENGTH { // wrap
			pos = 0
			input.Seek(0)
		}
	}
	// wildly fluctuating size:
	for i := int64(0); i < 100; i++ {
		size := r.Intn(10000)
		err = checkReadBytes(input, size+1, pos, t)
		if err != nil {
			return err
		}
		if pos += size + 1; int64(pos) >= TEST_FILE_LENGTH { // wrap
			pos = 0
			input.Seek(0)
		}
	}
	// constant small size (7 bytes):
	for i := 0; i < bufferSize; i++ {
		err = checkReadBytes(input, 7, pos, t)
		if err != nil {
			return err
		}
		if pos += 7; int64(pos) >= TEST_FILE_LENGTH { // wrap
			pos = 0
			input.Seek(0)
		}
	}
	return nil
}

var buffer []byte = make([]byte, 10)

func grow(buffer []byte, newCap int) []byte {
	if newCap <= cap(buffer) {
		return buffer[0:newCap]
	}
	ans := make([]byte, newCap)
	copy(ans, buffer)
	return ans
}

func checkReadBytes(input IndexInput, size, pos int, t *testing.T) error {
	// Just to see that "offset" is treated properly in readBytes(), we
	// add an arbitrary offset at the beginning of the array
	offset := size % 10 // arbitrary
	buffer = grow(buffer, offset+size)
	assertEquals(t, int64(pos), input.FilePointer())
	left := TEST_FILE_LENGTH - input.FilePointer()
	if left <= 0 {
		return nil
	} else if left < int64(size) {
		size = int(left)
	}
	if err := input.ReadBytes(buffer[offset : offset+size]); err != nil {
		return err
	}
	assertEquals(t, int64(pos+size), input.FilePointer())
	for i := 0; i < size; i++ {
		assertEquals(t, byten(int64(pos+i)), buffer[offset+i])
	}
	return nil
}

// This tests that attempts to readBytes() past an EOF will fail, while
// reads up to the EOF will succeed. The EOF is determined by the
// BufferedIndexInput's arbitrary length() value.
func testEOF(t *testing.T) {
	input := newMyBufferedIndexInput(1024)
	// see that we can read all the bytes at one go:
	if err := checkReadBytes(input, int(input.Length()), 0, t); err != nil {
		t.Error(err)
	}
	// go back and see that we can't read more than that, for small and
	// large overflows:
	pos := int(input.Length() - 10)
	input.Seek(int64(pos))
	if err := checkReadBytes(input, 10, pos, t); err != nil {
		t.Error(err)
	}
	input.Seek(int64(pos))
	if err := checkReadBytes(input, 11, pos, t); err == nil {
		t.Error("Block read past end of file")
	}
	input.Seek(int64(pos))
	if err := checkReadBytes(input, 50, pos, t); err == nil {
		t.Error("Block read past end of file")
	}
	input.Seek(int64(pos))
	if err := checkReadBytes(input, 100000, pos, t); err == nil {
		t.Error("Block read past end of file")
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

func newMyBufferedIndexInput(length int64) *MyBufferedIndexInput {
	ans := &MyBufferedIndexInput{pos: 0, length: length}
	ans.BufferedIndexInput = newBufferedIndexInputBySize(ans, fmt.Sprintf(
		"MyBufferedIndexInput(len=%v)", length), BUFFER_SIZE)
	return ans
}

func (in *MyBufferedIndexInput) readInternal(buf []byte) error {
	for i, _ := range buf {
		buf[i] = byten(in.pos)
		in.pos++
	}
	return nil
}

func (in *MyBufferedIndexInput) seekInternal(pos int64) error {
	in.pos = pos
	return nil
}

func (in *MyBufferedIndexInput) Close() error {
	return nil
}

func (in *MyBufferedIndexInput) Length() int64 {
	return in.length
}

func (in *MyBufferedIndexInput) Slice(desc string, offset, length int64) (IndexInput, error) {
	panic("not supported")
}

func (in *MyBufferedIndexInput) Clone() IndexInput {
	return &MyBufferedIndexInput{
		in.BufferedIndexInput.Clone(),
		in.pos,
		in.length,
	}
}
