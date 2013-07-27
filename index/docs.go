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

const (
	DOCS_POSITIONS_ENUM_FLAG_OFF_SETS = 1
	DOCS_POSITIONS_ENUM_FLAG_PAYLOADS = 2
)

type DocsAndPositionsEnum struct {
}
