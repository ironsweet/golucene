package packed

import (
	"github.com/balzaczyy/golucene/core/util"
)

// util/packed/PackedInts.java#Writer

/* A write-once Writer. */
type Writer interface {
	// Add a value to the stream.
	Add(v int64) error
	// Perform end-of-stream operations.
	Finish() error
}

// util/packed/PackedWriter.java

type PackedWriter struct {
}

func newPackedWriter(format PackedFormat, out util.DataOutput,
	valueCount, bitsPerValue, mem int) *PackedWriter {
	panic("not implemented yet")
}

func (w *PackedWriter) Add(v int64) error {
	panic("not implemented yet")
}

func (w *PackedWriter) Finish() error {
	panic("not implemented yet")
}
