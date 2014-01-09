package util

import (
	"io"
	"sync"
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
	// prints a message
	Message(component, message string)
	// returns true if messages are enabled and should be posted.
	IsEnabled(component string) bool
}

// Instance of InfoStream that does no logging at all.
var NO_OUTPUT = NoOutput(true)

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

var defaultInfoStream InfoStream = NO_OUTPUT
var defaultInfoStreamLock = &sync.Mutex{}

// The default InfoStream used by a newly instantiated classes.
func DefaultInfoStream() InfoStream {
	defaultInfoStreamLock.Lock() // synchronized
	defer defaultInfoStreamLock.Unlock()
	return defaultInfoStream
}

/*
Sets the default InfoStream used by a newly instantiated classes. It
cannot be nil, to disable logging use NO_OUTPUT.
*/
func SetDefaultInfoStream(infoStream InfoStream) {
	defaultInfoStreamLock.Lock() // synchronized
	defer defaultInfoStreamLock.Unlock()
	assert2(infoStream != nil, `Cannot set InfoStream default implementation to nil.
To disable logging use NO_OUTPUT.`)
	defaultInfoStream = infoStream
}

// util/PrintStreamInfoStream.java

// InfoStream implementation over an io.Writer such as os.Stdout
type PrintStreamInfoStream struct {
}

func NewPrintStreamInfoStream(w io.Writer) {
	panic("not implemented yet")
}

func (is *PrintStreamInfoStream) Message(component, message string) {
	panic("not implemented yet")
}

func (is *PrintStreamInfoStream) IsEnabled(component string) bool {
	return true
}

func (is *PrintStreamInfoStream) Close() error {
	if !is.isSystemStream() {
		return is.Close()
	}
	return nil
}

func (is *PrintStreamInfoStream) isSystemStream() bool {
	panic("not implemented yet")
}
