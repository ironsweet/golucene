package index

import (
	"lucene/search"
)

const (
	DOCS_ENUM_FLAG_FREQS = 1
)

var (
	DOCS_ENUM_EMPTY = DocsEnum{}
)

type DocsEnum struct {
	search.DocIdSetIterator
}
