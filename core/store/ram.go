package store

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
	"hash"
	"hash/crc32"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"
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
	*BaseDirectory

	sizeInBytes int64 // synchronized

	fileMap     map[string]*RAMFile // synchronized
	fileMapLock *sync.RWMutex
}

func NewRAMDirectory() *RAMDirectory {
	ans := &RAMDirectory{
		fileMap:     make(map[string]*RAMFile),
		fileMapLock: &sync.RWMutex{},
	}
	ans.DirectoryImpl = NewDirectoryImpl(ans)
	ans.BaseDirectory = NewBaseDirectory(ans)
	ans.SetLockFactory(newSingleInstanceLockFactory())
	return ans
}

func (d *RAMDirectory) LockID() string {
	return fmt.Sprintf("lucene-%v", util.ItoHex(int64(uintptr(unsafe.Pointer(&d)))))
}

func (rd *RAMDirectory) ListAll() (names []string, err error) {
	rd.EnsureOpen()
	rd.fileMapLock.RLock()
	defer rd.fileMapLock.RUnlock()
	names = make([]string, 0, len(rd.fileMap))
	for name, _ := range rd.fileMap {
		names = append(names, name)
	}
	return names, nil
}

// Returns true iff the named file exists in this directory
func (rd *RAMDirectory) FileExists(name string) bool {
	rd.EnsureOpen()
	rd.fileMapLock.RLock()
	defer rd.fileMapLock.RUnlock()
	_, ok := rd.fileMap[name]
	return ok
}

// Returns the length in bytes of a file in the directory.
func (rd *RAMDirectory) FileLength(name string) (length int64, err error) {
	rd.EnsureOpen()
	rd.fileMapLock.RLock()
	defer rd.fileMapLock.RUnlock()
	if file, ok := rd.fileMap[name]; ok {
		return file.Length(), nil
	}
	return 0, os.ErrNotExist
}

/*
Return total size in bytes of all files in this directory. This is
currently quantized to BUFFER_SIZE.
*/
func (rd *RAMDirectory) RamBytesUsed() int64 {
	rd.EnsureOpen()
	return atomic.LoadInt64(&rd.sizeInBytes)
}

/* Removes an existing file in the directory */
func (rd *RAMDirectory) DeleteFile(name string) error {
	rd.EnsureOpen()
	rd.fileMapLock.RLock()
	defer rd.fileMapLock.RUnlock()
	if file, ok := rd.fileMap[name]; ok {
		file.directory = nil
		atomic.AddInt64(&rd.sizeInBytes, -file.sizeInBytes)
		return nil
	}
	return errors.New(name)
}

// Creates a new, empty file in the directory with the given name.
// Returns a stream writing this file:
func (rd *RAMDirectory) CreateOutput(name string, context IOContext) (out IndexOutput, err error) {
	rd.EnsureOpen()
	file := rd.newRAMFile()
	rd.fileMapLock.Lock()
	defer rd.fileMapLock.Unlock()
	if existing, ok := rd.fileMap[name]; ok {
		atomic.AddInt64(&rd.sizeInBytes, -existing.sizeInBytes)
		existing.directory = nil
	}
	rd.fileMap[name] = file
	return NewRAMOutputStream(file, true), nil
}

// Returns a new RAMFile for storing data. This method can be
// overridden to return different RAMFile impls, that e.g. override
// RAMFile.newBuffer(int).
func (rd *RAMDirectory) newRAMFile() *RAMFile {
	return NewRAMFile(rd)
}

func (rd *RAMDirectory) Sync(names []string) error {
	return nil
}

// Returns a stream reading an existing file.
func (rd *RAMDirectory) OpenInput(name string, context IOContext) (in IndexInput, err error) {
	rd.EnsureOpen()
	if file, ok := rd.fileMap[name]; ok {
		return newRAMInputStream(name, file)
	}
	return nil, errors.New(name)
}

// Closes the store to future operations, releasing associated memroy.
func (rd *RAMDirectory) Close() error {
	rd.IsOpen = false
	rd.fileMapLock.Lock()
	defer rd.fileMapLock.Unlock()
	rd.fileMap = make(map[string]*RAMFile)
	return nil
}

/* test-only */
func (rd *RAMDirectory) GetRAMFile(name string) *RAMFile {
	rd.fileMapLock.Lock()
	defer rd.fileMapLock.Unlock()
	return rd.fileMap[name]
}

/* test-only */
func (d *RAMDirectory) PutRAMFile(name string, file *RAMFile) {
	d.fileMapLock.Lock()
	defer d.fileMapLock.Unlock()
	d.fileMap[name] = file
}

/* test-only */
func (rd *RAMDirectory) ChangeSize(diff int64) {
	atomic.AddInt64(&rd.sizeInBytes, diff)
}

func (rd *RAMDirectory) String() string {
	return fmt.Sprintf("RAMDirectory@%v", rd.DirectoryImpl.String())
}

// store/RAMFile.java

// Represents a file in RAM as a list of []byte buffers.
type RAMFile struct {
	sync.Locker
	buffers     [][]byte
	length      int64
	directory   *RAMDirectory
	sizeInBytes int64
	newBuffer   func(size int) []byte
}

func NewRAMFileBuffer() *RAMFile {
	return &RAMFile{
		Locker:    &sync.Mutex{},
		newBuffer: newBuffer,
	}
}

func NewRAMFile(directory *RAMDirectory) *RAMFile {
	return &RAMFile{
		Locker:    &sync.Mutex{},
		directory: directory,
		newBuffer: newBuffer,
	}
}

func (rf *RAMFile) Length() int64 {
	rf.Lock()
	defer rf.Unlock()
	return rf.length
}

func (rf *RAMFile) SetLength(length int64) {
	rf.Lock() // synchronized
	defer rf.Unlock()
	rf.length = length
}

func (rf *RAMFile) addBuffer(size int) []byte {
	buffer := rf.newBuffer(size)
	rf.Lock() // synchronized
	defer rf.Unlock()
	rf.buffers = append(rf.buffers, buffer)
	rf.sizeInBytes += int64(size)

	if rf.directory != nil {
		atomic.AddInt64(&rf.directory.sizeInBytes, int64(size))
	}
	return buffer
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

func (rf *RAMFile) RamBytesUsed() int64 {
	rf.Lock()
	defer rf.Unlock()
	return rf.sizeInBytes
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
	locksLock sync.Locker
	locks     map[string]bool
}

func newSingleInstanceLockFactory() *SingleInstanceLockFactory {
	return &SingleInstanceLockFactory{
		LockFactoryImpl: &LockFactoryImpl{},
		locksLock:       &sync.Mutex{},
		locks:           make(map[string]bool),
	}
}

func (fac *SingleInstanceLockFactory) Make(name string) Lock {
	// We do not use the LockPrefix at all, becaues the private map
	// instance effectively scopes the locking to this single Directory
	// instance.
	return newSingleInstanceLock(fac.locks, fac.locksLock, name)
}

func (fac *SingleInstanceLockFactory) Clear(name string) error {
	fac.locksLock.Lock() // synchronized
	defer fac.locksLock.Unlock()
	if _, ok := fac.locks[name]; ok {
		delete(fac.locks, name)
	}
	return nil
}

func (fac *SingleInstanceLockFactory) String() string {
	return fmt.Sprintf("SingleInstanceLockFactory@%v", fac.locks)
}

type SingleInstanceLock struct {
	*LockImpl
	name      string
	locksLock sync.Locker
	locks     map[string]bool
}

func newSingleInstanceLock(locks map[string]bool, locksLock sync.Locker, name string) *SingleInstanceLock {
	ans := &SingleInstanceLock{
		name:      name,
		locksLock: locksLock,
		locks:     locks,
	}
	ans.LockImpl = NewLockImpl(ans)
	return ans
}

func (lock *SingleInstanceLock) Obtain() (ok bool, err error) {
	lock.locksLock.Lock() // synchronized
	defer lock.locksLock.Unlock()
	lock.locks[lock.name] = true
	return true, nil
}

func (lock *SingleInstanceLock) Close() error {
	lock.locksLock.Lock() // synchronized
	defer lock.locksLock.Unlock()
	delete(lock.locks, lock.name)
	return nil
}

func (lock *SingleInstanceLock) IsLocked() bool {
	lock.locksLock.Lock() // synchronized
	defer lock.locksLock.Unlock()
	_, ok := lock.locks[lock.name]
	return ok
}

func (lock *SingleInstanceLock) String() string {
	return fmt.Sprintf("SingleInstanceLock: %v", lock.name)
}

// store/RAMInputStream.java

// A memory-resident IndexInput implementation.
type RAMInputStream struct {
	*IndexInputImpl

	file   *RAMFile
	length int64

	currentBuffer      []byte
	currentBufferIndex int

	bufferPosition int
	bufferStart    int64
	bufferLength   int
}

func newRAMInputStream(name string, f *RAMFile) (in *RAMInputStream, err error) {
	if !(f.length/BUFFER_SIZE < math.MaxInt32) {
		return nil, errors.New(fmt.Sprintf("RAMInputStream too large length=%v: %v", f.length, name))
	}

	in = &RAMInputStream{
		file:               f,
		length:             int64(f.length),
		currentBufferIndex: -1,
	}
	in.IndexInputImpl = NewIndexInputImpl(fmt.Sprintf("RAMInputStream(name=%v)", name), in)
	return in, nil
}

func (in *RAMInputStream) Close() error {
	return nil
}

func (in *RAMInputStream) Length() int64 {
	return in.length
}

func (in *RAMInputStream) ReadByte() (byte, error) {
	if in.bufferPosition >= in.bufferLength {
		in.currentBufferIndex++
		err := in.switchCurrentBuffer(true)
		if err != nil {
			return 0, err
		}
	}
	in.bufferPosition++
	return in.currentBuffer[in.bufferPosition-1], nil
}

func (in *RAMInputStream) ReadBytes(buf []byte) error {
	var offset = 0
	for limit := len(buf); limit > 0; {
		if in.bufferPosition >= in.bufferLength {
			in.currentBufferIndex++
			err := in.switchCurrentBuffer(true)
			if err != nil {
				return err
			}
		}

		bytesToCopy := in.bufferLength - in.bufferPosition
		if limit < bytesToCopy {
			bytesToCopy = limit
		}
		copy(buf[offset:], in.currentBuffer[in.bufferPosition:in.bufferPosition+bytesToCopy])
		offset += bytesToCopy
		limit -= bytesToCopy
		in.bufferPosition += bytesToCopy
	}
	return nil
}

func (in *RAMInputStream) switchCurrentBuffer(enforceEOF bool) error {
	in.bufferStart = int64(BUFFER_SIZE * in.currentBufferIndex)
	if in.bufferStart > in.length || in.currentBufferIndex >= in.file.numBuffers() {
		// end of file reached, no more buffer left
		if enforceEOF {
			return errors.New(fmt.Sprintf("read past EOF: %v", in))
		}
		// Force EOF if a read takes place at this position
		in.currentBufferIndex--
		in.bufferPosition = BUFFER_SIZE
	} else {
		in.currentBuffer = in.file.Buffer(in.currentBufferIndex)
		in.bufferPosition = 0
		bufLen := in.length - in.bufferStart
		if BUFFER_SIZE < bufLen {
			bufLen = BUFFER_SIZE
		}
		in.bufferLength = int(bufLen)
	}
	return nil
}

func (in *RAMInputStream) FilePointer() int64 {
	if in.currentBufferIndex < 0 {
		return 0
	}
	return in.bufferStart + int64(in.bufferPosition)
}

func (in *RAMInputStream) Seek(pos int64) error {
	if in.currentBuffer == nil || pos < in.bufferStart || pos >= in.bufferStart+BUFFER_SIZE {
		in.currentBufferIndex = int(pos / BUFFER_SIZE)
		err := in.switchCurrentBuffer(false)
		if err != nil {
			return err
		}
	}
	in.bufferPosition = int(pos % BUFFER_SIZE)
	return nil
}

func (in *RAMInputStream) Slice(desc string, offset, length int64) (IndexInput, error) {
	panic("not implemented yet")
}

func (in *RAMInputStream) Clone() IndexInput {
	panic("not implemented yet")
}

func (in *RAMInputStream) String() string {
	return fmt.Sprintf("%v;%v@[0-%v]", in.IndexInputImpl.String(), in.FilePointer(), in.length)
}

// store/RamOutputStream.java

/*
A memory-resident IndexOutput implementation
*/
type RAMOutputStream struct {
	*IndexOutputImpl

	file *RAMFile

	currentBuffer      []byte
	currentBufferIndex int

	bufferPosition int
	bufferStart    int64
	bufferLength   int

	crc hash.Hash32
}

/* Construct an empty output buffer. */
func NewRAMOutputStreamBuffer() *RAMOutputStream {
	return NewRAMOutputStream(NewRAMFileBuffer(), false)
}

func NewRAMOutputStream(f *RAMFile, checksum bool) *RAMOutputStream {
	// make sure that we switch to the first needed buffer lazily
	out := &RAMOutputStream{file: f, currentBufferIndex: -1}
	out.IndexOutputImpl = NewIndexOutput(out)
	if checksum {
		out.crc = newBufferedChecksum(crc32.NewIEEE())
	}
	return out
}

/* Copy the current contents of this buffer to the named output. */
func (out *RAMOutputStream) WriteTo(output util.DataOutput) error {
	err := out.Flush()
	if err != nil {
		return err
	}
	end := out.file.length
	pos := int64(0)
	buffer := 0
	for pos < end {
		length := BUFFER_SIZE
		nextPos := pos + int64(length)
		if nextPos > end { // at the last buffer
			length = int(end - pos)
		}
		err = output.WriteBytes(out.file.Buffer(buffer)[:length])
		if err != nil {
			return err
		}
		buffer++
		pos = nextPos
	}
	return nil
}

/* Copy the current contents of this buffer to output byte slice */
func (out *RAMOutputStream) WriteToBytes(bytes []byte) error {
	err := out.Flush()
	if err != nil {
		return err
	}
	end := out.file.length
	pos := int64(0)
	buffer := 0
	bytesUpto := 0
	for pos < end {
		length := BUFFER_SIZE
		nextPos := pos + int64(length)
		if nextPos > end {
			length = int(end - pos)
		}
		copy(bytes[bytesUpto:], out.file.Buffer(buffer)[:length])
		buffer++
		bytesUpto += length
		pos = nextPos
	}
	return nil
}

/* Resets this to an empty file. */
func (out *RAMOutputStream) Reset() {
	out.currentBuffer = nil
	out.currentBufferIndex = -1
	out.bufferPosition = 0
	out.bufferStart = 0
	out.bufferLength = 0
	out.file.SetLength(0)
	if out.crc != nil {
		out.crc.Reset()
	}
}

func (out *RAMOutputStream) Close() error {
	return out.Flush()
}

// func (out *RAMOutputStream) Length() (int64, error) {
// 	return out.file.length, nil
// }

func (out *RAMOutputStream) WriteByte(b byte) error {
	if out.bufferPosition == out.bufferLength {
		out.currentBufferIndex++
		out.switchCurrentBuffer()
	}
	if out.crc != nil {
		out.crc.Write([]byte{b})
	}
	out.currentBuffer[out.bufferPosition] = b
	out.bufferPosition++
	return nil
}

func (out *RAMOutputStream) WriteBytes(buf []byte) error {
	assert(buf != nil)
	if out.crc != nil {
		out.crc.Write(buf)
	}
	var offset = 0
	for limit := len(buf); limit > 0; {
		if out.bufferPosition == out.bufferLength {
			out.currentBufferIndex++
			out.switchCurrentBuffer()
		}

		bytesToCopy := len(out.currentBuffer) - out.bufferPosition
		if limit < bytesToCopy {
			bytesToCopy = limit
		}
		copy(out.currentBuffer[out.bufferPosition:], buf[offset:offset+bytesToCopy])
		offset += bytesToCopy
		limit -= bytesToCopy
		out.bufferPosition += bytesToCopy
	}
	return nil
}

func (out *RAMOutputStream) switchCurrentBuffer() {
	if out.currentBufferIndex == out.file.numBuffers() {
		out.currentBuffer = out.file.addBuffer(BUFFER_SIZE)
	} else {
		out.currentBuffer = out.file.Buffer(out.currentBufferIndex)
	}
	out.bufferPosition = 0
	out.bufferStart = BUFFER_SIZE * int64(out.currentBufferIndex)
	out.bufferLength = len(out.currentBuffer)
}

func (out *RAMOutputStream) setFileLength() {
	if pointer := out.bufferStart + int64(out.bufferPosition); pointer > int64(out.file.length) {
		out.file.SetLength(pointer)
	}
}

func (out *RAMOutputStream) Flush() error {
	out.setFileLength()
	return nil
}

func (out *RAMOutputStream) FilePointer() int64 {
	if out.currentBufferIndex < 0 {
		return 0
	}
	return out.bufferStart + int64(out.bufferPosition)
}

func (out *RAMOutputStream) Checksum() int64 {
	assert2(out.crc != nil, "internal RAMOutputStream created with checksum disabled")
	return int64(out.crc.Sum32())
}
