package index

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"log"
	"strconv"
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
		if field.name() == name && field.stringValue() != "" {
			return field.stringValue()
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
func newDocumentStoredFieldVisitor() *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{
		doc: NewDocument(),
	}
}

func (visitor *DocumentStoredFieldVisitor) binaryField(fi FieldInfo, value []byte) error {
	panic("not implemented yet")
	// visitor.doc.add(newStoredField(fieldInfo.name, value))
	// return nil
}

func (visitor *DocumentStoredFieldVisitor) stringField(fi FieldInfo, value string) error {
	ft := newFieldTypeFrom(TEXT_FIELD_TYPE_STORED)
	ft._storeTermVectors = fi.storeTermVector
	ft._indexed = fi.indexed
	ft._omitNorms = fi.omitNorms
	ft._indexOptions = fi.indexOptions
	visitor.doc.Add(newStringField(fi.name, value, ft))
	return nil
}

func (visitor *DocumentStoredFieldVisitor) intField(fi FieldInfo, value int) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) longField(fi FieldInfo, value int64) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) floatField(fi FieldInfo, value float32) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) doubleField(fi FieldInfo, value float64) error {
	panic("not implemented yet")
}

func (visitor *DocumentStoredFieldVisitor) needsField(fi FieldInfo) (status StoredFieldVisitorStatus, err error) {
	if visitor.fieldsToAdd == nil {
		status = STORED_FIELD_VISITOR_STATUS_YES
	} else if _, ok := visitor.fieldsToAdd[fi.name]; ok {
		status = STORED_FIELD_VISITOR_STATUS_YES
	} else {
		status = STORED_FIELD_VISITOR_STATUS_NO
	}
	return
}

func (visitor *DocumentStoredFieldVisitor) Document() *Document {
	return visitor.doc
}

// document/FieldType.java

// Data type of the numeric value
type NumericType int

const (
	FIELD_TYPE_NUMERIC_INT    = 1 // 32-bit integer numeric type
	FIELD_TYPE_NUMERIC_LONG   = 2 // 64-bit long numeric type
	FIELD_TYPE_NUMERIC_FLOAT  = 3 // 32-bit float numeric type
	FIELD_TYPE_NUMERIC_DOUBLE = 4 // 64-bit double numeric type
)

// Describes the properties of a field.
type FieldType struct {
	_indexed                  bool
	_stored                   bool
	_tokenized                bool
	_storeTermVectors         bool
	_storeTermVectorOffsets   bool
	_storeTermVectorPositions bool
	_storeTermVectorPayloads  bool
	_omitNorms                bool
	_indexOptions             IndexOptions
	numericType               NumericType
	frozen                    bool
	numericPrecisionStep      int
	_docValueType             DocValuesType
}

// Create a new mutable FieldType with all of the properties from <code>ref</code>
func newFieldTypeFrom(ref *FieldType) *FieldType {
	ft := newFieldType()
	ft._indexed = ref._indexed
	ft._stored = ref._stored
	ft._tokenized = ref._tokenized
	ft._storeTermVectors = ref._storeTermVectors
	ft._storeTermVectorOffsets = ref._storeTermVectorOffsets
	ft._storeTermVectorPositions = ref._storeTermVectorPositions
	ft._storeTermVectorPayloads = ref._storeTermVectorPayloads
	ft._omitNorms = ref._omitNorms
	ft._indexOptions = ref._indexOptions
	ft._docValueType = ref._docValueType
	ft.numericType = ref.numericType
	// Do not copy frozen!
	return ft
}

// Create a new FieldType with default properties.
func newFieldType() *FieldType {
	return &FieldType{
		_tokenized:           true,
		_indexOptions:        INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS,
		numericPrecisionStep: util.NUMERIC_PRECISION_STEP_DEFAULT,
	}
}

func (ft *FieldType) indexed() bool   { return ft._indexed }
func (ft *FieldType) stored() bool    { return ft._stored }
func (ft *FieldType) tokenized() bool { return ft._tokenized }

func (ft *FieldType) storeTermVectors() bool         { return ft._storeTermVectors }
func (ft *FieldType) storeTermVectorOffsets() bool   { return ft._storeTermVectorOffsets }
func (ft *FieldType) storeTermVectorPositions() bool { return ft._storeTermVectorPositions }
func (ft *FieldType) storeTermVectorPayloads() bool  { return ft._storeTermVectorPayloads }

func (ft *FieldType) omitNorms() bool             { return ft._omitNorms }
func (ft *FieldType) indexOptions() IndexOptions  { return ft._indexOptions }
func (ft *FieldType) docValueType() DocValuesType { return ft._docValueType }

// Prints a Field for human consumption.
func (ft *FieldType) String() string {
	var buf bytes.Buffer
	if ft.stored() {
		buf.WriteString("stored")
	}
	if ft.indexed() {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("indexed")
		if ft.tokenized() {
			buf.WriteString(",tokenized")
		}
		if ft.storeTermVectors() {
			buf.WriteString(",termVector")
		}
		if ft.storeTermVectorOffsets() {
			buf.WriteString(",termVectorOffsets")
		}
		if ft.storeTermVectorPositions() {
			buf.WriteString(",termVectorPosition")
			if ft.storeTermVectorPayloads() {
				buf.WriteString(",termVectorPayloads")
			}
		}
		if ft.omitNorms() {
			buf.WriteString(",omitNorms")
		}
		if ft.indexOptions() != INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS {
			fmt.Fprintf(&buf, ",indexOptions=%v", ft.indexOptions())
		}
		if ft.numericType != 0 {
			fmt.Fprintf(&buf, ",numericType=%v,numericPrecisionStep=%v", ft.numericType, ft.numericPrecisionStep)
		}
	}
	if ft.docValueType() != 0 {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		fmt.Fprintf(&buf, "docValueType=%v", ft.docValueType())
	}
	return buf.String()
}

// document/Field.java

type Field struct {
	_type  *FieldType  // Field's type
	_name  string      // Field's name
	_data  interface{} // Field's value
	_boost float32     // Field's boost

	/*
		Pre-analyzed tokenStream for indexed fields; this is
		separte from fieldsData because you are allowed to
		have both; eg maybe field has a String value but you
		customize how it's tokenized
	*/
	_tokenStream analysis.TokenStream

	internalTokenStream analysis.TokenStream
}

// Create field with String value
func newStringField(name, value string, ft *FieldType) *Field {
	assert2(ft._stored || ft._indexed,
		"it doesn't make sense to have a field that is neither indexed nor stored")
	assert2(ft._indexed || !ft._storeTermVectors,
		"can not store term vector information for a field that is not indexed")
	return &Field{_type: ft, _name: name, _data: value}
}

func (f *Field) stringValue() string {
	switch f._data.(type) {
	case string:
		return f._data.(string)
	case int:
		return strconv.Itoa(f._data.(int))
	default:
		log.Println("Unknown type", f._data)
		panic("not implemented yet")
	}
}

func assert2(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func (f *Field) readerValue() io.Reader {
	if v, ok := f._data.(io.Reader); ok {
		return v
	}
	return nil
}

func (f *Field) name() string {
	return f._name
}

func (f *Field) boost() float32 {
	return f._boost
}

func (f *Field) binaryValue() []byte {
	if v, ok := f._data.([]byte); ok {
		return v
	}
	return nil
}

func (f *Field) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v<%v:", f._type, f._name)
	if f._data != nil {
		fmt.Fprint(&buf, f._data)
	}
	fmt.Fprint(&buf, ">")
	return buf.String()
}

func (f *Field) fieldType() IndexableFieldType {
	return f._type
}

func (f *Field) tokenStream(analyzer analysis.Analyzer) (ts analysis.TokenStream, err error) {
	panic("not implemented yet")
}

// document/TextField.java

var (
	// Indexed, tokenized, not stored
	TEXT_FIELD_TYPE_NOT_STORED = func() *FieldType {
		ft := newFieldType()
		ft._indexed = true
		ft._tokenized = true
		ft.frozen = true
		return ft
	}()
	// Indexed, tokenized, stored
	TEXT_FIELD_TYPE_STORED = func() *FieldType {
		ft := newFieldType()
		ft._indexed = true
		ft._tokenized = true
		ft._stored = true
		ft.frozen = true
		return ft
	}()
)

type TextField struct {
	*Field
}

// document/StoredField.java

// Type for a stored-only field.
var STORED_FIELD_TYPE = func() *FieldType {
	ans := newFieldType()
	ans._stored = true
	return ans
}()

/*
A field whose value is stored so that IndexSearcher.doc()
and IndexReader.document() will return the field and its
value.
*/
type StoredField struct {
	*Field
}

/*
Create a stored-only field with the given binary value.

NOTE: the provided byte[] is not copied so be sure
not to change it until you're done with this field.
*/
// func newStoredField(name string, value []byte) *StoredField {
// 	return &StoredField{newStringField(name, value, STORED_FIELD_TYPE)}
// }
