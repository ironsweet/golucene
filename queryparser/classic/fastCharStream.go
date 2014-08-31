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
