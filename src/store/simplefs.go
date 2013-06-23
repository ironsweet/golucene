package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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
	fileLock sync.Locker
}

func newSimpleFSIndexInput(desc, path string, context IOContext, chunkSize int) (in *SimpleFSIndexInput, err error) {
	super, err := newFSIndexInput(desc, path, context, chunkSize)
	if err != nil {
		return nil, err
	}
	in = &SimpleFSIndexInput{super, &sync.Mutex{}}
	super.readInternal = func(b []byte, offset, length int) error {
		in.fileLock.Lock()
		defer in.fileLock.Unlock()

		// TODO make use of Go's relative Seek or ReadAt function
		position := in.off + in.FilePointer()
		_, err := in.file.Seek(position, 0)
		if err != nil {
			return err
		}

		if position+int64(length) > in.end {
			return errors.New(fmt.Sprintf("read past EOF: %v", in))
		}

		total := 0
		for {
			readLength := length - total
			if in.chunkSize < readLength {
				readLength = in.chunkSize
			}
			// FIXME verify slice is working
			i, err := in.file.Read(b[offset+total : offset+total+readLength])
			if err != nil {
				return errors.New(fmt.Sprintf("%v: %v", err, in))
			}
			total += i
			if total >= length {
				break
			}
		}
		return nil
	}
	super.seekInternal = func(position int64) {}
	return in, nil
}
