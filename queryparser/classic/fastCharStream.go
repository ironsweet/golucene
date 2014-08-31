package classic

import (
	"io"
)

type FastCharStream struct {
	input io.RuneReader // source of chars
}

func newFastCharStream(r io.RuneReader) *FastCharStream {
	return &FastCharStream{input: r}
}

func (cs *FastCharStream) readChar() (rune, error) {
	panic("not implemented yet")
}

func (cs *FastCharStream) beginToken() (rune, error) {
	panic("not implemented yet")
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
