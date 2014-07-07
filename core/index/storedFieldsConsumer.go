package index

// import (
// 	"github.com/balzaczyy/golucene/core/index/model"
// 	"github.com/balzaczyy/golucene/core/store"
// 	"github.com/balzaczyy/golucene/core/util"
// 	"sync"
// )

// type StoredFieldsConsumer interface {
// 	addField(docId int, field model.IndexableField, fieldInfo *model.FieldInfo)
// 	flush(state *model.SegmentWriteState) error
// 	abort()
// 	startDocument()
// 	finishDocument() error
// }

// /* Just switches between two DocFieldConsumers */
// type TwoStoredFieldsConsumers struct {
// 	first  StoredFieldsConsumer
// 	second StoredFieldsConsumer
// }

// func newTwoStoredFieldsConsumers(first, second StoredFieldsConsumer) *TwoStoredFieldsConsumers {
// 	return &TwoStoredFieldsConsumers{first, second}
// }

// func (p *TwoStoredFieldsConsumers) addField(docId int, field model.IndexableField, fieldInfo *model.FieldInfo) {
// 	// err := p.first.addField(docId, field, fieldInfo)
// 	// if err == nil {
// 	// 	err = p.second.addField(docId, field, fieldInfo)
// 	// }
// 	// return err
// 	p.first.addField(docId, field, fieldInfo)
// 	p.second.addField(docId, field, fieldInfo)
// }

// func (p *TwoStoredFieldsConsumers) flush(state *model.SegmentWriteState) error {
// 	err := p.first.flush(state)
// 	if err == nil {
// 		err = p.second.flush(state)
// 	}
// 	return err
// }

// func (p *TwoStoredFieldsConsumers) abort() {
// 	p.first.abort()
// 	p.second.abort()
// }

// func (p *TwoStoredFieldsConsumers) startDocument() {
// 	p.first.startDocument()
// 	p.second.startDocument()
// }

// func (p *TwoStoredFieldsConsumers) finishDocument() error {
// 	err := p.first.finishDocument()
// 	if err == nil {
// 		err = p.second.finishDocument()
// 	}
// 	return err
// }

// // index/StoredFieldsProcessor.java

// /* This is a StoredFieldsConsumer that writes stored fields */
// type StoredFieldsProcessor struct {
// 	sync.Locker

// 	fieldsWriter StoredFieldsWriter
// 	lastDocId    int

// 	docWriter *DocumentsWriterPerThread

// 	docState *docState
// 	codec    Codec

// 	numStoredFields int
// 	storedFields    []model.IndexableField
// 	fieldInfos      []*model.FieldInfo
// }

// func newStoredFieldsProcessor(docWriter *DocumentsWriterPerThread) *StoredFieldsProcessor {
// 	return &StoredFieldsProcessor{
// 		Locker:    &sync.Mutex{},
// 		docWriter: docWriter,
// 		docState:  docWriter.docState,
// 		codec:     docWriter.codec,
// 	}
// }

// func (p *StoredFieldsProcessor) reset() {
// 	p.numStoredFields = 0
// 	p.storedFields = nil
// 	p.fieldInfos = nil
// }

// func (p *StoredFieldsProcessor) startDocument() {
// 	p.reset()
// }

// func (p *StoredFieldsProcessor) flush(state *model.SegmentWriteState) (err error) {
// 	numDocs := state.SegmentInfo.DocCount()
// 	if numDocs > 0 {
// 		// It's possible that all documents seen in this segment hit
// 		// non-aborting errors, in which case we will not have yet init'd
// 		// the FieldsWriter:
// 		err = p.initFieldsWriter(state.Context)
// 		if err == nil {
// 			err = p.fill(numDocs)
// 		}
// 	}
// 	if w := p.fieldsWriter; w != nil {
// 		err = func() error {
// 			var success = false
// 			defer func() {
// 				if success {
// 					err = util.CloseWhileHandlingError(err, w)
// 				} else {
// 					util.CloseWhileSuppressingError(w)
// 				}
// 			}()

// 			err = w.Finish(state.FieldInfos, numDocs)
// 			if err != nil {
// 				return err
// 			}
// 			success = true
// 			return nil
// 		}()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return
// }

// func (p *StoredFieldsProcessor) initFieldsWriter(ctx store.IOContext) error {
// 	p.Lock()
// 	defer p.Unlock()
// 	if p.fieldsWriter == nil {
// 		var err error
// 		p.fieldsWriter, err = p.codec.StoredFieldsFormat().FieldsWriter(
// 			p.docWriter.directory, p.docWriter.segmentInfo, ctx)
// 		if err != nil {
// 			return err
// 		}
// 		p.lastDocId = 0
// 	}
// 	return nil
// }

// func (p *StoredFieldsProcessor) abort() {
// 	p.reset()

// 	if w := p.fieldsWriter; w != nil {
// 		assert(w != nil)
// 		w.Abort()
// 		p.fieldsWriter = nil
// 		p.lastDocId = 0
// 	}
// }

// /* Fills in any hold in the docIDs */
// func (p *StoredFieldsProcessor) fill(docId int) error {
// 	// We must "catch up" for all docs before us that had no stored fields:
// 	for p.lastDocId < docId {
// 		err := p.fieldsWriter.StartDocument(0)
// 		if err != nil {
// 			return err
// 		}
// 		p.lastDocId++
// 		err = p.fieldsWriter.FinishDocument()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (p *StoredFieldsProcessor) finishDocument() error {
// 	p.docWriter.testPoint("StoredFieldsWriter.finishDocument start")

// 	err := p.initFieldsWriter(store.IO_CONTEXT_DEFAULT)
// 	if err != nil {
// 		return err
// 	}
// 	err = p.fill(p.docState.docID)
// 	if err != nil {
// 		return err
// 	}

// 	if w := p.fieldsWriter; w != nil && p.numStoredFields > 0 {
// 		err = w.StartDocument(p.numStoredFields)
// 		if err != nil {
// 			return err
// 		}
// 		for i := 0; i < p.numStoredFields; i++ {
// 			err = w.WriteField(p.fieldInfos[i], p.storedFields[i])
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		err = w.FinishDocument()
// 		if err != nil {
// 			return err
// 		}
// 		p.lastDocId++
// 	}

// 	p.reset()
// 	p.docWriter.testPoint("StoredFieldsWriter.finishDocument end")
// 	return nil
// }

// func (p *StoredFieldsProcessor) addField(docId int, field model.IndexableField, fieldInfo *model.FieldInfo) {
// 	if field.FieldType().Stored() {
// 		if p.numStoredFields == len(p.storedFields) {
// 			newSize := util.Oversize(p.numStoredFields+1, util.NUM_BYTES_OBJECT_REF)
// 			newArray := make([]model.IndexableField, newSize)
// 			copy(newArray, p.storedFields)
// 			p.storedFields = newArray

// 			newInfoArray := make([]*model.FieldInfo, newSize)
// 			copy(newInfoArray, p.fieldInfos)
// 			p.fieldInfos = newInfoArray
// 		}

// 		p.storedFields[p.numStoredFields] = field
// 		p.fieldInfos[p.numStoredFields] = fieldInfo
// 		p.numStoredFields++

// 		p.docState.testPoint("StoredFieldsWriterPerThread.processFields.writeField")
// 	}
// }

// index/DocValuesProcessor.java

// type DocValuesProcessor struct {
// 	writers   map[string]DocValuesWriter
// 	bytesUsed util.Counter
// }

// func newDocValuesProcessor(bytesUsed util.Counter) *DocValuesProcessor {
// 	return &DocValuesProcessor{make(map[string]DocValuesWriter), bytesUsed}
// }

// func (p *DocValuesProcessor) startDocument() {}

// func (p *DocValuesProcessor) finishDocument() error { return nil }

// func (p *DocValuesProcessor) addField(docId int, field model.IndexableField, fieldInfo *model.FieldInfo) {
// 	if dvType := field.FieldType().DocValueType(); int(dvType) != 0 {
// 		fieldInfo.SetDocValueType(dvType)
// 		switch dvType {
// 		case model.DOC_VALUES_TYPE_BINARY:
// 			p.addBinaryField(fieldInfo, docId, field.BinaryValue())
// 		case model.DOC_VALUES_TYPE_SORTED:
// 			p.addSortedField(fieldInfo, docId, field.BinaryValue())
// 		case model.DOC_VALUES_TYPE_SORTED_SET:
// 			p.addSortedSetField(fieldInfo, docId, field.BinaryValue())
// 		case model.DOC_VALUES_TYPE_NUMERIC:
// 			p.addNumericField(fieldInfo, docId, field.NumericValue())
// 		default:
// 			assertn(false, "unrecognized DocValues.Type: %v", dvType)
// 		}
// 	}
// }

// func (p *DocValuesProcessor) flush(state *model.SegmentWriteState) (err error) {
// 	if len(p.writers) != 0 {
// 		codec := state.SegmentInfo.Codec().(Codec)
// 		var dvConsumer DocValuesConsumer
// 		dvConsumer, err = codec.DocValuesFormat().FieldsConsumer(state)
// 		if err != nil {
// 			return err
// 		}
// 		var success = false
// 		defer func() {
// 			if success {
// 				err = mergeError(err, util.Close(dvConsumer))
// 			} else {
// 				util.CloseWhileSuppressingError(dvConsumer)
// 			}
// 		}()

// 		for _, writer := range p.writers {
// 			writer.finish(state.SegmentInfo.DocCount())
// 			err = writer.flush(state, dvConsumer)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		// TODO: catch missing DV dields here? else we have nil/""
// 		// depending on how docs landed in segments? but we can't detect
// 		// all cases, and we should leave this behavior undefined. dv is
// 		// not "schemaless": it's column-stride.
// 		p.writers = make(map[string]DocValuesWriter)
// 		success = true
// 	}
// 	return nil
// }

// func (p *DocValuesProcessor) addBinaryField(fieldInfo *model.FieldInfo,
// 	docId int, value []byte) {
// 	panic("not implemented yet")
// }

// func (p *DocValuesProcessor) addSortedField(fieldInfo *model.FieldInfo,
// 	docId int, value []byte) {
// 	panic("not implemented yet")
// }

// func (p *DocValuesProcessor) addSortedSetField(fieldInfo *model.FieldInfo,
// 	docId int, value []byte) {
// 	panic("not implemented yet")
// }

// func (p *DocValuesProcessor) addNumericField(field *model.FieldInfo,
// 	docId int, value interface{}) {
// 	panic("not implemented yet")
// }

// func (p *DocValuesProcessor) abort() {
// 	for _, writer := range p.writers {
// 		writer.abort()
// 	}
// 	p.writers = make(map[string]DocValuesWriter)
// }
