package search

type DocIdSetIterator struct {
}

type DocsEnum struct {
	*DocIdSetIterator
}

type Scorer struct {
	*DocsEnum
	weight Weight
}
