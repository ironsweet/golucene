package index

// index/TermsHashConsumer.java

type TermsHashConsumer interface {
	abort()
}

// index/TermVectorsConsumer.java

type TermVectorsConsumer struct {
	writer    TermVectorsWriter
	docWriter *DocumentsWriterPerThread
	docState  *docState

	hasVectors       bool
	numVectorsFields int
	lastDocId        int
	perFields        []*TermVectorsConsumerPerField
}

func newTermVectorsConsumer(docWriter *DocumentsWriterPerThread) *TermVectorsConsumer {
	return &TermVectorsConsumer{
		docWriter: docWriter,
		docState:  docWriter.docState,
	}
}

func (tvc *TermVectorsConsumer) abort() {
	tvc.hasVectors = false

	if tvc.writer != nil {
		tvc.writer.abort()
		tvc.writer = nil
	}

	tvc.lastDocId = 0
	tvc.reset()
}

func (tvc *TermVectorsConsumer) reset() {
	tvc.perFields = nil
	tvc.numVectorsFields = 0
}

// index/FreqProxTermsWriter.java

type FreqProxTermsWriter struct {
}

func (w *FreqProxTermsWriter) abort() {}
