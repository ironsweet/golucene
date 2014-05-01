package index

type DocValuesWriter interface {
	abort()
	finish(numDoc int)
	flush(state SegmentWriteState, consumer DocValuesConsumer) error
}
