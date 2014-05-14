package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocConsumer.java

type DocConsumer interface {
	processDocument(fieldInfos *model.FieldInfosBuilder) error
	finishDocument() error
	flush(state SegmentWriteState) error
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
	consumer       DocFieldConsumer
	storedConsumer StoredFieldsConsumer
	codec          Codec

	// Holds all fields seen in current doc
	_fields    []*DocFieldProcessorPerField
	fieldCount int

	// Hash table for all fields ever seen
	fieldHash       []*DocFieldProcessorPerField
	hashMask        int
	totalFieldCount int

	fieldGen int

	docState *docState

	bytesUsed util.Counter
}

func newDocFieldProcessor(docWriter *DocumentsWriterPerThread,
	consumer DocFieldConsumer, storedConsumer StoredFieldsConsumer) *DocFieldProcessor {

	assert(storedConsumer != nil)
	return &DocFieldProcessor{
		fieldHash:      make([]*DocFieldProcessorPerField, 2),
		hashMask:       1,
		docState:       docWriter.docState,
		codec:          docWriter.codec,
		bytesUsed:      docWriter._bytesUsed,
		consumer:       consumer,
		storedConsumer: storedConsumer,
	}
}

func (p *DocFieldProcessor) flush(state SegmentWriteState) error {
	childFields := make(map[string]DocFieldConsumerPerField)
	for _, f := range p.fields() {
		childFields[f.fieldInfo().Name] = f
	}

	err := p.storedConsumer.flush(state)
	if err != nil {
		return err
	}
	err = p.consumer.flush(childFields, state)
	if err != nil {
		return err
	}

	// Impotant to save after asking consumer to flush so consumer can
	// alter the FieldInfo if necessary. E.g., FreqProxTermsWriter does
	// this with FieldInfo.storePayload.
	infosWriter := p.codec.FieldInfosFormat().FieldInfosWriter()
	assert(infosWriter != nil)
	return infosWriter(state.directory, state.segmentInfo.Name,
		state.fieldInfos, store.IO_CONTEXT_DEFAULT)
}

func (p *DocFieldProcessor) abort() {
	for _, field := range p.fieldHash {
		for field != nil {
			next := field.next
			field.abort()
			field = next
		}
	}
	p.storedConsumer.abort()
	p.consumer.abort()
	// assert2(err == nil, err.Error())
}

func (p *DocFieldProcessor) fields() []DocFieldConsumerPerField {
	var fields []DocFieldConsumerPerField
	for _, field := range p.fieldHash {
		for field != nil {
			fields = append(fields, field.consumer)
			field = field.next
		}
	}
	assert(len(fields) == p.totalFieldCount)
	return fields
}

func (p *DocFieldProcessor) rehash() {
	newHashSize := len(p.fieldHash) * 2
	assert(newHashSize > len(p.fieldHash)) // avoid overflow

	newHashArray := make([]*DocFieldProcessorPerField, newHashSize)

	// Rehash
	newHashMask := newHashSize - 1
	for _, fp0 := range p.fieldHash {
		for fp0 != nil {
			hashPos2 := hashstr(fp0.fieldInfo.Name) & newHashMask
			nextFP0 := fp0.next
			fp0.next = newHashArray[hashPos2]
			newHashArray[hashPos2] = fp0
			fp0 = nextFP0
		}
	}

	p.fieldHash = newHashArray
	p.hashMask = newHashMask
}

func (p *DocFieldProcessor) processDocument(fieldInfos *model.FieldInfosBuilder) error {
	p.consumer.startDocument()
	p.storedConsumer.startDocument()

	p.fieldCount = 0

	thisFieldGen := p.fieldGen
	p.fieldGen++

	// Absorb any new fields first seen in this document. Also absort
	// any changes to fields we had already seen before (e.g. suddenly
	// turning on norms or vectors, etc.)

	for _, field := range p.docState.doc {
		fieldName := field.Name()

		// Make sure we have a PerField allocated
		hashPos := hashstr(fieldName) & p.hashMask
		fp := p.fieldHash[hashPos]
		for fp != nil && fp.fieldInfo.Name != fieldName {
			fp = fp.next
		}

		if fp == nil {
			// TODO FI: we need to genericize the "flags" that a field
			// holds, and, how these flags are merged; it needs to be more
			// "pluggable" such that if I want to have a new "thing" my
			// Fields can do, I can easily add it
			fi := fieldInfos.AddOrUpdate(fieldName, field.FieldType())

			fp = newDocFieldProcessorPerField(p, fi)
			fp.next = p.fieldHash[hashPos]
			p.fieldHash[hashPos] = fp
			p.totalFieldCount++

			if p.totalFieldCount >= len(p.fieldHash)/2 {
				p.rehash()
			}
		} else {
			panic("not implemented yet")
		}

		if thisFieldGen != fp.lastGen {
			panic("not implemented yet")
		}

		fp.addField(field)
		p.storedConsumer.addField(p.docState.docID, field, fp.fieldInfo)
	}
	return nil
}

const primeRK = 16777619

/* simple string hash used by Go strings package */
func hashstr(sep string) int {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}
	return int(hash)
}

func (p *DocFieldProcessor) finishDocument() (err error) {
	defer func() {
		err = mergeError(err, p.consumer.finishDocument())
	}()
	return p.storedConsumer.finishDocument()
}

// index/DocFieldProcessorPerField.java

/* Holds all per thread, per field state. */
type DocFieldProcessorPerField struct {
	consumer  DocFieldConsumerPerField
	fieldInfo *model.FieldInfo

	next    *DocFieldProcessorPerField
	lastGen int // -1

	fieldCount int
	fields     []model.IndexableField
}

func newDocFieldProcessorPerField(docFieldProcessor *DocFieldProcessor,
	fieldInfo *model.FieldInfo) *DocFieldProcessorPerField {
	return &DocFieldProcessorPerField{
		consumer:  docFieldProcessor.consumer.addField(fieldInfo),
		fieldInfo: fieldInfo,
	}
}

func (f *DocFieldProcessorPerField) addField(field model.IndexableField) {
	if f.fieldCount == len(f.fields) {
		newSize := util.Oversize(f.fieldCount+1, util.NUM_BYTES_OBJECT_REF)
		newArray := make([]model.IndexableField, newSize)
		copy(newArray, f.fields)
		f.fields = newArray
	}
	f.fields[f.fieldCount] = field
	f.fieldCount++
}

func (f *DocFieldProcessorPerField) abort() {
	f.consumer.abort()
}
