package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Lock struct {
	self interface{}
}

type LockFactory struct {
	self       interface{}
	lockPrefix string
	Make       func(name string) Lock
	Clear      func(name string) error
}

type FSLockFactory struct {
	*LockFactory
	lockDir string // can not be set twice
}

type SimpleFSLock struct {
	*Lock
	file, dir string
}

type SimpleFSLockFactory struct {
	*FSLockFactory
}

func NewSimpleFSLockFactory(path string) SimpleFSLockFactory {
	ans := SimpleFSLockFactory{}

	origin := &LockFactory{}
	origin.self = ans
	origin.Make = func(name string) Lock {
		if origin.lockPrefix != "" {
			name = fmt.Sprintf("%v-%v", origin.lockPrefix, name)
		}
		ans := SimpleFSLock{nil, filepath.Join(path, name), name}
		ans.Lock = &Lock{ans}
		return *(ans.Lock)
	}
	origin.Clear = func(name string) error {
		if origin.lockPrefix != "" {
			name = fmt.Sprintf("%v-%v", origin.lockPrefix, name)
		}
		return os.Remove(filepath.Join(path, name))
	}

	ans.FSLockFactory = &FSLockFactory{origin, path}
	return ans
}

type Directory struct {
	isOpen      bool
	lockFactory LockFactory
	LockID      func() string
}

func (d *Directory) SetLockFactory(lockFactory LockFactory) {
	d.lockFactory = lockFactory
	d.lockFactory.lockPrefix = d.LockID()
}

func (d *Directory) ensureOpen() {
	if !d.isOpen {
		panic("this Directory is closed")
	}
}

type FSDirectory struct {
	*Directory
	path string
}

func (d *FSDirectory) SetLockFactory(lockFactory LockFactory) {
	d.Directory.SetLockFactory(lockFactory)

	// for filesystem based LockFactory, delete the lockPrefix, if the locks are placed
	// in index dir. If no index dir is given, set ourselves
	if lf, ok := lockFactory.self.(*FSLockFactory); ok {
		if lf.lockDir == "" {
			lf.lockDir = d.path
			lf.lockPrefix = ""
		} else if lf.lockDir == d.path {
			lf.lockPrefix = ""
		}
	}
}

func newFSDirectory(path string) (d FSDirectory, err error) {
	d = FSDirectory{}
	if f, err := os.Open(path); err == nil {
		fi, err := f.Stat()
		if err != nil {
			return d, err
		}
		if !fi.IsDir() {
			return d, errors.New(fmt.Sprintf("file '%v' exists but is not a directory", path))
		}
	}

	super := Directory{LockID: func() string {
		d.ensureOpen()
		var digest int
		for _, ch := range d.path {
			digest = 31*digest + int(ch)
		}
		return fmt.Sprintf("lucene-%v", strconv.FormatUint(uint64(digest), 10))
	}}
	d.Directory = &super
	// TODO default to native lock factory
	d.SetLockFactory(*(NewSimpleFSLockFactory(path).LockFactory))
	return d, nil
}

// TODO support lock factory
func OpenFSDirectory(path string) (d FSDirectory, err error) {
	// TODO support native implementations
	super, err := NewSimpleFSDirectory(path)
	if err != nil {
		return d, err
	}
	return *(super.FSDirectory), nil
}

type SimpleFSDirectory struct {
	*FSDirectory
}

func NewSimpleFSDirectory(path string) (d SimpleFSDirectory, err error) {
	d = SimpleFSDirectory{}
	super, err := newFSDirectory(path)
	if err != nil {
		return d, err
	}
	d.FSDirectory = &super
	return d, nil
}
