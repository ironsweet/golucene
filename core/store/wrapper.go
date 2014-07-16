package store

// store/TrackingDirectoryWrapper.java

import (
	"fmt"
	"sync"
)

/*
A delegating Directory that records which files were written to and deleted.
*/
type TrackingDirectoryWrapper struct {
	Directory
	sync.Locker
	createdFilenames map[string]bool // synchronized
}

func NewTrackingDirectoryWrapper(other Directory) *TrackingDirectoryWrapper {
	return &TrackingDirectoryWrapper{
		Directory:        other,
		Locker:           &sync.Mutex{},
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

func (w *TrackingDirectoryWrapper) CreateOutput(name string, ctx IOContext) (IndexOutput, error) {
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

func (w *TrackingDirectoryWrapper) Copy(to Directory, src, dest string, ctx IOContext) error {
	func() {
		w.Lock()
		defer w.Unlock()
		w.createdFilenames[dest] = true
	}()
	return w.Directory.Copy(to, src, dest, ctx)
}

func (w *TrackingDirectoryWrapper) EachCreatedFiles(f func(name string)) {
	w.Lock()
	defer w.Unlock()
	for name, _ := range w.createdFilenames {
		f(name)
	}
}

func (w *TrackingDirectoryWrapper) ContainsFile(name string) bool {
	w.Lock()
	defer w.Unlock()
	_, ok := w.createdFilenames[name]
	return ok
}
