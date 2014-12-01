package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// index/IndexFileDeleter.java

const VERBOSE_REF_COUNT = false

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
	// Files that we tried to delete but failed (likely because they
	// are open and we are running on Windows), so we will retry them
	// again later:
	deletable map[string]bool
	// Reference count for all files in the index.
	// Counts how many existing commits reference a file.
	refCounts map[string]*RefCount

	// Holds all commits (Segments_N) current in the index. This will
	// have just 1 commit if you are using the default delete policy (
	// KeepOnlyLastCommitDeletionPolicy). Other policies may leave
	// commit points live for longer in which case this list would be
	// longer than 1:
	commits []IndexCommit

	// Holds files we had incref'd from the previous non-commit checkpoint:
	lastFiles []string

	// Commits that the IndexDeletionPolicy have decided to delete:
	commitsToDelete []*CommitPoint

	infoStream util.InfoStream
	directory  store.Directory
	policy     IndexDeletionPolicy

	startingCommitDeleted bool
	lastSegmentInfos      *SegmentInfos

	writer *IndexWriter
}

/*
Initialize the deleter: find all previous commits in the Directory,
incref the files they reference, call the policy to let it delete
commits. This will remove any files not referenced by any of the
commits.
*/
func newIndexFileDeleter(directory store.Directory, policy IndexDeletionPolicy,
	segmentInfos *SegmentInfos, infoStream util.InfoStream, writer *IndexWriter,
	initialIndexExists bool) (*IndexFileDeleter, error) {

	assert(writer != nil)

	currentSegmentsFile := segmentInfos.SegmentsFileName()
	if infoStream.IsEnabled("IFD") {
		infoStream.Message("IFD", "init: current segments file is '%v'; deletionPolicy=%v",
			currentSegmentsFile, reflect.TypeOf(policy).Name())
	}

	fd := &IndexFileDeleter{
		infoStream: infoStream,
		writer:     writer,
		policy:     policy,
		directory:  directory,
		refCounts:  make(map[string]*RefCount),
	}

	// First pass: walk the files and initialize our ref counts:
	currentGen := segmentInfos.generation

	var currentCommitPoint *CommitPoint
	var files []string
	files, err := directory.ListAll()
	if _, ok := err.(*store.NoSuchDirectoryError); ok {
		// it means the directory is empty, so ignore it
		files = make([]string, 0)
	} else if err != nil {
		return nil, err
	}

	if currentSegmentsFile != "" {
		m := model.CODEC_FILE_PATTERN
		for _, filename := range files {
			if !strings.HasSuffix(filename, WRITE_LOCK_NAME) &&
				filename != INDEX_FILENAME_SEGMENTS_GEN &&
				(m.MatchString(filename) || strings.HasPrefix(filename, util.SEGMENTS)) {

				// Add this file to refCounts with initial count 0:
				fd.refCount(filename)

				if strings.HasPrefix(filename, util.SEGMENTS) {
					// This is a commit (segments or segments_N), and it's
					// valid (<= the max gen). Load it, then incref all files
					// it refers to:
					if infoStream.IsEnabled("IFD") {
						infoStream.Message("IFD", "init: load commit '%v'", filename)
					}
					sis := &SegmentInfos{}
					err := sis.Read(directory, filename)
					if os.IsNotExist(err) {
						// LUCENE-948: on NFS (and maybe others), if
						// you have writers switching back and forth
						// between machines, it's very likely that the
						// dir listing will be stale and will claim a
						// file segments_X exists when in fact it
						// doesn't.  So, we catch this and handle it
						// as if the file does not exist
						if infoStream.IsEnabled("IFD") {
							infoStream.Message("IFD",
								"init: hit FileNotFoundException when loading commit '%v'; skipping this commit point",
								filename)
						}
						sis = nil
					} else if err != nil {
						if GenerationFromSegmentsFileName(filename) <= currentGen {
							length, _ := directory.FileLength(filename)
							if length > 0 {
								return nil, err
							}
						}
						// Most likely we are opening an index that has an
						// aborted "future" commit, so suppress exc in this case
						sis = nil
					} else { // sis != nil
						commitPoint := newCommitPoint(fd.commitsToDelete, directory, sis)
						if sis.generation == segmentInfos.generation {
							currentCommitPoint = commitPoint
						}
						fd.commits = append(fd.commits, commitPoint)
						fd.incRef(sis, true)

						if fd.lastSegmentInfos == nil || sis.generation > fd.lastSegmentInfos.generation {
							fd.lastSegmentInfos = sis
						}
					}
				}
			}
		}
	}

	if currentCommitPoint == nil && currentSegmentsFile != "" && initialIndexExists {
		// We did not in fact see the segments_N file corresponding to
		// the segmentInfos that was passed in. Yet, it must exist,
		// because our caller holds the write lock. This can happen when
		// the directory listing was stale (e.g. when index accessed via
		// NFS client with stale directory listing cache). So we try now
		// to explicitly open this commit point:
		sis := &SegmentInfos{}
		err := sis.Read(directory, currentSegmentsFile)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to locate current segments_N file '%v'",
				currentSegmentsFile))
		}
		if infoStream.IsEnabled("IFD") {
			infoStream.Message("IFD", "forced open of current segments file %v",
				segmentInfos.SegmentsFileName())
		}
		currentCommitPoint = newCommitPoint(fd.commitsToDelete, directory, sis)
		fd.commits = append(fd.commits, currentCommitPoint)
		fd.incRef(sis, true)
	}

	// We keep commits list in sorted order (oldest to newest):
	util.TimSort(IndexCommits(fd.commits))

	// refCounts only includes "normal" filenames (does not include segments.gen, write.lock)
	files = nil
	for k, _ := range fd.refCounts {
		files = append(files, k)
	}
	inflateGens(segmentInfos, files, fd.infoStream)

	// Now delete anyting with ref count at 0. These are presumably
	// abandoned files e.g. due to crash of IndexWriter.
	for filename, rc := range fd.refCounts {
		if rc.count == 0 {
			if infoStream.IsEnabled("IFD") {
				infoStream.Message("IFD", "init: removing unreferenced file '%v'",
					filename)
			}
			fd.deleteFile(filename)
		}
	}

	// Finally, give policy a chance to remove things on startup:
	err = policy.onInit(fd.commits)
	if err != nil {
		return nil, err
	}

	// Always protect the incoming segmentInfos since sometime it may
	// not be the most recent commit
	err = fd.checkpoint(segmentInfos, false)
	if err != nil {
		return nil, err
	}

	fd.startingCommitDeleted = (currentCommitPoint != nil && currentCommitPoint.IsDeleted())

	fd.deleteCommits()

	return fd, nil
}

/*
Set all gens beyond what we currently see in the directory, to avoid
double-write in cases where the previous IndexWriter did not
gracefully close/rollback (e.g. os/machine crashed or lost power).
*/
func inflateGens(infos *SegmentInfos, files []string, infoStream util.InfoStream) {
	var maxSegmentGen int64 = math.MinInt64
	var maxSegmentName int64 = math.MinInt32

	// Confusingly, this is the union of liveDocs, field infos, doc
	// values (and maybe others, in the future) gens. THis is somewhat
	// messy, since it means DV updates will suddenly write to the next
	// gen after live docs' gen, for example, but we don't have the
	// APIs to ask the codec which file is which:
	maxPerSegmentGen := make(map[string]int64)

	for _, filename := range files {
		if filename == INDEX_FILENAME_SEGMENTS_GEN || filename == WRITE_LOCK_NAME {
			// do nothing
		} else if strings.HasPrefix(filename, INDEX_FILENAME_SEGMENTS) {
			if n := GenerationFromSegmentsFileName(filename); n > maxSegmentGen {
				maxSegmentGen = n
			}
		} else {
			segmentName := util.ParseSegmentName(filename)
			assert2(strings.HasPrefix(segmentName, "_"), "file=%v", filename)

			n, err := strconv.ParseInt(segmentName[1:], 36, 64)
			assert(err == nil)
			if n > maxSegmentName {
				maxSegmentName = n
			}

			curGen := maxPerSegmentGen[segmentName] // or zero if not exists
			if n := util.ParseGeneration(filename); n > curGen {
				curGen = n
			}

			maxPerSegmentGen[segmentName] = curGen
		}
	}
}

func (fd *IndexFileDeleter) ensureOpen() {
	fd.writer.ClosingControl.ensureOpen(false)
	// since we allow 'closing' state, we must still check this, we
	// could be closing because we hit unexpected error
	assert2(fd.writer.tragedy == nil,
		"refusing to delete any files: this IndexWriter hit an unrecoverable error\n%v",
		fd.writer.tragedy)
}

/*
Remove the CommitPoint(s) in the commitsToDelete list by decRef'ing
all files from each SegmentInfos.
*/
func (fd *IndexFileDeleter) deleteCommits() {
	if size := len(fd.commitsToDelete); size > 0 {
		// First decref all files that had been referred to by the
		// now-deleted commits:
		for _, commit := range fd.commitsToDelete {
			if fd.infoStream.IsEnabled("IFD") {
				fd.infoStream.Message("IFD", "deleteCommits: now decRef commit '%v'",
					commit.segmentsFileName)
			}
			fd.decRefFiles(commit.files)
		}
		fd.commitsToDelete = nil

		// Now compact commits to remove deleted ones (preserving the sort):
		var writeTo = 0
		for readFrom, commit := range fd.commits {
			if !commit.IsDeleted() && readFrom != writeTo {
				fd.commits[writeTo] = commit
				writeTo++
			}
		}
		for i, _ := range fd.commits[writeTo:] {
			fd.commits[i] = nil
		}
		fd.commits = fd.commits[:writeTo]
	}
}

/*
Writer calls this when it has hit an error and had to roll back, to
tell us that there may now be unreferenced files in the filesystem.
So we re-list the filesystem and delete such files. If segmentName is
non-empty, we only delete files correspoding to that segment.
*/
func (fd *IndexFileDeleter) refresh(segmentName string) error {
	// assert locked()

	var prefix1, prefix2 string
	if segmentName != "" {
		prefix1 = segmentName + "."
		prefix2 = segmentName + "_"
	}

	m := model.CODEC_FILE_PATTERN
	files, err := fd.directory.ListAll()
	if err != nil {
		return err
	}
	for _, filename := range files {
		_, hasRef := fd.refCounts[filename]
		if (segmentName == "" || strings.HasPrefix(filename, prefix1) ||
			strings.HasPrefix(filename, prefix2)) &&
			!strings.HasSuffix(filename, WRITE_LOCK_NAME) &&
			!hasRef && filename != INDEX_FILENAME_SEGMENTS_GEN &&
			(m.MatchString(filename) || strings.HasPrefix(filename, INDEX_FILENAME_SEGMENTS)) {

			// Unreferenced file, so remove it
			if fd.infoStream.IsEnabled("IFD") {
				fd.infoStream.Message("IFD",
					"refresh [prefix=%v]: removing newly created unreferenced file '%v'",
					segmentName, filename)
			}
			fd.deleteFile(filename)
		}
	}
	return nil
}

func (fd *IndexFileDeleter) refreshList() error {
	// set to nil so that we regenerate the list of pending files;
	// else we can accumulate some file more than once
	fd.deletable = nil
	return fd.refresh("")
}

func (fd *IndexFileDeleter) Close() error {
	// DecRef old files from the last checkpoint, if any:
	// assert locked()
	if len(fd.lastFiles) > 0 {
		fd.decRefFiles(fd.lastFiles)
		fd.lastFiles = nil
	}
	fd.deletePendingFiles()
	return nil
}

func (fd *IndexFileDeleter) deletePendingFiles() {
	// assert locked()
	if fd.deletable != nil {
		oldDeletable := fd.deletable
		fd.deletable = nil
		for filename, _ := range oldDeletable {
			if fd.infoStream.IsEnabled("IFD") {
				fd.infoStream.Message("IFD", "delete pending file %v", filename)
			}
			rc, ok := fd.refCounts[filename]
			assert2(!ok || rc.count <= 0,
				// LUCENE-5904: should never happen!  This means we are about to pending-delete a referenced index file
				"filename=%v is in pending delete list but also has refCount=%v",
				filename, rc.count)
			fd.deleteFile(filename)
		}
	}
}

/*
For definition of "check point" see IndexWriter comments:
"Clarification: Check Points (and commits)".

Writer calls this when it has made a "consistent change" to the index,
meaning new files are written to the index the in-memory SegmentInfos
have been modified to point to those files.

This may or may not be a commit (sgments_N may or may not have been
written).

We simply incref the files referenced by the new SegmentInfos and
decref the files we had previously seen (if any).

If this is a commit, we also call the policy to give it a chance to
remove other commits. If any commits are removed, we decref their
files as well.
*/
func (fd *IndexFileDeleter) checkpoint(segmentInfos *SegmentInfos, isCommit bool) error {
	// asset locked()
	start := time.Now()
	defer func() {
		if fd.infoStream.IsEnabled("IFD") {
			elapsed := time.Now().Sub(start)
			fd.infoStream.Message("IFD", "%v to checkpoint", elapsed)
		}
	}()
	if fd.infoStream.IsEnabled("IFD") {
		fd.infoStream.Message("IFD", "now checkpoint '%v' [%v segments; isCommit = %v]",
			fd.writer.readerPool.segmentsToString(fd.writer._toLiveInfos(segmentInfos).Segments),
			len(segmentInfos.Segments), isCommit)
	}

	// Try again now to delete any previously un-deletable files (
	// because they were in use, on Windows):
	fd.deletePendingFiles()

	// Incref the files:
	fd.incRef(segmentInfos, isCommit)

	if isCommit {
		// Append to our commits list:
		fd.commits = append(fd.commits, newCommitPoint(fd.commitsToDelete, fd.directory, segmentInfos))

		// Tell policy so it can remove commits:
		err := fd.policy.onCommit(fd.commits)
		if err != nil {
			return err
		}

		// Decref files for commits that were deleted by the policy:
		fd.deleteCommits()
	} else {
		// DecRef old files from the last checkpoint, if any:
		fd.decRefFiles(fd.lastFiles)
		fd.lastFiles = nil

		// Save files so we can decr on next checkpoint/commit:
		fd.lastFiles = append(fd.lastFiles, segmentInfos.files(fd.directory, false)...)
	}
	return nil
}

func (del *IndexFileDeleter) incRef(segmentInfos *SegmentInfos, isCommit bool) {
	// assert locked()
	// If this is a commit point, also incRef the segments_N file:
	files := segmentInfos.files(del.directory, isCommit)
	for _, filename := range files {
		del.incRefFile(filename)
	}
}

func (del *IndexFileDeleter) incRefFiles(files []string) {
	// assert locked
	for _, file := range files {
		del.incRefFile(file)
	}
}

func (del *IndexFileDeleter) incRefFile(filename string) {
	// assert locked
	rc := del.refCount(filename)
	if del.infoStream.IsEnabled("IFD") && VERBOSE_REF_COUNT {
		del.infoStream.Message("IFD", "  IncRef '%v': pre-incr count is %v",
			filename, rc.count)
	}
	rc.incRef()
}

func (fd *IndexFileDeleter) decRefFiles(files []string) {
	// assert locked()
	for _, file := range files {
		fd.decRefFile(file)
	}
}

func (fd *IndexFileDeleter) decRefFilesWhileSuppressingError(files []string) {
	for _, file := range files {
		fd.decRefFileWhileSuppressingError(file)
	}
}

func (fd *IndexFileDeleter) decRefFile(filename string) {
	//assert locked()
	rc := fd.refCount(filename)
	if fd.infoStream.IsEnabled("IFD") && VERBOSE_REF_COUNT {
		fd.infoStream.Message("IFD", "  DecRef '%v': pre-decr count is %v",
			filename, rc.count)
	}
	if rc.decRef() == 0 {
		// This file is no longer referenced by any past commit points
		// nor by the in-memory SegmentInfos:
		fd.deleteFile(filename)
		delete(fd.refCounts, filename)
	}
}

func (fd *IndexFileDeleter) decRefFileWhileSuppressingError(file string) {
	defer func() {
		recover()
	}()
	fd.decRefFile(file)
}

func (del *IndexFileDeleter) decRefInfos(infos *SegmentInfos) {
	del.decRefFiles(infos.files(del.directory, false))
}

// 529
func (del *IndexFileDeleter) exists(filename string) bool {
	if v, ok := del.refCounts[filename]; ok {
		return v.count > 0
	}
	return false
}

func (del *IndexFileDeleter) refCount(filename string) *RefCount {
	// assert Thread.holdsLock(del.writer) TODO GoLucene doesn't have this capability
	rc, ok := del.refCounts[filename]
	if !ok {
		rc = newRefCount(filename)
		// we should never incRef a file we are already wanting to delete
		assert2(del.deletable == nil || !del.deletable[filename],
			"file '%v' cannot be incRef'd: it's already pending delete",
			filename)
		del.refCounts[filename] = rc
	}
	return rc
}

/*
Deletes the specified files, but only if they are new (have not yet
been incref'd).
*/
func (fd *IndexFileDeleter) deleteNewFiles(files []string) {
	// assert locked
	for _, filename := range files {
		// NOTE: it's very unusual yet possible for the
		// refCount to be present and 0: it can happen if you
		// open IW on a crashed index, and it removes a bunch
		// of unref'd files, and then you add new docs / do
		// merging, and it reuses that segment name.
		// TestCrash.testCrashAfterReopen can hit this:
		if rf, ok := fd.refCounts[filename]; !ok || rf.count == 0 {
			if fd.infoStream.IsEnabled("IFD") {
				fd.infoStream.Message("IFD", "delete new file '%v'", filename)
			}
			fd.deleteFile(filename)
		}
	}
}

func (del *IndexFileDeleter) deleteFile(filename string) {
	//assert locked()
	del.ensureOpen()
	if del.infoStream.IsEnabled("IFD") {
		del.infoStream.Message("IFD", "delete '%v'", filename)
	}
	err := del.directory.DeleteFile(filename)
	if err != nil { // if delete fails
		if del.directory.FileExists(filename) {
			// Some operating systems (e.g. Windows) don't
			// permit a file to be deleted while it is opened
			// for read (e.g. by another process or thread). So
			// we assume that when a delete fails it is because
			// the file is open in another process, and queue
			// the file for subsequent deletion.
			if del.infoStream.IsEnabled("IFD") {
				del.infoStream.Message("IFD",
					"unable to remove file '%v': %v; will re-try later.",
					filename, err)
			}
			if del.deletable == nil {
				del.deletable = make(map[string]bool)
			}
			del.deletable[filename] = true
		}
	}
}

/*
Tracks the reference count for a single index file:
*/
type RefCount struct {
	// filename used only for better assert error messages
	filename string
	initDone bool
	count    int
}

func newRefCount(filename string) *RefCount {
	return &RefCount{filename: filename}
}

func (rf *RefCount) incRef() int {
	if !rf.initDone {
		rf.initDone = true
	} else {
		assert2(rf.count > 0, fmt.Sprintf("RefCount is 0 pre-increment for file %v", rf.filename))
	}
	rf.count++
	return rf.count
}

func (rf *RefCount) decRef() int {
	assert2(rf.count > 0, fmt.Sprintf("RefCount is 0 pre-decrement for file %v", rf.filename))
	rf.count--
	return rf.count
}

/*
Holds details for each commit point. This class is also passed to the
deletion policy. Note: this class has a natural ordering that is
inconsistent with equals.
*/
type CommitPoint struct {
	files            []string
	segmentsFileName string
	deleted          bool
	directory        store.Directory
	commitsToDelete  []*CommitPoint
	generation       int64
	userData         map[string]string
	segmentCount     int
}

func newCommitPoint(commitsToDelete []*CommitPoint, directory store.Directory,
	segmentInfos *SegmentInfos) *CommitPoint {
	return &CommitPoint{
		directory:        directory,
		commitsToDelete:  commitsToDelete,
		userData:         segmentInfos.userData,
		segmentsFileName: segmentInfos.SegmentsFileName(),
		generation:       segmentInfos.generation,
		files:            segmentInfos.files(directory, true),
		segmentCount:     len(segmentInfos.Segments),
	}
}

func (cp *CommitPoint) String() string {
	return fmt.Sprintf("IndexFileDeleter.CommitPoint(%v)", cp.segmentsFileName)
}

func (cp *CommitPoint) SegmentCount() int {
	return cp.segmentCount
}

func (cp *CommitPoint) SegmentsFileName() string {
	return cp.segmentsFileName
}

func (cp *CommitPoint) FileNames() []string {
	return cp.files
}

func (cp *CommitPoint) Directory() store.Directory {
	return cp.directory
}

func (cp *CommitPoint) Generation() int64 {
	return cp.generation
}

func (cp *CommitPoint) UserData() map[string]string {
	return cp.userData
}

func (cp *CommitPoint) Delete() {
	if !cp.deleted {
		cp.deleted = true
		cp.commitsToDelete = append(cp.commitsToDelete, cp)
	}
}

func (cp *CommitPoint) IsDeleted() bool {
	return cp.deleted
}
