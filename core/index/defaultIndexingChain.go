package index

import (
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

/* Default general purpose indexing chain, which handles indexing all types of fields */
type DefaultIndexChain struct {
	bytesUsed  util.Counter
	docState   *docState
	docWriter  *DocumentsWriterPerThread
	fieldInfos *FieldInfosBuilder

	// Writes postings and term vectors:
	termsHash TermsHash

	nextFieldGen int64

	// Holds fields seen in each document
	fields []*PerField
}

func newDefaultIndexingChain(docWriter *DocumentsWriterPerThread) *DefaultIndexChain {
	termVectorsWriter := newTermVectorsConsumer(docWriter)
	return &DefaultIndexChain{
		docWriter:  docWriter,
		fieldInfos: docWriter.fieldInfos,
		docState:   docWriter.docState,
		bytesUsed:  docWriter._bytesUsed,
		termsHash:  newFreqProxTermsWriter(docWriter, termVectorsWriter),
		fields:     make([]*PerField, 1),
	}
}

func (c *DefaultIndexChain) flush(state *SegmentWriteState) error {
	panic("not implemented yet")
}

/*
Catch up for all docs before us that had no stored fields, or hit
non-aborting errors before writing stored fields.
*/
func (c *DefaultIndexChain) fillStoredFields(docId int) error {
	panic("not implemented yet")
}

func (c *DefaultIndexChain) abort() {
	panic("not implemented yet")
}

/* Calls StoredFieldsWriter.startDocument, aborting the segment if it hits any error. */
func (c *DefaultIndexChain) startStoredFields() error {
	panic("not implemented yet")
}

/* Calls StoredFieldsWriter.finishDocument(), aborting the segment if it hits any error. */
func (c *DefaultIndexChain) finishStoredFields() error {
	panic("not implemented yet")
}

func (c *DefaultIndexChain) processDocument() (err error) {
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

func (c *DefaultIndexChain) processField(field IndexableField,
	fieldGen int64, fieldCount int) (int, error) {
	panic("not implemented yet")
}

type PerField struct {
	*DefaultIndexChain // acess at least docState, termsHash.
}

func (f *PerField) finish() error {
	panic("not implemented yet")
}
