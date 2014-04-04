package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
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
	/** Performs any initialization, such as reading and
	 *  verifying the header from the provided terms
	 *  dictionary {@link IndexInput}. */
	Init(termsIn store.IndexInput) error
	/** Return a newly created empty TermState */
	NewTermState() *BlockTermState
	/** Actually decode metadata for next term */
	nextTerm(fieldInfo model.FieldInfo, state *BlockTermState) error
	/** Must fully consume state, since after this call that
	 *  TermState may be reused. */
	docs(fieldInfo model.FieldInfo, state *BlockTermState, skipDocs util.Bits, reuse DocsEnum, flags int) (de DocsEnum, err error)
	// docsAndPositions(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits)
	/** Returns approximate RAM bytes used */
	// RamBytesUsed() int64
	/** Reads data for all terms in the next block; this
	 *  method should merely load the byte[] blob but not
	 *  decode, which is done in {@link #nextTerm}. */
	ReadTermsBlock(termsIn store.IndexInput, fieldInfo model.FieldInfo, termState *BlockTermState) error
}
