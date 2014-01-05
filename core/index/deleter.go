package index

// index/IndexFileDeleter.java

/*
This class keeps track of each SegmentInfos instance that is still
"live", either because it corresponds to a segments_N file in the
Directory (a "commit", i.e. a commited egmentInfos) or because it's
an in-memory SegmentInfos that a writer is actively updating but has
not yet committed. This class uses simple reference counting to map
the live SegmentInfos instances to individual files in the Directory.

The same directory file maybe referenced by more than one IndexCommit,
i.e. more than one SegmentInfos. Therefore we count how many commits
reference each file. When all the commits referencing a certain file
have been deleted, the refcount for that file becomes zero, and the
file is deleted.

A separate deletion policy interface (IndexDeletionPolicy) is
consulted on creation (onInit) and once per commit (onCommit), to
decide when a commit should be removed.

It is the business of the IndexDeletionPolicy to choose when to
delete commit points. The actual mechanics of file deletion, retrying,
etc, derived from the deletion of commit points is the business of
the IndexFileDeleter.

The current default deletion policy is KeepOnlyLastCommitDeletionPolicy,
which  removes all prior commits when a new commit has completed.
This matches the bahavior before 2.2.

Note that you must hold the write.lock before instantiating this
class. It opens segments_N file(s) directly with no retry logic.
*/
type IndexFileDeleter struct {
}

//L432
/*
For definition of "check point" see IndexWriter comments:
"Clarification: Check Points (and commits)".

Writer calls this when it has made a "consistent change" to the index,
meaning new files are written to the index the in-memory SegmentInfos
have been modified to point to those files.

This may or may not be a commit (sgments_N may or may not have been
	written).

	WAe simply incref the files referenced by the new SegmentInfos and
	decref the files we had previously seen (if any).

	If this is a commit, we also call the policy to give it a chance to
	remove other commits. If any commits are removed, we decref their
	files as well.
*/
func (del *IndexFileDeleter) checkpoint(segmentInfos *SegmentInfos, isCommit bool) error {
	panic("not implemented yet")
}
