package test_framework

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	. "github.com/balzaczyy/golucene/test_framework/util"
	"io"
	"log"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
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
	*BaseDirectoryWrapperImpl
	sync.Locker                     // simulate Java's synchronized keyword
	myLockFactory store.LockFactory // overrides LockFactory

	randomErrorRate       float64
	randomErrorRateOnOpen float64
	randomState           *rand.Rand
	noDeleteOpenFile      bool
	preventDoubleWrite    bool
	trackDiskUsage        bool
	wrapLockFactory       bool
	unSyncedFiles         map[string]bool
	createdFiles          map[string]bool
	openFilesForWrite     map[string]bool
	openLocks             map[string]bool // synchronized
	openLocksLock         sync.Locker
	crashed               bool // volatile
	throttledOutput       *ThrottledIndexOutput
	throttling            Throttling

	inputCloneCount int32 // atomic

	// use this for tracking files for crash.
	// additionally: provides debugging information in case you leave one open
	openFileHandles map[io.Closer]error // synchronized

	// NOTE: we cannot intialize the map here due to the order in which our
	// constructor actually does this member initialization vs when it calls
	// super. It seems like super is called, then our members are initialzed.
	//
	// Ian: it's not the case for golucene BUT I have no idea why it stays...
	openFiles map[string]int

	// Only tracked if noDeleteOpenFile is true: if an attempt is made to delete
	// an open file, we entroll it here.
	openFilesDeleted map[string]bool

	failOnCreateOutput               bool
	failOnOpenInput                  bool
	assertNoUnreferencedFilesOnClose bool

	failures []*Failure
}

func (mdw *MockDirectoryWrapper) init() {
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
		noDeleteOpenFile:                 true,
		preventDoubleWrite:               true,
		trackDiskUsage:                   false,
		wrapLockFactory:                  true,
		openFilesForWrite:                make(map[string]bool),
		openLocks:                        make(map[string]bool),
		openLocksLock:                    &sync.Mutex{},
		throttling:                       THROTTLING_SOMETIMES,
		inputCloneCount:                  0,
		openFileHandles:                  make(map[io.Closer]error),
		failOnCreateOutput:               true,
		failOnOpenInput:                  true,
		assertNoUnreferencedFilesOnClose: true,
	}
	ans.BaseDirectoryWrapperImpl = NewBaseDirectoryWrapper(delegate)
	ans.Locker = &sync.Mutex{}
	// must make a private random since our methods are called from different
	// methods; else test failures may not be reproducible from the original
	// seed
	ans.randomState = rand.New(rand.NewSource(random.Int63()))
	ans.throttledOutput = newThrottledIndexOutput(
		mBitsToBytes(40+ans.randomState.Intn(10)), 5+ans.randomState.Int63n(5), nil)
	// force wrapping of LockFactory
	ans.myLockFactory = newMockLockFactoryWrapper(ans, delegate.LockFactory())
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
func (mdw *MockDirectoryWrapper) mustSync() bool {
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) Sync(names []string) error {
	w.Lock() // synchronized
	defer w.Unlock()
	w.maybeYield()
	err := w.maybeThrowDeterministicException()
	if err != nil {
		return err
	}
	if w.crashed {
		return errors.New("cannot sync after crash")
	}
	// don't wear out out hardware so much intests.
	if Rarely(w.randomState) || w.mustSync() {
		for _, name := range names {
			// randomly fail with IOE on any file
			err = w.maybeThrowIOException(name)
			if err != nil {
				return err
			}
			err = w.Directory.Sync([]string{name})
			if err != nil {
				return err
			}
			delete(w.unSyncedFiles, name)
		}
	} else {
		for _, name := range names {
			delete(w.unSyncedFiles, name)
		}
	}
	return nil
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

func (w *MockDirectoryWrapper) maybeThrowIOException(message string) error {
	if w.randomState.Float64() < w.randomErrorRate {
		if message != "" {
			message = fmt.Sprintf(" (%v)", message)
		}
		if VERBOSE {
			log.Printf("MockDirectoryWrapper: now return random error%v", message)
			debug.PrintStack()
		}
		return errors.New(fmt.Sprintf("a random IO error%v", message))
	}
	return nil
}

func (w *MockDirectoryWrapper) maybeThrowIOExceptionOnOpen(name string) error {
	if w.randomState.Float64() < w.randomErrorRateOnOpen {
		if VERBOSE {
			log.Printf("MockDirectoryWrapper: now return random error during open file=%v", name)
			debug.PrintStack()
		}
		if w.randomState.Intn(2) == 0 {
			return errors.New(fmt.Sprintf("a random IO error (%v)", name))
		}
		return os.ErrNotExist
	}
	return nil
}

func (w *MockDirectoryWrapper) DeleteFile(name string) error {
	w.maybeYield()
	return w.deleteFile(name, false)
}

/*
sets the cause of the incoming ioe to be the stack trace when the
offending file name was opened
*/
func (w *MockDirectoryWrapper) fillOpenTrace(err error, name string, input bool) error {
	w.Lock()
	defer w.Unlock()
	return w._fillOpenTrace(err, name, input)
}

func (w *MockDirectoryWrapper) _fillOpenTrace(err error, name string, input bool) error {
	for closer, cause := range w.openFileHandles {
		v, ok := closer.(*MockIndexInputWrapper)
		if input && ok && v.name == name {
			err = mergeError(err, cause)
			break
		} else {
			v2, ok := closer.(*MockIndexOutputWrapper)
			if !input && ok && v2.name == name {
				err = mergeError(err, cause)
				break
			}
		}
	}
	return err
}

func mergeError(err, err2 error) error {
	if err == nil {
		return err2
	} else {
		return errors.New(fmt.Sprintf("%v\n  %v", err, err2))
	}
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

	if _, ok := w.unSyncedFiles[name]; ok {
		delete(w.unSyncedFiles, name)
	}
	if !forced && w.noDeleteOpenFile {
		if _, ok := w.openFiles[name]; ok {
			w.openFilesDeleted[name] = true
			return w._fillOpenTrace(errors.New(fmt.Sprintf(
				"MockDirectoryWrapper: file  '%v' is still open: cannot delete",
				name)), name, true)
		}
		delete(w.openFilesDeleted, name)
	}
	return w.Directory.DeleteFile(name)
}

func (w *MockDirectoryWrapper) CreateOutput(name string, context store.IOContext) (store.IndexOutput, error) {
	w.Lock() // synchronized
	defer w.Unlock()

	err := w.maybeThrowDeterministicException()
	if err != nil {
		return nil, err
	}
	err = w.maybeThrowIOExceptionOnOpen(name)
	if err != nil {
		return nil, err
	}
	w.maybeYield()
	if w.failOnCreateOutput {
		if err = w.maybeThrowDeterministicException(); err != nil {
			return nil, err
		}
	}
	if w.crashed {
		return nil, errors.New("cannot createOutput after crash")
	}
	w.init()
	if _, ok := w.createdFiles[name]; w.preventDoubleWrite && ok && name != "segments.gen" {
		return nil, errors.New(fmt.Sprintf("file %v was already written to", name))
	}
	if _, ok := w.openFiles[name]; w.noDeleteOpenFile && ok {
		return nil, errors.New(fmt.Sprintf("MockDirectoryWraper: file %v is still open: cannot overwrite", name))
	}

	if w.crashed {
		return nil, errors.New("cannot createOutput after crash")
	}
	w.unSyncedFiles[name] = true
	w.createdFiles[name] = true

	if ramdir, ok := w.Directory.(*store.RAMDirectory); ok {
		file := store.NewRAMFile(ramdir)
		existing := ramdir.GetRAMFile(name)

		// Enforce write once:
		if existing != nil && name != "segments.gen" && w.preventDoubleWrite {
			return nil, errors.New(fmt.Sprintf("file %v already exists", name))
		} else {
			if existing != nil {
				ramdir.ChangeSize(-existing.SizeInBytes())
				// existing.directory = nil
			}
			ramdir.PutRAMFile(name, file)
		}
	}
	log.Printf("MDW: create %v", name)
	delegateOutput, err := w.Directory.CreateOutput(name, NewIOContext(w.randomState, context))
	if err != nil {
		return nil, err
	}
	if w.randomState.Intn(10) == 0 {
		// once ina while wrap the IO in a buffered IO with random buffer sizes
		delegateOutput = newBufferedIndexOutputWrapper(
			1+w.randomState.Intn(store.DEFAULT_BUFFER_SIZE), delegateOutput)
	}
	io := newMockIndexOutputWrapper(w, name, delegateOutput)
	w._addFileHandle(io, name, HANDLE_OUTPUT)
	w.openFilesForWrite[name] = true

	// throttling REALLY slows down tests, so don't do it very often for SOMETIMES
	if _, ok := w.Directory.(*store.RateLimitedDirectoryWrapper); w.throttling == THROTTLING_ALWAYS ||
		(w.throttling == THROTTLING_SOMETIMES && w.randomState.Intn(50) == 0) && !ok {
		if VERBOSE {
			log.Println(fmt.Sprintf("MockDirectoryWrapper: throttling indexOutpu (%v)", name))
		}
		return w.throttledOutput.newFromDelegate(io), nil
	}
	return io, nil
}

type Handle int

const (
	HANDLE_INPUT  = Handle(1)
	HANDLE_OUTPUT = Handle(2)
	HANDLE_SLICE  = Handle(3)
)

func handleName(handle Handle) string {
	switch handle {
	case HANDLE_INPUT:
		return "Input"
	case HANDLE_OUTPUT:
		return "Output"
	case HANDLE_SLICE:
		return "Slice"
	}
	panic("should not be here")
}

func (w *MockDirectoryWrapper) addFileHandle(c io.Closer, name string, handle Handle) {
	w.Lock() // synchronized
	defer w.Unlock()
	w._addFileHandle(c, name, handle)
}

func (w *MockDirectoryWrapper) _addFileHandle(c io.Closer, name string, handle Handle) {
	if v, ok := w.openFiles[name]; ok {
		w.openFiles[name] = v + 1
	} else {
		w.openFiles[name] = 1
	}
	w.openFileHandles[c] = errors.New(fmt.Sprintf("unclosed Index %v: %v", handleName(handle), name))
}

func (w *MockDirectoryWrapper) OpenInput(name string, context store.IOContext) (ii store.IndexInput, err error) {
	w.Lock() // synchronized
	defer w.Unlock()

	if err = w.maybeThrowDeterministicException(); err != nil {
		return
	}
	if err = w.maybeThrowIOExceptionOnOpen(name); err != nil {
		return
	}
	w.maybeYield()
	if w.failOnOpenInput {
		if err = w.maybeThrowDeterministicException(); err != nil {
			return
		}
	}
	if !w.Directory.FileExists(name) {
		return nil, errors.New(fmt.Sprintf("%v in dir=%v", name, w.Directory))
	}

	// cannot open a file for input if it's still open for output,
	//except for segments.gen and segments_N
	if _, ok := w.openFilesForWrite[name]; ok && strings.HasPrefix(name, "segments") {
		err = w.fillOpenTrace(errors.New(fmt.Sprintf(
			"MockDirectoryWrapper: file '%v' is still open for writing", name)), name, false)
		return
	}

	var delegateInput store.IndexInput
	delegateInput, err = w.Directory.OpenInput(name, NewIOContext(w.randomState, context))
	if err != nil {
		return
	}

	randomInt := w.randomState.Intn(500)
	if randomInt == 0 {
		if VERBOSE {
			log.Printf("MockDirectoryWrapper: using SlowClosingMockIndexInputWrapper for file %v", name)
		}
		panic("not implemented yet")
	} else if randomInt == 1 {
		if VERBOSE {
			log.Printf("MockDirectoryWrapper: using SlowOpeningMockIndexInputWrapper for file %v", name)
		}
		panic("not implemented yet")
	} else {
		ii = newMockIndexInputWrapper(w, name, delegateInput)
	}
	w.addFileHandle(ii, name, HANDLE_INPUT)
	return ii, nil
}

func (w *MockDirectoryWrapper) Close() error {
	// files that we tried to delete, but couldn't because reader were open
	// all that matters is that we tried! (they will eventually go away)
	w.Lock()
	pendingDeletions := make(map[string]bool)
	for k, v := range w.openFilesDeleted {
		pendingDeletions[k] = v
	}
	w.Unlock()

	w.maybeYield()

	w.Lock()
	if w.openFiles == nil {
		w.openFiles = make(map[string]int)
		w.openFilesDeleted = make(map[string]bool)
	}
	nOpenFiles := len(w.openFiles)
	w.Unlock()

	if w.noDeleteOpenFile && nOpenFiles > 0 {
		panic("not implemented yet")
	}

	w.openLocksLock.Lock()
	nOpenLocks := len(w.openLocks)
	w.openLocksLock.Unlock()

	if w.noDeleteOpenFile && nOpenLocks > 0 {
		panic(fmt.Sprintf("MockDirectoryWrapper: cannot close: there are still open locks: %v", w.openLocks))
	}

	w.isOpen = false
	if w.checkIndexOnClose {
		w.randomErrorRate = 0
		w.randomErrorRateOnOpen = 0
		if ok, err := index.IsIndexExists(w); err != nil {
			return err
		} else if ok {
			log.Println("\nNOTE: MockDirectoryWrapper: now crash")
			err = w.Crash() // corrupt any unsynced-files
			if err != nil {
				return err
			}
			log.Println("\nNOTE: MockDirectoryWrapper: now run CheckIndex")
			_, err = CheckIndex(w, w.crossCheckTermVectorsOnClose)
			if err != nil {
				return err
			}

			// TODO: factor this out / share w/ TestIW.assertNoUnreferencedFiles
			if w.assertNoUnreferencedFilesOnClose {
				// now look for unreferenced files: discount ones that we tried to delete but could not
				all, err := w.ListAll()
				if err != nil {
					return err
				}
				allFiles := make(map[string]bool)
				for _, name := range all {
					allFiles[name] = true
				}
				for name, _ := range pendingDeletions {
					delete(allFiles, name)
				}
				startFiles := make([]string, 0, len(allFiles))
				for k, _ := range allFiles {
					startFiles = append(startFiles, k)
				}
				iwc := index.NewIndexWriterConfig(TEST_VERSION_CURRENT, nil)
				iwc.SetIndexDeletionPolicy(index.NO_DELETION_POLICY)
				iw, err := index.NewIndexWriter(w.Directory, iwc)
				if err != nil {
					return err
				}
				err = iw.Rollback()
				if err != nil {
					return err
				}
				endFiles, err := w.Directory.ListAll()
				if err != nil {
					return err
				}

				hasSegmentsGenFile := sort.SearchStrings(endFiles, index.INDEX_FILENAME_SEGMENTS_GEN) >= 0
				if pendingDeletions["segments.gen"] && hasSegmentsGenFile {
					panic("not implemented yet")
				}

				// its possible we cannot delete the segments_N on windows if someone has it open and
				// maybe other files too, depending on timing. normally someone on windows wouldnt have
				// an issue (IFD would nuke this stuff eventually), but we pass NoDeletionPolicy...
				for _, file := range pendingDeletions {
					log.Println(file)
					panic("not implemented yet")
				}

				sort.Strings(startFiles)
				startFiles = uniqueStrings(startFiles)
				sort.Strings(endFiles)
				endFiles = uniqueStrings(endFiles)

				if !reflect.DeepEqual(startFiles, endFiles) {
					panic("not implemented")
				}

				ir1, err := index.OpenDirectoryReader(w)
				if err != nil {
					return err
				}
				numDocs1 := ir1.NumDocs()
				err = ir1.Close()
				if err != nil {
					return err
				}
				iw, err = index.NewIndexWriter(w, index.NewIndexWriterConfig(TEST_VERSION_CURRENT, nil))
				if err != nil {
					return err
				}
				err = iw.Close()
				if err != nil {
					return err
				}
				ir2, err := index.OpenDirectoryReader(w)
				if err != nil {
					return err
				}
				numDocs2 := ir2.NumDocs()
				err = ir2.Close()
				if err != nil {
					return err
				}
				assert2(numDocs1 == numDocs2, fmt.Sprintf("numDocs changed after opening/closing IW: before=%v after=%v", numDocs1, numDocs2))
			}
		}
	}
	return w.Directory.Close()
}

func assert2(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func uniqueStrings(a []string) []string {
	ans := make([]string, 0, len(a)) // inefficient for fewer unique items
	for _, v := range a {
		if n := len(ans); n == 0 || ans[n-1] != v {
			ans = append(ans, v)
		}
	}
	return ans
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

/*
Objects that represent fail-lable conditions. Objects of a derived
class are created and registered with teh mock directory. After
register, each object will be invoked once for each first write of a
file, giving the object a chance to throw an IO error.
*/
type Failure struct {
	// eval is called on the first write of every new file
	eval   func(dir *MockDirectoryWrapper) error
	doFail bool
}

/*
reset should set the state of the failure to its default (freshly
constructed) state. Reset is convenient for tests that want to create
one failure object and then reuse it in mutiple cases. This, combined
with the fact that Failure subclasses are often anonymous classes
makes reset difficult to do otherwise.

A typical example of use is

		failure := &Failure { eval: func(dir *MockDirectoryWrapper) { ... } }
		...
		mock.failOn(failure.reset())

*/
func (f *Failure) Reset() *Failure { return f }
func (f *Failure) SetDoFail()      { f.doFail = true }
func (f *Failure) ClearDoFail()    { f.doFail = false }

/*
add a Failure object to the list of objects to be evaluated at every
potential failure opint
*/
func (w *MockDirectoryWrapper) failOn(fail *Failure) {
	w.failures = append(w.failures, fail)
}

// Iterate through the failures list, giving each object a
// chance to return an error
func (w *MockDirectoryWrapper) maybeThrowDeterministicException() error {
	for _, f := range w.failures {
		if err := f.eval(w); err != nil {
			return err
		}
	}
	return nil
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

func (w *MockDirectoryWrapper) FileLength(name string) (n int64, err error) {
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
	return w.lockFactory().Make(name)
}

func (w *MockDirectoryWrapper) ClearLock(name string) error {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	return w.lockFactory().Clear(name)
}

func (w *MockDirectoryWrapper) SetLockFactory(lockFactory store.LockFactory) {
	w.Lock() // synchronized
	defer w.Unlock()

	w.maybeYield()
	// sneaky: we must pass the original this way to the dir, because
	// some impls (e.g. FSDir) do instanceof here
	w.Directory.SetLockFactory(lockFactory)
	// now set out wrapped factory here
	w.myLockFactory = newMockLockFactoryWrapper(w, w.Directory.LockFactory())
}

func (w *MockDirectoryWrapper) LockFactory() store.LockFactory {
	w.Lock() // synchronized
	defer w.Unlock()

	return w.lockFactory()
}

func (w *MockDirectoryWrapper) lockFactory() store.LockFactory {
	w.maybeYield()
	if w.wrapLockFactory {
		return w.myLockFactory
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

type BufferedIndexOutputWrapper struct {
	*store.BufferedIndexOutput
	io store.IndexOutput
}

func newBufferedIndexOutputWrapper(bufferSize int, io store.IndexOutput) *BufferedIndexOutputWrapper {
	ans := &BufferedIndexOutputWrapper{}
	ans.BufferedIndexOutput = store.NewBufferedIndexOutput(bufferSize, ans)
	ans.io = io
	return ans
}

// util/ThrottledIndexOutput.java

const DEFAULT_MIN_WRITTEN_BYTES = 024

// Intentionally slow IndexOutput for testing.
type ThrottledIndexOutput struct {
	*store.IndexOutputImpl
	bytesPerSecond   int
	delegate         store.IndexOutput
	flushDelayMillis int64
	closeDelayMillis int64
	seekDelayMillis  int64
	pendingBytes     int64
	minBytesWritten  int64
	timeElapsed      int64
	bytes            []byte
}

func (out *ThrottledIndexOutput) newFromDelegate(output store.IndexOutput) *ThrottledIndexOutput {
	ans := &ThrottledIndexOutput{
		delegate:         output,
		bytesPerSecond:   out.bytesPerSecond,
		flushDelayMillis: out.flushDelayMillis,
		closeDelayMillis: out.closeDelayMillis,
		seekDelayMillis:  out.seekDelayMillis,
		minBytesWritten:  out.minBytesWritten,
		bytes:            make([]byte, 1),
	}
	ans.IndexOutputImpl = store.NewIndexOutput(ans)
	return ans
}

func newThrottledIndexOutput(bytesPerSecond int, delayInMillis int64, delegate store.IndexOutput) *ThrottledIndexOutput {
	assert(bytesPerSecond > 0)
	ans := &ThrottledIndexOutput{
		delegate:         delegate,
		bytesPerSecond:   bytesPerSecond,
		flushDelayMillis: delayInMillis,
		closeDelayMillis: delayInMillis,
		seekDelayMillis:  delayInMillis,
		minBytesWritten:  DEFAULT_MIN_WRITTEN_BYTES,
		bytes:            make([]byte, 1),
	}
	ans.IndexOutputImpl = store.NewIndexOutput(ans)
	return ans
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func mBitsToBytes(mBits int) int {
	return mBits * 125000
}

func (out *ThrottledIndexOutput) Close() error {
	<-time.After(time.Duration(out.closeDelayMillis + out.delay(true)))
	return out.delegate.Close()
}

func (out *ThrottledIndexOutput) WriteByte(b byte) error {
	out.bytes[0] = b
	return out.WriteBytes(out.bytes)
}

func (out *ThrottledIndexOutput) WriteBytes(buf []byte) error {
	before := time.Now()
	// TODO: sometimes, write only half the bytes, then sleep, then 2nd
	// half, then sleep, so we sometimes interrupt having only written
	// not all bytes
	err := out.delegate.WriteBytes(buf)
	if err != nil {
		return err
	}
	out.timeElapsed += int64(time.Now().Sub(before))
	out.pendingBytes += int64(len(buf))
	<-time.After(time.Duration(out.delay(false)))
	return nil
}

func (out *ThrottledIndexOutput) delay(closing bool) int64 {
	if out.pendingBytes > 0 && (closing || out.pendingBytes > out.minBytesWritten) {
		actualBps := (out.timeElapsed / out.pendingBytes) * 1000000000
		if actualBps > int64(out.bytesPerSecond) {
			expected := out.pendingBytes * 1000 / int64(out.bytesPerSecond)
			delay := expected - (out.timeElapsed / 1000000)
			out.pendingBytes = 0
			out.timeElapsed = 0
			return delay
		}
	}
	return 0
}

func (out *ThrottledIndexOutput) CopyBytes(input util.DataInput, numBytes int64) error {
	return out.delegate.CopyBytes(input, numBytes)
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

func (lock *MockLock) Release() error {
	if err := lock.delegate.Release(); err != nil {
		return err
	}
	lock.dir.openLocksLock.Lock()
	defer lock.dir.openLocksLock.Unlock()
	delete(lock.dir.openLocks, lock.name)
	return nil
}

func (lock *MockLock) IsLocked() bool {
	return lock.delegate.IsLocked()
}

// store/MockIndexInputWrapper.java

/*
Used by MockDirectoryWrapper to create an input stream that keeps
track of when it's been closed.
*/
type MockIndexInputWrapper struct {
	store.IndexInput // delegate
	dir              *MockDirectoryWrapper
	name             string
	isClone          bool
	closed           bool
}

func newMockIndexInputWrapper(dir *MockDirectoryWrapper, name string, delegate store.IndexInput) *MockIndexInputWrapper {
	panic("not implemented yet")
}

func (w *MockIndexInputWrapper) ensureOpen() {
	assert2(!w.closed, "Abusing closed IndexInput!")
}

func (w *MockIndexInputWrapper) Clone() store.IndexInput {
	panic("not implemented yet")
}

func (w *MockIndexInputWrapper) FilePointer() int64 {
	w.ensureOpen()
	return w.IndexInput.FilePointer()
}

func (w *MockIndexInputWrapper) Seek(pos int64) error {
	w.ensureOpen()
	return w.IndexInput.Seek(pos)
}

// store/MockIndexOutputWrapper.java

/*
Used by MockRAMDirectory to create an output stream that will throw
an error on fake disk full, track max disk space actually used, and
maybe throw random IO errors.
*/
type MockIndexOutputWrapper struct {
	store.IndexOutput // delegate
	dir               *MockDirectoryWrapper
	first             bool
	name              string
	singleByte        []byte
}

func newMockIndexOutputWrapper(dir *MockDirectoryWrapper, name string, delegate store.IndexOutput) *MockIndexOutputWrapper {
	return &MockIndexOutputWrapper{
		IndexOutput: delegate,
		name:        name,
		dir:         dir,
		first:       true,
		singleByte:  make([]byte, 1),
	}
}
