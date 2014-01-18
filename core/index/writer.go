package index

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"sync"
)

// index/IndexCommit.java

/*
Expert: represents a single commit into an index as seen by the
IndexDeletionPolicy or IndexReader.

Changes to the content of an index are made visible only after the
writer who made that change commits by writing a new segments file
(segments_N). This point in time, when the action of writing of a new
segments file to the directory is completed, is an index commit.

Each index commit oint has a unique segments file associated with it.
The segments file associated with a later index commit point would
have a larger N.
*/
type IndexCommit interface {
	// Get the segments file (segments_N) associated with the commit point.
	SegmentsFileName() string
	// Returns all index files referenced by this commit point.
	FileNames() (names []string, err error)
	// Returns the Directory for the index.
	Directory() store.Directory
	/*
		Delete this commit point. This only applies when using the commit
		point in the context of IndexWriter's IndexDeletionPolicy.

		Upon calling this, the writer is notified that this commit point
		should be deleted.

		Decision that a commit-point should be deleted is taken by the
		IndexDeletionPolicy in effect and therefore this should only be
		called by its onInit() or onCommit() methods.
	*/
	Delete()
	// Returns true if this commit should be deleted; this is only used
	// by IndexWriter after invoking the IndexDeletionPolicy.
	IsDeleted() bool
	// returns number of segments referenced by this commit.
	SegmentCount() int
	// Returns the generation (the _N in segments_N) for this IndexCommit
	Generation() int64
	// Returns userData, previously passed to SetCommitData(map) for this commit.
	UserData() (data map[string]string, err error)
}

type IndexCommits []IndexCommit

func (s IndexCommits) Len() int      { return len(s) }
func (s IndexCommits) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s IndexCommits) Less(i, j int) bool {
	if s[i].Directory() != s[j].Directory() {
		panic("cannot compare IndexCommits from different Directory instances")
	}
	return s[i].Generation() < s[j].Generation()
}

// index/LiveIndexWriterConfig.java

// Used by search package to assign a default similarity
var DefaultSimilarity func() Similarity

/*
Holds all the configuration used by IndexWriter with few setters for
settings that can be changed on an IndexWriter instance "live".

All the fields are either readonly or volatile.
*/
type LiveIndexWriterConfig struct {
	analyzer analysis.Analyzer

	maxBufferedDocs         int
	ramBufferSizeMB         float64
	maxBufferedDeleteTerms  int
	readerTermsIndexDivisor int
	mergedSegmentWarmer     IndexReaderWarmer
	termIndexInterval       int
	// TODO: this should be private to the codec, not settable here

	// controlling when commit points are deleted.
	delPolicy IndexDeletionPolicy

	// IndexCommit that IndexWriter is opened on.
	commit IndexCommit

	// OpenMode that IndexWriter is opened with.
	openMode OpenMode

	// Similarity to use when encoding norms.
	similarity Similarity

	// MergeScheduler to use for running merges.
	mergeScheduler MergeScheduler

	// Timeout when trying to obtain the write lock on init.
	writeLockTimeout int64

	// IndexingChain that determines how documents are indexed.
	indexingChain IndexingChain

	// Codec used to write new segments.
	codec Codec

	// InfoStream for debugging messages.
	infoStream util.InfoStream

	// MergePolicy for selecting merges.
	mergePolicy MergePolicy

	// DocumentsWriterPerThreadPool to control how goroutines are
	// allocated to DocumentsWriterPerThread.
	indexerThreadPool *DocumentsWriterPerThreadPool

	// True if readers should be pooled.
	readerPooling bool

	// FlushPolicy to control when segments are flushed.
	flushPolicy FlushPolicy

	// Sets the hard upper bound on RAM usage for a single segment,
	// after which the segment is forced to flush.
	perRoutineHardLimitMB int

	// Version that IndexWriter should emulate.
	matchVersion util.Version

	// True is segment flushes should use compound file format
	useCompoundFile bool
}

// used by IndexWriterConfig
func newLiveIndexWriterConfig(analyzer analysis.Analyzer, matchVersion util.Version) *LiveIndexWriterConfig {
	assert(DefaultSimilarity != nil)
	assert(DefaultCodec != nil)
	return &LiveIndexWriterConfig{
		analyzer:                analyzer,
		matchVersion:            matchVersion,
		ramBufferSizeMB:         DEFAULT_RAM_BUFFER_SIZE_MB,
		maxBufferedDocs:         DEFAULT_MAX_BUFFERED_DOCS,
		maxBufferedDeleteTerms:  DEFAULT_MAX_BUFFERED_DELETE_TERMS,
		readerTermsIndexDivisor: DEFAULT_READER_TERMS_INDEX_DIVISOR,
		termIndexInterval:       DEFAULT_TERM_INDEX_INTERVAL, // TODO: this should be private to the codec, not settable here
		delPolicy:               DEFAULT_DELETION_POLICY,
		useCompoundFile:         DEFAULT_USE_COMPOUND_FILE_SYSTEM,
		openMode:                OPEN_MODE_CREATE_OR_APPEND,
		similarity:              DefaultSimilarity(),
		mergeScheduler:          NewConcurrentMergeScheduler(),
		writeLockTimeout:        WRITE_LOCK_TIMEOUT,
		indexingChain:           defaultIndexingChain,
		codec:                   DefaultCodec(),
		infoStream:              util.DefaultInfoStream(),
		mergePolicy:             NewTieredMergePolicy(),
		flushPolicy:             newFlushByRamOrCountsPolicy(),
		readerPooling:           DEFAULT_READER_POOLING,
		indexerThreadPool:       newThreadAffinityDocumentsWriterPerThreadPool(DEFAULT_MAX_THREAD_STATES),
		perRoutineHardLimitMB:   DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB,
	}
}

// Creates a new config that handles the live IndexWriter settings.
func newLiveIndexWriterConfigFrom(config *IndexWriterConfig) *LiveIndexWriterConfig {
	return &LiveIndexWriterConfig{
		maxBufferedDeleteTerms:  config.maxBufferedDeleteTerms,
		maxBufferedDocs:         config.maxBufferedDocs,
		mergedSegmentWarmer:     config.mergedSegmentWarmer,
		ramBufferSizeMB:         config.ramBufferSizeMB,
		readerTermsIndexDivisor: config.readerTermsIndexDivisor,
		termIndexInterval:       config.termIndexInterval,
		matchVersion:            config.matchVersion,
		analyzer:                config.analyzer,
		delPolicy:               config.delPolicy,
		commit:                  config.commit,
		openMode:                config.openMode,
		similarity:              config.similarity,
		mergeScheduler:          config.mergeScheduler,
		writeLockTimeout:        config.writeLockTimeout,
		indexingChain:           config.indexingChain,
		codec:                   config.codec,
		infoStream:              config.infoStream,
		mergePolicy:             config.mergePolicy,
		indexerThreadPool:       config.indexerThreadPool,
		readerPooling:           config.readerPooling,
		flushPolicy:             config.flushPolicy,
		perRoutineHardLimitMB:   config.perRoutineHardLimitMB,
		useCompoundFile:         config.useCompoundFile,
	}
}

// L358
/*
Determines the minimal number of documents required before the
buffered in-memory documents are flushed as a new Segment. Large
values generally give faster indexing.

When this is set, the writer will flush every maxBufferedDocs added
documents. Pass in DISABLE_AUTO_FLUSH to prevent triggering a flush
due to number of buffered documents. Note that if flushing by RAM
usage is also enabled, then the flush will be triggered by whichever
comes first.

Disabled by default (writer flushes by RAM usage).

Takes effect immediately, but only the next time a document is added,
updated or deleted.
*/
func (conf *LiveIndexWriterConfig) SetMaxBufferedDocs(maxBufferedDocs int) *LiveIndexWriterConfig {
	assert2(maxBufferedDocs == DISABLE_AUTO_FLUSH || maxBufferedDocs >= 2,
		"maxBufferedDocs must at least be 2 when enabled")
	assert2(maxBufferedDocs != DISABLE_AUTO_FLUSH || conf.ramBufferSizeMB != DISABLE_AUTO_FLUSH,
		"at least one of ramBufferSize and maxBufferedDocs must be enabled")
	conf.maxBufferedDocs = maxBufferedDocs
	return conf
}

/*
Sets the merged segment warmer.

Take effect on the next merge.
*/
func (conf *LiveIndexWriterConfig) SetMergedSegmentWarmer(mergeSegmentWarmer IndexReaderWarmer) *LiveIndexWriterConfig {
	conf.mergedSegmentWarmer = mergeSegmentWarmer
	return conf
}

/*
Sets the termsIndeDivisor passed to any readers that IndexWriter
opens, for example when applying deletes or creating a near-real-time
reader in OpenDirectoryReader(). If you pass -1, the terms index
won't be loaded by the readers. This is only useful in advanced
siguations when you will only .Next() through all terms; attempts to
seek will hit an error.

takes effect immediately, but only applies to readers opened after
this call

NOTE: divisor settings > 1 do not apply to all PostingsFormat
implementation, including the default one in this release. It only
makes sense for terms indexes that can efficiently re-sample terms at
load time.
*/
func (conf *LiveIndexWriterConfig) SetReaderTermsIndexDivisor(divisor int) *LiveIndexWriterConfig {
	assert2(divisor > 0 || divisor == -1, fmt.Sprintf(
		"divisor must be >= 1, or -1 (got %v)", divisor))
	conf.readerTermsIndexDivisor = divisor
	return conf
}

/*
Sets if the IndexWriter should pack newly written segments in a
compound file. Default is true.

Use false for batch indexing with very large ram buffer settings.

Note: To control compound file usage during segment merges see
SetNoCFSRatio() and SetMaxCFSSegmentSizeMB(). This setting only
applies to newly created segment.
*/
func (conf *LiveIndexWriterConfig) SetUseCompoundFile(useCompoundFile bool) *LiveIndexWriterConfig {
	conf.useCompoundFile = useCompoundFile
	return conf
}

// index/IndexWriterConfig.java

// Specifies the open mode for IndeWriter
type OpenMode int

const (
	// Creates a new index or overwrites an existing one.
	OPEN_MODE_CREATE = OpenMode(1)
	// Opens an existing index.
	OPEN_MODE_APPEND = OpenMode(2)
	// Creates a new index if one does not exist,
	// otherwise it opens the index and documents will be appended.
	OPEN_MODE_CREATE_OR_APPEND = OpenMode(3)
)

// Default value is 32.
const DEFAULT_TERM_INDEX_INTERVAL = 32 // TODO: this should be private to the codec, not settable here

// Denotes a flush trigger is disabled.
const DISABLE_AUTO_FLUSH = -1

// Disabled by default (because IndexWriter flushes by RAM usage by default).
const DEFAULT_MAX_BUFFERED_DELETE_TERMS = DISABLE_AUTO_FLUSH

// Disabled by default (because IndexWriter flushes by RAM usage by default).
const DEFAULT_MAX_BUFFERED_DOCS = DISABLE_AUTO_FLUSH

// Default value is 16 MB (which means flush when buffered docs
// consume approximately 16 MB RAM)
const DEFAULT_RAM_BUFFER_SIZE_MB = 16

// Default value for the write lock timeout (1,000 ms)
const WRITE_LOCK_TIMEOUT = 1000

const DEFAULT_READER_POOLING = false

// Default value is 1.
const DEFAULT_READER_TERMS_INDEX_DIVISOR = DEFAULT_TERMS_INDEX_DIVISOR

// Default value is 1945.
const DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB = 1945

// The maximum number of simultaneous threads that may be indexing
// documents at once in IndexWriter; if more than this many threads
// arrive they will wait for others to finish. Default value is 8.
const DEFAULT_MAX_THREAD_STATES = 8

// Default value for compound file system for newly written segments
// (set to true). For batch indexing with very large ram buffers use
// false.
const DEFAULT_USE_COMPOUND_FILE_SYSTEM = true

/*
Holds all the configuration that is used to create an IndexWriter. Once
IndexWriter has been created with this object, changes to this object will not
affect the IndexWriter instance. For that, use LiveIndexWriterConfig that is
returned from IndexWriter.Config().

All setter methods return IndexWriterConfig to allow chaining settings
conveniently, for example:

		conf := NewIndexWriterConfig(analyzer)
						.setter1()
						.setter2()
*/
type IndexWriterConfig struct {
	*LiveIndexWriterConfig
	writer *util.SetOnce
}

// Sets the IndexWriter this config is attached to.
func (conf *IndexWriterConfig) setIndexWriter(writer *IndexWriter) *IndexWriterConfig {
	conf.writer.Set(writer)
	return conf
}

/*
Creates a new config that with defaults that match the specified
Version as well as the default Analyzer. If matchVersion is >= 3.2,
TieredMergePolicy is used for merging; else LogByteSizeMergePolicy.
Note that TieredMergePolicy is free to select non-contiguous merges,
which means docIDs may not remain monotonic over time. If this is a
problem, you should switch to LogByteSizeMergePolicy or
LogDocMergePolicy.
*/
func NewIndexWriterConfig(matchVersion util.Version, analyzer analysis.Analyzer) *IndexWriterConfig {
	return &IndexWriterConfig{
		LiveIndexWriterConfig: newLiveIndexWriterConfig(analyzer, matchVersion),
		writer:                util.NewSetOnce(),
	}
}

func (conf *IndexWriterConfig) Clone() *IndexWriterConfig {
	panic("not implemented yet")
}

/*
Expert: allows an optional IndexDeletionPolicy implementation to be
specified. You can use this to control when prior commits are deleted
from the index. The default policy is
KeepOnlyLastCommitDeletionPolicy which removes all prior commits as
soon as a new commit is done (this matches behavior before 2.2).
Creating your own policy can allow you to explicitly keep previous
"point in time" commits alive in the index for some time, to allow
readers to refresh to the new commit without having the old commit
deleted out from under them. This is necessary on filesystems like
NFS that do not support "delete on last close" semantics, which
Lucene's "point in time" search normally relies on.

NOTE: the deletion policy can not be nil
*/
func (conf *IndexWriterConfig) SetIndexDeletionPolicy(delPolicy IndexDeletionPolicy) *IndexWriterConfig {
	if delPolicy == nil {
		panic("indexDeletionPolicy must not be nil")
	}
	conf.delPolicy = delPolicy
	return conf
}

type Similarity interface {
	ComputeNorm(fs *FieldInvertState) int64
}

// L259
/*
Expert: set the Similarity implementation used by this IndexWriter.

NOTE: the similarity cannot be nil.

Only takes effect when IndexWriter is first created.
*/
func (conf *IndexWriterConfig) SetSimilarity(similarity Similarity) *IndexWriterConfig {
	assert2(similarity != nil, "similarity must not be nil")
	conf.similarity = similarity
	return conf
}

/*
Expert: sets the merge scheduler used by this writer. The default is
ConcurentMergeScheduler.

NOTE: the merge scheduler cannot be nil.

Only takes effect when IndexWriter is first created.
*/
func (conf *IndexWriterConfig) SetMergeScheduler(mergeScheduler MergeScheduler) *IndexWriterConfig {
	assert2(mergeScheduler != nil, "mergeScheduler must not be nil")
	conf.mergeScheduler = mergeScheduler
	return conf
}

/*
Expert: MergePolicy is invoked whenver there are changes to the
segments in the index. Its role is to select which merges to do, if
any, and return a MergeSpecification describing the merges. It also
selects merges to do for forceMerge.
*/
func (conf *IndexWriterConfig) SetMergePolicy(mergePolicy MergePolicy) *IndexWriterConfig {
	assert2(mergePolicy != nil, "mergePolicy must not be nil")
	conf.mergePolicy = mergePolicy
	return conf
}

// L406
/*
By default, IndexWriter does not pool the SegmentReaders it must open
for deletions and merging, unless a near-real-time reader has been
obtained by calling openDirectoryReader(IndexWriter, bool). This
method lets you enable pooling without getting a near-real-time
reader. NOTE: if you set this to false, IndexWriter will still pool
readers once openDirectoryReader(IndexWriter, bool) is called.
*/
func (conf *IndexWriterConfig) SetReaderPooling(readerPooling bool) *IndexWriterConfig {
	conf.readerPooling = readerPooling
	return conf
}

// L478
func (conf *IndexWriterConfig) InfoStream() util.InfoStream {
	return conf.infoStream
}

// L548
func (conf *IndexWriterConfig) SetMaxBufferedDocs(maxBufferedDocs int) *IndexWriterConfig {
	conf.LiveIndexWriterConfig.SetMaxBufferedDocs(maxBufferedDocs)
	return conf
}

func (conf *IndexWriterConfig) SetMergedSegmentWarmer(mergeSegmentWarmer IndexReaderWarmer) *IndexWriterConfig {
	conf.LiveIndexWriterConfig.SetMergedSegmentWarmer(mergeSegmentWarmer)
	return conf
}

func (conf *IndexWriterConfig) SetReaderTermsIndexDivisor(divisor int) *IndexWriterConfig {
	conf.LiveIndexWriterConfig.SetReaderTermsIndexDivisor(divisor)
	return conf
}

func (conf *IndexWriterConfig) SetUseCompoundFile(useCompoundFile bool) *IndexWriterConfig {
	conf.LiveIndexWriterConfig.SetUseCompoundFile(useCompoundFile)
	return conf
}

func (conf *IndexWriterConfig) String() string {
	panic("not implemented yet")
}

// index/IndexWriter.java

// Use a seprate goroutine to protect closing control
type ClosingControl struct {
	_closed   bool // volatile
	_closing  bool // volatile
	isRunning bool
	closer    chan func() (bool, error)
	done      chan error
}

func newClosingControl() *ClosingControl {
	ans := &ClosingControl{
		closer: make(chan func() (bool, error)),
		done:   make(chan error),
	}
	go ans.daemon()
	return ans
}

func (cc *ClosingControl) daemon() {
	var err error
	for cc.isRunning {
		err = nil
		select {
		case f := <-cc.closer:
			if !cc._closed {
				cc._closing = true
				cc._closed, err = f()
				cc._closing = false
			}
			cc.done <- err
		}
	}
}

// Used internally to throw an AlreadyClosedError if this IndexWriter
// has been closed or is in the process of closing.
func (cc *ClosingControl) ensureOpen(failIfClosing bool) {
	assert2(cc._closed || failIfClosing && cc._closing, "this IndexWriter is closed")
}

func (cc *ClosingControl) close(f func() (ok bool, err error)) error {
	if cc._closed {
		return nil // already closed
	}
	cc.closer <- f
	return <-cc.done
}

// Name of the write lock in the index.
const WRITE_LOCK_NAME = "write.lock"

/*
Absolute hard maximum length for a term, in bytes once encoded as
UTF8. If a term arrives from the analyzer longer than this length,
it is skipped and a message is printed to infoStream, if set.
*/
const MAX_TERM_LENGTH = MAX_TERM_LENGTH_UTF8

/*
An IndexWriter creates and maintains an index.

The OpenMode option on IndexWriterConfig.SetOpenMode() determines
whether a new index is created, or whether an existing index is
opened. Note that you can open an index with OPEN_MODE_CREATE even
while readers are using the index. The old readers will continue to
search the "point in time" snapshot they had opened, and won't see
the newly created index until they re-open. If OPEN_MODE_CREATE_OR_APPEND
is used, IndexWriter will create a new index if there is not already
an index at the provided path and otherwise open th existing index.

In either case, documents are added with AddDocument() and removed
with DeleteDocumentsByQuery(). A document can be updated with
UpdateDocuments() (which just deletes and then adds the entire
document). When finished adding, deleting and updating documents,
Close() should be called.

...
*/
type IndexWriter struct {
	sync.Locker
	*ClosingControl

	hitOOM bool // volatile

	directory store.Directory   // where this index resides
	analyzer  analysis.Analyzer // how to analyze text

	rollbackSegments []*SegmentInfoPerCommit // list of segmentInfo we will fallback to if the commit fails

	segmentInfos         *SegmentInfos // the segments
	globalFieldNumberMap *FieldNumbers

	docWriter  *DocumentsWriter
	eventQueue *list.List
	deleter    *IndexFileDeleter

	// used by forceMerge to note those needing merging
	segmentsToMerge map[*SegmentInfoPerCommit]bool

	writeLock store.Lock

	closed  bool // volatile
	closing bool // volatile

	// Holds all SegmentInfo instances currently involved in merges
	mergingSegments map[*SegmentInfoPerCommit]bool

	mergePolicy     MergePolicy
	mergeScheduler  MergeScheduler
	pendingMerges   *list.List
	runningMerges   map[*OneMerge]bool
	mergeExceptions []*OneMerge

	flushCount        int // atomic
	flushDeletesCount int // atomic

	readerPool            *ReaderPool
	bufferedDeletesStream *BufferedDeletesStream

	// This is a "write once" variable (like the organic dye on a DVD-R
	// that may or may not be heated by a laser and then cooled to
	// permanently record the event): it's false, until Reader() is
	// called for the first time, at which point it's switched to true
	// and never changes back to false. Once this is true, we hold open
	// and reuse SegmentReader instances internally for applying
	// deletes, doing merges, and reopening near real-time readers.
	poolReaders bool

	// The instance that we passed to the constructor. It is saved only
	// in order to allow users to query an IndexWriter settings.
	config *LiveIndexWriterConfig

	codec Codec // for writing new segments

	// If non-nil, information about merges will be printed to this.
	infoStream util.InfoStream

	// A hook for extending classes to execute operations after pending
	// and deleted documents have been flushed ot the Directory but
	// before the change is committed (new segments_N file written).
	doAfterFlush func() error
	// A hook for extending classes to execute operations before
	// pending added and deleted documents are flushed to the Directory.
	doBeforFlush func() error

	// Used only by commit and prepareCommit, below; lock order is
	// commitLock -> IW
	commitLock sync.Locker

	// Ensures only one flush() is actually flushing segments at a time:
	fullFlushLock sync.Locker

	// Called internally if any index state has changed.
	changed chan bool
}

// L421
type ReaderPool struct {
	owner *IndexWriter
	sync.Locker
	readerMap map[*SegmentInfoPerCommit]*ReadersAndLiveDocs
}

func newReaderPool(owner *IndexWriter) *ReaderPool {
	return &ReaderPool{
		owner:     owner,
		Locker:    &sync.Mutex{},
		readerMap: make(map[*SegmentInfoPerCommit]*ReadersAndLiveDocs),
	}
}

// Remove all our references to readers, and commits any pending changes.
func (pool *ReaderPool) dropAll(doSave bool) error {
	pool.Lock() // synchronized
	defer pool.Unlock()
	panic("not implemented yet")
}

// Commit live docs changes for the segment readers for the provided infos.
func (pool *ReaderPool) commit(infos *SegmentInfos) error {
	pool.Lock() // synchronized
	defer pool.Unlock()
	panic("not implemented yet")
}

// Obtain a readersAndLiveDocs instance from the ReaderPool. If
// create is true, you must later call release().
func (pool *ReaderPool) get(info *SegmentInfoPerCommit, create bool) *ReadersAndLiveDocs {
	pool.Lock() // synchronized
	defer pool.Unlock()
	panic("not implemented yet")
}

// Obtain the number of deleted docs for a pooled reader. If the
// reader isn't being pooled, the segmentInfo's delCount is returned.
func (w *IndexWriter) numDeletedDocs(info *SegmentInfoPerCommit) int {
	panic("not implemented yet")
}

// Used internally to throw an AlreadyClosedError if this IndexWriter
// has been closed or is in the process of closing.
func (w *IndexWriter) ensureOpenOrPanic(failIfClosing bool) {
	w.Lock() // volatile check
	defer w.Unlock()
	if w.closed || failIfClosing && w.closing {
		panic("this IndexWriter is closed")
	}
}

/*
Used internally to throw an AlreadyClosedError if this IndexWriter
has been closed or is in the process of closing.

Calls ensureOpen(true).
*/
func (w *IndexWriter) ensureOpen() {
	w.ensureOpenOrPanic(true)
}

/*
Constructs a new IndexWriter per the settings given in conf. If you want to
make "live" changes to this writer instance, use Config().

NOTE: after this writer is created, the given configuration instance cannot be
passed to another writer. If you intend to do so, you should clone it
beforehand.
*/
func NewIndexWriter(d store.Directory, conf *IndexWriterConfig) (w *IndexWriter, err error) {
	ans := &IndexWriter{
		Locker:         &sync.Mutex{},
		ClosingControl: newClosingControl(),

		segmentsToMerge: make(map[*SegmentInfoPerCommit]bool),
		mergingSegments: make(map[*SegmentInfoPerCommit]bool),
		pendingMerges:   list.New(),
		runningMerges:   make(map[*OneMerge]bool),
		mergeExceptions: make([]*OneMerge, 0),

		config:         newLiveIndexWriterConfigFrom(conf),
		directory:      d,
		analyzer:       conf.analyzer,
		infoStream:     conf.infoStream,
		mergePolicy:    conf.mergePolicy,
		mergeScheduler: conf.mergeScheduler,
		codec:          conf.codec,

		bufferedDeletesStream: newBufferedDeletesStream(conf.infoStream),
		poolReaders:           conf.readerPooling,

		writeLock: d.MakeLock(WRITE_LOCK_NAME),

		changed: make(chan bool),
	}
	ans.readerPool = newReaderPool(ans)

	conf.setIndexWriter(ans)
	ans.mergePolicy.SetIndexWriter(ans)

	// obtain write lock
	if ok, err := ans.writeLock.ObtainWithin(conf.writeLockTimeout); !ok || err != nil {
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("Index locked for write: %v", ans.writeLock))
	}

	var success bool = false
	defer func() {
		if !success {
			if ans.infoStream.IsEnabled("IW") {
				ans.infoStream.Message("IW", "init: hit exception on init; releasing write lock")
			}
			ans.writeLock.Release() // don't mask the original exception
			ans.writeLock = nil
		}
	}()

	var create bool
	switch conf.openMode {
	case OPEN_MODE_CREATE:
		create = true
	case OPEN_MODE_APPEND:
		create = false
	default:
		// CREATE_OR_APPEND - create only if an index does not exist
		ok, err := IsIndexExists(d)
		if err != nil {
			return nil, err
		}
		create = !ok
	}

	// If index is too old, reading the segments will return
	// IndexFormatTooOldError
	ans.segmentInfos = &SegmentInfos{}

	var initialIndexExists bool = true

	if create {
		// Try to read first. This is to allow create against an index
		// that's currently open for searching. In this case we write the
		// next segments_N file with no segments:
		err = ans.segmentInfos.ReadAll(d)
		if err == nil {
			ans.segmentInfos.Clear()
		} else {
			// Likely this means it's a fresh directory
			initialIndexExists = false
			err = nil
		}

		// Record that we have a change (zero out all segments) pending:
		ans.changed <- true
	} else {
		err = ans.segmentInfos.ReadAll(d)
		if err != nil {
			return
		}

		if commit := conf.commit; commit != nil {
			// Swap out all segments, but, keep metadta in SegmentInfos,
			// like version & generation, to preserve write-once. This is
			// important if readers are open against the future commit
			// points.
			assert2(commit.Directory() == d,
				"IndexCommit's directory doesn't match my directory")
			oldInfos := &SegmentInfos{}
			ans.segmentInfos.replace(oldInfos)
			ans.changed <- true
			ans.infoStream.Message("IW", fmt.Sprintf(
				"init: loaded commit '%v'", commit.SegmentsFileName()))
		}
	}

	ans.rollbackSegments = ans.segmentInfos.createBackupSegmentInfos()

	// start with previous field numbers, but new FieldInfos
	ans.globalFieldNumberMap, err = ans.fieldNumberMap()
	if err != nil {
		return
	}
	ans.config.flushPolicy.init(ans.config)
	ans.docWriter = newDocumentsWriter(ans, ans.config, d)
	ans.eventQueue = ans.docWriter.events

	// Default deleter (for backwards compatibility) is
	// KeepOnlyLastCommitDeleter:
	ans.deleter, err = newIndexFileDeleter(d, conf.delPolicy,
		ans.segmentInfos, ans.infoStream, ans, initialIndexExists)
	if err != nil {
		return
	}

	if ans.deleter.startingCommitDeleted {
		// Deletion policy deleted the "head" commit point. We have to
		// mark outsef as changed so that if we are closed w/o any
		// further changes we write a new segments_N file.
		ans.changed <- true
	}

	ans.infoStream.Message("IW", fmt.Sprintf("init: create=%v", create))
	ans.messageState()

	success = true
	return ans, nil
}

func (w *IndexWriter) fieldInfos(info *SegmentInfo) (infos FieldInfos, err error) {
	panic("not implemented yet")
}

/*
Loads or returns the alread loaded the global field number map for
this SegmentInfos. If this SegmentInfos has no global field number
map the returned instance is empty.
*/
func (w *IndexWriter) fieldNumberMap() (m *FieldNumbers, err error) {
	panic("not implemented yet")
}

func (w *IndexWriter) messageState() {
	panic("not implemented yet")
}

/*
Commits all changes to an index, wait for pending merges to complete,
and closes all associate files.

This is a "slow graceful shutdown" which may take a long time
especially if a big merge is pending. If you only want to close
resources, use rollback(). If you only want to commit pending changes
and close resources, see closeAndWait().

Note that this may be a costly operation, so, try to re-use a single
writer instead of closing and opening a new one. See commit() for
caveats about write caching done by some IO devices.

If an error is hit during close, e.g., due to disk full or some other
reason, then both the on-disk index and the internal state of the
IndexWriter instance will be consistent. However, the close will not
be complete even though part of it (flushing buffered documents) may
have succeeded, so the write lock will still be held.

If you can correct the underlying cause (e.g., free up some disk
space) then you can call close() again. Failing that, if you want to
force the write lock to be released (dangerous, because ou may then
lose buffered docs in the IndexWriter instance) then you can do
something like this:

	defer func() {
		if IsDirectoryLocked(directory) {
			UnlockDIrectory(directory)
		}
	}
	err = writer.Close()

after which, you must be certain not ot use the writer instance
anymore.

NOTE: if this method hits a memory issue, you should immediately
close the writer, again. See above for details. But it's probably
impossible for GoLucene.
*/
func (w *IndexWriter) Close() error {
	return w.CloseAndWait(true)
}

/*
Closes the index with or without waiting for currently running merges
to finish. This is only meaningful when using a MergeScheduler that
runs  merges in background threads.

NOTE: if this method hits a memory issue, you should immediately
close the writer, again. See above for details. But it's probably
impossible for GoLucene.

NOTE: it is dangerous to always call closeAndWait(false), especially
when IndexWriter is not open for very long, because this can result
in "merge starvation" whereby long merges will never have a chance to
finish. This will cause too many segments in your index over time.
*/
func (w *IndexWriter) CloseAndWait(waitForMerge bool) error {
	// Ensure that only one goroutine actaully gets to do the closing,
	// and make sure no commit is also in progress:
	w.commitLock.Lock()
	defer w.commitLock.Unlock()
	return w.close(func() (ok bool, err error) {
		// If any methods have hit memory issue, then abort on
		// close, in case the internal state of IndexWriter or
		// DocumentsWriter is corrupt
		if w.hitOOM {
			return w.rollbackInternal()
		}
		return w.closeInternal(waitForMerge, true)
	})
}

/*
Returns true if this goroutine should attempt to close, or false if
IndexWriter is now closed; else, waits until another thread finishes
closing.
*/

func (w *IndexWriter) closeInternal(waitForMerges bool, doFlush bool) (ok bool, err error) {
	panic("not implemented yet")
}

// Retuns the Directory used by this index.
func (w *IndexWriter) Directory() store.Directory {
	return w.directory
}

// L1201
/*
Adds a document to this index.

Note that if an Error is hit (for example disk full) then the index
will be consistent, but this document may not have been added.
Furthermore, it's possible the index will have one segment in
non-compound format even when using compound files (when a merge has
partially succeeded).

This method periodically flushes pending documents to the Directory
(see flush()), and also periodically triggers segment merges in the
index according to the MergePolicy in use.

Merges temporarily consume space in the directory. The amount of
space required is up to 1X the size of all segments being merged,
when no readers/searchers are open against the index, and up to 2X
the size of all segments being merged when readers/searchers are open
against the index (see forceMerge() for details). The sequence of
primitive merge operations performed is governed by the merge policy.

Note that each term in the document can be no longer than 16383
characters, otherwise error will be returned.

Note that it's possible to creat an invalid Unicode string in Java if
a UTF16 surrogate pair is malformed. In this case, the invalid
characters are silently replaced with the Unicode replacement
character U+FFFD.

NOTE: if this method hits a memory issue, you should immediately
close the writer. See above for details.
*/
func (w *IndexWriter) AddDocument(doc []IndexableField) error {
	return w.AddDocumentWithAnalyzer(doc, w.analyzer)
}

/*
Adds a document to this index, using the provided analyzer instead of
the value of Analyzer().

See AddDocument() for details on index and IndexWriter state after an
error, and flushing/merging temporary free space requirements.

NOTE: if this method hits a memory issue, you hsould immediately
close the writer. See above for details.
*/
func (w *IndexWriter) AddDocumentWithAnalyzer(doc []IndexableField, analyzer analysis.Analyzer) error {
	return w.UpdateDocument(nil, doc, analyzer)
}

// L1545
/*
Updates a document by first deleting the document(s) containing term
and then adding the new document. The delete and then add are atomic
as seen by a reader on the same index (flush may happen only after
the add).

NOTE: if this method hits a memory issue, you should immediately
close he write. See above for details.
*/
func (w *IndexWriter) UpdateDocument(term *Term, doc []IndexableField, analyzer analysis.Analyzer) error {
	panic("not implemented yet")
}

func (w *IndexWriter) newSegmentName() string {
	panic("not implemented yet")
}

/*
Forces merge policy to merge segments until there are <=
maxNumSegments. The actual merge to be executed are determined by the
MergePolicy.

This is a horribly costly operation, especially when you pass a small
maxNumSegments; usually you should only call this if the index is
static (will no longer be changed).

Note that this requires up to 2X the index size free space in your
Directory (3X if you're using compound file format). For example, if
your index size is 10 MB, then you need up to 20 MB free for this to
complete (30 MB if you're using compound file format). Also, it's
best to call commit() afterwards, to allow IndexWriter to free up
disk space.

If some but not all readers re-open while merging is underway, this
will cause > 2X temporary space to be consumed as those new readers
will then hold open the temporary segments at that time. it is best
not to re-open readers while merging is running.

The actual temporary usage could be much less than these figures (it
depends on many factors).

In general, once this completes, the total size of the index will be
less than the size of the starting index. It could be quite a bit
smaller (if there were many pending deletes) or just slightly smaller.

If an error is hit, for example, due to disk full, the index will not
be corrupted and no documents will be list. However, it may have been
partially merged (some segments were merged but not all), and it's
possible that one of the segments in the index will be in
non-compound format even when using compound file format. This will
occur when the error is hit during conversion of the segment into
compound format.

This call will merge those segments present in the index when call
started. If other routines are still adding documents and flushing
segments, those newly created segments will not be merged unless you
call forceMerge again.

NOTE: if this method hits a memory issue, you should immediately
close the writer.

NOTE: if you call CloseAndWait() with false, which aborts all running
merges, then any routine still running this method might hit a
MergeAbortedError.
*/
func (w *IndexWriter) forceMerge(maxNumSegments int) error {
	return w.forceMergeAndWait(maxNumSegments, true)
}

/*
Just like forceMerge(), except you can specify whether the call
should block until all merging completes. This is only meaningful
with  a Mergecheduler that is able to run merges in background
routines.

NOTE: if this method hits a memory issue, you should immediately
close the writer.
*/
func (w *IndexWriter) forceMergeAndWait(maxNumSegments int, doWait bool) error {
	panic("not implemented yet")
}

// Returns true if any merges in pendingMerges or runningMerges
// are maxNumSegments merges.
func (w *IndexWriter) maxSegmentsMergePending() bool {
	w.Lock() // synchronized
	defer w.Unlock()

	panic("not implemented yet")
}

func (w *IndexWriter) maybeMerge(trigger *MergeTrigger, maxNumSegments int) (err error) {
	w.ensureOpenOrPanic(false)
	if err = w.updatePendingMerges(trigger, maxNumSegments); err == nil {
		err = w.mergeScheduler.Merge(w)
	}
	return
}

func (w *IndexWriter) updatePendingMerges(trigger *MergeTrigger, maxNumSegments int) error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Experts: to be used by a MergePolicy to avoid selecting merges for
segments already being merged. The returned collection is not cloned,
and thus is only safe to access if you hold IndexWriter's lock (which
you do when IndexWriter invokes the MergePolicy).
*/

func (w *IndexWriter) MergingSegments() []*SegmentInfoPerCommit {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Expert: the MergeScheduler calls this method to retrieve the next
merge requested by the MergePolicy.
*/
func (w *IndexWriter) nextMerge() *OneMerge {
	w.Lock() // synchronized
	defer w.Unlock()

	if w.pendingMerges.Len() == 0 {
		return nil
	}
	// Advance the merge from pending to running
	merge := w.pendingMerges.Front().Value.(*OneMerge)
	w.pendingMerges.Remove(w.pendingMerges.Front())
	w.runningMerges[merge] = true
	return merge
}

// Expert: returns true if there are merges waiting to be scheduled.
func (w *IndexWriter) hasPendingMerges() bool {
	return w.pendingMerges.Len() > 0
}

/*
Close the IndexWriter without committing any changes that have
occurred since the last commit (or since it was opened, if commit
hasn't been called). This removes any temporary files that had been
created, after which the state of the index will be the same as it
was when commit() was last called or when this writer was first
opened. This also clears a previous call to prepareCommit()
*/
func (w *IndexWriter) Rollback() error {
	w.ensureOpen()
	return w.close(w.rollbackInternal)
}

func (w *IndexWriter) rollbackInternal() (ok bool, err error) {
	panic("not implemented yet")
}

func (w *IndexWriter) finishMerges(waitForMerge bool) {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Wait for any currently outstanding merges to finish.

It is guaranteed that any merges started prior to calling this method
will have completed once this method completes.
*/
func (w *IndexWriter) waitForMerges() {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Called whenever the SegmentInfos has been updatd and the index files
referenced exist (correctly) in the index directory.
*/
func (w *IndexWriter) checkpoint() error {
	w.Lock() // synchronized
	defer w.Unlock()
	w.changed <- true
	return w.deleter.checkpoint(w.segmentInfos, false)
}

/*
Atomically adds the segment private delete packet and publishes the
flushed segments SegmentInfo to the index writer.
*/
func (w *IndexWriter) publishFlushedSegment(newSegment *SegmentInfoPerCommit,
	packet *FrozenBufferedDeletes, globalPacket *FrozenBufferedDeletes) error {
	panic("not implemented yet")
}

func (w *IndexWriter) resetMergeExceptions() {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

func (w *IndexWriter) prepareCommitInternal() error {
	panic("not implemented yet")
}

/*
Commits all pending changes (added & deleted documents, segment
merges, added indexes, etc.) to the index, and syncs all referenced
index files, such that a reader will see the changes and the index
updates will survive an OS or machine crash or power loss. Note that
this does not wait for any running background merges to finish. This
may be a costly operation, so you should test the cost in your
application and do it only when really necessary.

Note that this operation calls Directory.sync on the index files.
That call  should not return until the file contents & metadata are
on stable storage. For FSDirectory, this calls the OS's fsync. But,
beware: some hardware devices may in fact cache writes even during
fsync, and return before the bits are actually on stable storage, to
give the appearance of faster performance. If you have such a device,
and it does not hav a battery backup (for example) then on power loss
it may still lose data. Lucene cannot guarantee consistency on such
devices.

NOTE: if this method hits a memory issue, you should immediately
close the writer.
*/
func (w *IndexWriter) Commit() error {
	w.ensureOpen()
	return w.commitInternal()
}

func (w *IndexWriter) commitInternal() error {
	panic("not implemented yet")
}

func (w *IndexWriter) finishCommit() error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Flush all in-memory buffered updates (adds and deletes) to the
Directory.
*/
func (w *IndexWriter) flush(triggerMerge bool, applyAllDeletes bool) error {
	panic("not implemented yet")
}

func (w *IndexWriter) doFlush(applyAllDeletes bool) (ok bool, err error) {
	panic("not implemented yet")
}

func (w *IndexWriter) maybeApplyDeletes(applyAllDeletes bool) error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

func (w *IndexWriter) applyAllDeletes() error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

// L3440
/*
Merges the indicated segments, replacing them in the stack with a
single segment.
*/
func (w *IndexWriter) merge(merge *OneMerge) error {
	panic("not implemented yet")
}

func setDiagnostics(info SegmentInfo, source string) {
	setDiagnosticsAndDetails(info, source, nil)
}

func setDiagnosticsAndDetails(info SegmentInfo, source string, details map[string]string) {
	panic("not implemented yet")
}

// Returns a string description of all segments, for debugging.
func (w *IndexWriter) segString() string {
	return w.segmentsToString(w.segmentInfos.Segments)
}

// returns a string description of the specified segments, for debugging.
func (w *IndexWriter) segmentsToString(infos []*SegmentInfoPerCommit) string {
	panic("not implemented yet")
}

// Returns a string description of the specified segment, for debugging.
func (w *IndexWriter) SegmentToString(info *SegmentInfoPerCommit) string {
	panic("not implemented yet")
}

// called only from assert
func (w *IndexWriter) fileExist(toSync *SegmentInfos) (ok bool, err error) {
	panic("not implemented yet")
}

// For infoStream output
func (w *IndexWriter) toLiveInfos(sis *SegmentInfos) *SegmentInfos {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

/*
Walk through all files referenced by the current segmentInfos and ask
the  Directory to sync each file, if it wans't already. If that
succeeds, then we prepare a new segments_N file but do not fully
commit it.
*/
func (w *IndexWriter) startCommit(toSync *SegmentInfos) error {
	panic("not implemented yet")
}

/*
Used only  by assert for testing. Current points:
- startDoFlush
- startCommitMerge
- startStartCommit
- midStartCommit
- midStartCommit2
- midStartCommitSuccess
- finishStartCommit
- startCommitMergeDeletes
- startMergeInit
- DocumentsWriter.ThreadState.init start
*/
func (w *IndexWriter) testPoint(message string) bool {
	panic("not implemented yet")
}

/*
NOTE: this method creates a compound file for all files returned by
info.files(). While, generally, this may include separate norms and
deleteion files, this SegmentInfos must not reference such files when
this method is called, because they are not allowed within a compound
file.
*/
func createCompoundFile(infoStream util.InfoStream,
	directory store.Directory,
	checkAbort *CheckAbort,
	info SegmentInfo,
	context store.IOContext) (names []string, err error) {
	panic("not implemented yet")
}

// Tries to delete the given files if unreferenced.
func (w *IndexWriter) deleteNewFiles(files []string) error {
	w.Lock() // synchronized
	defer w.Unlock()
	panic("not implemented yet")
}

func (w *IndexWriter) purge(forced bool) (n int, err error) {
	panic("not implemented yet")
}

func (w *IndexWriter) doAfterSegmentFlushed(triggerMerge bool, forcePurge bool) error {
	panic("not implemented yet")
}

func (w *IndexWriter) processEvents(triggerMerge bool, forcePurge bool) (ok bool, err error) {
	panic("not implemented yet")
}

func (w *IndexWriter) processEventsQueue(queue *list.List, triggerMerge, forcePurge bool) (ok, bool, err error) {
	panic("not implemented yet")
}

/*
Interface for internal atomic events. See DocumentsWriter fo details.
Events are executed concurrently and no order is guaranteed. Each
event should only rely on the serializeability within its process
method. All actions that must happen before or after a certain action
must be encoded inside the process() method.
*/
type Event func(writer *IndexWriter, triggerMerge, clearBuffers bool)

/*
If openDirectoryReader() has been called (ie, this writer is in near
real-time mode), then after a merge comletes, this class can be
invoked to warm the reader on the newly merged segment, before the
merge commits. This is not required for near real-time search, but
will reduce search latency on opening a new near real-time reader
after a merge completes.

NOTE: warm is called before any deletes have been carried over to the
merged segment.
*/
type IndexReaderWarmer interface {
	// Invoked on the AtomicReader for the newly merged segment, before
	// that segment is made visible to near-real-time readers.
	warm(reader AtomicReader) error
}

// index/ReadersAndLiveDocs.java

/*
Used by IndexWriter to hold open SegmentReaders (for searching or
merging), plus pending deletes, for a given segment.
*/
type ReadersAndLiveDocs struct{}

// index/BufferedDeletesStream.java

/*
Tracks the stream of BufferedDeletes. When DocumentsWriterPerThread
flushes, its buffered deletes are appended to this stream. We later
apply these deletes (resolve them to the actual docIDs, per segment)
when a merge is started (only to the to-be-merged segments). We also
apply to all segments when NRT reader is pulled, commit/close is
called, or when too many deletes are buffered and must be flushed (by
RAM usage or by count).

Each packet is assigned a generation, and each flushed or merged
segment is also assigned a generation, so we can track when
BufferedDeletes packets to apply to any given segment.
*/
type BufferedDeletesStream struct {
	// TODO: maybe linked list?
	deletes []*FrozenBufferedDeletes

	// Starts at 1 so that SegmentInfos that have never had deletes
	// applied (whose bufferedDelGen defaults to 0) will be correct:
	nextGen int64

	// used only by assert
	lastDeleteTerm *Term

	infoStream util.InfoStream
	bytesUsed  int64 // atomic
	numTerms   int   // atomic
}

func newBufferedDeletesStream(infoStream util.InfoStream) *BufferedDeletesStream {
	return &BufferedDeletesStream{
		deletes:    make([]*FrozenBufferedDeletes, 0),
		nextGen:    1,
		infoStream: infoStream,
	}
}
