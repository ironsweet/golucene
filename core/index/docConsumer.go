package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	// "github.com/balzaczyy/golucene/core/store"
	// "github.com/balzaczyy/golucene/core/util"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocument() error
	flush(state *model.SegmentWriteState) error
	abort()
}

// // index/DocFieldProcessor.java

// /*
// This is a DocConsumer that gathers all fields under the same name,
// and calls per-field consumers to process field by field. This class
// doesn't do any "real" work of its own: it just forwards the fields to
// a DocFieldConsumer.
// */
// type DocFieldProcessor struct {
// 	consumer       DocFieldConsumer
// 	storedConsumer StoredFieldsConsumer
// 	codec          Codec

// 	// Holds all fields seen in current doc
// 	_fields    []*DocFieldProcessorPerField
// 	fieldCount int

// 	// Hash table for all fields ever seen
// 	fieldHash       []*DocFieldProcessorPerField
// 	hashMask        int
// 	totalFieldCount int

// 	fieldGen int

// 	docState *docState

// 	bytesUsed util.Counter
// }

// func newDocFieldProcessor(docWriter *DocumentsWriterPerThread,
// 	consumer DocFieldConsumer, storedConsumer StoredFieldsConsumer) *DocFieldProcessor {

// 	assert(storedConsumer != nil)
// 	return &DocFieldProcessor{
// 		_fields:        make([]*DocFieldProcessorPerField, 1),
// 		fieldHash:      make([]*DocFieldProcessorPerField, 2),
// 		hashMask:       1,
// 		docState:       docWriter.docState,
// 		codec:          docWriter.codec,
// 		bytesUsed:      docWriter._bytesUsed,
// 		consumer:       consumer,
// 		storedConsumer: storedConsumer,
// 	}
// }

// func (p *DocFieldProcessor) flush(state *model.SegmentWriteState) error {
// 	childFields := make(map[string]DocFieldConsumerPerField)
// 	for _, f := range p.fields() {
// 		childFields[f.fieldInfo().Name] = f
// 	}

// 	err := p.storedConsumer.flush(state)
// 	if err != nil {
// 		return err
// 	}
// 	err = p.consumer.flush(childFields, state)
// 	if err != nil {
// 		return err
// 	}

// 	// Impotant to save after asking consumer to flush so consumer can
// 	// alter the FieldInfo if necessary. E.g., FreqProxTermsWriter does
// 	// this with FieldInfo.storePayload.
// 	infosWriter := p.codec.FieldInfosFormat().FieldInfosWriter()
// 	assert(infosWriter != nil)
// 	return infosWriter(state.Directory, state.SegmentInfo.Name,
// 		state.FieldInfos, store.IO_CONTEXT_DEFAULT)
// }

// func (p *DocFieldProcessor) abort() {
// 	for _, field := range p.fieldHash {
// 		for field != nil {
// 			next := field.next
// 			field.abort()
// 			field = next
// 		}
// 	}
// 	p.storedConsumer.abort()
// 	p.consumer.abort()
// 	// assert2(err == nil, err.Error())
// }

// func (p *DocFieldProcessor) fields() []DocFieldConsumerPerField {
// 	var fields []DocFieldConsumerPerField
// 	for _, field := range p.fieldHash {
// 		for field != nil {
// 			fields = append(fields, field.consumer)
// 			field = field.next
// 		}
// 	}
// 	assert(len(fields) == p.totalFieldCount)
// 	return fields
// }

// func (p *DocFieldProcessor) rehash() {
// 	newHashSize := len(p.fieldHash) * 2
// 	assert(newHashSize > len(p.fieldHash)) // avoid overflow

// 	newHashArray := make([]*DocFieldProcessorPerField, newHashSize)

// 	// Rehash
// 	newHashMask := newHashSize - 1
// 	for _, fp0 := range p.fieldHash {
// 		for fp0 != nil {
// 			hashPos2 := hashstr(fp0.fieldInfo.Name) & newHashMask
// 			nextFP0 := fp0.next
// 			fp0.next = newHashArray[hashPos2]
// 			newHashArray[hashPos2] = fp0
// 			fp0 = nextFP0
// 		}
// 	}

// 	p.fieldHash = newHashArray
// 	p.hashMask = newHashMask
// }

// func (p *DocFieldProcessor) processDocument(fieldInfos *model.FieldInfosBuilder) error {
// 	p.consumer.startDocument()
// 	p.storedConsumer.startDocument()

// 	p.fieldCount = 0

// 	thisFieldGen := p.fieldGen
// 	p.fieldGen++

// 	// Absorb any new fields first seen in this document. Also absort
// 	// any changes to fields we had already seen before (e.g. suddenly
// 	// turning on norms or vectors, etc.)

// 	for _, field := range p.docState.doc {
// 		fieldName := field.Name()

// 		// Make sure we have a PerField allocated
// 		hashPos := hashstr(fieldName) & p.hashMask
// 		fp := p.fieldHash[hashPos]
// 		for fp != nil && fp.fieldInfo.Name != fieldName {
// 			fp = fp.next
// 		}

// 		if fp == nil {
// 			// TODO FI: we need to genericize the "flags" that a field
// 			// holds, and, how these flags are merged; it needs to be more
// 			// "pluggable" such that if I want to have a new "thing" my
// 			// Fields can do, I can easily add it
// 			fi := fieldInfos.AddOrUpdate(fieldName, field.FieldType())

// 			fp = newDocFieldProcessorPerField(p, fi)
// 			fp.next = p.fieldHash[hashPos]
// 			p.fieldHash[hashPos] = fp
// 			p.totalFieldCount++

// 			if p.totalFieldCount >= len(p.fieldHash)/2 {
// 				p.rehash()
// 			}
// 		} else {
// 			panic("not implemented yet")
// 		}

// 		if thisFieldGen != fp.lastGen {
// 			// First time we're seeing this field for this doc
// 			fp.fieldCount = 0

// 			if p.fieldCount == len(p._fields) {
// 				newSize := len(p._fields) * 2
// 				newArray := make([]*DocFieldProcessorPerField, newSize)
// 				copy(newArray, p._fields[:p.fieldCount])
// 				p._fields = newArray
// 			}

// 			p._fields[p.fieldCount] = fp
// 			p.fieldCount++
// 			fp.lastGen = thisFieldGen
// 		}

// 		fp.addField(field)
// 		p.storedConsumer.addField(p.docState.docID, field, fp.fieldInfo)
// 	}

// 	// If we are writing vectors then we must visit fields in sorted
// 	// order so they are written in sorted order. TODO: we actually
// 	// only need to sort the subset of fields that have vectors enabled;
// 	// we could save [small amount of] CPU here.
// 	util.IntroSort(ByNameDocFieldProcessorPerFields(p._fields[:p.fieldCount]))
// 	for _, perField := range p._fields[:p.fieldCount] {
// 		err := perField.consumer.processFields(perField.fields, perField.fieldCount)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	if prefix, is := p.docState.maxTermPrefix, p.docState.infoStream; prefix != "" && is.IsEnabled("IW") {
// 		is.Message("IW",
// 			"WARNING: document contains at least one immense term (whose UTF8 encoding is longer than the max length %v), all of which were skipped.  Please correct the analyzer to not produce such terms.  The prefix of the first immense term is: '%v...'",
// 			MAX_TERM_LENGTH_UTF8,
// 			prefix)
// 		p.docState.maxTermPrefix = ""
// 	}

// 	return nil
// }

// type ByNameDocFieldProcessorPerFields []*DocFieldProcessorPerField

// func (a ByNameDocFieldProcessorPerFields) Len() int      { return len(a) }
// func (a ByNameDocFieldProcessorPerFields) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
// func (a ByNameDocFieldProcessorPerFields) Less(i, j int) bool {
// 	return a[i].fieldInfo.Name < a[j].fieldInfo.Name
// }

// func (p *DocFieldProcessor) finishDocument() (err error) {
// 	defer func() {
// 		err = mergeError(err, p.consumer.finishDocument())
// 	}()
// 	return p.storedConsumer.finishDocument()
// }

// // index/DocFieldProcessorPerField.java

// /* Holds all per thread, per field state. */
// type DocFieldProcessorPerField struct {
// 	consumer  DocFieldConsumerPerField
// 	fieldInfo *model.FieldInfo

// 	next    *DocFieldProcessorPerField
// 	lastGen int // -1

// 	fieldCount int
// 	fields     []model.IndexableField
// }

// func newDocFieldProcessorPerField(docFieldProcessor *DocFieldProcessor,
// 	fieldInfo *model.FieldInfo) *DocFieldProcessorPerField {
// 	return &DocFieldProcessorPerField{
// 		consumer:  docFieldProcessor.consumer.addField(fieldInfo),
// 		lastGen:   -1,
// 		fieldInfo: fieldInfo,
// 	}
// }

// func (f *DocFieldProcessorPerField) addField(field model.IndexableField) {
// 	if f.fieldCount == len(f.fields) {
// 		newSize := util.Oversize(f.fieldCount+1, util.NUM_BYTES_OBJECT_REF)
// 		newArray := make([]model.IndexableField, newSize)
// 		copy(newArray, f.fields)
// 		f.fields = newArray
// 	}
// 	f.fields[f.fieldCount] = field
// 	f.fieldCount++
// }

// func (f *DocFieldProcessorPerField) abort() {
// 	f.consumer.abort()
// }
