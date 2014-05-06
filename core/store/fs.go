package store

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
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
	sync.Locker
	path           string
	staleFiles     map[string]bool // synchronized, files written, but not yet sync'ed
	staleFilesLock *sync.RWMutex
	chunkSize      int
}

// TODO support lock factory
func newFSDirectory(self Directory, path string) (d *FSDirectory, err error) {
	d = &FSDirectory{
		DirectoryImpl:  NewDirectoryImpl(self),
		Locker:         &sync.Mutex{},
		path:           path,
		staleFiles:     make(map[string]bool),
		staleFilesLock: &sync.RWMutex{},
		chunkSize:      math.MaxInt32,
	}

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
	d.DirectoryImpl.SetLockFactory(lockFactory)

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
	_, err := os.Stat(name)
	return err != nil
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
	if os.IsExist(err) {
		err = os.Remove(filename)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot overwrite %v/%v: %v", d.path, name, err))
		}
	}
	return nil
}

func (d *FSDirectory) onIndexOutputClosed(io *FSIndexOutput) {
	d.staleFilesLock.Lock()
	defer d.staleFilesLock.Unlock()
	d.staleFiles[io.name] = true
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
	panic("not implemented yet")
}

func (d *FSDirectory) String() string {
	return fmt.Sprintf("FSDirectory@%v", d.DirectoryImpl.String())
}

type FSIndexInput struct {
	*BufferedIndexInput
	file      *os.File
	isClone   bool
	chunkSize int
	off       int64
	end       int64
}

func newFSIndexInput(desc, path string, context IOContext, chunkSize int) (*FSIndexInput, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	ans := &FSIndexInput{nil, f, false, chunkSize, 0, fi.Size()}
	ans.BufferedIndexInput = newBufferedIndexInput(ans, desc, context)
	return ans, nil
}

func newFSIndexInputFromFileSlice(desc string, f *os.File, off, length int64, bufferSize, chunkSize int) *FSIndexInput {
	ans := &FSIndexInput{nil, f, true, chunkSize, off, off + length}
	ans.BufferedIndexInput = newBufferedIndexInputBySize(ans, desc, bufferSize)
	return ans
}

func (in *FSIndexInput) Close() error {
	// only close the file if this is not a clone
	if !in.isClone {
		in.file.Close()
	}
	return nil
}

func (in *FSIndexInput) Clone() IndexInput {
	return &FSIndexInput{
		in.BufferedIndexInput.Clone(),
		in.file,
		true,
		in.chunkSize,
		in.off,
		in.end}
}

func (in *FSIndexInput) Length() int64 {
	return in.end - in.off
}

func (in *FSIndexInput) String() string {
	return fmt.Sprintf("%v, off=%v, end=%v", in.BufferedIndexInput.String(), in.off, in.end)
}

/*
The 'maximum' chunk size is 8192 bytes, because Lucene Java's malloc
limitation. In GoLucene, it's not required but not tested either.
*/
const CHUNK_SIZE = 8192

/*
Writes output with File.Write([]byte) (int, error)
*/
type FSIndexOutput struct {
	*BufferedIndexOutput
	parent *FSDirectory
	name   string
	file   *os.File
	isOpen bool // volatile
}

func newFSIndexOutput(parent *FSDirectory, name string) (*FSIndexOutput, error) {
	file, err := os.OpenFile(filepath.Join(parent.path, name), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}
	out := &FSIndexOutput{
		parent: parent,
		name:   name,
		file:   file,
		isOpen: true,
	}
	out.BufferedIndexOutput = NewBufferedIndexOutput(CHUNK_SIZE, out)
	return out, nil
}

func (out *FSIndexOutput) FlushBuffer(b []byte) error {
	assert(out.isOpen)
	offset, size := 0, len(b)
	for size > 0 {
		toWrite := CHUNK_SIZE
		if size < toWrite {
			toWrite = size
		}
		_, err := out.file.Write(b[offset : offset+toWrite])
		if err != nil {
			return err
		}
		offset += toWrite
		size -= toWrite
	}
	assert(size == 0)
	return nil
}

func (out *FSIndexOutput) Close() error {
	out.parent.onIndexOutputClosed(out)
	// only close the file if it has not been closed yet
	if out.isOpen {
		var err error
		defer func() {
			out.isOpen = false
			util.CloseWhileHandlingError(err, out.file)
		}()
		err = out.BufferedIndexOutput.Close()
	}
	return nil
}

func (out *FSIndexOutput) Length() (int64, error) {
	info, err := out.file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
