package test_framework

import (
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
)

// store/BaseDirectoryWrapper.java

/*
Calls check index on close.
do NOT make any methods in this class synchronized, volatile
do NOT import anything from the concurrency package.
no randoms, no nothing.
*/
type BaseDirectoryWrapper interface {
	store.Directory
	IsOpen() bool
}

type BaseDirectoryWrapperImpl struct {
	*store.DirectoryImpl
	// our in directory
	delegate                     store.Directory
	checkIndexOnClose            bool
	crossCheckTermVectorsOnClose bool
}

func NewBaseDirectoryWrapper(delegate store.Directory) *BaseDirectoryWrapperImpl {
	ans := &BaseDirectoryWrapperImpl{nil, delegate, true, true}
	ans.Directory = store.NewDirectoryImpl(ans)
	return ans
}

func (dw *BaseDirectoryWrapperImpl) IsOpen() bool {
	return dw.DirectoryImpl.IsOpen
}

func (dw *BaseDirectoryWrapperImpl) Close() error {
	dw.DirectoryImpl.IsOpen = false
	if dw.checkIndexOnClose {
		ok, err := index.IsIndexExists(dw)
		if err == nil && ok {
			_, err = CheckIndex(dw, dw.crossCheckTermVectorsOnClose)
		}
		if err != nil {
			return err
		}
	}
	return dw.delegate.Close()
}
