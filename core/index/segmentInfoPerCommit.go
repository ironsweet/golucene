package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
)

// index/SegmentInfoPerCommit.java

// Embeds a [read-only] SegmentInfo and adds per-commit fields.
type SegmentInfoPerCommit struct {
	// The SegmentInfo that we wrap.
	info *SegmentInfo
	// How many deleted docs in the segment:
	delCount int
	// Generation number of the live docs file (-1 if there are no deletes yet)
	delGen int64
	// Normally 1+delGen, unless an exception was hit on last attempt to write:
	nextWriteDelGen int64

	sizeInBytes int64 // volatile

	// NOTE: only used by in-RAM by IW to track buffered deletes;
	// this is never written to/read from the Directory
	bufferedDeletesGen int64
}

func NewSegmentInfoPerCommit(info *SegmentInfo, delCount int, delGen int64) *SegmentInfoPerCommit {
	nextWriteDelGen := int64(1)
	if delGen != -1 {
		nextWriteDelGen = delGen + 1
	}
	return &SegmentInfoPerCommit{
		info:            info,
		delCount:        delCount,
		delGen:          delGen,
		nextWriteDelGen: nextWriteDelGen,
		sizeInBytes:     -1,
	}
}

/* Called when we succeed in writing deletes */
func (info *SegmentInfoPerCommit) advanceDelGen() {
	info.delGen, info.nextWriteDelGen = info.nextWriteDelGen, info.delGen+1
	info.sizeInBytes = -1
}

/*
Called if there was an error while writing deletes, so that we don't
try to write to the same file more than once.
*/
func (info *SegmentInfoPerCommit) advanceNextWriteDelGen() {
	info.nextWriteDelGen++
}

/*
Returns total size in bytes of all files for this segment.

NOTE: This value is not correct for 3.0 segments that have shared
docstores. To get correct value, upgrade.
*/
func (si *SegmentInfoPerCommit) SizeInBytes() (sum int64, err error) {
	if si.sizeInBytes == -1 {
		sum = 0
		for _, fileName := range si.Files() {
			d, err := si.info.dir.FileLength(fileName)
			if err != nil {
				return 0, err
			}
			sum += d
		}
		si.sizeInBytes = sum
	}
	return si.sizeInBytes, nil
}

// Returns all files in use by this segment.
func (si *SegmentInfoPerCommit) Files() []string {
	// Start from the wrapped info's files:
	files := make(map[string]bool)
	for name, _ := range si.info.Files() {
		files[name] = true
	}

	// Must separately add any live docs files
	// si.info.codec.LiveDocsFormat().Files(si, files)
	panic("not implemented yet")

	ans := make([]string, 0, len(files))
	for s, _ := range files {
		ans = append(ans, s)
	}
	return ans
}

func (si *SegmentInfoPerCommit) setBufferedDeletesGen(v int64) {
	si.bufferedDeletesGen = v
	si.sizeInBytes = -1
}

// Returns true if there are any deletions for the segment at this
// commit.
func (si *SegmentInfoPerCommit) HasDeletions() bool {
	return si.delGen != -1
}

func (si *SegmentInfoPerCommit) StringOf(dir store.Directory, pendingDelCount int) string {
	return si.info.StringOf(dir, si.delCount+pendingDelCount)
}

func (si *SegmentInfoPerCommit) String() string {
	s := si.info.StringOf(si.info.dir, si.delCount)
	if si.delGen != -1 {
		s = fmt.Sprintf("%v:delGen=%v", s, si.delGen)
	}
	return s
}

func (si *SegmentInfoPerCommit) Clone() *SegmentInfoPerCommit {
	// Not clear that we need ot carry over nextWriteDelGen (i.e. do we
	// ever clone after a failed write and before the next successful
	// write?), but just do it to be safe:
	return &SegmentInfoPerCommit{
		info:            si.info,
		delCount:        si.delCount,
		delGen:          si.delGen,
		nextWriteDelGen: si.nextWriteDelGen,
	}
}
