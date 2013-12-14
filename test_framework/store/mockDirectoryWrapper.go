package store

import (
	cs "github.com/balzaczyy/golucene/core/store"
	"math/rand"
	"sync"
)

type MockDirectoryWrapper struct {
	*BaseDirectoryWrapper
	sync.Locker // simulate Java's synchronized keyword

	randomState        *rand.Rand
	noDeleteOpenFile   bool
	preventDoubleWrite bool
	trackDiskUsage     bool
	wrapLockFactory    bool
	unSyncedFiles      map[string]bool
	createdFiles       map[string]bool

	openFilesForWrite map[string]bool
	// openLocks map[string]bool // synchronized

	throttling Throttling

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

func NewMockDirectoryWrapper(random *rand.Rand, delegate cs.Directory) *MockDirectoryWrapper {
	ans := MockDirectoryWrapper{
		noDeleteOpenFile:   true,
		preventDoubleWrite: true,
		trackDiskUsage:     false,
		wrapLockFactory:    true,
		openFilesForWrite:  make(map[string]bool),
		// openLocks: make(map[string]bool),
		throttling:      THROTTLING_SOMETIMES,
		inputCloneCount: 0,
		// openFileHandles: make(map[io.Closer]error),
	}
	ans.BaseDirectoryWrapper = NewBaseDirectoryWrapper(delegate)
	ans.Locker = &sync.Mutex{}
	// must make a private random since our methods are called from different
	// methods; else test failures may not be reproducible from the original
	// seed
	ans.randomState = rand.New(rand.NewSource(random.Int63()))
	// ans.throttledOutput =
	panic("not implemented yet")
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
