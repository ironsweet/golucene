package classic

import (
	"io"
)

type FastCharStream struct {
	buffer []rune

	bufferLength   int
	bufferPosition int

	tokenStart  int
	bufferStart int

	input io.RuneReader // source of chars
}

func newFastCharStream(r io.RuneReader) *FastCharStream {
	return &FastCharStream{input: r}
}

func (cs *FastCharStream) readChar() (rune, error) {
	panic("not implemented yet")
}

func (cs *FastCharStream) beginToken() (rune, error) {
	cs.tokenStart = cs.bufferPosition
	return cs.readChar()
}

func (cs *FastCharStream) backup(amount int) {
	panic("not implemented yet")
}

func (cs *FastCharStream) image() string {
	panic("not implemented yet")
}

func (cs *FastCharStream) endColumn() int {
	panic("not implemented yet")
}

func (cs *FastCharStream) endLine() int {
	panic("not implemented yet")
}
