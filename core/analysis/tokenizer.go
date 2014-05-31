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
	input io.ReadCloser
}

/* Constructs a token stream processing the given input. */
func NewTokenizer(input io.ReadCloser) *Tokenizer {
	assert2(input != nil, "input must not be nil")
	return &Tokenizer{
		TokenStreamImpl: NewTokenStream(),
		input:           input,
	}
}

func (t *Tokenizer) Close() error {
	if t.input != nil {
		err := t.input.Close()
		if err != nil {
			return err
		}
		t.input = nil
	}
	return nil
}

/*
Return the corrected offset. If input is a CharFilter subclass, this
method calls CharFilter.correctOffset(), else returns currentOff.
*/
func (t *Tokenizer) CorrectOffset(currentOff int) int {
	assert2(t.input != nil, "this tokenizer is closed")
	if v, ok := t.input.(CharFilterService); ok {
		return v.CorrectOffset(currentOff)
	}
	return currentOff
}

/*
Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in
its tokenStream method) will use this to re-use a previously created
tokenizer.
*/
func (t *Tokenizer) SetReader(input io.ReadCloser) error {
	panic("not implemented yet")
}
