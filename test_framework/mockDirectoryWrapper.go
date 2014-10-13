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
	// "strings"
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
	*BaseDirectoryWrapperImpl
	sync.Locker                     // simulate Java's synchronized keyword
	myLockFactory store.LockFactory // overrides LockFactory

	isLocked bool // workaround re-entrant lock

	maxSize int64

	// Max actual bytes used. This is set by MockRAMOutputStream
	maxUsedSize           int64
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
	ans.throttledOutput = NewThrottledIndexOutput(
		MBitsToBytes(40+ans.randomState.Intn(10)), 5+ans.randomState.Int63n(5), nil)
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
	var delegate = mdw.Directory
	for {
		if v, ok := delegate.(*store.RateLimitedDirectoryWrapper); ok {
			delegate = v.Directory
		} else if v, ok := delegate.(*store.TrackingDirectoryWrapper); ok {
			delegate = v.Directory
		} else {
			break
		}
	}
	_, ok := delegate.(*store.NRTCachingDirectory)
	return ok
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

func (w *MockDirectoryWrapper) sizeInBytes() (int64, error) {
	w.Lock()
	defer w.Unlock()

	// v, ok := w.Directory.(*store.RAMDirectory)
	// if ok {
	// 	return v.RamBytesUsed(), nil
	// }
	// hack
	panic("not implemented yet")
}

// Simulates a crash of OS or machine by overwriting unsynced files.
func (w *MockDirectoryWrapper) Crash() error {
	w.Lock() // synchronized
	defer w.Unlock()
	return w._crash()
}

func (w *MockDirectoryWrapper) _crash() error {
	panic("not implemented yet")
	// w.crashed = true
	// w.openFiles = make(map[string]int)
	// w.openFilesForWrite = make(map[string]bool)
	// w.openFilesDeleted = make(map[string]bool)
	// files := w.unSyncedFiles
	// w.unSyncedFiles = make(map[string]bool)
	// // first force-close all files, so we can corrupt on windows etc.
	// // clone the file map, as these guys want to remove themselves on close.
	// m := make(map[io.Closer]error)
	// for k, v := range w.openFileHandles {
	// 	m[k] = v
	// }
	// for f, _ := range m {
	// 	f.Close() // ignore error
	// }

	// for name, _ := range files {
	// 	var action string
	// 	var err error
	// 	switch w.randomState.Intn(5) {
	// 	case 0:
	// 		action = "deleted"
	// 		err = w.deleteFile(name, true)
	// 	case 1:
	// 		action = "zeroes"
	// 		// Zero out file entirely
	// 		var length int64
	// 		length, err = w._fileLength(name)
	// 		if err == nil {
	// 			zeroes := make([]byte, 256)
	// 			var upto int64 = 0
	// 			var out store.IndexOutput
	// 			out, err = w.BaseDirectoryWrapperImpl.CreateOutput(name, NewDefaultIOContext(w.randomState))
	// 			if err == nil {
	// 				for upto < length && err == nil {
	// 					limit := length - upto
	// 					if int64(len(zeroes)) < limit {
	// 						limit = int64(len(zeroes))
	// 					}
	// 					err = out.WriteBytes(zeroes[:limit])
	// 					upto += limit
	// 				}
	// 				if err == nil {
	// 					err = out.Close()
	// 				}
	// 			}
	// 		}
	// 	case 2:
	// 		action = "partially truncated"
	// 		// Partially Truncate the file:

	// 		// First, make temp file and copy only half this file over:
	// 		var tempFilename string
	// 		for {
	// 			tempFilename = fmt.Sprintf("%v", w.randomState.Int())
	// 			if !w.BaseDirectoryWrapperImpl.FileExists(tempFilename) {
	// 				break
	// 			}
	// 		}
	// 		var tempOut store.IndexOutput
	// 		if tempOut, err = w.BaseDirectoryWrapperImpl.CreateOutput(tempFilename, NewDefaultIOContext(w.randomState)); err == nil {
	// 			var ii store.IndexInput
	// 			if ii, err = w.BaseDirectoryWrapperImpl.OpenInput(name, NewDefaultIOContext(w.randomState)); err == nil {
	// 				if err = tempOut.CopyBytes(ii, ii.Length()/2); err == nil {
	// 					if err = tempOut.Close(); err == nil {
	// 						if err = ii.Close(); err == nil {
	// 							// Delete original and copy bytes back:
	// 							if err = w.deleteFile(name, true); err == nil {
	// 								var out store.IndexOutput
	// 								if out, err = w.BaseDirectoryWrapperImpl.CreateOutput(name, NewDefaultIOContext(w.randomState)); err == nil {
	// 									if ii, err = w.BaseDirectoryWrapperImpl.OpenInput(tempFilename, NewDefaultIOContext(w.randomState)); err == nil {
	// 										if err = out.CopyBytes(ii, ii.Length()); err == nil {
	// 											if err = out.Close(); err == nil {
	// 												if err = ii.Close(); err == nil {
	// 													err = w.deleteFile(tempFilename, true)
	// 												}
	// 											}
	// 										}
	// 									}
	// 								}
	// 							}
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	case 3:
	// 		// the file survived intact:
	// 		action = "didn't change"
	// 	default:
	// 		action = "fully truncated"
	// 		// totally truncate the file to zero bytes
	// 		if err = w.deleteFile(name, true); err == nil {
	// 			var out store.IndexOutput
	// 			if out, err = w.BaseDirectoryWrapperImpl.CreateOutput(name, NewDefaultIOContext(w.randomState)); err == nil {
	// 				if err = out.SetLength(0); err == nil {
	// 					err = out.Close()
	// 				}
	// 			}
	// 		}
	// 	}
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if VERBOSE {
	// 		log.Printf("MockDirectoryWrapper: %v unsynced file: %v", action, name)
	// 	}
	// }
	// return nil
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
	if !w.isLocked {
		w.Lock() // synchronized
		defer w.Unlock()
	}

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
	panic("not implemented yet")
	// if !w.isLocked {
	// 	w.Lock() // synchronized
	// 	defer w.Unlock()
	// }

	// err := w.maybeThrowDeterministicException()
	// if err != nil {
	// 	return nil, err
	// }
	// err = w.maybeThrowIOExceptionOnOpen(name)
	// if err != nil {
	// 	return nil, err
	// }
	// w.maybeYield()
	// if w.failOnCreateOutput {
	// 	if err = w.maybeThrowDeterministicException(); err != nil {
	// 		return nil, err
	// 	}
	// }
	// if w.crashed {
	// 	return nil, errors.New("cannot createOutput after crash")
	// }
	// w.init()
	// if _, ok := w.createdFiles[name]; w.preventDoubleWrite && ok && name != "segments.gen" {
	// 	return nil, errors.New(fmt.Sprintf("file %v was already written to", name))
	// }
	// if _, ok := w.openFiles[name]; w.noDeleteOpenFile && ok {
	// 	return nil, errors.New(fmt.Sprintf("MockDirectoryWraper: file %v is still open: cannot overwrite", name))
	// }

	// if w.crashed {
	// 	return nil, errors.New("cannot createOutput after crash")
	// }
	// w.unSyncedFiles[name] = true
	// w.createdFiles[name] = true

	// if ramdir, ok := w.Directory.(*store.RAMDirectory); ok {
	// 	file := store.NewRAMFile(ramdir)
	// 	existing := ramdir.GetRAMFile(name)

	// 	// Enforce write once:
	// 	if existing != nil && name != "segments.gen" && w.preventDoubleWrite {
	// 		return nil, errors.New(fmt.Sprintf("file %v already exists", name))
	// 	} else {
	// 		if existing != nil {
	// 			ramdir.ChangeSize(-existing.SizeInBytes())
	// 			// existing.directory = nil
	// 		}
	// 		ramdir.PutRAMFile(name, file)
	// 	}
	// }
	// log.Printf("MDW: create %v", name)
	// delegateOutput, err := w.Directory.CreateOutput(name, NewIOContext(w.randomState, context))
	// if err != nil {
	// 	return nil, err
	// }
	// assert(delegateOutput != nil)
	// if w.randomState.Intn(10) == 0 {
	// 	// once ina while wrap the IO in a buffered IO with random buffer sizes
	// 	delegateOutput = newBufferedIndexOutputWrapper(
	// 		1+w.randomState.Intn(store.DEFAULT_BUFFER_SIZE), delegateOutput)
	// 	assert(delegateOutput != nil)
	// }
	// io := newMockIndexOutputWrapper(w, name, delegateOutput)
	// w._addFileHandle(io, name, HANDLE_OUTPUT)
	// w.openFilesForWrite[name] = true

	// // throttling REALLY slows down tests, so don't do it very often for SOMETIMES
	// if _, ok := w.Directory.(*store.RateLimitedDirectoryWrapper); w.throttling == THROTTLING_ALWAYS ||
	// 	(w.throttling == THROTTLING_SOMETIMES && w.randomState.Intn(50) == 0) && !ok {
	// 	if VERBOSE {
	// 		log.Println(fmt.Sprintf("MockDirectoryWrapper: throttling indexOutpu (%v)", name))
	// 	}
	// 	return w.throttledOutput.NewFromDelegate(io), nil
	// }
	// return io, nil
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
	if !w.isLocked {
		w.Lock() // synchronized
		defer w.Unlock()
	}
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
	panic("not implemented yet")
	// if !w.isLocked {
	// 	w.Lock() // synchronized
	// 	defer w.Unlock()
	// }

	// if err = w.maybeThrowDeterministicException(); err != nil {
	// 	return
	// }
	// if err = w.maybeThrowIOExceptionOnOpen(name); err != nil {
	// 	return
	// }
	// w.maybeYield()
	// if w.failOnOpenInput {
	// 	if err = w.maybeThrowDeterministicException(); err != nil {
	// 		return
	// 	}
	// }
	// if !w.Directory.FileExists(name) {
	// 	return nil, errors.New(fmt.Sprintf("%v in dir=%v", name, w.Directory))
	// }

	// // cannot open a file for input if it's still open for output,
	// // except for segments.gen and segments_N
	// if _, ok := w.openFilesForWrite[name]; ok && strings.HasPrefix(name, "segments") {
	// 	err = w.fillOpenTrace(errors.New(fmt.Sprintf(
	// 		"MockDirectoryWrapper: file '%v' is still open for writing", name)), name, false)
	// 	return
	// }

	// var delegateInput store.IndexInput
	// delegateInput, err = w.Directory.OpenInput(name, NewIOContext(w.randomState, context))
	// if err != nil {
	// 	return
	// }

	// randomInt := w.randomState.Intn(500)
	// if randomInt == 0 {
	// 	if VERBOSE {
	// 		log.Printf("MockDirectoryWrapper: using SlowClosingMockIndexInputWrapper for file %v", name)
	// 	}
	// 	panic("not implemented yet")
	// } else if randomInt == 1 {
	// 	if VERBOSE {
	// 		log.Printf("MockDirectoryWrapper: using SlowOpeningMockIndexInputWrapper for file %v", name)
	// 	}
	// 	panic("not implemented yet")
	// } else {
	// 	ii = newMockIndexInputWrapper(w, name, delegateInput)
	// }
	// w._addFileHandle(ii, name, HANDLE_INPUT)
	// return ii, nil
}

// L594
/*
Like recomputeSizeInBytes(), but uses actual file lengths rather than
buffer allocations (which are quantized up to nearest
RAMOutputStream.BUFFER_SIZE (now 1024) bytes).
*/
func (w *MockDirectoryWrapper) recomputeActualSizeInBytes() (int64, error) {
	w.Lock()
	defer w.Unlock()
	// if _, ok := w.Directory.(store.RAMDirectory); !ok {
	// 	return w.sizeInBytes()
	// }
	panic("not implemented yet")
}

func (w *MockDirectoryWrapper) Close() error {
	w.Lock()
	w.isLocked = true
	defer func() {
		w.isLocked = false
		w.Unlock()
	}()

	// files that we tried to delete, but couldn't because reader were open
	// all that matters is that we tried! (they will eventually go away)
	pendingDeletions := make(map[string]bool)
	for k, v := range w.openFilesDeleted {
		pendingDeletions[k] = v
	}

	w.maybeYield()

	if w.openFiles == nil {
		w.openFiles = make(map[string]int)
		w.openFilesDeleted = make(map[string]bool)
	}
	nOpenFiles := len(w.openFiles)

	if w.noDeleteOpenFile && nOpenFiles > 0 {
		// print the first one as its very verbose otherwise
		var cause error
		for _, v := range w.openFileHandles {
			cause = v
			break
		}
		panic(mergeError(errors.New(fmt.Sprintf(
			"MockDirectoryWrapper: cannot close: there are still open files: %v",
			w.openFiles)), cause))
	}

	nOpenLocks := func() int {
		w.openLocksLock.Lock()
		defer w.openLocksLock.Unlock()
		return len(w.openLocks)
	}()
	if w.noDeleteOpenFile && nOpenLocks > 0 {
		panic(fmt.Sprintf("MockDirectoryWrapper: cannot close: there are still open locks: %v", w.openLocks))
	}

	w.isOpen = false
	if w.checkIndexOnClose {
		w.randomErrorRate = 0
		w.randomErrorRateOnOpen = 0
		files, err := w._ListAll()
		if err != nil {
			return err
		}
		if index.IsIndexFileExists(files) {
			fmt.Println("\nNOTE: MockDirectoryWrapper: now crash")
			err = w._crash() // corrupt any unsynced-files
			if err != nil {
				return err
			}
			fmt.Println("\nNOTE: MockDirectoryWrapper: now run CheckIndex")
			w.Unlock() // CheckIndex may access synchronized method
			CheckIndex(w, w.crossCheckTermVectorsOnClose)
			w.Lock() // CheckIndex may access synchronized method

			// TODO: factor this out / share w/ TestIW.assertNoUnreferencedFiles
			if w.assertNoUnreferencedFilesOnClose {
				// now look for unreferenced files: discount ones that we tried to delete but could not
				all, err := w._ListAll()
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

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
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

	w._removeOpenFile(c, name)
}

func (w *MockDirectoryWrapper) _removeOpenFile(c io.Closer, name string) {
	if v, ok := w.openFiles[name]; ok {
		if v == 1 {
			delete(w.openFiles, name)
		} else {
			w.openFiles[name] = v - 1
		}
	}
	delete(w.openFileHandles, c)
}

func (w *MockDirectoryWrapper) removeIndexOutput(out store.IndexOutput, name string) {
	w.Lock() // synchronized
	defer w.Unlock()

	delete(w.openFilesForWrite, name)
	w._removeOpenFile(out, name)
}

func (w *MockDirectoryWrapper) removeIndexInput(in store.IndexInput, name string) {
	if !w.isLocked {
		w.Lock() // synchronized
		defer w.Unlock()
	}

	w._removeOpenFile(in, name)
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

func (w *MockDirectoryWrapper) ListAll() ([]string, error) {
	if !w.isLocked {
		w.Lock() // synchronized
		defer w.Unlock()
	}
	return w._ListAll()
}

func (w *MockDirectoryWrapper) _ListAll() ([]string, error) {
	w.maybeYield()
	return w.Directory.ListAll()
}

func (w *MockDirectoryWrapper) FileExists(name string) bool {
	if !w.isLocked {
		w.Lock() // synchronized
		defer w.Unlock()
	}

	w.maybeYield()
	return w.Directory.FileExists(name)
}

func (w *MockDirectoryWrapper) FileLength(name string) (int64, error) {
	w.Lock() // synchronized
	defer w.Unlock()

	return w._fileLength(name)
}

func (w *MockDirectoryWrapper) _fileLength(name string) (int64, error) {
	w.maybeYield()
	return w.Directory.FileLength(name)
}

func (w *MockDirectoryWrapper) MakeLock(name string) store.Lock {
	if !w.isLocked {
		w.Lock() // synchronized
		defer w.Unlock()
	}

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
	w.isLocked = true
	defer func() {
		w.isLocked = false
		w.Unlock()
	}()

	w.maybeYield()
	// randomize the IOContext here?
	return w.Directory.Copy(to, src, dest, context)
}

// func (w *MockDirectoryWrapper) CreateSlicer(name string,
// 	context store.IOContext) (store.IndexInputSlicer, error) {

// 	if !w.isLocked {
// 		w.Lock() // synchronized
// 		defer w.Unlock()
// 	}

// 	w.maybeYield()
// 	if !w.Directory.FileExists(name) {
// 		return nil, errors.New(fmt.Sprintf("File not found: %v", name))
// 	}
// 	// cannot open a file for input if it's still open for output,
// 	// except for segments.gen and segments_N
// 	if _, ok := w.openFilesForWrite[name]; ok && !strings.HasPrefix(name, "segments") {
// 		return nil, w.fillOpenTrace(errors.New(fmt.Sprintf(
// 			"MockDirectoryWrapper: file '%v' is still open for writing", name)), name, false)
// 	}

// 	delegateHandle, err := w.Directory.CreateSlicer(name, context)
// 	if err != nil {
// 		return nil, err
// 	}
// 	handle := &myIndexInputSlicer{w, delegateHandle, name, false}
// 	w._addFileHandle(handle, name, HANDLE_SLICE)
// 	return handle, nil
// }

// type myIndexInputSlicer struct {
// 	owner          *MockDirectoryWrapper
// 	delegateHandle store.IndexInputSlicer
// 	name           string
// 	isClosed       bool
// }

// func (s *myIndexInputSlicer) Close() error {
// 	if !s.isClosed {
// 		err := s.delegateHandle.Close()
// 		if err != nil {
// 			return err
// 		}
// 		s.owner.removeOpenFile(s, s.name)
// 		s.isClosed = true
// 	}
// 	return nil
// }

// func (s *myIndexInputSlicer) OpenSlice(desc string, offset, length int64) store.IndexInput {
// 	s.owner.maybeYield()
// 	slice := s.delegateHandle.OpenSlice(desc, offset, length)
// 	ii := newMockIndexInputWrapper(s.owner, s.name, slice)
// 	s.owner.addFileHandle(ii, s.name, HANDLE_INPUT)
// 	return ii
// }

// func (s *myIndexInputSlicer) OpenFullSlice() store.IndexInput {
// 	s.owner.maybeYield()
// 	slice := s.delegateHandle.OpenFullSlice()
// 	ii := newMockIndexInputWrapper(s.owner, s.name, slice)
// 	s.owner.addFileHandle(ii, s.name, HANDLE_INPUT)
// 	return ii
// }

// type BufferedIndexOutputWrapper struct {
// 	*store.BufferedIndexOutput
// 	io store.IndexOutput
// }

// func newBufferedIndexOutputWrapper(bufferSize int, io store.IndexOutput) *BufferedIndexOutputWrapper {
// 	ans := &BufferedIndexOutputWrapper{}
// 	ans.BufferedIndexOutput = store.NewBufferedIndexOutput(bufferSize, ans)
// 	ans.io = io
// 	return ans
// }

// func (w *BufferedIndexOutputWrapper) Length() (int64, error) {
// 	return w.io.Length()
// }

// func (w *BufferedIndexOutputWrapper) FlushBuffer(buf []byte) error {
// 	return w.io.WriteBytes(buf)
// }

// // func (w *BufferedIndexOutputWrapper) Flush() error {
// // 	defer w.io.Flush()
// // 	return w.BufferedIndexOutput.Flush()
// // }

// func (w *BufferedIndexOutputWrapper) Close() error {
// 	defer w.io.Close()
// 	return w.BufferedIndexOutput.Close()
// }

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
	panic("not implemented yet")
	// assert(name != "")
	// ans := &MockLock{
	// 	delegate: delegate,
	// 	name:     name,
	// 	dir:      dir,
	// }
	// ans.LockImpl = store.NewLockImpl(ans)
	// return ans
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

func (lock *MockLock) Close() error {
	panic("not implemented yet")
	// if err := lock.delegate.Release(); err != nil {
	// 	return err
	// }
	// lock.dir.openLocksLock.Lock()
	// defer lock.dir.openLocksLock.Unlock()
	// delete(lock.dir.openLocks, lock.name)
	// return nil
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
	*store.IndexInputImpl
	dir      *MockDirectoryWrapper
	delegate store.IndexInput
	name     string
	isClone  bool
	closed   bool
}

func newMockIndexInputWrapper(dir *MockDirectoryWrapper,
	name string, delegate store.IndexInput) *MockIndexInputWrapper {
	ans := &MockIndexInputWrapper{nil, dir, delegate, name, false, false}
	ans.IndexInputImpl = store.NewIndexInputImpl(fmt.Sprintf(
		"MockIndexInputWrapper(name=%v delegate=%v)",
		name, delegate), ans)
	return ans
}

func (w *MockIndexInputWrapper) Close() (err error) {
	panic("not implemented yet")
	// defer func() {
	// 	w.closed = true
	// 	err2 := w.delegate.Close()
	// 	err = mergeError(err, err2)
	// 	if err2 == nil {
	// 		// Pending resolution on LUCENE-686 we may want to remove the
	// 		// conditional check so we also track that all clones get closed:
	// 		if !w.isClone {
	// 			w.dir.removeIndexInput(w, w.name)
	// 		}
	// 	}
	// }()
	// // turn on the following to look for leaks closing inputs, after
	// // fixing TestTransactions
	// // return w.dir.maybeThrowDeterministicException()
	// return nil
}

func (w *MockIndexInputWrapper) ensureOpen() {
	assert2(!w.closed, "Abusing closed IndexInput!")
}

func (w *MockIndexInputWrapper) Clone() store.IndexInput {
	panic("not implemented yet")
}

func (w *MockIndexInputWrapper) FilePointer() int64 {
	w.ensureOpen()
	return w.delegate.FilePointer()
}

func (w *MockIndexInputWrapper) Seek(pos int64) error {
	w.ensureOpen()
	return w.delegate.Seek(pos)
}

func (w *MockIndexInputWrapper) Length() int64 {
	w.ensureOpen()
	return w.delegate.Length()
}

func (w *MockIndexInputWrapper) ReadByte() (byte, error) {
	w.ensureOpen()
	return w.delegate.ReadByte()
}

func (w *MockIndexInputWrapper) ReadBytes(buf []byte) error {
	w.ensureOpen()
	return w.delegate.ReadBytes(buf)
}

func (w *MockIndexInputWrapper) ReadShort() (int16, error) {
	w.ensureOpen()
	return w.delegate.ReadShort()
}

func (w *MockIndexInputWrapper) ReadInt() (int32, error) {
	w.ensureOpen()
	return w.delegate.ReadInt()
}

func (w *MockIndexInputWrapper) ReadLong() (int64, error) {
	w.ensureOpen()
	return w.delegate.ReadLong()
}

func (w *MockIndexInputWrapper) ReadString() (string, error) {
	w.ensureOpen()
	return w.delegate.ReadString()
}

func (w *MockIndexInputWrapper) ReadStringStringMap() (map[string]string, error) {
	w.ensureOpen()
	return w.delegate.ReadStringStringMap()
}

func (w *MockIndexInputWrapper) ReadVInt() (int32, error) {
	w.ensureOpen()
	return w.delegate.ReadVInt()
}

func (w *MockIndexInputWrapper) ReadVLong() (int64, error) {
	w.ensureOpen()
	return w.delegate.ReadVLong()
}

func (w *MockIndexInputWrapper) String() string {
	return fmt.Sprintf("MockIndexInputWrapper(%v)", w.delegate)
}

// store/MockIndexOutputWrapper.java

/*
Used by MockRAMDirectory to create an output stream that will throw
an error on fake disk full, track max disk space actually used, and
maybe throw random IO errors.
*/
type MockIndexOutputWrapper struct {
	*store.IndexOutputImpl
	dir        *MockDirectoryWrapper
	delegate   store.IndexOutput
	first      bool
	name       string
	singleByte []byte
}

func newMockIndexOutputWrapper(dir *MockDirectoryWrapper, name string, delegate store.IndexOutput) *MockIndexOutputWrapper {
	assert(delegate != nil)
	ans := &MockIndexOutputWrapper{
		dir:        dir,
		name:       name,
		delegate:   delegate,
		first:      true,
		singleByte: make([]byte, 1),
	}
	ans.IndexOutputImpl = store.NewIndexOutput(ans)
	return ans
}

func (w *MockIndexOutputWrapper) checkCrashed() error {
	// If MockRAMDir crashed since we were opened, then don't write anything
	if w.dir.crashed {
		return errors.New(fmt.Sprintf("MockRAMDirectory was crashed; cannot write to %v", w.name))
	}
	return nil
}

func (w *MockIndexOutputWrapper) checkDiskFull(buf []byte, in util.DataInput) (err error) {
	panic("not implemented yet")
	// var freeSpace int64 = 0
	// if w.dir.maxSize > 0 {
	// 	sizeInBytes, err := w.dir.sizeInBytes()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	freeSpace = w.dir.maxSize - sizeInBytes
	// }
	// var realUsage int64 = 0

	// // Enforce disk full:
	// if w.dir.maxSize > 0 && freeSpace <= int64(len(buf)) {
	// 	// Compute the real disk free. This will greatly slow down our
	// 	// test but makes it more accurate:
	// 	realUsage, err = w.dir.recomputeActualSizeInBytes()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	freeSpace = w.dir.maxSize - realUsage
	// }

	// if w.dir.maxSize > 0 && freeSpace <= int64(len(buf)) {
	// 	if freeSpace > 0 {
	// 		realUsage += freeSpace
	// 		if buf != nil {
	// 			w.delegate.WriteBytes(buf[:freeSpace])
	// 		} else {
	// 			w.CopyBytes(in, int64(len(buf)))
	// 		}
	// 	}
	// 	if realUsage > w.dir.maxUsedSize {
	// 		w.dir.maxUsedSize = realUsage
	// 	}
	// 	n, err := w.dir.recomputeActualSizeInBytes()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fLen, err := w.delegate.Length()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	message := fmt.Sprintf("fake disk full at %v bytes when writing %v (file length=%v",
	// 		n, w.name, fLen)
	// 	if freeSpace > 0 {
	// 		message += fmt.Sprintf("; wrote %v of %v bytes", freeSpace, len(buf))
	// 	}
	// 	message += ")"
	// 	if VERBOSE {
	// 		log.Println("MDW: now throw fake disk full")
	// 		debug.PrintStack()
	// 	}
	// 	return errors.New(message)
	// }
	// return nil
}

func (w *MockIndexOutputWrapper) Close() (err error) {
	panic("not implemented yet")
	// defer func() {
	// 	err2 := w.delegate.Close()
	// 	if err2 != nil {
	// 		err = mergeError(err, err2)
	// 		return
	// 	}
	// 	if w.dir.trackDiskUsage {
	// 		// Now compute actual disk usage & track the maxUsedSize
	// 		// in the MDW:
	// 		size, err2 := w.dir.recomputeActualSizeInBytes()
	// 		if err2 != nil {
	// 			err = mergeError(err, err2)
	// 			return
	// 		}
	// 		if size > w.dir.maxUsedSize {
	// 			w.dir.maxUsedSize = size
	// 		}
	// 	}
	// 	w.dir.removeIndexOutput(w, w.name)
	// }()
	// return w.dir.maybeThrowDeterministicException()
}

func (w *MockIndexOutputWrapper) Flush() error {
	panic("not implemented yet")
	// defer w.delegate.Flush()
	// return w.dir.maybeThrowDeterministicException()
}

func (w *MockIndexOutputWrapper) WriteByte(b byte) error {
	w.singleByte[0] = b
	return w.WriteBytes(w.singleByte)
}

func (w *MockIndexOutputWrapper) WriteBytes(buf []byte) error {
	assert(w.delegate != nil)
	err := w.checkCrashed()
	if err != nil {
		return err
	}
	err = w.checkDiskFull(buf, nil)
	if err != nil {
		return err
	}

	if w.dir.randomState.Intn(200) == 0 {
		half := len(buf) / 2
		err = w.delegate.WriteBytes(buf[:half])
		if err != nil {
			return err
		}
		runtime.Gosched()
		err = w.delegate.WriteBytes(buf[half:])
		if err != nil {
			return err
		}
	} else {
		err = w.delegate.WriteBytes(buf)
		if err != nil {
			return err
		}
	}

	if w.first {
		// Maybe throw random error; only do this on first write to a new file:
		w.first = false
		err = w.dir.maybeThrowIOException(w.name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *MockIndexOutputWrapper) FilePointer() int64 {
	return w.delegate.FilePointer()
}

func (w *MockIndexOutputWrapper) Length() (int64, error) {
	panic("not implemented yet")
	// return w.delegate.Length()
}

func (w *MockIndexOutputWrapper) CopyBytes(input util.DataInput, numBytes int64) error {
	panic("not implemented yet")
}

func (w *MockIndexOutputWrapper) String() string {
	return fmt.Sprintf("MockIndexOutputWrapper(%v,%v)", reflect.TypeOf(w.delegate), w.delegate)
}
