package index

type InvertedDocConsumer interface{}

/*
This class implements InvertedDocConsumer, which is passed each token
produced by the analyzer on each field. It stores these tokens in a
hash table, and allocates separate bytes streams per token. Consumers
of this class, e.g., FreqproxTermsWriter and TermVectorsConsumer,
write their own byte streams under each term.
*/
type TermsHash struct {
}

func newTermsHash(docWriter *DocumentsWriterPerThread,
	consumer TermsHashConsumer, tackAllocations bool,
	nextTermsHash *TermsHash) *TermsHash {

	panic("not implemented yet")
}
