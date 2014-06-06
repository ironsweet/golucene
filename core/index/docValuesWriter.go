package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

type DocValuesWriter interface {
	abort()
	finish(int)
	flush(SegmentWriteState, DocValuesConsumer) error
}

// index/NumericDocValuesWriter.java

/* Buffers up pending long per doc, then flushes when segment flushes. */
type NumericDocValuesWriter struct {
	fieldInfo *model.FieldInfo
}

func newNumericDocValuesWriter(fieldInfo *model.FieldInfo,
	iwBytesUsed util.Counter, trackDocsWithField bool) *NumericDocValuesWriter {
	panic("not implemented yet")
}

func (w *NumericDocValuesWriter) addValue(docId int, value int64) {
	panic("not implemented yet")
}

func (w *NumericDocValuesWriter) finish(numDoc int) {}

func (w *NumericDocValuesWriter) flush(state SegmentWriteState, dvConsumer DocValuesConsumer) error {
	maxDoc := state.segmentInfo.DocCount()
	dvConsumer.AddNumericField(w.fieldInfo, newNumericIterator(maxDoc, w))
	return nil
}

func (w *NumericDocValuesWriter) abort() {}

/* Iterates over the values we have in ram */
type NumericIterator struct {
}

func newNumericIterator(maxDoc int, owner *NumericDocValuesWriter) func() (interface{}, bool) {
	panic("not implemented yet")
}
