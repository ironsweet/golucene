package store

import (
	"errors"
	"fmt"
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
		err = os.Mkdir(lock.dir, 0755)
		if err != nil { // IO error
			return
		}
	} else { // IO error
		return
	}
	var f *os.File
	if f, err = os.Create(lock.file); err == nil {
		fmt.Printf("File '%v' is created.\n", f.Name())
		ok = true
		defer f.Close()
	}
	return

}

func (lock *SimpleFSLock) Close() error {
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

func (d *SimpleFSDirectory) OpenInput(name string, context IOContext) (IndexInput, error) {
	d.EnsureOpen()
	fpath := filepath.Join(d.path, name)
	// fmt.Printf("Opening %v...\n", fpath)
	return newSimpleFSIndexInput(fmt.Sprintf("SimpleFSIndexInput(path='%v')", fpath), fpath, context)
}

// func (d *SimpleFSDirectory) CreateSlicer(name string, ctx IOContext) (slicer IndexInputSlicer, err error) {
// 	d.EnsureOpen()
// 	f, err := os.Open(filepath.Join(d.path, name))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &fileIndexInputSlicer{f, ctx, d.chunkSize}, nil
// }

// type fileIndexInputSlicer struct {
// 	file      *os.File
// 	ctx       IOContext
// 	chunkSize int
// }

// func (s *fileIndexInputSlicer) Close() error {
// 	err := s.file.Close()
// 	if err != nil {
// 		fmt.Printf("Closing %v failed: %v\n", s.file.Name(), err)
// 	}
// 	return err
// }

// func (s *fileIndexInputSlicer) OpenSlice(desc string, offset, length int64) IndexInput {
// 	return newSimpleFSIndexInputFromFileSlice(fmt.Sprintf("SimpleFSIndexInput(%v in path='%v' slice=%v:%v)",
// 		desc, s.file.Name(), offset, offset+length),
// 		s.file, offset, length, bufferSize(s.ctx), s.chunkSize)
// }

// func (s *fileIndexInputSlicer) OpenFullSlice() IndexInput {
// 	fi, err := s.file.Stat()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return s.OpenSlice("full-slice", 0, fi.Size())
// }

/*
The maximum chunk size is 8192 bytes, becaues Java RamdomAccessFile
mallocs a native buffer outside of stack if the read buffer size is
larger. GoLucene takes the same default value.
TODO: test larger value here
*/
const CHUNK_SIZE = 8192

type SimpleFSIndexInput struct {
	*BufferedIndexInput
	fileLock sync.Locker
	// the file channel we will read from
	file *os.File
	// is this instance a clone and hence does not own the file to close it
	isClone bool
	// start offset: non-zero in the slice case
	off int64
	// end offset (start+length)
	end int64
}

func newSimpleFSIndexInput(desc, path string, ctx IOContext) (*SimpleFSIndexInput, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fstat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	ans := new(SimpleFSIndexInput)
	ans.BufferedIndexInput = newBufferedIndexInput(ans, desc, ctx)
	ans.file = f
	ans.off = 0
	ans.end = fstat.Size()
	ans.fileLock = &sync.Mutex{}
	return ans, nil
}

func newSimpleFSIndexInputFromFileSlice(desc string, file *os.File, off, length int64, bufferSize int) *SimpleFSIndexInput {
	ans := new(SimpleFSIndexInput)
	ans.BufferedIndexInput = newBufferedIndexInputBySize(ans, desc, bufferSize)
	ans.file = file
	ans.off = off
	ans.end = off + length
	ans.isClone = true
	return ans
}

func (in *SimpleFSIndexInput) Close() error {
	if !in.isClone {
		return in.file.Close()
	}
	return nil
}

func (in *SimpleFSIndexInput) Clone() IndexInput {
	ans := &SimpleFSIndexInput{
		in.BufferedIndexInput.Clone(),
		in.fileLock,
		in.file,
		true,
		in.off,
		in.end,
	}
	ans.spi = ans
	return ans
}

func (in *SimpleFSIndexInput) Slice(desc string, offset, length int64) (IndexInput, error) {
	assert2(offset >= 0 && length >= 0 && offset+length <= in.Length(),
		"slice() %v out of bounds: %v", desc, in)
	ans := newSimpleFSIndexInputFromFileSlice(desc, in.file, in.off+offset, length, in.bufferSize)
	ans.fileLock = in.fileLock // share same file lock
	return ans, nil
}

func (in *SimpleFSIndexInput) Length() int64 {
	return in.end - in.off
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
		if CHUNK_SIZE < readLength {
			readLength = CHUNK_SIZE
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

func (in *SimpleFSIndexInput) seekInternal(pos int64) error { return nil }
