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

type NoOutput bool

func (is NoOutput) Message(component, message string) {
	panic("message() should not be called when isEnabled returns false")
}

func (is NoOutput) IsEnabled(component string) bool {
	return false
}

func (is NoOutput) Close() error { return nil }

func (is NoOutput) Clone() InfoStream {
	return is
}

var NO_OUTPUT = NoOutput(true)

var DefaultInfoStream = func() InfoStream {
	return NO_OUTPUT
}
