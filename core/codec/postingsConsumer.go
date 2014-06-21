package codec

/*
Abstract API that consumes postings for an individual term.

The lifecycle is:
	- PostingsConsumer is returned for each term by
	TermsConsumer.startTerm().
	- startDoc() is called for each document where the term occurs,
	specifying id and term frequency for that document.
	- If positions are enabled for the field, then addPosition() will
	be called for each occurrence in the document.
	- finishDoc() is called when the producer is done adding positions
	to the document.
*/
type PostingsConsumer interface {
	// Adds a new doc in this term. freq will be -1 when term
	// frequencies are omitted for the field.
	StartDoc(docId, freq int) error
	// Add a new position & payload, and start/end offset. A nil
	// payload means no payload; a non-nil payload with zero length
	// also means no payload. Caller may reuse the BytesRef for the
	// payload between calls (method must fully consume the payload).
	// startOffset and endOffset will be -1 when offsets are not indexed.
	AddPosition(position int, payload []byte, startOffset, endOffset int) error
	// Called when we are done adding positions & payloads for each doc.
	FinishDoc() error
}
