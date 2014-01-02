package test_framework

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
	. "github.com/balzaczyy/golucene/test_framework/util"
)

// util/_TestUtil.java

func CheckIndex(dir store.Directory, crossCheckTermVectors bool) (status index.CheckIndexStatus, err error) {
	panic("not implemented yet")
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
