package test_framework

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"io"
	"math/rand"
	"runtime"
	"sync"
)

// store/MockDirectoryWrapper.java

/*
This is a Directory wrapper that adds methods intended to be used
only by unit tests. It also adds a number of fatures useful for
testing:

1. Instances created by newDirectory() are tracked to ensure they are
closed by the test.
2. When a MockDirectoryWrapper is closed, it returns an error if it
has any open files against it (with a stacktrace indicating where
they were opened from)
3. When a MockDirectoryWrapper is closed, it runs CheckIndex to test
if the index was corrupted.
4. MockDirectoryWrapper simulates some "features" of Windows, such as
refusing to write/delete to open files.
*/
type MockDirectoryWrapper struct {
	self *store.DirectoryImpl // overrides LockFactory
	*BaseDirectoryWrapperImpl
	sync.Locker // simulate Java's synchronized keyword

	randomState        *rand.Rand
	noDeleteOpenFile   bool
	preventDoubleWrite bool
	trackDiskUsage     bool
	wrapLockFactory    bool
	unSyncedFiles      map[string]bool
	createdFiles       map[string]bool
	openFilesForWrite  map[string]bool
	openLocks          map[string]bool // synchronized
	openLocksLock      sync.Locker
	crashed            bool // volatile
	throttledOutput    *ThrottledIndexOutput
	throttling         Throttling

	inputCloneCount int // atomic

	// use this for tracking files for crash.
	// additionally: provides debugging information in case you leave one open
	// openFileHandles map[io.Closeable]error // synchronized

	// NOTE: we cannot intialize the map here due to the order in which our
	// constructor actually does this member initialization vs when it calls
	// super. It seems like super is called, then our members are initialzed.
	//
	// Ian: it's not the case for golucene BUT I have no idea why it stays...
	openFiles map[string]int

	// Only tracked if noDeleteOpenFile is true: if an attempt is made to delete
	// an open file, we entroll it here.
	openFilesDeleted map[string]bool
}

func (mdw *MockDirectoryWrapper) init() {
	mdw.Lock() // synchronized
	defer mdw.Unlock()

	if mdw.openFiles == nil {
		mdw.openFiles = make(map[string]int)
		mdw.openFilesDeleted = make(map[string]bool)
	}

	if mdw.createdFiles == nil {
		mdw.createdFiles = make(map[string]bool)
	}
	if mdw.unSyncedFiles == nil {
		mdw.unSyncedFiles = make(map[string]bool)
	}
}

func NewMockDirectoryWrapper(random *rand.Rand, delegate store.Directory) *MockDirectoryWrapper {
	ans := &MockDirectoryWrapper{
		noDeleteOpenFile:   true,
		preventDoubleWrite: true,
		trackDiskUsage:     false,
		wrapLockFactory:    true,
		openFilesForWrite:  make(map[string]bool),
		openLocks:          make(map[string]bool),
		openLocksLock:      &sync.Mutex{},
		throttling:         THROTTLING_SOMETIMES,
		inputCloneCount:    0,
		// openFileHandles: make(map[io.Closer]error),
	}
	ans.self = store.NewDirectoryImpl(ans)
	ans.BaseDirectoryWrapperImpl = NewBaseDirectoryWrapper(delegate)
	ans.Locker = &sync.Mutex{}
	// must make a private random since our methods are called from different
	// methods; else test failures may not be reproducible from the original
	// seed
	ans.randomState = rand.New(rand.NewSource(random.Int63()))
	ans.throttledOutput = newThrottledIndexOutput(
		mBitsToBytes(40+ans.randomState.Intn(10)), 5+ans.randomState.Int63n(5), nil)
	// force wrapping of LockFactory
	fac := newMockLockFactoryWrapper(ans, delegate.LockFactory())
	oldId := fac.LockPrefix()
	ans.self.SetLockFactory(fac)
	fac.SetLockPrefix(oldId) // workaround
	ans.init()
	return ans
}

// Controlling hard disk throttling
// Set via setThrottling()
// WARNING: can make tests very slow.
type Throttling int

const (
	// always emulate a slow hard disk. Cold be very slow!
	THROTTLING_ALWAYS = Throttling(1)
	// sometimes (2% of the time) emulate a slow hard disk.
	THROTTLING_SOMETIMES = Throttling(2)
	// never throttle output
	THROTTLING_NEVER = Throttling(3)
)

func (mdw *MockDirectoryWrapper) SetThrottling(throttling Throttling) {
	mdw.throttling = throttling
}

/*
Returns true if delegate must sync its files. Currently, only
NRTCachingDirectory requires sync'ing its files because otherwise
they are cached in an internal RAMDirectory. If other directories
requires that too, they should be added to this method.
*/
func (mdw *MockDirectoryWrapper) mustSync() {
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) Sync(names []string) error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) String() string {
	// NOTE: do not maybeYield here, since it consumes randomness and
	// can thus (unexpectedly during debugging) change the behavior of
	// a seed maybeYield()
	return fmt.Sprintf("MockDirWrapper(%v)", w.Directory)
}

// Simulates a crash of OS or machine by overwriting unsynced files.
func (w *MockDirectoryWrapper) Crash() error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) DeleteFile(name string) error {
	w.maybeYield()
	return w.deleteFile(name, false)
}

func (w *MockDirectoryWrapper) maybeYield() {
	if w.randomState.Intn(2) == 0 {
		runtime.Gosched()
	}
}

func (w *MockDirectoryWrapper) deleteFile(name string, forced bool) error {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()

	err := w.maybeThrowDeterministicException()
	if err != nil {
		return err
	}

	if w.crashed && !forced {
		return errors.New("cannot delete after crash")
	}

	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) CreateOutput(name string, context store.IOContext) (out store.IndexOutput, err error) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

type Handle int

const (
	HANDLE_INPUT  = Handle(1)
	HANDLE_OUTPUT = Handle(2)
	HANDLE_SLICE  = Handle(3)
)

func (w *MockDirectoryWrapper) addFileHandle(c io.Closer, name string, handle Handle) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) OpenInput(name string, context store.IOContext) (in store.IndexInput, err error) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) Close() error {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) removeOpenFile(c io.Closer, name string) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) removeIndexOutput(out store.IndexOutput, name string) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) removeIndexInput(in store.IndexInput, name string) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

// Iterate through the failures list, giving each object a
// chance to return an error
func (w *MockDirectoryWrapper) maybeThrowDeterministicException() error {
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) ListAll() (names []string, err error) {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	return w.Directory.ListAll()
}

func (w *MockDirectoryWrapper) FileExists(name string) bool {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	return w.Directory.FileExists(name)
}

func (w *MockDirectoryWrapper) FileLength(name string) int64 {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	// return w.Directory.FileLength(name)
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) MakeLock(name string) store.Lock {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	return w.LockFactory().Make(name)
}

func (w *MockDirectoryWrapper) ClearLock(name string) error {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	return w.LockFactory().Clear(name)
}

func (w *MockDirectoryWrapper) SetLockFactory(lockFactory store.LockFactory) {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	// sneaky: we must pass the original this way to the dir, because
	// some impls (e.g. FSDir) do instanceof here
	w.Directory.SetLockFactory(lockFactory)
	// now set out wrapped factory here
	fac := newMockLockFactoryWrapper(w, w.Directory.LockFactory())
	oldId := fac.LockPrefix()
	w.self.SetLockFactory(fac)
	fac.SetLockPrefix(oldId) // workaround
}

func (w *MockDirectoryWrapper) LockFactory() store.LockFactory {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	if w.wrapLockFactory {
		return w.self.LockFactory()
	} else {
		return w.Directory.LockFactory()
	}
}

func (w *MockDirectoryWrapper) LockID() string {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	return w.Directory.LockID()
}

func (w *MockDirectoryWrapper) Copy(to store.Directory, src string, dest string, context store.IOContext) error {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	// randomize the IOContext here?
	panic("not implemented yet")
	// return w.Directory.Copy(to, src, dest, context)
}

func (w *MockDirectoryWrapper) CreateSlicer(name string, context store.IOContext) (slicer store.IndexInputSlicer, err error) {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

// util/ThrottledIndexOutput.java

const DEFAULT_MIN_WRITTEN_BYTES = 024

// Intentionally slow IndexOutput for testing.
type ThrottledIndexOutput struct {
	bytesPerSecond   int
	delegate         store.IndexOutput
	flushDelayMillis int64
	closeDelayMillis int64
	seekDelayMillis  int64
	minBytesWritten  int64
}

func newThrottledIndexOutput(bytesPerSecond int, delayInMillis int64, delegate store.IndexOutput) *ThrottledIndexOutput {
	assert(bytesPerSecond > 0)
	return &ThrottledIndexOutput{
		delegate:         delegate,
		bytesPerSecond:   bytesPerSecond,
		flushDelayMillis: delayInMillis,
		closeDelayMillis: delayInMillis,
		seekDelayMillis:  delayInMillis,
		minBytesWritten:  DEFAULT_MIN_WRITTEN_BYTES,
	}
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func mBitsToBytes(mBits int) int {
	return mBits * 125000
}

// store/MockLockFactoryWrapper.java

// Used by MockDirectoryWrapper to wrap another factory and track
// open locks.
type MockLockFactoryWrapper struct {
	store.LockFactory
	dir *MockDirectoryWrapper
}

func newMockLockFactoryWrapper(dir *MockDirectoryWrapper, delegate store.LockFactory) *MockLockFactoryWrapper {
	return &MockLockFactoryWrapper{delegate, dir}
}

func (w *MockLockFactoryWrapper) Make(lockName string) store.Lock {
	return newMockLock(w.dir, w.LockFactory.Make(lockName), lockName)
}

func (w *MockLockFactoryWrapper) Clear(lockName string) error {
	err := w.LockFactory.Clear(lockName)
	if err != nil {
		return err
	}
	w.dir.openLocksLock.Lock()
	defer w.dir.openLocksLock.Unlock()
	delete(w.dir.openLocks, lockName)
	return nil
}

func (w *MockLockFactoryWrapper) String() string {
	return fmt.Sprintf("MockLockFactoryWrapper(%v)", w.LockFactory)
}

type MockLock struct {
	*store.LockImpl
	delegate store.Lock
	name     string
	dir      *MockDirectoryWrapper
}

func newMockLock(dir *MockDirectoryWrapper, delegate store.Lock, name string) *MockLock {
	ans := &MockLock{
		delegate: delegate,
		name:     name,
		dir:      dir,
	}
	ans.LockImpl = store.NewLockImpl(ans)
	return ans
}

func (lock *MockLock) Obtain() (ok bool, err error) {
	ok, err = lock.delegate.Obtain()
	if err != nil {
		return
	}
	if ok {
		lock.dir.openLocksLock.Lock()
		defer lock.dir.openLocksLock.Unlock()
		lock.dir.openLocks[lock.name] = true
	}
	return ok, nil
}

func (lock *MockLock) Release() {
	lock.delegate.Release()
	lock.dir.openLocksLock.Lock()
	defer lock.dir.openLocksLock.Unlock()
	delete(lock.dir.openLocks, lock.name)
}

func (lock *MockLock) IsLocked() bool {
	return lock.delegate.IsLocked()
}
