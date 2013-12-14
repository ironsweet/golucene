package store

import (
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/test_framework/util"
)

// store/BaseDirectoryWrapper.java

/*
Calls check index on close.
do NOT make any methods in this class synchronized, volatile
do NOT import anything from the concurrency package.
no randoms, no nothing.
*/
type BaseDirectoryWrapper struct {
	*store.DirectoryImpl
	// our in directory
	delegate                     store.Directory
	checkIndexOnClose            bool
	crossCheckTermVectorsOnClose bool
}

func NewBaseDirectoryWrapper(delegate store.Directory) *BaseDirectoryWrapper {
	ans := &BaseDirectoryWrapper{nil, delegate, true, true}
	ans.Directory = store.NewDirectoryImpl(ans)
	return ans
}

func (dw *BaseDirectoryWrapper) IsOpen() bool {
	return dw.DirectoryImpl.IsOpen
}

func (dw *BaseDirectoryWrapper) Close() error {
	dw.DirectoryImpl.IsOpen = false
	if dw.checkIndexOnClose {
		ok, err := index.IsIndexExists(dw)
		if err == nil && ok {
			_, err = util.CheckIndex(dw, dw.crossCheckTermVectorsOnClose)
		}
		if err != nil {
			return err
		}
	}
	return dw.delegate.Close()
}

func (dw *BaseDirectoryWrapper) ClearLock(name string) error {
	return dw.delegate.ClearLock(name)
}
