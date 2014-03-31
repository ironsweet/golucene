package index

type InvertedDocEndConsumer interface {
	abort()
}

/*
Writes norms. Each thread X field accumlates the norms for the
doc/fields it saw, then the flush method below merges all of these
 together into a single _X.nrm file.
*/
type NormsConsumer struct {
}

func (nc *NormsConsumer) abort() {}
