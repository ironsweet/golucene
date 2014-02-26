package index

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

// index/KeepOnlyLastCommitDeletionPolicy.java

/*
This IndexDeletionPolicy implementation that keeps only the most
recent commit and immediately removes all prior commits after a new
commit is done. This is the default deletion policy.
*/
type KeepOnlyLastCommitDeletionPolicy bool

// Deletes all commits except the most recent one.
func (p KeepOnlyLastCommitDeletionPolicy) onInit(commits []IndexCommit) error {
	return p.onCommit(commits)
}

// Deletes all commits except the most recent one.
func (p KeepOnlyLastCommitDeletionPolicy) onCommit(commits []IndexCommit) error {
	// Note that len(commits) should normally be 2 (if not called by
	// onInit above).
	for i, limit := 0, len(commits); i < limit-1; i++ {
		commits[i].Delete()
	}
	return nil
}

func (p KeepOnlyLastCommitDeletionPolicy) Clone() IndexDeletionPolicy {
	return p
}

const DEFAULT_DELETION_POLICY = KeepOnlyLastCommitDeletionPolicy(true)
