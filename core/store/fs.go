package store

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

type NoSuchDirectoryError struct {
	msg string
}

func newNoSuchDirectoryError(msg string) *NoSuchDirectoryError {
	return &NoSuchDirectoryError{msg}
}

func (err *NoSuchDirectoryError) Error() string {
	return err.msg
}

type FSDirectory struct {
	*DirectoryImpl
	path      string
	chunkSize int
}

// TODO support lock factory
func newFSDirectory(self Directory, path string) (d *FSDirectory, err error) {
	d = &FSDirectory{}
	d.DirectoryImpl = NewDirectoryImpl(self)
	d.path = path
	d.chunkSize = math.MaxInt32

	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return d, newNoSuchDirectoryError(fmt.Sprintf("file '%v' exists but is not a directory", path))
	}

	// TODO default to native lock factory
	d.SetLockFactory(NewSimpleFSLockFactory(path))
	return d, nil
}

func OpenFSDirectory(path string) (d Directory, err error) {
	// TODO support native implementations
	super, err := NewSimpleFSDirectory(path)
	if err != nil {
		return nil, err
	}
	return super, nil
}

func (d *FSDirectory) SetLockFactory(lockFactory LockFactory) error {
	d.DirectoryImpl.SetLockFactory(lockFactory)

	// for filesystem based LockFactory, delete the lockPrefix, if the locks are placed
	// in index dir. If no index dir is given, set ourselves
	if lf, ok := lockFactory.(*FSLockFactory); ok {
		if lf.lockDir == "" {
			lf.lockDir = d.path
			lf.lockPrefix = ""
		} else if lf.lockDir == d.path {
			lf.lockPrefix = ""
		}
	}
	return nil
}

func FSDirectoryListAll(path string) (paths []string, err error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, newNoSuchDirectoryError(fmt.Sprintf("directory '%v' does not exist", path))
	} else if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if !fi.IsDir() {
		return nil, newNoSuchDirectoryError(fmt.Sprintf("file '%v' exists but is not a directory", path))
	}

	// Exclude subdirs
	return f.Readdirnames(0)
}

func (d *FSDirectory) ListAll() (paths []string, err error) {
	d.ensureOpen()
	return FSDirectoryListAll(d.path)
}

func (d *FSDirectory) FileExists(name string) bool {
	d.ensureOpen()
	_, err := os.Stat(name)
	return err != nil
}

func (d *FSDirectory) LockID() string {
	d.ensureOpen()
	var digest int
	for _, ch := range d.path {
		digest = 31*digest + int(ch)
	}
	return fmt.Sprintf("lucene-%v", strconv.FormatUint(uint64(digest), 10))
}

type FSIndexInput struct {
	*BufferedIndexInput
	file      *os.File
	isClone   bool
	chunkSize int
	off       int64
	end       int64
}

func newFSIndexInput(desc, path string, context IOContext, chunkSize int) (in *FSIndexInput, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	super := newBufferedIndexInput(desc, context)
	ans := &FSIndexInput{super, f, false, chunkSize, 0, fi.Size()}
	ans.LengthCloser = ans
	return ans, nil
}

func newFSIndexInputFromFileSlice(desc string, f *os.File, off, length int64, bufferSize, chunkSize int) *FSIndexInput {
	super := newBufferedIndexInputBySize(desc, bufferSize)
	ans := &FSIndexInput{super, f, true, chunkSize, off, off + length}
	ans.LengthCloser = ans
	return ans
}

func (in *FSIndexInput) Length() int64 {
	return in.end - in.off
}

func (in *FSIndexInput) Close() error {
	// only close the file if this is not a clone
	if !in.isClone {
		in.file.Close()
	}
	return nil
}

func (in *FSIndexInput) Clone() IndexInput {
	clone := &FSIndexInput{
		in.BufferedIndexInput.Clone().(*BufferedIndexInput),
		in.file,
		true,
		in.chunkSize,
		in.off,
		in.end}
	clone.LengthCloser = clone
	return clone
}

func (in *FSIndexInput) String() string {
	return fmt.Sprintf("%v, off=%v, end=%v", in.BufferedIndexInput.String(), in.off, in.end)
}
