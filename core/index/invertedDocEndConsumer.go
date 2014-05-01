package index

type InvertedDocEndConsumer interface {
	flush(fieldsToHash map[string]InvertedDocEndConsumerPerField, state SegmentWriteState) error
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

func (nc *NormsConsumer) flush(fieldsToFlush map[string]InvertedDocEndConsumerPerField, state SegmentWriteState) error {
	panic("not implemented yet")
}
