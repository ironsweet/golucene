package store

import (
	"fmt"
	"os"
	"path/filepath"
)

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

type SimpleFSDirectory struct {
	*FSDirectory
}

func NewSimpleFSDirectory(path string) (d *SimpleFSDirectory, err error) {
	super, err := newFSDirectory(path)
	if err != nil {
		return nil, err
	}
	super.OpenInput = func(name string, context IOContext) (in *IndexInput, err error) {
		super.ensureOpen()
		fpath := filepath.Join(path, name)
		sin, err := newSimpleFSIndexInput(fmt.Sprintf("SimpleFSIndexInput(path='%v')", fpath),
			fpath, context, super.chunkSize)
		return sin.IndexInput, err
	}
	return &SimpleFSDirectory{super}, nil
}

type SimpleFSIndexInput struct {
	*FSIndexInput
}

func newSimpleFSIndexInput(desc, path string, context IOContext, chunkSize int) (in *SimpleFSIndexInput, err error) {
	super, err := newFSIndexInput(desc, path, context, chunkSize)
	if err != nil {
		return nil, err
	}
	return &SimpleFSIndexInput{super}, nil
}
