package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/util"
	"reflect"
)

// index/LiveIndexWriterConfig.java

/*
Holds all the configuration used by IndexWriter with few setters for
settings that can be changed on an IndexWriter instance "live".

All the fields are either readonly or volatile.
*/
type LiveIndexWriterConfig interface {
	TermIndexInterval() int
	MaxBufferedDocs() int
	RAMBufferSizeMB() float64
	Similarity() Similarity
	Codec() Codec
	MergePolicy() MergePolicy
	indexingChain() IndexingChain
	RAMPerThreadHardLimitMB() int
	flushPolicy() FlushPolicy
	InfoStream() util.InfoStream
	indexerThreadPool() *DocumentsWriterPerThreadPool
	UseCompoundFile() bool
}

type LiveIndexWriterConfigImpl struct {
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
	_indexingChain IndexingChain

	// Codec used to write new segments.
	codec Codec

	// InfoStream for debugging messages.
	infoStream util.InfoStream

	// MergePolicy for selecting merges.
	mergePolicy MergePolicy

	// DocumentsWriterPerThreadPool to control how goroutines are
	// allocated to DocumentsWriterPerThread.
	_indexerThreadPool *DocumentsWriterPerThreadPool

	// True if readers should be pooled.
	readerPooling bool

	// FlushPolicy to control when segments are flushed.
	_flushPolicy FlushPolicy

	// Sets the hard upper bound on RAM usage for a single segment,
	// after which the segment is forced to flush.
	perRoutineHardLimitMB int

	// Version that IndexWriter should emulate.
	matchVersion util.Version

	// True is segment flushes should use compound file format
	useCompoundFile bool // volatile

	// True if merging should check integrity of segments before merge
	checkIntegrityAtMerge bool // volatile
}

// used by IndexWriterConfig
func newLiveIndexWriterConfig(analyzer analysis.Analyzer,
	matchVersion util.Version) *LiveIndexWriterConfigImpl {

	assert(DefaultSimilarity != nil)
	assert(DefaultCodec != nil)
	return &LiveIndexWriterConfigImpl{
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
		_indexingChain:          defaultIndexingChain,
		codec:                   DefaultCodec(),
		infoStream:              util.DefaultInfoStream(),
		mergePolicy:             NewTieredMergePolicy(),
		_flushPolicy:            newFlushByRamOrCountsPolicy(),
		readerPooling:           DEFAULT_READER_POOLING,
		_indexerThreadPool:      NewDocumentsWriterPerThreadPool(DEFAULT_MAX_THREAD_STATES),
		perRoutineHardLimitMB:   DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB,
		checkIntegrityAtMerge:   DEFAULT_CHECK_INTEGRITY_AT_MERGE,
	}
}

// Creates a new config that handles the live IndexWriter settings.
// func newLiveIndexWriterConfigFrom(config *IndexWriterConfig) *LiveIndexWriterConfigImpl {
// 	return &LiveIndexWriterConfig{
// 		maxBufferedDeleteTerms:  config.maxBufferedDeleteTerms,
// 		maxBufferedDocs:         config.maxBufferedDocs,
// 		mergedSegmentWarmer:     config.mergedSegmentWarmer,
// 		ramBufferSizeMB:         config.ramBufferSizeMB,
// 		readerTermsIndexDivisor: config.readerTermsIndexDivisor,
// 		termIndexInterval:       config.termIndexInterval,
// 		matchVersion:            config.matchVersion,
// 		analyzer:                config.analyzer,
// 		delPolicy:               config.delPolicy,
// 		commit:                  config.commit,
// 		openMode:                config.openMode,
// 		similarity:              config.similarity,
// 		mergeScheduler:          config.mergeScheduler,
// 		writeLockTimeout:        config.writeLockTimeout,
// 		indexingChain:           config.indexingChain,
// 		codec:                   config.codec,
// 		infoStream:              config.infoStream,
// 		mergePolicy:             config.mergePolicy,
// 		indexerThreadPool:       config.indexerThreadPool,
// 		readerPooling:           config.readerPooling,
// 		flushPolicy:             config.flushPolicy,
// 		perRoutineHardLimitMB:   config.perRoutineHardLimitMB,
// 		useCompoundFile:         config.useCompoundFile,
// 	}
// }

func (conf *LiveIndexWriterConfigImpl) TermIndexInterval() int {
	return conf.termIndexInterval
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
func (conf *LiveIndexWriterConfigImpl) SetMaxBufferedDocs(maxBufferedDocs int) *LiveIndexWriterConfigImpl {
	assert2(maxBufferedDocs == DISABLE_AUTO_FLUSH || maxBufferedDocs >= 2,
		"maxBufferedDocs must at least be 2 when enabled")
	assert2(maxBufferedDocs != DISABLE_AUTO_FLUSH || conf.ramBufferSizeMB != DISABLE_AUTO_FLUSH,
		"at least one of ramBufferSize and maxBufferedDocs must be enabled")
	conf.maxBufferedDocs = maxBufferedDocs
	return conf
}

/* Returns the number of buffered added documents that will trigger a flush if enabled. */
func (conf *LiveIndexWriterConfigImpl) MaxBufferedDocs() int {
	return conf.maxBufferedDocs
}

/*
Expert: MergePolicy is invoked whenver there are changes to the
segments in the index. Its role is to select which merges to do, if
any, and return a MergeSpecification describing the merges. It also
selects merges to do for forceMerge.
*/
func (conf *LiveIndexWriterConfigImpl) SetMergePolicy(mergePolicy MergePolicy) *LiveIndexWriterConfigImpl {
	assert2(mergePolicy != nil, "mergePolicy must not be nil")
	conf.mergePolicy = mergePolicy
	return conf
}

func (conf *LiveIndexWriterConfigImpl) RAMBufferSizeMB() float64 {
	return conf.ramBufferSizeMB
}

/*
Sets the merged segment warmer.

Take effect on the next merge.
*/
func (conf *LiveIndexWriterConfigImpl) SetMergedSegmentWarmer(mergeSegmentWarmer IndexReaderWarmer) *LiveIndexWriterConfigImpl {
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
func (conf *LiveIndexWriterConfigImpl) SetReaderTermsIndexDivisor(divisor int) *LiveIndexWriterConfigImpl {
	assert2(divisor > 0 || divisor == -1, fmt.Sprintf(
		"divisor must be >= 1, or -1 (got %v)", divisor))
	conf.readerTermsIndexDivisor = divisor
	return conf
}

func (conf *LiveIndexWriterConfigImpl) Similarity() Similarity {
	return conf.similarity
}

/* Returns the current Codec. */
func (conf *LiveIndexWriterConfigImpl) Codec() Codec {
	return conf.codec
}

// L477
/* Returns the current MergePolicy in use by this writer. */
func (conf *LiveIndexWriterConfigImpl) MergePolicy() MergePolicy {
	return conf.mergePolicy
}

/* Returns the configured DocumentsWriterPerThreadPool instance. */
func (conf *LiveIndexWriterConfigImpl) indexerThreadPool() *DocumentsWriterPerThreadPool {
	return conf._indexerThreadPool
}

func (conf *LiveIndexWriterConfigImpl) indexingChain() IndexingChain {
	return conf._indexingChain
}

func (conf *LiveIndexWriterConfigImpl) RAMPerThreadHardLimitMB() int {
	return conf.perRoutineHardLimitMB
}

func (conf *LiveIndexWriterConfigImpl) flushPolicy() FlushPolicy {
	return conf._flushPolicy
}

/* Returns InfoStream used for debugging. */
func (conf *LiveIndexWriterConfigImpl) InfoStream() util.InfoStream {
	return conf.infoStream
}

/*
Sets if the IndexWriter should pack newly written segments in a
compound file. Default is true.

Use false for batch indexing with very large ram buffer settings.

Note: To control compound file usage during segment merges see
SetNoCFSRatio() and SetMaxCFSSegmentSizeMB(). This setting only
applies to newly created segment.
*/
func (conf *LiveIndexWriterConfigImpl) SetUseCompoundFile(useCompoundFile bool) *LiveIndexWriterConfigImpl {
	conf.useCompoundFile = useCompoundFile
	return conf
}

func (conf *LiveIndexWriterConfigImpl) UseCompoundFile() bool {
	return conf.useCompoundFile
}

func (conf *LiveIndexWriterConfigImpl) String() string {
	return fmt.Sprintf(`matchVersion=%v
analyzer=%v
ramBufferSizeMB=%v
maxBufferedDocs=%v
maxBufferedDeleteTerms=%v
mergedSegmentWarmer=%v
readerTermsIndexDivisor=%v
termIndexInterval=%v
delPolicy=%v
commit=%v
openMode=%v
similarity=%v
mergeScheduler=%v
default WRITE_LOCK_TIMEOUT=%v
writeLockTimeout=%v
codec=%v
infoStream=%v
mergePolicy=%v
indexerThreadPool=%v
readerPooling=%v
perThreadHardLimitMB=%v
useCompoundFile=%v
checkIntegrityAtMerge=%v
`, conf.matchVersion, reflect.TypeOf(conf.analyzer),
		conf.ramBufferSizeMB, conf.maxBufferedDocs,
		conf.maxBufferedDeleteTerms, reflect.TypeOf(conf.mergedSegmentWarmer),
		conf.readerTermsIndexDivisor, conf.termIndexInterval,
		reflect.TypeOf(conf.delPolicy), conf.commit,
		conf.openMode, reflect.TypeOf(conf.similarity),
		conf.mergeScheduler, WRITE_LOCK_TIMEOUT,
		conf.writeLockTimeout, conf.codec,
		reflect.TypeOf(conf.infoStream), conf.mergePolicy,
		conf.indexerThreadPool, conf.readerPooling,
		conf.perRoutineHardLimitMB, conf.useCompoundFile,
		conf.checkIntegrityAtMerge)
}
