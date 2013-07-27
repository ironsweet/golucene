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

func NewSimpleFSLockFactory(path string) *SimpleFSLockFactory {
	ans := &SimpleFSLockFactory{}
	ans.FSLockFactory = newFSLockFactory()
	ans.setLockDir(path)
	return ans
}

func (f *SimpleFSLockFactory) make(name string) Lock {
	if f.lockPrefix != "" {
		name = fmt.Sprintf("%v-%v", f.lockPrefix, name)
	}
	ans := SimpleFSLock{nil, filepath.Join(f.lockDir, name), name}
	ans.Lock = &Lock{ans}
	return *(ans.Lock)
}

func (f *SimpleFSLockFactory) clear(name string) error {
	if f.lockPrefix != "" {
		name = fmt.Sprintf("%v-%v", f.lockPrefix, name)
	}
	return os.Remove(filepath.Join(f.lockDir, name))
}

type SimpleFSDirectory struct {
	*FSDirectory
}

func NewSimpleFSDirectory(path string) (d *SimpleFSDirectory, err error) {
	d = &SimpleFSDirectory{}
	d.FSDirectory, err = newFSDirectory(d, path)
	if err != nil {
		return nil, err
	}
	return
}

func (d *SimpleFSDirectory) OpenInput(name string, context IOContext) (in IndexInput, err error) {
	d.ensureOpen()
	fpath := filepath.Join(d.path, name)
	sin, err := newSimpleFSIndexInput(fmt.Sprintf("SimpleFSIndexInput(path='%v')", fpath),
		fpath, context, d.chunkSize)
	return sin.BufferedIndexInput, err
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
	super.readInternal = func(buf []byte) error {
		length := len(buf)
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
			i, err := in.file.Read(buf[total : total+readLength])
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
