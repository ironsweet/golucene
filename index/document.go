package index

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
type Document []IndexableField

/** Constructs a new document with no fields. */
func newDocument() Document {
	return Document(make([]IndexableField, 0))
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
func (doc Document) add(field IndexableField) {
	doc = append(doc, field)
}

/** Returns the string value of the field with the given name if any exist in
 * this document, or null.  If multiple fields exist with this name, this
 * method returns the first value added. If only binary fields with this name
 * exist, returns null.
 * For {@link IntField}, {@link LongField}, {@link
 * FloatField} and {@link DoubleField} it returns the string value of the number. If you want
 * the actual numeric field instance back, use {@link #getField}.
 */
func (doc Document) get(name string) string {
	for _, field := range doc {
		if field.name() == name && field.stringValue() != "" {
			return field.stringValue()
		}
	}
	return ""
}

// document/DocumentStoredFieldVisitor.java
/** A {@link StoredFieldVisitor} that creates a {@link
 *  Document} containing all stored fields, or only specific
 *  requested fields provided to {@link #DocumentStoredFieldVisitor(Set)}.
 *  <p>
 *  This is used by {@link IndexReader#document(int)} to load a
 *  document.
 *
 * @lucene.experimental */
type DocumentStoredFieldVisitor struct {
	doc         Document
	fieldsToAdd []string
}

/** Load all stored fields. */
func newDocumentStoredFieldVisitor() *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{
		doc: newDocument(),
	}
}

func (visitor *DocumentStoredFieldVisitor) binaryField(fieldInfo FieldInfo, value []byte) error {
	visitor.doc.add(newStoredField(fieldInfo.name, value))
}

func (visitor *DocumentStoredFieldVisitor) Document() Document {
	return visitor.doc
}
