package index

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
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

// index/IndexDeletionPolicy.java

/*
Expert: policy for deletion of stale index commits.

Implement this interface, and pass it to one of the IndexWriter or
IndexReader constructors, to customize when older point-in-time-commits
are deleted from the index directory. The default deletion policy is
KeepOnlyLastCommitDeletionPolicy, which always remove old commits as
soon as a new commit is done (this matches the behavior before 2.2).

One expected use case for this ( and the reason why it was first
created) is to work around problems with an index directory accessed
via filesystems like NFS because NFS does not provide the "delete on
last close" semantics that Lucene's "point in time" search normally
relies on. By implementing a custom deletion policy, such as "a
commit is only removed once it has been stale for more than X
minutes", you can give your readers time to refresh to the new commit
before IndexWriter removes the old commits. Note that doing so will
increase the storage requirements of the index. See [LUCENE-710] for
details.

Implementers of sub-classes should make sure that Clone() returns an
independent instance able to work with any other IndexWriter or
Directory instance.
*/
type IndexDeletionPolicy interface {
	/*
		This is called once when a writer is first instantiated to give the
		policy a chance to remove old commit points.

		The writer locates all index commits present in the index directory
		and calls this method. The policy may choose to delete some of the
		commit points, doing so by calling method delete() of IndexCommit.

		Note: the last CommitPoint is the most recent one, i.e. the "front
		index state". Be careful not to delete it, unless you know for sure
		what you are doing, and unless you can afford to lose the index
		content while doing that.
	*/
	onInit(commits []IndexCommit) error
	/*
	  This is called each time the writer completed a commit. This
	  gives the policy a chance to remove old commit points with each
	  commit.

	  The policy may now choose to delete old commit points by calling
	  method Delete() of IndexCommit.

	  This method is only called when Commit() or Close() is called, or
	  possibly not at all if the Rollback() is called.

	  Note: the last CommitPoint is the most recent one, i.e. the
	  "front index state". Be careful not to delete it, unless you know
	  for sure what you are doing, and unless you can afford to lose
	  the index content while doing that.
	*/
	onCommit(commits []IndexCommit) error
}

// index/NoDeletionPolicy.java

// An IndexDeletionPolicy which keeps all index commits around, never
// deleting them. This class is a singleton and can be accessed by
// referencing INSTANCE.
type NoDeletionPolicy bool

func (p NoDeletionPolicy) onCommit(commits []IndexCommit) error { return nil }
func (p NoDeletionPolicy) onInit(commits []IndexCommit) error   { return nil }
func (p NoDeletionPolicy) Clone() IndexDeletionPolicy           { return p }

const NO_DELETION_POLICY = NoDeletionPolicy(true)

// index/LiveIndexWriterConfig.java

/*
Holds all the configuration used by IndexWriter with few setters for
settings that can be changed on an IndexWriter instance "live".
*/
type LiveIndexWriterConfig struct {
	delPolicy IndexDeletionPolicy // volatile
}

// index/IndexWriterConfig.java

// Specifies the open mode for IndeWriter
type OpenMode int

const (
	// Creates a new index or overwrites an existing one.
	OPEN_MODE_CREATE = 1
	// Opens an existing index.
	OPEN_MODE_APPEND = 2
	// Creates a new index if one does not exist,
	// otherwise it opens the index and documents will be appended.
	OPEN_MODE_CREATE_OR_APPEND = 3
)

// Default value for the write lock timeout (1,000 ms)
const WRITE_LOCK_TIMEOUT = 1000

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

// index/IndexWriter.java

type IndexWriter struct {
}

/*
Constructs a new IndexWriter per the settings given in conf. If you want to
make "live" changes to this writer instance, use Config().

NOTE: after this writer is created, the given configuration instance cannot be
passed to another writer. If you intend to do so, you should clone it
beforehand.
*/
func NewIndexWriter(d store.Directory, conf *IndexWriterConfig) (w *IndexWriter, err error) {
	panic("not implemented yet")
}

// L906
func (w *IndexWriter) Close() error {
	panic("not implemented yet")
}

// L1201
func (w *IndexWriter) AddDocument(doc []IndexableField) error {
	panic("not implemented yet")
}

// L2001
/*
Close the IndexWriter without committing any changes that have
occurred since the last commit (or since it was opened, if commit
hasn't been called). This removes any temporary files that had been
created, after which the state of the index will be the same as it
was when commit() was last called or when this writer was first
opened. This also clears a previous call to prepareCommit()
*/
func (w *IndexWriter) Rollback() error {
	panic("not implemented yet")
}
