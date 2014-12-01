package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
)

type DocValuesWriter interface {
	finish(int)
	flush(*SegmentWriteState, DocValuesConsumer) error
}

// index/NumericDocValuesWriter.java

const MISSING int64 = 0

/* Buffers up pending long per doc, then flushes when segment flushes. */
type NumericDocValuesWriter struct {
	pending       packed.PackedLongValuesBuilder
	iwBytesUsed   util.Counter
	bytesUsed     int64
	docsWithField *util.FixedBitSet
	fieldInfo     *FieldInfo
}

func newNumericDocValuesWriter(fieldInfo *FieldInfo,
	iwBytesUsed util.Counter, trackDocsWithField bool) *NumericDocValuesWriter {
	ans := &NumericDocValuesWriter{
		fieldInfo:   fieldInfo,
		iwBytesUsed: iwBytesUsed,
	}
	if trackDocsWithField {
		ans.docsWithField = util.NewFixedBitSetOf(64)
	}
	ans.pending = packed.DeltaPackedBuilder(packed.PackedInts.COMPACT)
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
	if w.docsWithField != nil {
		w.docsWithField = util.EnsureFixedBitSet(w.docsWithField, docId)
		w.docsWithField.Set(docId)
	}

	w.updateBytesUsed()
}

func (w *NumericDocValuesWriter) docsWithFieldBytesUsed() int64 {
	// size of the []int64 + some overhead
	if w.docsWithField == nil {
		return 0
	}
	return util.SizeOf(w.docsWithField.Bits()) + 64
}

func (w *NumericDocValuesWriter) updateBytesUsed() {
	newBytesUsed := w.pending.RamBytesUsed() + w.docsWithFieldBytesUsed()
	w.iwBytesUsed.AddAndGet(newBytesUsed - w.bytesUsed)
	w.bytesUsed = newBytesUsed
}

func (w *NumericDocValuesWriter) finish(numDoc int) {}

func (w *NumericDocValuesWriter) flush(state *SegmentWriteState,
	dvConsumer DocValuesConsumer) error {

	maxDoc := state.SegmentInfo.DocCount()
	values := w.pending.Build()
	dvConsumer.AddNumericField(w.fieldInfo, func() func() (interface{}, bool) {
		return newNumericIterator(maxDoc, values, w.docsWithField)
	})
	return nil
}

/* Iterates over the values we have in ram */
type NumericIterator struct{}

func newNumericIterator(maxDoc int,
	values packed.PackedLongValues,
	docsWithFields *util.FixedBitSet) func() (interface{}, bool) {

	upto, size := 0, int(values.Size())
	iter := values.Iterator()
	return func() (interface{}, bool) {
		if upto >= maxDoc {
			return nil, false
		}
		var value interface{}
		if upto < size {
			v, _ := iter()
			if docsWithFields == nil || docsWithFields.At(upto) {
				value = v
			} else {
				value = nil
			}
		} else if docsWithFields != nil {
			value = nil
		} else {
			value = MISSING
		}
		upto++
		return value, true
	}
}
