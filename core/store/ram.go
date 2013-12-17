package store

import (
	"sync"
)

// store/RAMDirectory.java

/*
A memory-resident Directory implementation. Locking implementation
is by default the SingleInstanceLockFactory but can be changed with
SetLockFactory().

Warning: This class is not intended to work with huge indexes.
Everything beyond several hundred megabytes will waste resources (GC
cycles), becaues it uses an internal buffer size of 1024 bytes,
producing millions of byte[1024] arrays. This class is optimized for
small memory-resident indexes. It also has bad concurrency on
multithreaded environments.

It is recommended to materialze large indexes on disk and use
MMapDirectory, which is a high-performance directory implementation
working diretly on the file system cache of the operating system, so
copying dat to Java heap space is not useful.
*/
type RAMDirectory struct {
	*DirectoryImpl

	fileMap     map[string]RAMFile // synchronized
	sizeInBytes int64              // synchronized
}

func NewRAMDirectory() *RAMDirectory {
	ans := &RAMDirectory{}
	ans.DirectoryImpl = NewDirectoryImpl(ans)
	ans.SetLockFactory(newSingleInstanceLockFactory())
	return ans
}

func (rd *RAMDirectory) ListAll() (names []string, err error) {
	panic("not implemented yet")
}

// Returns true iff the named file exists in this directory
func (rd *RAMDirectory) FileExists(name string) bool {
	panic("not implemented yet")
}

// Returns the length in bytes of a file in the directory.
func (rd *RAMDirectory) FileLength(name string) (length int64, err error) {
	panic("not implemented yet")
}

// Removes an existing file in the directory
func (rd *RAMDirectory) DeleteFile(name string) error {
	panic("not implemented yet")
}

// Creates a new, empty file in the directory with the given name.
// Returns a stream writing this file:
func (rd *RAMDirectory) CreateOutput(name string, context IOContext) (out IndexOutput, err error) {
	panic("not implemented yet")
}

// Returns a new RAMFile for storing data. This method can be
// overridden to return different RAMFile impls, that e.g. override
// RAMFile.newBuffer(int).
func (rd *RAMDirectory) newRAMFile() *RAMFile {
	return newRAMFile(rd)
}

// Returns a stream reading an existing file.
func (rd *RAMDirectory) OpenInput(name string, context IOContext) (in IndexInput, err error) {
	panic("not implemented yet")
}

// Closes the store to future operations, releasing associated memroy.
func (rd *RAMDirectory) Close() error {
	rd.IsOpen = false
	panic("not implemented yet")
}

// store/RAMFile.java

// Represents a file in RAM as a list of []byte buffers.
type RAMFile struct {
	sync.Locker
	buffers     [][]byte
	length      int
	directory   *RAMDirectory
	sizeInBytes int64
	newBuffer   func(size int) []byte
}

func newRAMFileBuffer() *RAMFile {
	return &RAMFile{
		Locker:    &sync.Mutex{},
		newBuffer: newBuffer,
	}
}

func newRAMFile(directory *RAMDirectory) *RAMFile {
	return &RAMFile{
		Locker:    &sync.Mutex{},
		directory: directory,
		newBuffer: newBuffer,
	}
}

func (rf *RAMFile) SetLength(length int) {
	rf.Lock()
	defer rf.Unlock()
	rf.length = length
}

func (rf *RAMFile) addBuffer(size int) []byte {
	panic("not implemented yet")
}

func (rf *RAMFile) Buffer(index int) []byte {
	rf.Lock()
	defer rf.Unlock()
	return rf.buffers[index]
}

func (rf *RAMFile) numBuffers() int {
	rf.Lock()
	defer rf.Unlock()
	return len(rf.buffers)
}

// Expert: allocate a new buffer.
// Subclasses can allocate differently
func newBuffer(size int) []byte {
	return make([]byte, size)
}

// store/SingleInstanceLockFactory.java

/*
Implements LockFactory for a single in-process instance, meaning all
locking will take place through this one instance. Only use this
LockFactory when you are certain all IndexReaders and IndexWriters
for a given index are running against a single shared in-process
Directory instance. This is currently the default locking for
RAMDirectory.
*/
type SingleInstanceLockFactory struct {
	*LockFactoryImpl
	locks map[string]bool
}

func newSingleInstanceLockFactory() *SingleInstanceLockFactory {
	return &SingleInstanceLockFactory{LockFactoryImpl: &LockFactoryImpl{}}
}

func (fac *SingleInstanceLockFactory) Make(name string) Lock {
	panic("not implemented yet")
}

func (fac *SingleInstanceLockFactory) Clear(name string) error {
	panic("not implemented yet")
}
