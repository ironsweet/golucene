package util

import (
	"io"
)

// util/InfoStream.java

/*
Debugging API for Lucene classes such as IndexWriter and SegmentInfos.

NOTE: Enabling infostreams may cause performance degradation in some
components.
*/

type InfoStream interface {
	io.Closer
	Clone() InfoStream
	Message(component, message string)
	IsEnabled(component string) bool
}

var DefaultInfoStream = func() InfoStream {
	panic("not implemented yet")
}
