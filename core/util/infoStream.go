package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// util/InfoStream.java

/*
Debugging API for Lucene classes such as IndexWriter and SegmentInfos.

NOTE: Enabling infostreams may cause performance degradation in some
components.
*/
type InfoStream interface {
	io.Closer
	// Clone() InfoStream
	// prints a message
	Message(component, message string, args ...interface{})
	// returns true if messages are enabled and should be posted.
	IsEnabled(component string) bool
}

// Instance of InfoStream that does no logging at all.
var NO_OUTPUT = NoOutput(true)

type NoOutput bool

func (is NoOutput) Message(component, message string, args ...interface{}) {
	panic("message() should not be called when isEnabled returns false")
}

func (is NoOutput) IsEnabled(component string) bool {
	return false
}

func (is NoOutput) Close() error { return nil }

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
	log.Print("Setting default infoStream...")
	defaultInfoStreamLock.Lock() // synchronized
	defer defaultInfoStreamLock.Unlock()
	assert2(infoStream != nil, `Cannot set InfoStream default implementation to nil.
To disable logging use NO_OUTPUT.`)
	defaultInfoStream = infoStream
}

// util/PrintStreamInfoStream.java

var MESSAGE_ID int32 // atomic
const FORMAT = "2006/01/02 15:04:05"

/*
InfoStream implementation over an io.Writer such as os.Stdout
*/
type PrintStreamInfoStream struct {
	stream    io.Writer
	messageId int32
}

func NewPrintStreamInfoStream(w io.Writer) *PrintStreamInfoStream {
	return &PrintStreamInfoStream{w, atomic.AddInt32(&MESSAGE_ID, 1) - 1}
}

func (is *PrintStreamInfoStream) Message(component,
	message string, args ...interface{}) {
	fmt.Fprintf(is.stream, "%4v %v [%v] %v\n", component, is.messageId,
		time.Now().Format(FORMAT), fmt.Sprintf(message, args...))
}

func (is *PrintStreamInfoStream) IsEnabled(component string) bool {
	return "TP" != component
}

func (is *PrintStreamInfoStream) Close() error {
	if !is.isSystemStream() {
		return is.Close()
	}
	return nil
}

func (is *PrintStreamInfoStream) isSystemStream() bool {
	return is.stream == os.Stdout || is.stream == os.Stderr
}
