package packed

// util/packed/PackedInts.java#Writer

/* A write-once Writer. */
type Writer interface {
	// Add a value to the stream.
	Add(v int64) error
	// The number of bits per value.
	BitsPerValue() int
	// Perform end-of-stream operations.
	Finish() error
}

// util/packed/PackedWriter.java

type PackedWriter struct {
	bitsPerValue int
}

func newPackedWriter(format PackedFormat, out DataOutput,
	valueCount, bitsPerValue, mem int) *PackedWriter {
	panic("not implemented yet")
}

func (w *PackedWriter) Add(v int64) error {
	panic("not implemented yet")
}

func (w *PackedWriter) BitsPerValue() int {
	return w.bitsPerValue
}

func (w *PackedWriter) Finish() error {
	panic("not implemented yet")
}
