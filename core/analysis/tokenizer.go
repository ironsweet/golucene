package analysis

import (
	"io"
)

// analysis/Tokenizer.java

/*
A Tokenizer is a TokenStream whose input is a Reader.

This is an abstract class; subclasses must override IncrementToken()

NOTE: Subclasses overriding IncrementToken() must call
Attributes().ClearAttributes() before setting attributes.
*/
type Tokenizer struct {
	*TokenStreamImpl
	// The text source for this Tokenizer
	input io.ReadCloser
	// Pending reader: not actually assigned to input until reset()
	inputPending io.ReadCloser
}

/* Constructs a token stream processing the given input. */
func NewTokenizer(input io.ReadCloser) *Tokenizer {
	assert2(input != nil, "input must not be nil")
	return &Tokenizer{
		TokenStreamImpl: NewTokenStream(),
		inputPending:    input,
		input:           ILLEGAL_STATE_READER,
	}
}

func (t *Tokenizer) Close() error {
	err := t.input.Close()
	if err != nil {
		return err
	}
	t.inputPending = ILLEGAL_STATE_READER
	t.input = ILLEGAL_STATE_READER
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

var ILLEGAL_STATE_READER = new(illegalStateReader)

type illegalStateReader struct{}

func (r *illegalStateReader) Read(p []byte) (int, error) {
	panic("TokenStream contract violation: reset()/close() call missing, " +
		"reset() called multiple times, or subclass does not call super.reset(). " +
		"Please see Javadocs of TokenStream class for more information about the correct consuming workflow.")
}

func (r *illegalStateReader) Close() error { return nil }
