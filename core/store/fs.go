package store

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	// "time"
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

type FSDirectorySPI interface {
	OpenInput(string, IOContext) (IndexInput, error)
}

type FSDirectory struct {
	*DirectoryImpl
	*BaseDirectory
	FSDirectorySPI
	sync.Locker
	path           string
	staleFiles     map[string]bool // synchronized, files written, but not yet sync'ed
	staleFilesLock *sync.RWMutex
	chunkSize      int
}

// TODO support lock factory
func newFSDirectory(spi FSDirectorySPI, path string) (d *FSDirectory, err error) {
	d = &FSDirectory{
		Locker:         &sync.Mutex{},
		path:           path,
		staleFiles:     make(map[string]bool),
		staleFilesLock: &sync.RWMutex{},
		chunkSize:      math.MaxInt32,
	}
	d.DirectoryImpl = NewDirectoryImpl(d)
	d.BaseDirectory = NewBaseDirectory(d)
	d.FSDirectorySPI = spi

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

func (d *FSDirectory) SetLockFactory(lockFactory LockFactory) {
	d.BaseDirectory.SetLockFactory(lockFactory)

	// for filesystem based LockFactory, delete the lockPrefix, if the locks are placed
	// in index dir. If no index dir is given, set ourselves
	// TODO change FSDirectory to interface
	if lf, ok := lockFactory.(*SimpleFSLockFactory); ok {
		if lf.lockDir == "" {
			lf.lockDir = d.path
			lf.lockPrefix = ""
		} else if lf.lockDir == d.path {
			lf.lockPrefix = ""
		}
	}
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
	d.EnsureOpen()
	return FSDirectoryListAll(d.path)
}

func (d *FSDirectory) FileExists(name string) bool {
	d.EnsureOpen()
	_, err := os.Stat(filepath.Join(d.path, name))
	return err == nil || os.IsExist(err)
}

// Returns the length in bytes of a file in the directory.
func (d *FSDirectory) FileLength(name string) (n int64, err error) {
	d.EnsureOpen()
	fi, err := os.Stat(filepath.Join(d.path, name))
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// Removes an existing file in the directory.
func (d *FSDirectory) DeleteFile(name string) (err error) {
	d.EnsureOpen()
	if err = os.Remove(filepath.Join(d.path, name)); err == nil {
		d.staleFilesLock.Lock()
		defer d.staleFilesLock.Unlock()
		delete(d.staleFiles, name)
	}
	return
}

/*
Creates an IndexOutput for the file with the given name.
*/
func (d *FSDirectory) CreateOutput(name string, ctx IOContext) (out IndexOutput, err error) {
	d.EnsureOpen()
	err = d.ensureCanWrite(name)
	if err != nil {
		return nil, err
	}
	return newFSIndexOutput(d, name)
}

func (d *FSDirectory) ensureCanWrite(name string) error {
	err := os.MkdirAll(d.path, os.ModeDir|0660)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot create directory %v: %v", d.path, err))
	}

	filename := filepath.Join(d.path, name)
	_, err = os.Stat(filename)
	if err == nil || os.IsExist(err) {
		err = os.Remove(filename)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot overwrite %v/%v: %v", d.path, name, err))
		}
	}
	return nil
}

/*
Sub classes should call this method on closing an open IndexOuput,
reporting the name of the file that was closed. FSDirectory needs
this information to take care of syncing stale files.
*/
func (d *FSDirectory) onIndexOutputClosed(name string) {
	d.staleFilesLock.Lock()
	defer d.staleFilesLock.Unlock()
	d.staleFiles[name] = true
}

func (d *FSDirectory) Sync(names []string) (err error) {
	d.EnsureOpen()

	toSync := make(map[string]bool)
	d.staleFilesLock.RLock()
	for _, name := range names {
		if _, ok := d.staleFiles[name]; ok {
			continue
		}
		toSync[name] = true
	}
	d.staleFilesLock.RUnlock()

	for name, _ := range toSync {
		err = d.fsync(name)
		if err != nil {
			return err
		}
	}

	// fsync the directory itself, but only if there was any file
	// fsynced before (otherwise it can happen that the directory does
	// not yet exist)!
	if len(toSync) > 0 {
		err = util.Fsync(d.path, true)
		if err != nil {
			return err
		}
	}

	for name, _ := range toSync {
		delete(d.staleFiles, name)
	}
	return
}

func (d *FSDirectory) LockID() string {
	d.EnsureOpen()
	var digest int
	for _, ch := range d.path {
		digest = 31*digest + int(ch)
	}
	return fmt.Sprintf("lucene-%v", strconv.FormatUint(uint64(digest), 10))
}

func (d *FSDirectory) Close() error {
	d.Lock() // synchronized
	defer d.Unlock()
	d.IsOpen = false
	return nil
}

func (d *FSDirectory) fsync(name string) error {
	// var err, err2 error
	// var success = false
	// var retryCount = 0
	// for !success && retryCount < 5 {
	// 	retryCount++
	// 	if err2, success = func() (error, bool) {
	// 		file, err := os.Open(filepath.Join(d.path, name))
	// 		if err != nil {
	// 			return err, false
	// 		}
	// 		defer file.Close()
	// 		return file.Sync(), true
	// 	}(); err2 != nil {
	// 		if err == nil {
	// 			err = err2
	// 		}
	// 		time.Sleep(5 * time.Millisecond)
	// 	}
	// }
	// return err

	// Go's fsync doesn't work the same way as Java
	// disabled since I got "fsync: Access is denied" all the time
	return nil
}

func (d *FSDirectory) String() string {
	return fmt.Sprintf("FSDirectory@%v", d.DirectoryImpl.String())
}

// type FSIndexInput struct {
// 	*BufferedIndexInput
// 	file      *os.File
// 	isClone   bool
// 	chunkSize int
// 	off       int64
// 	end       int64
// }

// func newFSIndexInput(desc, path string, context IOContext, chunkSize int) (*FSIndexInput, error) {
// 	f, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fi, err := f.Stat()
// 	if err != nil {
// 		return nil, err
// 	}
// 	ans := &FSIndexInput{nil, f, false, chunkSize, 0, fi.Size()}
// 	ans.BufferedIndexInput = newBufferedIndexInput(ans, desc, context)
// 	return ans, nil
// }

// func newFSIndexInputFromFileSlice(desc string, f *os.File, off, length int64, bufferSize, chunkSize int) *FSIndexInput {
// 	ans := &FSIndexInput{nil, f, true, chunkSize, off, off + length}
// 	ans.BufferedIndexInput = newBufferedIndexInputBySize(ans, desc, bufferSize)
// 	return ans
// }

// func (in *FSIndexInput) Close() error {
// 	// only close the file if this is not a clone
// 	if !in.isClone {
// 		in.file.Close()
// 	}
// 	return nil
// }

// func (in *FSIndexInput) Clone() IndexInput {
// 	return &FSIndexInput{
// 		in.BufferedIndexInput.Clone(),
// 		in.file,
// 		true,
// 		in.chunkSize,
// 		in.off,
// 		in.end}
// }

// func (in *FSIndexInput) Length() int64 {
// 	return in.end - in.off
// }

// func (in *FSIndexInput) String() string {
// 	return fmt.Sprintf("%v, off=%v, end=%v", in.BufferedIndexInput.String(), in.off, in.end)
// }

type FilteredWriteCloser struct {
	io.WriteCloser
	f func(p []byte) (int, error)
}

func filter(w io.WriteCloser, f func(p []byte) (int, error)) *FilteredWriteCloser {
	return &FilteredWriteCloser{w, f}
}

func (w *FilteredWriteCloser) Write(p []byte) (int, error) {
	return w.f(p)
}

/*
Writes output with File.Write([]byte) (int, error)
*/
type FSIndexOutput struct {
	*OutputStreamIndexOutput
	*FSDirectory
	name string
}

func newFSIndexOutput(parent *FSDirectory, name string) (*FSIndexOutput, error) {
	file, err := os.OpenFile(filepath.Join(parent.path, name), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}
	return &FSIndexOutput{
		newOutputStreamIndexOutput(filter(file, func(p []byte) (int, error) {
			// This implementation ensures, that we never write more than CHUNK_SIZE bytes:
			chunk := CHUNK_SIZE
			offset, length := 0, len(p)
			for length > 0 {
				if length < chunk {
					chunk = length
				}
				n, err := file.Write(p[offset : offset+chunk])
				if err != nil {
					return offset, err
				}
				length -= n
				offset += n
			}
			return offset, nil
		}), CHUNK_SIZE),
		parent,
		name,
	}, nil
}

func (out *FSIndexOutput) Close() (err error) {
	defer func() {
		err = out.OutputStreamIndexOutput.Close()
	}()
	out.onIndexOutputClosed(out.name)
	return nil
}
