package index

import (
	"github.com/balzaczyy/golucene/store"
	"io"
)

// PostingReaderBase.javaz

/*
The core terms dictionaries (BlockTermsReader,
BlockTreeTermsReader) interacts with a single instnce
of this class to manage creation of DocsEnum and
DocsAndPositionsEnum instances. It provides an
IndexInput (termsIn) where this class may read any
previously stored data that it had written in its
corresponding PostingsWrierBase at indexing
time.
*/
type PostingsReaderBase interface {
	io.Closer
	/*
		Performs any initialization, such as reading and
		verifying the header from the provided terms
		dictionary IndexInput.
	*/
	Init(termsIn store.IndexInput) error
	// Return a newly created empty BlockTermState
	NewTermState() TermState
	// nextTerm(fieldInfo FieldInfo, state BlockTermState)
	// docs(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits, reuse DocsEnum, flags int)
	// docsAndPositions(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits)
}
