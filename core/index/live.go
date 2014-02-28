package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

// index/LiveIndexWriterConfig.java

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
		indexerThreadPool:       NewDocumentsWriterPerThreadPool(DEFAULT_MAX_THREAD_STATES),
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
