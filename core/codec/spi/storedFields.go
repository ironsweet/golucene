package spi

import (
	"github.com/balzaczyy/golucene/core/index/model"
	"io"
)

// index/StoredFieldsVisitor.java

type StoredFieldVisitor interface {
	BinaryField(fi *model.FieldInfo, value []byte) error
	StringField(fi *model.FieldInfo, value string) error
	IntField(fi *model.FieldInfo, value int) error
	LongField(fi *model.FieldInfo, value int64) error
	FloatField(fi *model.FieldInfo, value float32) error
	DoubleField(fi *model.FieldInfo, value float64) error
	NeedsField(fi *model.FieldInfo) (StoredFieldVisitorStatus, error)
}

type StoredFieldVisitorStatus int

const (
	STORED_FIELD_VISITOR_STATUS_YES  = StoredFieldVisitorStatus(1)
	STORED_FIELD_VISITOR_STATUS_NO   = StoredFieldVisitorStatus(2)
	STORED_FIELD_VISITOR_STATUS_STOP = StoredFieldVisitorStatus(3)
)

// codecs/StoredFieldsReader.java

type StoredFieldsReader interface {
	io.Closer
	VisitDocument(n int, visitor StoredFieldVisitor) error
	Clone() StoredFieldsReader
}

// codecs/StoredFieldsWriter.java

/*
Codec API for writing stored fields:

1. For every document, StartDocument() is called, informing the Codec
how many fields will be written.
2. WriteField() is called for each field in the document.
3. After all documents have been writen, Finish() is called for
verification/sanity-checks.
4. Finally the writer is closed.
*/
type StoredFieldsWriter interface {
	io.Closer
	// Called before writing the stored fields of te document.
	// WriteField() will be called for each stored field. Note that
	// this is called even if the document has no stored fields.
	StartDocument() error
	// Called when a document and all its fields have been added.
	FinishDocument() error
	// Writes a single stored field.
	WriteField(info *model.FieldInfo, field model.IndexableField) error
	// Aborts writing entirely, implementation should remove any
	// partially-written files, etc.
	Abort()
	// Called before Close(), passing in the number of documents that
	// were written. Note that this is intentionally redundant
	// (equivalent to the number of calls to startDocument(int)), but a
	// Codec should check that this is the case to detect the JRE bug
	// described in LUCENE-1282.
	Finish(fis model.FieldInfos, numDocs int) error
}
