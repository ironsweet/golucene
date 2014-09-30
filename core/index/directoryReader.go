package index

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	// "io"
	"errors"
	"strings"
)

const DEFAULT_TERMS_INDEX_DIVISOR = 1

type DirectoryReader interface {
	IndexReader
	// doOpenIfChanged() error
	// doOpenIfChanged(c IndexCommit) error
	// doOpenIfChanged(w IndexWriter, c IndexCommit) error
	Version() int64
	IsCurrent() bool
}

type DirectoryReaderImpl struct {
	*BaseCompositeReader
	directory store.Directory
}

func newDirectoryReader(spi BaseCompositeReaderSPI, directory store.Directory, segmentReaders []AtomicReader) *DirectoryReaderImpl {
	// log.Printf("Initializing DirectoryReader with %v segment readers...", len(segmentReaders))
	readers := make([]IndexReader, len(segmentReaders))
	for i, v := range segmentReaders {
		readers[i] = v
	}
	return &DirectoryReaderImpl{
		BaseCompositeReader: newBaseCompositeReader(spi, readers),
		directory:           directory,
	}
}

func OpenDirectoryReader(directory store.Directory) (r DirectoryReader, err error) {
	return openStandardDirectoryReader(directory, nil, DEFAULT_TERMS_INDEX_DIVISOR)
}

/*
Returns true if an index likely exists at the specified directory. Note that
if a corrupt index exists, or if an index in the process of committing
*/
func IsIndexExists(directory store.Directory) (ok bool, err error) {
	// LUCENE-2812, LUCENE-2727, LUCENE-4738: this logic will
	// return true in cases that should arguably be false,
	// such as only IW.prepareCommit has been called, or a
	// corrupt first commit, but it's too deadly to make
	// this logic "smarter" and risk accidentally returning
	// false due to various cases like file description
	// exhaustion, access denied, etc., because in that
	// case IndexWriter may delete the entire index.  It's
	// safer to err towards "index exists" than try to be
	// smart about detecting not-yet-fully-committed or
	// corrupt indices.  This means that IndexWriter will
	// throw an exception on such indices and the app must
	// resolve the situation manually:
	var files []string
	files, err = directory.ListAll()
	if _, ok := err.(*store.NoSuchDirectoryError); ok {
		// Directory does not exist --> no index exists
		return false, nil
	} else if err != nil {
		return false, err
	}
	return IsIndexFileExists(files), nil
}

/* No lock is required */
func IsIndexFileExists(files []string) bool {
	// Defensive: maybe a Directory impl returns null
	// instead of throwing NoSuchDirectoryException:
	if files != nil {
		prefix := INDEX_FILENAME_SEGMENTS + "_"
		for _, file := range files {
			if strings.HasPrefix(file, prefix) || file == INDEX_FILENAME_SEGMENTS_GEN {
				return true
			}
		}
	}
	return false
}

type StandardDirectoryReader struct {
	*DirectoryReaderImpl
	writer       *IndexWriter // NRT
	segmentInfos *SegmentInfos
}

// TODO support IndexWriter
func newStandardDirectoryReader(directory store.Directory, readers []AtomicReader,
	sis *SegmentInfos, termInfosIndexDivisor int, applyAllDeletes bool) *StandardDirectoryReader {
	// log.Printf("Initializing StandardDirectoryReader with %v sub readers...", len(readers))
	ans := &StandardDirectoryReader{segmentInfos: sis}
	ans.DirectoryReaderImpl = newDirectoryReader(ans, directory, readers)
	return ans
}

func openStandardDirectoryReader(directory store.Directory,
	commit IndexCommit, termInfosIndexDivisor int) (r DirectoryReader, err error) {
	// log.Print("Initializing SegmentsFile...")
	obj, err := NewFindSegmentsFile(directory, func(segmentFileName string) (interface{}, error) {
		sis := &SegmentInfos{}
		err := sis.Read(directory, segmentFileName)
		if err != nil {
			return nil, err
		}
		// log.Printf("Found %v segments...", len(sis.Segments))
		readers := make([]AtomicReader, len(sis.Segments))
		for i := len(sis.Segments) - 1; i >= 0; i-- {
			sr, err := NewSegmentReader(sis.Segments[i], termInfosIndexDivisor, store.IO_CONTEXT_READ)
			if err != nil {
				for _, r := range readers {
					if r != nil {
						util.CloseWhileSuppressingError(r)
					}
				}
				return nil, err
			}
			readers[i] = sr
		}
		// log.Printf("Obtained %v SegmentReaders.", len(readers))
		return newStandardDirectoryReader(directory, readers, sis, termInfosIndexDivisor, false), nil
	}).run(commit)
	if err != nil {
		return nil, err
	}
	return obj.(*StandardDirectoryReader), err
}

func (r *StandardDirectoryReader) String() string {
	var buf bytes.Buffer
	buf.WriteString("StandardDirectoryReader(")
	segmentsFile := r.segmentInfos.SegmentsFileName()
	if segmentsFile != "" {
		fmt.Fprintf(&buf, "%v:%v", segmentsFile, r.segmentInfos.version)
	}
	// if r.writer != nil {
	// fmt.Fprintf(w, "%v", r.writer)
	// }
	for _, v := range r.getSequentialSubReaders() {
		fmt.Fprintf(&buf, " %v", v)
	}
	buf.WriteString(")")
	return buf.String()
}

func (r *StandardDirectoryReader) Version() int64 {
	r.ensureOpen()
	return r.segmentInfos.version
}

func (r *StandardDirectoryReader) IsCurrent() bool {
	r.ensureOpen()
	// if writer == nill || writer.IsClosed() {
	// Fully read the segments file: this ensures that it's
	// completely written so that if
	// IndexWriter.prepareCommit has been called (but not
	// yet commit), then the reader will still see itself as
	// current:
	sis := SegmentInfos{}
	sis.ReadAll(r.directory)

	// we loaded SegmentInfos from the directory
	return sis.version == r.segmentInfos.version
	// } else {
	// return writer.nrtIsCurrent(r.segmentInfos)
	// }
}

func (r *StandardDirectoryReader) doClose() error {
	var firstErr error
	for _, r := range r.getSequentialSubReaders() {
		// try to close each reader, even if an error is returned
		func() {
			defer func() {
				if err := recover(); err != nil && firstErr == nil {
					if s, ok := err.(string); ok {
						firstErr = errors.New(s)
					} else {
						firstErr = errors.New(fmt.Sprintf("%v", err))
					}
				}
			}()
			r.decRef()
		}()
	}

	if w := r.writer; w != nil {
		panic("not implemented yet")
		// Since we just closed, writer may now be able to delete unused files:
		w.deletePendingFiles()
	}

	return firstErr
}
