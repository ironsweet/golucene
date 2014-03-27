package index

// index/TermsHashConsumer.java

type TermsHashConsumer interface {
}

// index/TermVectorsConsumer.java

type TermVectorsConsumer struct {
	docWriter *DocumentsWriterPerThread
	docState  *docState
}

func newTermVectorsConsumer(docWriter *DocumentsWriterPerThread) *TermVectorsConsumer {
	return &TermVectorsConsumer{
		docWriter: docWriter,
		docState:  docWriter.docState,
	}
}

// index/FreqProxTermsWriter.java

type FreqProxTermsWriter struct {
}
