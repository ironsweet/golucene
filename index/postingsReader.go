package index

import (
	"github.com/balzaczyy/golucene/store"
	"io"
)

// PostingReaderBase.java

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
	/**		Performs any initialization, such as reading and
	* verifying the header from the provided terms
	* dictionary IndexInput.	*/
	Init(termsIn store.IndexInput) error
	// Return a newly created empty BlockTermState
	NewTermState() *BlockTermState
	// nextTerm(fieldInfo FieldInfo, state BlockTermState)
	// docs(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits, reuse DocsEnum, flags int)
	// docsAndPositions(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits)
	/** Returns approximate RAM bytes used */
	// RamBytesUsed() int64
	/** Reads data for all terms in the next block; this
	 *  method should merely load the byte[] blob but not
	 *  decode, which is done in {@link #nextTerm}. */
	ReadTermsBlock(termsIn store.IndexInput, fieldInfo FieldInfo, termState *BlockTermState) error
}
