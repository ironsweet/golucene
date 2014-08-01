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
	ans := &TermVectorsConsumer{
		docWriter: docWriter,
	}
	ans.TermsHashImpl = newTermsHash(ans, docWriter, false, nil)
	return ans
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

func (c *TermVectorsConsumer) initTermVectorsWriter() error {
	if c.writer == nil {
		panic("not implemented yet")
	}
	return nil
}

type TermVectorsConsumerPerFields []*TermVectorsConsumerPerField

func (a TermVectorsConsumerPerFields) Len() int      { return len(a) }
func (a TermVectorsConsumerPerFields) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TermVectorsConsumerPerFields) Less(i, j int) bool {
	return a[i].fieldInfo.Name < a[j].fieldInfo.Name
}

func (c *TermVectorsConsumer) finishDocument() (err error) {
	c.docWriter.testPoint("TermVectorsTermsWriter.finishDocument start")

	if !c.hasVectors {
		return
	}

	// Fields in term vectors are UTF16 sorted: (?)
	util.IntroSort(TermVectorsConsumerPerFields(c.perFields[:c.numVectorsFields]))

	if err = c.initTermVectorsWriter(); err != nil {
		return
	}

	if err = c.fill(c.docState.docID); err != nil {
		return
	}

	// Append term vectors to the real outputs:
	if err = c.writer.StartDocument(c.numVectorsFields); err != nil {
		return
	}
	for i := 0; i < c.numVectorsFields; i++ {
		if err = c.perFields[i].finishDocument(); err != nil {
			return
		}
	}
	if err = c.writer.FinishDocument(); err != nil {
		return
	}

	assert2(c.lastDocId == c.docState.docID,
		"lastDocID=%v docState.docID=%v",
		c.lastDocId, c.docState.docID)

	c.lastDocId++

	c.TermsHashImpl.reset()
	c.resetFields()
	c.docWriter.testPoint("TermVectorsTermsWriter.finishDocument end")
	return
}

func (tvc *TermVectorsConsumer) abort() {
	tvc.hasVectors = false

	defer func() {
		if tvc.writer != nil {
			tvc.writer.Abort()
			tvc.writer = nil
		}

		tvc.lastDocId = 0
		tvc.reset()
	}()

	tvc.TermsHashImpl.abort()
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
	ans := &FreqProxTermsWriter{}
	ans.TermsHashImpl = newTermsHash(ans, docWriter, true, termVectors)
	return ans
}

func (w *FreqProxTermsWriter) flush(fieldsToFlush map[string]TermsHashPerField,
	state *model.SegmentWriteState) (err error) {

	if err = w.TermsHashImpl.flush(fieldsToFlush, state); err != nil {
		return
	}

	// Gather all FieldData's that have postings, across all ThreadStates
	var allFields []*FreqProxTermsWriterPerField

	for _, f := range fieldsToFlush {
		if perField := f.(*FreqProxTermsWriterPerField); perField.bytesHash.Size() > 0 {
			allFields = append(allFields, perField)
		}
	}

	// Sort by field name
	util.IntroSort(FreqProxTermsWriterPerFields(allFields))

	var consumer FieldsConsumer
	if consumer, err = state.SegmentInfo.Codec().(Codec).PostingsFormat().FieldsConsumer(state); err != nil {
		return
	}

	var success = false
	defer func() {
		if success {
			err = util.Close(consumer)
		} else {
			util.CloseWhileSuppressingError(consumer)
		}
	}()

	var termsHash TermsHash
	// Current writer chain:
	// FieldsConsumer
	// -> IMPL: FormatPostingsTermsDictWriter
	// -> TermsConsumer
	// -> IMPL: FormatPostingsTermsDictWriter.TermsWriter
	// -> DocsConsumer
	// -> IMPL: FormatPostingsDocWriter
	// -> PositionsConsumer
	// -> IMPL: FormatPostingsPositionsWriter

	for _, fieldWriter := range allFields {
		fieldInfo := fieldWriter.fieldInfo

		// If this field has postings then add them to the segment
		if err = fieldWriter.flush(fieldInfo.Name, consumer, state); err != nil {
			return
		}

		assert(termsHash == nil || termsHash == fieldWriter.termsHash)
		termsHash = fieldWriter.termsHash
		fieldWriter.reset()
	}

	if termsHash != nil {
		termsHash.reset()
	}
	success = true
	return nil
}

type FreqProxTermsWriterPerFields []*FreqProxTermsWriterPerField

func (a FreqProxTermsWriterPerFields) Len() int      { return len(a) }
func (a FreqProxTermsWriterPerFields) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a FreqProxTermsWriterPerFields) Less(i, j int) bool {
	return a[i].fieldInfo.Name < a[j].fieldInfo.Name
}

func (w *FreqProxTermsWriter) addField(invertState *FieldInvertState,
	fieldInfo *model.FieldInfo) TermsHashPerField {

	return newFreqProxTermsWriterPerField(invertState, w, fieldInfo,
		w.nextTermsHash.addField(invertState, fieldInfo))
}
