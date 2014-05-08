package store

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// store/SimpleFSLockFactory.java

type SimpleFSLock struct {
	*LockImpl
	file, dir string
}

func newSimpleFSLock(lockDir, lockFileName string) *SimpleFSLock {
	ans := &SimpleFSLock{
		dir:  lockDir,
		file: filepath.Join(lockDir, lockFileName),
	}
	ans.LockImpl = NewLockImpl(ans)
	return ans
}

func (lock *SimpleFSLock) Obtain() (ok bool, err error) {
	// Ensure that lockDir exists and is a directory:
	var fi os.FileInfo
	fi, err = os.Stat(lock.dir)
	if err == nil { // exists
		if !fi.IsDir() {
			err = errors.New(fmt.Sprintf("Found regular file where directory expected: %v", lock.dir))
			return
		}
	} else if os.IsNotExist(err) {
		err = os.Mkdir(lock.dir, 0600)
		if err != nil { // IO error
			return
		}
	} else { // IO error
		return
	}
	var f *os.File
	if f, err = os.OpenFile(lock.file, os.O_CREATE, 0666); err == nil {
		log.Printf("File '%v' is created.", f.Name())
		ok = true
		defer f.Close()
	}
	return

}

func (lock *SimpleFSLock) Release() (err error) {
	return os.Remove(lock.file)
}

func (lock *SimpleFSLock) IsLocked() bool {
	f, err := os.Open(lock.file)
	if err == nil {
		defer f.Close()
	}
	return err == nil || os.IsExist(err)
}

func (lock *SimpleFSLock) String() string {
	return fmt.Sprintf("SimpleFSLock@%v", lock.file)
}

/*
Implements LockFactory using os.Create().

NOTE: This API may has the same issue as the one in Lucene Java that
the write lock may not be released when Go program exists abnormally.

When this happens, an error is returned when trying to create a
writer, in which case you need to explicitly clear the lock file
first. You can either manually remove the file, or use
UnlockDirectory() API. But, first be certain that no writer is in
fact writing to the index otherwise you can easily corrupt your index.

If you suspect that this or any other LockFactory is not working
properly in your environment, you can easily test it by using
VerifyingLockFactory, LockVerifyServer and LockStressTest.
*/
type SimpleFSLockFactory struct {
	*FSLockFactory
}

func NewSimpleFSLockFactory(path string) *SimpleFSLockFactory {
	ans := &SimpleFSLockFactory{}
	ans.FSLockFactory = newFSLockFactory()
	ans.setLockDir(path)
	return ans
}

func (f *SimpleFSLockFactory) Make(name string) Lock {
	if f.lockPrefix != "" {
		name = fmt.Sprintf("%v-%v", f.lockPrefix, name)
	}
	return newSimpleFSLock(f.lockDir, name)
}

func (f *SimpleFSLockFactory) Clear(name string) error {
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
	log.Printf("Opening %v...", name)
	d.EnsureOpen()
	fpath := filepath.Join(d.path, name)
	sin, err := newSimpleFSIndexInput(fmt.Sprintf("SimpleFSIndexInput(path='%v')", fpath),
		fpath, context, d.chunkSize)
	return sin, err
}

func (d *SimpleFSDirectory) CreateSlicer(name string, ctx IOContext) (slicer IndexInputSlicer, err error) {
	d.EnsureOpen()
	f, err := os.Open(filepath.Join(d.path, name))
	if err != nil {
		return nil, err
	}
	return &fileIndexInputSlicer{f, ctx, d.chunkSize}, nil
}

type fileIndexInputSlicer struct {
	file      *os.File
	ctx       IOContext
	chunkSize int
}

func (s *fileIndexInputSlicer) Close() error {
	return s.file.Close()
}

func (s *fileIndexInputSlicer) openSlice(desc string, offset, length int64) IndexInput {
	return newSimpleFSIndexInputFromFileSlice(fmt.Sprintf("SimpleFSIndexInput(%v in path='%v' slice=%v:%v)",
		desc, s.file.Name(), offset, offset+length),
		s.file, offset, length, bufferSize(s.ctx), s.chunkSize)
}

func (s *fileIndexInputSlicer) openFullSlice() IndexInput {
	fi, err := s.file.Stat()
	if err != nil {
		panic(err)
	}
	return s.openSlice("full-slice", 0, fi.Size())
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
	in.SeekReader = in
	return in, nil
}

func newSimpleFSIndexInputFromFileSlice(desc string, file *os.File, off, length int64, bufferSize, chunkSize int) *SimpleFSIndexInput {
	super := newFSIndexInputFromFileSlice(desc, file, off, length, bufferSize, chunkSize)
	ans := &SimpleFSIndexInput{super, &sync.Mutex{}}
	ans.SeekReader = ans
	return ans
}

func (in *SimpleFSIndexInput) readInternal(buf []byte) error {
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

func (in *SimpleFSIndexInput) seekInternal(pos int64) error {
	return nil // nothing
}

func (in *SimpleFSIndexInput) Clone() IndexInput {
	ans := &SimpleFSIndexInput{in.FSIndexInput.Clone().(*FSIndexInput), &sync.Mutex{}}
	ans.SeekReader = ans
	return ans
}
