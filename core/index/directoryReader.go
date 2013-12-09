package index

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"log"
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

func newDirectoryReader(self IndexReader, directory store.Directory, segmentReaders []AtomicReader) *DirectoryReaderImpl {
	log.Printf("Initializing DirectoryReader with %v segment readers...", len(segmentReaders))
	readers := make([]IndexReader, len(segmentReaders))
	for i, v := range segmentReaders {
		readers[i] = v
	}
	ans := &DirectoryReaderImpl{directory: directory}
	ans.BaseCompositeReader = newBaseCompositeReader(self, readers)
	return ans
}

func OpenDirectoryReader(directory store.Directory) (r DirectoryReader, err error) {
	return openStandardDirectoryReader(directory, DEFAULT_TERMS_INDEX_DIVISOR)
}

type StandardDirectoryReader struct {
	*DirectoryReaderImpl
	segmentInfos SegmentInfos
}

// TODO support IndexWriter
func newStandardDirectoryReader(directory store.Directory, readers []AtomicReader,
	sis SegmentInfos, termInfosIndexDivisor int, applyAllDeletes bool) *StandardDirectoryReader {
	log.Printf("Initializing StandardDirectoryReader with %v sub readers...", len(readers))
	ans := &StandardDirectoryReader{segmentInfos: sis}
	ans.DirectoryReaderImpl = newDirectoryReader(ans, directory, readers)
	return ans
}

// TODO support IndexCommit
func openStandardDirectoryReader(directory store.Directory,
	termInfosIndexDivisor int) (r DirectoryReader, err error) {
	log.Print("Initializing SegmentsFile...")
	obj, err := NewFindSegmentsFile(directory, func(segmentFileName string) (obj interface{}, err error) {
		sis := &SegmentInfos{}
		err = sis.Read(directory, segmentFileName)
		if err != nil {
			return nil, err
		}
		log.Printf("Found %v segments...", len(sis.Segments))
		readers := make([]AtomicReader, len(sis.Segments))
		for i := len(sis.Segments) - 1; i >= 0; i-- {
			sr, err := NewSegmentReader(sis.Segments[i], termInfosIndexDivisor, store.IO_CONTEXT_READ)
			readers[i] = sr
			if err != nil {
				rs := make([]io.Closer, len(readers))
				for i, v := range readers {
					rs[i] = v
				}
				return nil, util.CloseWhileHandlingError(err, rs...)
			}
		}
		log.Printf("Obtained %v SegmentReaders.", len(readers))
		return newStandardDirectoryReader(directory, readers, *sis, termInfosIndexDivisor, false), nil
	}).run()
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
	panic("not implemented yet")
}
