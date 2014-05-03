package index

// index/InvertedDocEndConsumerPerField.java

type InvertedDocEndConsumerPerField interface {
	// finish() error
	abort()
}

// index/NormsConsumerPerField.java

type NormsConsumerPerField struct {
	consumer *NumericDocValuesWriter
}

func (nc *NormsConsumerPerField) flush(state SegmentWriteState, normsWriter DocValuesConsumer) error {
	panic("not implemented yet")
}

func (nc *NormsConsumerPerField) isEmpty() bool {
	return nc.consumer == nil
}

func (nc *NormsConsumerPerField) abort() {}
