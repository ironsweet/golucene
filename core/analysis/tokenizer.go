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
	Input io.RuneReader
	// Pending reader: not actually assigned to input until reset()
	inputPending io.RuneReader
}

/* Constructs a token stream processing the given input. */
func NewTokenizer(input io.RuneReader) *Tokenizer {
	assert2(input != nil, "input must not be nil")
	return &Tokenizer{
		TokenStreamImpl: NewTokenStream(),
		inputPending:    input,
		Input:           ILLEGAL_STATE_READER,
	}
}

func (t *Tokenizer) Close() error {
	if v, ok := t.Input.(io.Closer); ok {
		err := v.Close()
		if err != nil {
			return err
		}
	} // optional close
	t.inputPending = ILLEGAL_STATE_READER
	t.Input = ILLEGAL_STATE_READER
	return nil
}

/*
Return the corrected offset. If input is a CharFilter subclass, this
method calls CharFilter.correctOffset(), else returns currentOff.
*/
func (t *Tokenizer) CorrectOffset(currentOff int) int {
	assert2(t.Input != nil, "this tokenizer is closed")
	if v, ok := t.Input.(CharFilterService); ok {
		return v.CorrectOffset(currentOff)
	}
	return currentOff
}

/*
Expert: Set a new reader on the Tokenizer. Typically, an analyzer (in
its tokenStream method) will use this to re-use a previously created
tokenizer.
*/
func (t *Tokenizer) SetReader(input io.RuneReader) error {
	assert2(input != nil, "input must not be nil")
	assert2(t.Input == ILLEGAL_STATE_READER, "TokenStream contract violation: close() call missing")
	t.inputPending = input
	return nil
}

func (t *Tokenizer) Reset() error {
	t.Input = t.inputPending
	t.inputPending = ILLEGAL_STATE_READER
	return nil
}

var ILLEGAL_STATE_READER = new(illegalStateReader)

type illegalStateReader struct{}

func (r *illegalStateReader) ReadRune() (rune, int, error) {
	panic("TokenStream contract violation: reset()/close() call missing, " +
		"reset() called multiple times, or subclass does not call super.reset(). " +
		"Please see Javadocs of TokenStream class for more information about the correct consuming workflow.")
}

func (r *illegalStateReader) String() string { return "ILLEGAL_STATE_READRE" }
