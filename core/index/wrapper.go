package index

// store/TrackingDirectoryWrapper.java

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"sync"
)

/*
A delegating Directory that records which files were written to and deleted.
*/
type TrackingDirectoryWrapper struct {
	sync.Locker
	store.Directory
	createdFilenames map[string]bool // synchronized
}

func newTrackingDirectoryWrapper(other store.Directory) *TrackingDirectoryWrapper {
	return &TrackingDirectoryWrapper{
		Locker:           &sync.Mutex{},
		Directory:        other,
		createdFilenames: make(map[string]bool),
	}
}

func (w *TrackingDirectoryWrapper) DeleteFile(name string) error {
	func() {
		w.Lock()
		defer w.Unlock()
		delete(w.createdFilenames, name)
	}()
	return w.Directory.DeleteFile(name)
}

func (w *TrackingDirectoryWrapper) CreateOutput(name string, ctx store.IOContext) (store.IndexOutput, error) {
	func() {
		w.Lock()
		defer w.Unlock()
		w.createdFilenames[name] = true
	}()
	return w.Directory.CreateOutput(name, ctx)
}

func (w *TrackingDirectoryWrapper) String() string {
	return fmt.Sprintf("TrackingDirectoryWrapper(%v)", w.Directory)
}

func (w *TrackingDirectoryWrapper) Copy(to store.Directory, src, dest string, ctx store.IOContext) error {
	func() {
		w.Lock()
		defer w.Unlock()
		w.createdFilenames[dest] = true
	}()
	return w.Directory.Copy(to, src, dest, ctx)
}

func (w *TrackingDirectoryWrapper) createdFiles() map[string]bool {
	panic("not implemented yet")
}
