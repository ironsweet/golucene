package spi

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"io"
)

// codecs/TermVectorsReader.java

type TermVectorsReader interface {
	io.Closer
	Get(doc int) model.Fields
	Clone() TermVectorsReader
}

// codecs/TermVectorsWriter.java

/*
Codec API for writing term vecrors:

1. For every document, StartDocument() is called, informing the Codec
how may fields will be written.
2. StartField() is called for each field in the document, informing
the codec how many terms will be written for that field, and whether
or not positions, offsets, or payloads are enabled.
3. Within each field, StartTerm() is called for each term.
4. If offsets and/or positions are enabled, then AddPosition() will
be called for each term occurrence.
5. After all documents have been written, Finish() is called for
verification/sanity-checks.
6. Finally the writer is closed.
*/
type TermVectorsWriter interface {
	io.Closer
	// Called before writing the term vectors of the document.
	// startField() will be called numVectorsFields times. Note that if
	// term vectors are enabled, this is called even if the document
	// has no vector fields, in this case numVectorFields will be zero.
	StartDocument(int) error
	// Called after a doc and all its fields have been added
	FinishDocument() error
	// Aborts writing entirely, implementation should remove any
	// partially-written files, etc.
	Abort()
	// Called before Close(), passing in the number of documents that
	// were written. Note that this is intentionally redendant
	// (equivalent to the number of calls to startDocument(int)), but a
	// Codec should check that this is the case to detect the JRE bug
	// described in LUCENE-1282.
	Finish(model.FieldInfos, int) error
}
