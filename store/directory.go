package store

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

const (
	IO_CONTEXT_TYPE_MERGE   = 1
	IO_CONTEXT_TYPE_READ    = 2
	IO_CONTEXT_TYPE_FLUSH   = 4
	IO_CONTEXT_TYPE_DEFAULT = 8
)

type IOContextType int

var (
	IO_CONTEXT_DEFAULT  = NewIOContextFromType(IOContextType(IO_CONTEXT_TYPE_DEFAULT))
	IO_CONTEXT_READONCE = NewIOContextBool(true)
	IO_CONTEXT_READ     = NewIOContextBool(false)
)

type IOContext struct {
	context IOContextType
	// mergeInfo MergeInfo
	// flushInfo FlushInfo
	readOnce bool
}

func NewIOContextForFlush(flushInfo FlushInfo) IOContext {
	return IOContext{IOContextType(IO_CONTEXT_TYPE_FLUSH), false}
}

func NewIOContextFromType(context IOContextType) IOContext {
	return IOContext{context, false}
}

func NewIOContextBool(readOnce bool) IOContext {
	return IOContext{IOContextType(IO_CONTEXT_TYPE_READ), readOnce}
}

func NewIOContextForMerge(mergeInfo MergeInfo) IOContext {
	return IOContext{IOContextType(IO_CONTEXT_TYPE_MERGE), false}
}

type FlushInfo struct {
	numDocs              int
	estimatedSegmentSize int64
}

type MergeInfo struct {
	totalDocCount       int
	estimatedMergeBytes int64
	isExternal          bool
	mergeMaxNumSegments int
}

type Lock struct {
	self interface{}
}

type LockFactory interface {
	make(name string) Lock
	clear(name string) error
	setLockPrefix(prefix string)
	getLockPrefix() string
}

type LockFactoryImpl struct {
	lockPrefix string
}

func (f *LockFactoryImpl) setLockPrefix(prefix string) {
	f.lockPrefix = prefix
}

func (f *LockFactoryImpl) getLockPrefix() string {
	return f.lockPrefix
}

type FSLockFactory struct {
	*LockFactoryImpl
	lockDir string // can not be set twice
}

func newFSLockFactory() *FSLockFactory {
	ans := &FSLockFactory{}
	ans.LockFactoryImpl = &LockFactoryImpl{}
	return ans
}

func (f *FSLockFactory) setLockDir(lockDir string) {
	if f.lockDir != "" {
		panic("You can set the lock directory for this factory only once.")
	}
	f.lockDir = lockDir
}

func (f *FSLockFactory) getLockDir() string {
	return f.lockDir
}

func (f *FSLockFactory) clear(name string) error {
	panic("invalid")
}

func (f *FSLockFactory) make(name string) Lock {
	panic("invalid")
}

type Directory interface {
	io.Closer
	// Files related methods
	listAll() (paths []string, err error)
	fileExists(name string) bool
	// DeleteFile(name string) error
	// FileLength(name string) int64
	// CreateOutput(name string, ctx, IOContext) (out IndexOutput, err error)
	// Sync(names []string) error
	openInput(name string, context IOContext) (in IndexInput, err error)
	// Locks related methods
	makeLock(name string) Lock
	clearLock(name string) error
	setLockFactory(lockFactory LockFactory) error
	getLockFactory() LockFactory
	getLockID() string
	// Utilities
	// Copy(to Directory, src, dest string, ctx IOContext) error
	// Experimental methods
	createSlicer(name string, ctx IOContext) (slicer IndexInputSlicer, err error)
	// Private methods
	ensureOpen()
}

type DirectoryImpl struct {
	Directory
	isOpen      bool
	lockFactory LockFactory
}

func newDirectoryImpl(self Directory) *DirectoryImpl {
	return &DirectoryImpl{Directory: self, isOpen: true}
}

func (d *DirectoryImpl) makeLock(name string) Lock {
	return d.lockFactory.make(name)
}

func (d *DirectoryImpl) clearLock(name string) error {
	if d.lockFactory != nil {
		return d.lockFactory.clear(name)
	}
	return nil
}

func (d *DirectoryImpl) setLockFactory(lockFactory LockFactory) error {
	// assert lockFactory != nil
	d.lockFactory = lockFactory
	d.lockFactory.setLockPrefix(d.getLockID())
	return nil
}

func (d *DirectoryImpl) getLockFactory() LockFactory {
	return d.lockFactory
}

func (d *DirectoryImpl) getLockID() string {
	return d.String()
}

func (d *DirectoryImpl) String() string {
	return fmt.Sprintf("Directory lockFactory=%v", d.lockFactory)
}

func (d *DirectoryImpl) createSlicer(name string, context IOContext) (is IndexInputSlicer, err error) {
	d.ensureOpen()
	base, err := d.Directory.openInput(name, context)
	if err != nil {
		return nil, err
	}
	return simpleIndexInputSlicer{base}, nil
}

func (d *DirectoryImpl) ensureOpen() {
	if !d.isOpen {
		panic("this Directory is closed")
	}
}

type IndexInputSlicer interface {
	io.Closer
	openSlice(desc string, offset, length int64) IndexInput
	openFullSlice() IndexInput
}

type simpleIndexInputSlicer struct {
	base IndexInput
}

func (is simpleIndexInputSlicer) openSlice(desc string, offset, length int64) IndexInput {
	return newSlicedIndexInput(fmt.Sprintf("SlicedIndexInput(%v in %v)", desc, is.base),
		is.base, offset, length).BufferedIndexInput
}

func (is simpleIndexInputSlicer) Close() error {
	return is.base.Close()
}

func (is simpleIndexInputSlicer) openFullSlice() IndexInput {
	return is.base
}

type SlicedIndexInput struct {
	*BufferedIndexInput
	base       IndexInput
	fileOffset int64
	length     int64
}

func newSlicedIndexInput(desc string, base IndexInput, fileOffset, length int64) SlicedIndexInput {
	return newSlicedIndexInputBySize(desc, base, fileOffset, length, BUFFER_SIZE)
}

func newSlicedIndexInputBySize(desc string, base IndexInput, fileOffset, length int64, bufferSize int) SlicedIndexInput {
	return SlicedIndexInput{
		BufferedIndexInput: newBufferedIndexInputBySize(fmt.Sprintf(
			"SlicedIndexInput(%v in %v slice=%v:%v)", desc, base, fileOffset, fileOffset+length), bufferSize),
		base: base, fileOffset: fileOffset, length: length}
}

type FSDirectory struct {
	*DirectoryImpl
	path      string
	chunkSize int
}

// TODO support lock factory
func newFSDirectory(self Directory, path string) (d *FSDirectory, err error) {
	d = &FSDirectory{}
	d.DirectoryImpl = newDirectoryImpl(self)
	d.path = path
	d.chunkSize = math.MaxInt32

	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return d, errors.New(fmt.Sprintf("file '%v' exists but is not a directory", path))
	}

	// TODO default to native lock factory
	d.setLockFactory(NewSimpleFSLockFactory(path))
	return d, nil
}

func OpenFSDirectory(path string) (d FSDirectory, err error) {
	// TODO support native implementations
	super, err := NewSimpleFSDirectory(path)
	if err != nil {
		return d, err
	}
	return *(super.FSDirectory), nil
}

func (d *FSDirectory) setLockFactory(lockFactory LockFactory) error {
	d.DirectoryImpl.setLockFactory(lockFactory)

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
		return nil, errors.New(fmt.Sprintf("directory '%v' does not exist", path))
	} else if err != nil {
		panic(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if !fi.IsDir() {
		return nil, errors.New(fmt.Sprintf("file '%v' exists but is not a directory", path))
	}

	// Exclude subdirs
	return f.Readdirnames(0)
}

func (d *FSDirectory) listAll() (paths []string, err error) {
	d.ensureOpen()
	return FSDirectoryListAll(d.path)
}

func (d *FSDirectory) fileExists(name string) bool {
	d.ensureOpen()
	_, err := os.Stat(name)
	return err != nil
}

func (d *FSDirectory) getLockID() string {
	d.ensureOpen()
	var digest int
	for _, ch := range d.path {
		digest = 31*digest + int(ch)
	}
	return fmt.Sprintf("lucene-%v", strconv.FormatUint(uint64(digest), 10))
}
