package analysis

/*
A TokenFilter is a TokenStream whose input is another TokenStream.

This is an abstract class; subclasses must override IncrementToken().
*/
type TokenFilter struct {
	*TokenStreamImpl
	input TokenStream
}

/* Construct a token stream filtering the given input. */
func NewTokenFilter(input TokenStream) *TokenFilter {
	return &TokenFilter{
		TokenStreamImpl: NewTokenStreamWith(input.Attributes()),
		input:           input,
	}
}

func (f *TokenFilter) End() error {
	return f.input.End()
}

func (f *TokenFilter) Close() error {
	return f.input.Close()
}

func (f *TokenFilter) Reset() error {
	return f.input.Reset()
}
