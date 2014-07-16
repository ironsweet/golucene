package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/index/model"
	// "github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// index/TermsHashConsumer.java

// type TermsHashConsumer interface {
// 	flush(map[string]TermsHashConsumerPerField, *model.SegmentWriteState) error
// 	abort()
// 	startDocument()
// 	finishDocument(*TermsHash) error
// 	addField(*TermsHashPerField, *model.FieldInfo) TermsHashConsumerPerField
// }

// index/TermVectorsConsumer.java

type TermVectorsConsumer struct {
	*TermsHashImpl

	writer TermVectorsWriter

	docWriter *DocumentsWriterPerThread

	hasVectors       bool
	numVectorsFields int
	lastDocId        int
	perFields        []*TermVectorsConsumerPerField
}

func newTermVectorsConsumer(docWriter *DocumentsWriterPerThread) *TermVectorsConsumer {
	return &TermVectorsConsumer{
		TermsHashImpl: newTermsHash(docWriter, false, nil),
		docWriter:     docWriter,
	}
}

func (tvc *TermVectorsConsumer) flush(fieldsToFlush map[string]TermsHashPerField,
	state *model.SegmentWriteState) (err error) {
	if tvc.writer != nil {
		numDocs := state.SegmentInfo.DocCount()
		assert(numDocs > 0)
		// At least one doc in this run had term vectors enabled
		func() {
			defer func() {
				err = mergeError(err, util.Close(tvc.writer))
				tvc.writer = nil
				tvc.lastDocId = 0
				tvc.hasVectors = false
			}()

			err = tvc.fill(numDocs)
			if err == nil {
				err = tvc.writer.Finish(state.FieldInfos, numDocs)
			}
		}()
		if err != nil {
			return err
		}
	}

	return
}

/*
Fills in no-term-vectors for all docs we haven't seen since the last
doc that had term vectors.
*/
func (c *TermVectorsConsumer) fill(docId int) error {
	for c.lastDocId < docId {
		c.writer.StartDocument(0)
		err := c.writer.FinishDocument()
		if err != nil {
			return err
		}
		c.lastDocId++
	}
	return nil
}

func (c *TermVectorsConsumer) finishDocument() error {
	panic("not implemented yet")
	// c.docWriter.testPoint("TermVectorsTermsWriter.finishDocument start")

	// if !c.hasVectors {
	// 	return nil
	// }

	// var err error
	// // initTermVectorsWriter
	// if c.writer == nil {
	// 	w := c.docWriter
	// 	ctx := store.NewIOContextForFlush(&store.FlushInfo{
	// 		w.numDocsInRAM,
	// 		w.bytesUsed(),
	// 	})
	// 	c.writer, err = w.codec.TermVectorsFormat().VectorsWriter(w.directory, w.segmentInfo, ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.lastDocId = 0
	// }

	// err = c.fill(c.docState.docID)
	// if err != nil {
	// 	return err
	// }

	// // Append term vectors to the real outputs:
	// err = c.writer.startDocument(c.numVectorsFields)
	// if err != nil {
	// 	return err
	// }
	// for i := 0; i < c.numVectorsFields; i++ {
	// 	err = c.perFields[i].finishDocument()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// err = c.writer.finishDocument()
	// if err != nil {
	// 	return err
	// }

	// assertn(c.lastDocId == c.docState.docID, "lastDocID=%v docState.docID=%v", c.lastDocId, c.docState.docID)

	// c.lastDocId++

	// termsHash.reset()
	// c.reset()
	// c.docWriter.testPoint("TermVectorsTermsWriter.finishDocument end")
	// return nil
}

func (tvc *TermVectorsConsumer) abort() {
	panic("not implemented yet")
	// tvc.hasVectors = false

	// if tvc.writer != nil {
	// 	tvc.writer.abort()
	// 	tvc.writer = nil
	// }

	// tvc.lastDocId = 0
	// tvc.reset()
}

func (tvc *TermVectorsConsumer) resetFields() {
	tvc.perFields = nil
	tvc.numVectorsFields = 0
}

func (tvc *TermVectorsConsumer) addField(invertState *FieldInvertState,
	fieldInfo *model.FieldInfo) TermsHashPerField {
	return newTermVectorsConsumerPerField(invertState, tvc, fieldInfo)
}

func (c *TermVectorsConsumer) addFieldToFlush(fieldToFlush *TermVectorsConsumerPerField) {
	panic("not implemented yet")
}

func (c *TermVectorsConsumer) startDocument() {
	c.resetFields()
	c.numVectorsFields = 0
}

// func (c *TermVectorsConsumer) clearLastVectorFieldName() bool {
// 	c.lastVectorFieldName = ""
// 	return true
// }

// index/FreqProxTermsWriter.java

type FreqProxTermsWriter struct {
	*TermsHashImpl
}

func newFreqProxTermsWriter(docWriter *DocumentsWriterPerThread, termVectors TermsHash) *FreqProxTermsWriter {
	return &FreqProxTermsWriter{
		newTermsHash(docWriter, true, termVectors),
	}
}

func (w *FreqProxTermsWriter) flush(fieldsToFlush map[string]TermsHashPerField,
	state *model.SegmentWriteState) (err error) {
	panic("not implemented yet")
	// // Gather all FieldData's that have postings, across all ThreadStates
	// var allFields []*FreqProxTermsWriterPerField

	// for _, f := range fieldsToFlush {
	// 	perField := f.(*FreqProxTermsWriterPerField)
	// 	if perField.termsHashPerField.bytesHash.Size() > 0 {
	// 		allFields = append(allFields, perField)
	// 	}
	// }

	// // Sort by field name
	// util.IntroSort(FreqProxTermsWriterPerFields(allFields))

	// var consumer FieldsConsumer
	// consumer, err = state.SegmentInfo.Codec().(Codec).PostingsFormat().FieldsConsumer(state)
	// if err != nil {
	// 	return err
	// }

	// var success = false
	// defer func() {
	// 	if success {
	// 		err = mergeError(err, util.Close(consumer))
	// 	} else {
	// 		util.CloseWhileSuppressingError(consumer)
	// 	}
	// }()

	// var termsHash *TermsHash
	// // Current writer chain:
	// // FieldsConsumer
	// // -> IMPL: FormatPostingsTermsDictWriter
	// // -> TermsConsumer
	// // -> IMPL: FormatPostingsTermsDictWriter.TermsWriter
	// // -> DocsConsumer
	// // -> IMPL: FormatPostingsDocWriter
	// // -> PositionsConsumer
	// // -> IMPL: FormatPostingsPositionsWriter

	// for _, fieldWriter := range allFields {
	// 	fieldInfo := fieldWriter.fieldInfo

	// 	// If this field has postings then add them to the segment
	// 	err = fieldWriter.flush(fieldInfo.Name, consumer, state)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	perField := fieldWriter.termsHashPerField
	// 	assert(termsHash == nil || termsHash == perField.termsHash)
	// 	termsHash = perField.termsHash
	// 	numPostings := perField.bytesHash.Size()
	// 	perField.reset()
	// 	perField.shrinkHash(numPostings)
	// 	fieldWriter.reset()
	// }

	// if termsHash != nil {
	// 	termsHash.reset()
	// }
	// success = true
	// return
}

type FreqProxTermsWriterPerFields []*FreqProxTermsWriterPerField

func (a FreqProxTermsWriterPerFields) Len() int      { return len(a) }
func (a FreqProxTermsWriterPerFields) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a FreqProxTermsWriterPerFields) Less(i, j int) bool {
	return a[i].fieldInfo.Name < a[j].fieldInfo.Name
}

func (w *FreqProxTermsWriter) addField(invertState *FieldInvertState,
	fieldInfo *model.FieldInfo) TermsHashPerField {
	panic("not implemented yet")
	// return newFreqProxTermsWriterPerField(invertState, w, fieldInfo, w.nex)
}
