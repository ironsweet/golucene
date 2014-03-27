package index

// index/DocConsumer.java

type DocConsumer interface {
	// processDocuments(fieldInfos *FieldInfosBuilder) error
	// finishDocument() error
	// flush(state *SegmentWriteState) error
	abort()
}

// index/DocFieldProcessor.java

/*
This is a DocConsumer that gathers all fields under the same name,
and calls per-field consumers to process field by field. This class
doesn't do any "real" work of its own: it just forwards the fields to
a DocFieldConsumer.
*/
type DocFieldProcessor struct {
}

func newDocFieldProcessor(docWriter *DocumentsWriterPerThread,
	consumer DocFieldConsumer, storedConsumer StoredFieldsConsumer) *DocFieldProcessor {

	panic("not implemented yet")
}

func (p *DocFieldProcessor) abort() {
	panic("not implemented yet")
}
