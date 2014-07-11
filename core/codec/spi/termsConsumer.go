package spi

import (
	. "github.com/balzaczyy/golucene/core/codec"
)

/*
Abstract API that consumes terms for an individual field.

The lifecycle is:
	- TermsConsumer is returned for each field by FieldsConsumer.addField().
	- TermsConsumer returns a PostingsCOnsumer for each term in startTerm().
	- When the producer (e.g. IndexWriter) is done adding documents for
	the term, it calls finishTerm(), passing in the accumulated term statistics.
	- Producer calls finish() with the accumulated collection
	statistics when it is finished adding terms to the field.
*/
type TermsConsumer interface {
	// Starts a ew term in this field; this may be called with no
	// corresponding call to finish if the term had no docs.
	StartTerm([]byte) (PostingsConsumer, error)
	// Finishes the current term; numDocs must be > 0.
	// stats.totalTermFreq will be -1 when term frequencies are omitted
	// for the field.
	FinishTerm([]byte, *TermStats) error
	// Called when we are done adding terms to this field.
	// sumTotalTermFreq will be -1 when term frequencies are omitted
	// for the field.
	Finish(sumTotalTermFreq, sumDocFreq int64, docCount int) error
	// Return the BytesRef comparator used to sort terms before feeding
	// to this API.
	Comparator() func(a, b []byte) bool
}
