package util

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
	"io"
	"testing"
)

// util/_TestUtil.java

// Returns a temp directory, based on the given description. Creates the directory.
func TempDir(desc string) string {
	if len(desc) < 3 {
		panic("description must be at least 3 characters")
	}
	panic("not implemented yet")
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

type T struct {
	delegate *testing.T
}

func WrapT(t *testing.T) *T {
	return &T{t}
}

func (c *T) Error(args ...interface{}) {
	c.delegate.Error(args)
}

// util/CloseableDirectory.java

// Attempts to close a BaseDirectoryWrapper
type CloseableDirectory struct {
	dir           BaseDirectoryWrapper
	failureMarker *TestRuleMarkFailure
}

type BaseDirectoryWrapper interface {
	io.Closer
	IsOpen() bool
}

func NewCloseableDirectory(dir BaseDirectoryWrapper,
	failureMarker *TestRuleMarkFailure) *CloseableDirectory {
	return &CloseableDirectory{dir, failureMarker}
}

func (cd *CloseableDirectory) Close() error {
	// We only attempt to check open/closed state if there were no other test
	// failures.
	// TODO: perform real close of the delegate: LUCENE-4058
	// defer cd.dir.Close()
	if cd.failureMarker.WasSuccessful() && cd.dir.IsOpen() {
		panic(fmt.Sprintf("Directory not closed: %v", cd.dir))
	}
	return nil
}
