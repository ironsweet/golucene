package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

/* Default general purpose indexing chain, which handles indexing all types of fields */
type DefaultIndexingChain struct {
	bytesUsed  util.Counter
	docState   *docState
	docWriter  *DocumentsWriterPerThread
	fieldInfos *FieldInfosBuilder

	// Writes postings and term vectors:
	termsHash TermsHash

	storedFieldsWriter StoredFieldsWriter // lazy init
	lastStoredDocId    int

	nextFieldGen int64

	// Holds fields seen in each document
	fields []*PerField
}

func newDefaultIndexingChain(docWriter *DocumentsWriterPerThread) *DefaultIndexingChain {
	termVectorsWriter := newTermVectorsConsumer(docWriter)
	return &DefaultIndexingChain{
		docWriter:  docWriter,
		fieldInfos: docWriter.fieldInfos,
		docState:   docWriter.docState,
		bytesUsed:  docWriter._bytesUsed,
		termsHash:  newFreqProxTermsWriter(docWriter, termVectorsWriter),
		fields:     make([]*PerField, 1),
	}
}

// TODO: can we remove this lazy-init / make cleaner / do it another way...?
func (c *DefaultIndexingChain) initStoredFieldsWriter() (err error) {
	if c.storedFieldsWriter == nil {
		c.storedFieldsWriter, err = c.docWriter.codec.StoredFieldsFormat().FieldsWriter(
			c.docWriter.directory, c.docWriter.segmentInfo, store.IO_CONTEXT_DEFAULT)
	}
	return
}

func (c *DefaultIndexingChain) flush(state *SegmentWriteState) error {
	panic("not implemented yet")
}

/*
Catch up for all docs before us that had no stored fields, or hit
non-aborting errors before writing stored fields.
*/
func (c *DefaultIndexingChain) fillStoredFields(docId int) (err error) {
	for err == nil && c.lastStoredDocId < docId {
		err = c.startStoredFields()
		if err == nil {
			err = c.finishStoredFields()
		}
	}
	return
}

func (c *DefaultIndexingChain) abort() {
	panic("not implemented yet")
}

/* Calls StoredFieldsWriter.startDocument, aborting the segment if it hits any error. */
func (c *DefaultIndexingChain) startStoredFields() (err error) {
	var success = false
	defer func() {
		if !success {
			c.docWriter.setAborting()
		}
	}()

	if err = c.initStoredFieldsWriter(); err != nil {
		return
	}
	if err = c.storedFieldsWriter.StartDocument(); err != nil {
		return
	}
	success = true

	c.lastStoredDocId++
	return nil
}

/* Calls StoredFieldsWriter.finishDocument(), aborting the segment if it hits any error. */
func (c *DefaultIndexingChain) finishStoredFields() error {
	panic("not implemented yet")
}

func (c *DefaultIndexingChain) processDocument() (err error) {
	// How many indexed field names we've seen (collapses multiple
	// field instances by the same name):
	fieldCount := 0

	fieldGen := c.nextFieldGen
	c.nextFieldGen++

	// NOTE: we need to passes here, in case there are multi-valued
	// fields, because we must process all instances of a given field
	// at once, since the anlayzer is free to reuse TOkenStream across
	// fields (i.e., we cannot have more than one TokenStream running
	// "at once"):

	c.termsHash.startDocument()

	if err = c.fillStoredFields(c.docState.docID); err != nil {
		return
	}
	if err = c.startStoredFields(); err != nil {
		return
	}

	if err = func() error {
		defer func() {
			if !c.docWriter.aborting {
				// Finish each indexed field name seen in the document:
				for _, field := range c.fields[:fieldCount] {
					err = mergeError(err, field.finish())
				}
				err = mergeError(err, c.finishStoredFields())
			}
		}()

		for _, field := range c.docState.doc {
			if fieldCount, err = c.processField(field, fieldGen, fieldCount); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return
	}

	var success = false
	defer func() {
		if !success {
			// Must abort, on the possibility that on-disk term vectors are now corrupt:
			c.docWriter.setAborting()
		}
	}()

	if err = c.termsHash.finishDocument(); err != nil {
		return
	}
	success = true
	return nil
}

func (c *DefaultIndexingChain) processField(field IndexableField,
	fieldGen int64, fieldCount int) (int, error) {
	panic("not implemented yet")
}

type PerField struct {
	*DefaultIndexingChain // acess at least docState, termsHash.
}

func (f *PerField) finish() error {
	panic("not implemented yet")
}
