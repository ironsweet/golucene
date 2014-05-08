package test_framework

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	. "github.com/balzaczyy/golucene/test_framework/util"
	"log"
)

// util/_TestUtil.java

func CheckIndex(dir store.Directory, crossCheckTermVectors bool) *index.CheckIndexStatus {
	var buf bytes.Buffer
	checker := index.NewCheckIndex(dir, crossCheckTermVectors, &buf)
	indexStatus := checker.CheckIndex(nil)
	if indexStatus == nil || !indexStatus.Clean {
		fmt.Println("CheckIndex failed")
		fmt.Println(buf.String())
		panic("CheckIndex failed")
	}
	if INFOSTREAM {
		log.Println(buf.String())
	}
	return indexStatus
}

// util/CloseableDirectory.java

// Attempts to close a BaseDirectoryWrapper
func NewCloseableDirectory(dir BaseDirectoryWrapper, failureMarker *TestRuleMarkFailure) func() error {
	return func() error {
		// We only attempt to check open/closed state if there were no other test
		// failures.
		// TODO: perform real close of the delegate: LUCENE-4058
		// defer dir.Close()
		if failureMarker.WasSuccessful() && dir.IsOpen() {
			panic(fmt.Sprintf("Directory not closed: %v", dir))
		}
		return nil
	}
}

// util/NullInfoStream.java

// Prints nothing. Just to make sure tests pass w/ and w/o enabled
// InfoStream without actually making noise.
type NullInfoStream int

var nullInfoStream = NullInfoStream(0)

func NewNullInfoStream() util.InfoStream {
	return nullInfoStream
}

func (is NullInfoStream) Message(component, message string, args ...interface{}) {
	assert(component != "")
	assert(message != "")
}

func (is NullInfoStream) IsEnabled(component string) bool {
	assert(component != "")
	return true // to actually enable logging, we just ignore on message()
}

func (is NullInfoStream) Close() error {
	return nil
}

func (is NullInfoStream) Clone() util.InfoStream {
	return is
}

// func assert(ok bool) {
// 	if !ok {
// 		panic("assert fail")
// 	}
// }
