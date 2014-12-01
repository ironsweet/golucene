package index

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
)

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
Default value for calling checkIntegrity() before merging segments
(set to false). You can set this to true for additional safety.
*/
const DEFAULT_CHECK_INTEGRITY_AT_MERGE = false

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
	*LiveIndexWriterConfigImpl
	writer *util.SetOnce
}

// Sets the IndexWriter this config is attached to.
func (conf *IndexWriterConfig) setIndexWriter(writer *IndexWriter) *IndexWriterConfig {
	conf.writer.Set(writer)
	return conf
}

// L523
/*
Information about merges, deletes and a message when maxFieldLength
is reached will be printed to this. Must not be nil, but NO_OUTPUT
may be used to surpress output.
*/
func (conf *IndexWriterConfig) SetInfoStream(infoStream util.InfoStream) *IndexWriterConfig {
	assert2(infoStream != nil, "Cannot set InfoStream implementation to null. "+
		"To disable logging use InfoStream.NO_OUTPUT")
	conf.infoStream = infoStream
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
		LiveIndexWriterConfigImpl: newLiveIndexWriterConfig(analyzer, matchVersion),
		writer: util.NewSetOnce(),
	}
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

// L310
func (conf *IndexWriterConfig) MergePolicy() MergePolicy {
	return conf.mergePolicy
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
	conf.LiveIndexWriterConfigImpl.SetMaxBufferedDocs(maxBufferedDocs)
	return conf
}

func (conf *IndexWriterConfig) SetMergedSegmentWarmer(mergeSegmentWarmer IndexReaderWarmer) *IndexWriterConfig {
	conf.LiveIndexWriterConfigImpl.SetMergedSegmentWarmer(mergeSegmentWarmer)
	return conf
}

func (conf *IndexWriterConfig) SetReaderTermsIndexDivisor(divisor int) *IndexWriterConfig {
	conf.LiveIndexWriterConfigImpl.SetReaderTermsIndexDivisor(divisor)
	return conf
}

func (conf *IndexWriterConfig) SetUseCompoundFile(useCompoundFile bool) *IndexWriterConfig {
	conf.LiveIndexWriterConfigImpl.SetUseCompoundFile(useCompoundFile)
	return conf
}

func (conf *IndexWriterConfig) String() string {
	panic("not implemented yet")
}
