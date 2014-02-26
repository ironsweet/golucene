package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// index/DocumentsWriterFlushControl.java

/*
This class controls DocumentsWriterPerThread (DWPT) flushing during
indexing. It tracks the memory consumption per DWPT and uses a
configured FlushPolicy to decide if a DWPT must flush.

In addition to the FlushPolicy the flush control might set certain
DWPT as flush pending iff a DWPT exceeds the RAMPerThreadHardLimitMB()
to prevent address space exhaustion.
*/
type DocumentsWriterFlushControl struct {
	hardMaxBytesPerDWPT int64

	stallControl  *DocumentsWriterStallControl
	perThreadPool *DocumentsWriterPerThreadPool
	flushPolicy   FlushPolicy

	documentsWriter       *DocumentsWriter
	config                *LiveIndexWriterConfig
	bufferedDeletesStream *BufferedDeletesStream
	infoStream            util.InfoStream
}

func newDocumentsWriterFlushControl(documentsWriter *DocumentsWriter,
	config *LiveIndexWriterConfig, bufferedDeletesStream *BufferedDeletesStream) *DocumentsWriterFlushControl {
	return &DocumentsWriterFlushControl{
		infoStream:            config.infoStream,
		stallControl:          newDocumentsWriterStallControl(),
		perThreadPool:         documentsWriter.perThreadPool,
		flushPolicy:           documentsWriter.flushPolicy,
		config:                config,
		hardMaxBytesPerDWPT:   int64(config.perRoutineHardLimitMB) * 1024 * 1024,
		documentsWriter:       documentsWriter,
		bufferedDeletesStream: bufferedDeletesStream,
	}
}
