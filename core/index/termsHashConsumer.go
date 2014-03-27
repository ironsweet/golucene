package index

// index/TermsHashConsumer.java

type TermsHashConsumer interface {
}

// index/TermVectorsConsumer.java

type TermVectorsConsumer struct {
}

func newTermVectorsConsumer(docWriter *DocumentsWriterPerThread) *TermVectorsConsumer {
	panic("not implemented yet")
}

// index/FreqProxTermsWriter.java

type FreqProxTermsWriter struct {
}
