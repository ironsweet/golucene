package index

import (
	"io"
)

type StoredFieldsReader interface {
	io.Closer
	visitDocument(n int, visitor StoredFieldVisitor) error
	clone() StoredFieldsReader
}

type TermVectorsReader interface {
	io.Closer
	get(doc int) Fields
	clone() TermVectorsReader
}
