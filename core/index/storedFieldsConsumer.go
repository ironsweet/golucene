package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

type StoredFieldsConsumer interface {
}

/* Just switches between two DocFieldConsumers */
type TwoStoredFieldsConsumers struct {
}

func newTwoStoredFieldsConsumers(first, second StoredFieldsConsumer) *TwoStoredFieldsConsumers {
	panic("Not implemented yet")
}

// index/StoredFieldsProcessor.java

/* This is a StoredFieldsConsumer that writes stored fields */
type StoredFieldsProcessor struct {
}

func newStoredFieldsProcessor(docWriter *DocumentsWriterPerThread) *StoredFieldsProcessor {
	panic("not implemented yet")
}

// index/DocValuesProcessor.java

type DocValuesProcessor struct {
}

func newDocValuesProcessor(bytesUsed util.Counter) *DocValuesProcessor {
	panic("not implemented yet")
}
