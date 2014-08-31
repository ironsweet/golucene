package document

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
)

// document/Document.java
/** Documents are the unit of indexing and search.
 *
 * A Document is a set of fields.  Each field has a name and a textual value.
 * A field may be {@link org.apache.lucene.index.IndexableFieldType#stored() stored} with the document, in which
 * case it is returned with search hits on the document.  Thus each document
 * should typically contain one or more stored fields which uniquely identify
 * it.
 *
 * <p>Note that fields which are <i>not</i> {@link org.apache.lucene.index.IndexableFieldType#stored() stored} are
 * <i>not</i> available in documents retrieved from the index, e.g. with {@link
 * ScoreDoc#doc} or {@link IndexReader#document(int)}.
 */
type Document struct {
	fields []IndexableField
}

/** Constructs a new document with no fields. */
func NewDocument() *Document {
	return &Document{make([]IndexableField, 0)}
}

func (doc *Document) Fields() []IndexableField {
	return doc.fields
}

/**
 * <p>Adds a field to a document.  Several fields may be added with
 * the same name.  In this case, if the fields are indexed, their text is
 * treated as though appended for the purposes of search.</p>
 * <p> Note that add like the removeField(s) methods only makes sense
 * prior to adding a document to an index. These methods cannot
 * be used to change the content of an existing index! In order to achieve this,
 * a document has to be deleted from an index and a new changed version of that
 * document has to be added.</p>
 */
func (doc *Document) Add(field IndexableField) {
	doc.fields = append(doc.fields, field)
}

/*
Returns the string value of the field with the given name if any exist in
this document, or null.  If multiple fields exist with this name, this
method returns the first value added. If only binary fields with this name
exist, returns null.

For IntField, LongField, FloatField, and DoubleField, it returns the string
value of the number. If you want the actual numeric field instance back, use
getField().
*/
func (doc *Document) Get(name string) string {
	for _, field := range doc.fields {
		if field.Name() == name && field.StringValue() != "" {
			return field.StringValue()
		}
	}
	return ""
}

// document/DocumentStoredFieldVisitor.java
/*
A StoredFieldVisitor that creates a Document containing all
stored fields, or only specific requested fields provided
to DocumentStoredFieldVisitor.

This is used by IndexReader.Document() to load a document.
*/
type DocumentStoredFieldVisitor struct {
	*StoredFieldVisitorAdapter
	doc         *Document
	fieldsToAdd map[string]bool
}

/** Load all stored fields. */
func NewDocumentStoredFieldVisitor() *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{
		doc: NewDocument(),
	}
}

func (visitor *DocumentStoredFieldVisitor) BinaryField(fi *FieldInfo, value []byte) error {
	panic("not implemented yet")
	// visitor.doc.add(newStoredField(fieldInfo.name, value))
	// return nil
}

func (visitor *DocumentStoredFieldVisitor) StringField(fi *FieldInfo, value string) error {
	ft := NewFieldTypeFrom(TEXT_FIELD_TYPE_STORED)
	ft.storeTermVectors = fi.HasVectors()
	ft.indexed = fi.IsIndexed()
	ft._omitNorms = fi.OmitsNorms()
	ft._indexOptions = fi.IndexOptions()
	visitor.doc.Add(NewFieldFromString(fi.Name, value, ft))
	return nil
}

func (visitor *DocumentStoredFieldVisitor) IntField(fi *FieldInfo, value int) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) LongField(fi *FieldInfo, value int64) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) FloatField(fi *FieldInfo, value float32) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) DoubleField(fi *FieldInfo, value float64) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) NeedsField(fi *FieldInfo) (status StoredFieldVisitorStatus, err error) {
	if visitor.fieldsToAdd == nil {
		status = STORED_FIELD_VISITOR_STATUS_YES
	} else if _, ok := visitor.fieldsToAdd[fi.Name]; ok {
		status = STORED_FIELD_VISITOR_STATUS_YES
	} else {
		status = STORED_FIELD_VISITOR_STATUS_NO
	}
	return
}

func (visitor *DocumentStoredFieldVisitor) Document() *Document {
	return visitor.doc
}

type StoredFieldVisitorAdapter struct{}

func (va *StoredFieldVisitorAdapter) BinaryField(fi *FieldInfo, value []byte) error  { return nil }
func (va *StoredFieldVisitorAdapter) StringField(fi *FieldInfo, value string) error  { return nil }
func (va *StoredFieldVisitorAdapter) IntField(fi *FieldInfo, value int) error        { return nil }
func (va *StoredFieldVisitorAdapter) LongField(fi *FieldInfo, value int64) error     { return nil }
func (va *StoredFieldVisitorAdapter) FloatField(fi *FieldInfo, value float32) error  { return nil }
func (va *StoredFieldVisitorAdapter) DoubleField(fi *FieldInfo, value float64) error { return nil }
