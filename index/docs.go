package index

type DocIdSetIterator interface {
	DocId() int
	Freq() int
	NextDoc() (doc int, more bool)
}

const (
	DOCS_ENUM_FLAG_FREQS = 1
)

var (
	DOCS_ENUM_EMPTY = DocsEnum{}
)

type DocsEnum struct {
	DocIdSetIterator
}
