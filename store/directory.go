package store

import (
	"errors"
	"fmt"
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

type Directory struct {
	isOpen      bool
	lockFactory LockFactory
	LockID      func() string
	ListAll     func() (paths []string, err error)
	FileExists  func(name string) bool
	OpenInput   func(name string, context IOContext) (in IndexInput, err error)
}

func (d *Directory) SetLockFactory(lockFactory LockFactory) {
	d.lockFactory = lockFactory
	d.lockFactory.lockPrefix = d.LockID()
}

func (d *Directory) String() string {
	return fmt.Sprintf("Directory lockFactory=%v", d.lockFactory)
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

func (d *Directory) createSlicer(name string, context IOContext) (is IndexInputSlicer, err error) {
	d.ensureOpen()
	base, err := d.OpenInput(name, context)
	if err != nil {
		return nil, err
	}
	return simpleIndexInputSlicer{base}, nil
}

func (d *Directory) ensureOpen() {
	if !d.isOpen {
		panic("this Directory is closed")
	}
}

type FSDirectory struct {
	*Directory
	path      string
	chunkSize int
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

func newFSDirectory(path string) (d *FSDirectory, err error) {
	d = &FSDirectory{chunkSize: math.MaxInt32}
	if f, err := os.Open(path); err == nil {
		fi, err := f.Stat()
		if err != nil {
			return d, err
		}
		if !fi.IsDir() {
			return d, errors.New(fmt.Sprintf("file '%v' exists but is not a directory", path))
		}
	}

	super := Directory{ListAll: func() (paths []string, err error) {
		d.ensureOpen()
		return ListAll(d.path)
	}, FileExists: func(name string) bool {
		d.ensureOpen()
		if _, err := os.Stat(name); err != nil {
			return true
		}
		return false
	}, LockID: func() string {
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

func ListAll(path string) (paths []string, err error) {
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
