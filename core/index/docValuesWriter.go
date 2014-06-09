package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
)

type DocValuesWriter interface {
	abort()
	finish(int)
	flush(SegmentWriteState, DocValuesConsumer) error
}

// index/NumericDocValuesWriter.java

const MISSING int64 = 0

/* Buffers up pending long per doc, then flushes when segment flushes. */
type NumericDocValuesWriter struct {
	pending            *packed.AppendingDeltaPackedLongBuffer
	iwBytesUsed        util.Counter
	bytesUsed          int64
	docsWithField      *util.OpenBitSet
	fieldInfo          *model.FieldInfo
	trackDocsWithField bool
}

func newNumericDocValuesWriter(fieldInfo *model.FieldInfo,
	iwBytesUsed util.Counter, trackDocsWithField bool) *NumericDocValuesWriter {
	ans := &NumericDocValuesWriter{
		docsWithField:      util.NewOpenBitSet(),
		fieldInfo:          fieldInfo,
		iwBytesUsed:        iwBytesUsed,
		trackDocsWithField: trackDocsWithField,
	}
	ans.pending = packed.NewAppendingDeltaPackedLongBufferWithOverhead(packed.PackedInts.COMPACT)
	ans.bytesUsed = ans.pending.RamBytesUsed() + ans.docsWithFieldBytesUsed()
	ans.iwBytesUsed.AddAndGet(ans.bytesUsed)
	return ans
}

func (w *NumericDocValuesWriter) addValue(docId int, value int64) {
	assert2(int64(docId) >= w.pending.Size(),
		"DocValuesField '%v' appears more than once in this document (only one value is allowed per field)",
		w.fieldInfo.Name)

	// Fill in any holes
	for i := int(w.pending.Size()); i < docId; i++ {
		w.pending.Add(MISSING)
	}

	w.pending.Add(value)
	if w.trackDocsWithField {
		w.docsWithField.Set(int64(docId))
	}

	w.updateBytesUsed()
}

func (w *NumericDocValuesWriter) docsWithFieldBytesUsed() int64 {
	// size of the []int64 + some overhead
	return util.SizeOf(w.docsWithField.RealBits()) + 64
}

func (w *NumericDocValuesWriter) updateBytesUsed() {
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
