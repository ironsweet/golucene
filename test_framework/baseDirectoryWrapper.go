package test_framework

import (
	"fmt"
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
	isOpen                       bool
	store.Directory              // our delegate in directory
	checkIndexOnClose            bool
	crossCheckTermVectorsOnClose bool
}

func NewBaseDirectoryWrapper(delegate store.Directory) *BaseDirectoryWrapperImpl {
	return &BaseDirectoryWrapperImpl{
		Directory:                    delegate,
		checkIndexOnClose:            true,
		crossCheckTermVectorsOnClose: true,
	}
}

func (dw *BaseDirectoryWrapperImpl) Close() error {
	dw.isOpen = false
	if dw.checkIndexOnClose {
		ok, err := index.IsIndexExists(dw)
		if err != nil {
			return err
		}
		if ok {
			CheckIndex(dw, dw.crossCheckTermVectorsOnClose)
		}
	}
	return dw.Directory.Close()
}

func (dw *BaseDirectoryWrapperImpl) IsOpen() bool {
	return dw.isOpen
}

func (dw *BaseDirectoryWrapperImpl) String() string {
	return fmt.Sprintf("BaseDirectoryWrapper(%v)", dw.Directory)
}

func (dw *BaseDirectoryWrapperImpl) ensureOpen() {
	assert2(dw.isOpen, "this Directory is closed")
}
