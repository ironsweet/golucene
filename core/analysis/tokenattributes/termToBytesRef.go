package tokenattributes

/*
This attribute is requested by TermsHashPerField to index the
contents. This attribute can be used to customize the final []byte
encoding of terms.

Consumers of this attribute call BytesRef() up-front, and then invoke
FillBytesRef() for each term. Examle:

	termAtt := tokenStream.Get("TermToBytesRefAttribute")
	bytes := termAtt.BytesRef();

	var err error
	var ok bool
	for ok, err = termAtt.IncrementToken(); ok && err == nil; ok, err = termAtt.IncrementToken() {

		// you must call termAtt.FillBytesRef() before doing something with the bytes.
		// this encodes the term value (internally it might be a []rune, etc) into the bytes.
		hashCode := termAtt.FillBytesRef()

		if isIntersting(bytes) {

			// becaues the bytes are reused by the attribute (like CharTermAttribute's []rune buffer),
			// you should make a copy if you need persistent access to the bytes, otherwise they will
			// be rewritten across calls to IncrementToken()

			clone := make([]byte, len(bytes))
			copy(clone, bytes)
			doSomethingWith(cone)
		}
	}
	...
*/
type TermToBytesRefAttribute interface {
	// Updates the bytes BytesRef() to contain this term's final
	// encoding, and returns its hashcode.
	FillBytesRef() int
	// Retrieve this attribute's BytesRef. The bytes are updated from
	// the current term when the consumer calls FillBytesRef().
	BytesRef() []byte
}
