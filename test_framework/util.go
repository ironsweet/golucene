package test_framework

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
	"io/ioutil"
	"os"
)

// util/_TestUtil.java

// Returns a temp directory, based on the given description. Creates the directory.
func TempDir(desc string) string {
	if len(desc) < 3 {
		panic("description must be at least 3 characters")
	}
	// Ian: I prefer Go's own way to obtain temp folder instead of Lucene's method
	f, err := ioutil.TempDir("", desc)
	if err != nil {
		panic(err)
	}
	f += ".tmp" // add suffix
	closeAfterSuite(NewCloseableFile(f, suiteFailureMarker))
	return f
}

func CheckIndex(dir store.Directory, crossCheckTermVectors bool) (status index.CheckIndexStatus, err error) {
	panic("not implemented yet")
}

// util/TestRuleMarkFailure.java

type TestRuleMarkFailure struct {
	*T
	failures bool
}

func (tr *TestRuleMarkFailure) Error(args ...interface{}) {
	tr.failures = true
	tr.T.Error(args)
}

// Check if this object had any marked failures.
func (tr *TestRuleMarkFailure) hadFailures() bool {
	return tr.failures
}

// Check if this object was sucessful
func (tr *TestRuleMarkFailure) WasSuccessful() bool {
	return tr.hadFailures()
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

// util/CloseableFile.java

// A Closeable that attempts to remove a given file/folder
func NewCloseableFile(file string, failureMarker *TestRuleMarkFailure) func() error {
	return func() error {
		// only if there were no other test failures.
		if failureMarker.WasSuccessful() {
			os.RemoveAll(file) // ignore any error
			// no re-check
		}
		return nil
	}
}
