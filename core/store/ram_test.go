package store

import (
	"testing"
)

func TestIO(t *testing.T) {
	filename := "a.txt"
	testdata := "hello world"

	dir := NewRAMDirectory()
	func() {
		out, err := dir.CreateOutput(filename, IO_CONTEXT_DEFAULT)
		assert2(err == nil, "%v", err)
		defer out.Close()

		err = out.WriteString(testdata)
		assert2(err == nil, "%v", err)
	}()

	f := dir.GetRAMFile(filename)
	assertEquals(t, f.length, int64(len(testdata))+1)
	t.Log(f.buffers)

	in, err := dir.OpenInput(filename, IO_CONTEXT_DEFAULT)
	assert2(err == nil, "%v", err)
	s, err := in.ReadString()
	assert2(err == nil, "%v", err)
	assertEquals(t, s, testdata)
}
