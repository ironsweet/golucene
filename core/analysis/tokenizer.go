package analysis

import (
	"io"
)

/*
A Tokenizer is a TokenStream whose input is a Reader.

This is an abstract class; subclasses must override IncrementToken()

NOTE: Subclasses overriding IncrementToken() must call
Attributes().ClearAttributes() before setting attributes.
*/
type Tokenizer struct {
	*TokenStreamImpl
}

/* Constructs a token stream processing the given input. */
func NewTokenizer(input io.ReadCloser) *Tokenizer {
	assert2(input != nil, "inpu tmust not be nil")
	return &Tokenizer{
		TokenStreamImpl: NewTokenStream(),
	}
}

/*
Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in
its tokenStream method) will use this to re-use a previously created
tokenizer.
*/
func (t *Tokenizer) SetReader(input io.ReadCloser) error {
	panic("not implemented yet")
}
