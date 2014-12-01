package index

import (
	"errors"
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"runtime/debug"
	"strconv"
)

// index/CheckIndex.java

/* Returned from checkIndex() detailing the health and status of the index */
type CheckIndexStatus struct {
	// True if no problems found with the index.
	Clean bool

	// True if we were unable to locate and load the segments_N file.
	MissingSegments bool

	// True if we were unable to open the segments_N file.
	cantOpenSegments bool

	// True if we were unable to read the versioin number from segments_N file.
	missingSegmentVersion bool

	// Name of latest segments_N file in the index.
	segmentsFilename string

	// Number of segments in the index
	numSegments int

	// True if the index was created with a newer version of Lucene than the CheckIndex tool.
	toolOutOfDate bool

	// List of SegmentInfoStatus instances, detailing status of each segment.
	segmentInfos []*SegmentInfoStatus

	// Directory index is in.
	dir store.Directory

	// SegmentInfos instance containing only segments that had no
	// problems (this is used with the fixIndex() method to repare the
	// index)
	newSegments *SegmentInfos

	// How many documents will be lost to bad segments.
	totLoseDocCount int

	// How many bad segments were found.
	numBadSegments int

	// Whether the SegmentInfos.counter is greater than any of the segments' names.
	validCounter bool

	// The greatest segment name.
	maxSegmentName int

	// Holds the userData of the last commit in the index
	userData map[string]string
}

/* Holds the status of each segment in the index. */
type SegmentInfoStatus struct {
	// Name of the segment
	name string

	// Codec used to read this segment.
	codec Codec

	// Document count (does not take deletions into account).
	docCount int

	// True if segment is compound file format.
	compound bool

	// Number of files referenced by this segment.
	numFiles int

	// Net size (MB) of the files referenced by this segment.
	sizeMB float64

	// True if this segment has pending deletions.
	hasDeletions bool

	// Current deletions generation.
	deletionsGen int64

	// Number of deleted documents.
	numDeleted int

	// True if we were able to open an AR on this segment.
	openReaderPassed bool

	// Number of fields in this segment.
	numFields int

	// Map that includes certain debugging details that IW records into each segment it creates
	diagnostics map[string]string

	// Status for testing of field norms (nil if field norms could not be tested).
	fieldNormStatus *FieldNormStatus

	// Status for testing of indexed terms (nil if indexed terms could not be tested).
	termIndexStatus *TermIndexStatus

	// Status for testing of stored fields (nil if stored fields could not be tested).
	storedFieldStatus *StoredFieldStatus

	// Status for testing term vectors (nil if term vectors could not be tested).
	termVectorStatus *TermVectorStatus

	// Status for testing of DocVlaues (nil if DocValues could not be tested).
	docValuesStatus *DocValuesStatus
}

type FieldNormStatus struct {
	err error
}

type TermIndexStatus struct {
	err error
}

type StoredFieldStatus struct {
	err error
}

type TermVectorStatus struct {
	err error
}

type DocValuesStatus struct {
	err error
}

/*
Basic tool and API to check the health of an index and write a new
segments file that removes reference to problematic segments.

As this tool checks every byte in the index, on a large index it can
take a long time to run.
*/
type CheckIndex struct {
	infoStream            io.Writer
	dir                   store.Directory
	crossCheckTermVectors bool
	failFast              bool
}

func NewCheckIndex(dir store.Directory, crossCheckTermVectors bool, infoStream io.Writer) *CheckIndex {
	return &CheckIndex{
		infoStream: infoStream,
		dir:        dir,
		crossCheckTermVectors: crossCheckTermVectors,
	}
}

func (ch *CheckIndex) msg(msg string, args ...interface{}) {
	fmt.Fprintf(ch.infoStream, msg, args...)
	fmt.Fprintln(ch.infoStream)
}

/*
Returns a Status instance detailing the state of the index.

As this method checks every byte in the specified segments, on a
large index it can take quite a long time to run.

WARNING: make sure you only call this when the index is not opened
by any writer.
*/
func (ch *CheckIndex) CheckIndex(onlySegments []string) *CheckIndexStatus {
	sis := &SegmentInfos{}
	result := &CheckIndexStatus{
		dir: ch.dir,
	}
	err := sis.ReadAll(ch.dir)
	if err != nil {
		if ch.failFast {
			panic("niy")
		}
		fmt.Fprintln(ch.infoStream, "ERROR: could not read any segments file in directory")
		debug.PrintStack()
		result.MissingSegments = true
		return result
	}

	// find the oldest and newest segment versions
	var oldest util.Version
	var newest util.Version
	var oldSegs string
	for _, si := range sis.Segments {
		if version := si.Info.Version(); len(version) != 0 {
			if len(oldest) == 0 || !version.OnOrAfter(oldest) {
				oldest = version
			}
			if len(newest) == 0 || version.OnOrAfter(newest) {
				newest = version
			}
		} else {
			// pre-3.1 segment
			oldSegs = "pre-3.1"
		}
	}

	numSegments := len(sis.Segments)
	segmentsFilename := sis.SegmentsFileName()
	// note: we only read the format byte (required preamble) here!
	input, err := ch.dir.OpenInput(segmentsFilename, store.IO_CONTEXT_READONCE)
	if err != nil {
		if ch.failFast {
			panic("niy")
		}
		fmt.Fprintln(ch.infoStream, "ERROR: could not open segments file in directory")
		debug.PrintStack()
		result.cantOpenSegments = true
		return result
	}
	defer input.Close() // ignore error

	_, err = input.ReadInt()
	if err != nil {
		if ch.failFast {
			panic("niy")
		}
		fmt.Fprintln(ch.infoStream, "ERROR: could not read segment file version in directory")
		debug.PrintStack()
		result.missingSegmentVersion = true
		return result
	}

	var sFormat string
	var skip = false

	result.segmentsFilename = segmentsFilename
	result.numSegments = numSegments
	result.userData = sis.userData
	var userDataStr string
	if len(sis.userData) > 0 {
		userDataStr = fmt.Sprintf(" userData=%v", sis.userData)
	}

	var versionStr string
	if oldSegs != "" {
		if len(newest) != 0 {
			versionStr = fmt.Sprintf("versions=[%v .. %v]", oldSegs, newest)
		} else {
			versionStr = fmt.Sprintf("version=%v", oldSegs)
		}
	} else if len(newest) != 0 { // implies oldest is set
		if newest.Equals(oldest) {
			versionStr = fmt.Sprintf("version=%v", oldest)
		} else {
			versionStr = fmt.Sprintf("versions=[%v .. %v]", oldest, newest)
		}
	}

	ch.msg("Segments file=%v numSegments=%v %v format=%v%v",
		segmentsFilename, numSegments, versionStr, sFormat, userDataStr)

	names := make(map[string]bool)
	if onlySegments != nil {
		for _, name := range onlySegments {
			names[name] = true
		}
		panic("not implemented yet")
	}

	if skip {
		ch.msg(
			"\nERROR: this index appears to be created by a newer version of Lucene than this tool was compiled on; please re-compile this tool on the matching version of Lucene; exiting")
		result.toolOutOfDate = true
		return result
	}

	result.newSegments = sis.Clone()
	result.newSegments.Clear()
	result.maxSegmentName = -1

	for i, info := range sis.Segments {
		segmentName, err := strconv.ParseInt(info.Info.Name[1:], 36, 32)
		if err != nil {
			panic(err) // impossible
		}
		if int(segmentName) > result.maxSegmentName {
			result.maxSegmentName = int(segmentName)
		}
		if _, ok := names[info.Info.Name]; !ok {
			continue
		}
		segInfoStat := new(SegmentInfoStatus)
		result.segmentInfos = append(result.segmentInfos, segInfoStat)
		infoDocCount := info.Info.DocCount()
		ch.msg("  %v of %v: name=%v docCount=%v ",
			1+i, numSegments, info.Info.Name, infoDocCount)
		segInfoStat.name = info.Info.Name
		segInfoStat.docCount = infoDocCount

		version := info.Info.Version()
		if infoDocCount <= 0 && version.OnOrAfter(util.VERSION_45) {
			panic(fmt.Sprintf("illegal number of documents: maxDoc=%v", infoDocCount))
		}

		toLoseDocCount := infoDocCount
		err = func() error {
			assert2(len(version) != 0, "pre 4.0 is not supported yet")
			ch.msg("    version=%v", version)
			codec := info.Info.Codec().(Codec)
			ch.msg("    codec = %v", codec)
			segInfoStat.codec = codec
			ch.msg("    compound = %v", info.Info.IsCompoundFile())
			segInfoStat.compound = info.Info.IsCompoundFile()
			ch.msg("    numFiles = %v", len(info.Files()))
			segInfoStat.numFiles = len(info.Files())
			n, err := info.SizeInBytes()
			if err != nil {
				return err
			}
			segInfoStat.sizeMB = float64(n) / (1024 * 1024)
			if v := info.Info.Attribute("Lucene3xSegmentInfoFormat.dsoffset"); v == "" {
				// don't print size in bytes if it's a 3.0 segment iwht shared docstores
				ch.msg("    size (MB) = %v", segInfoStat.sizeMB)
			}

			diagnostics := info.Info.Diagnostics()
			segInfoStat.diagnostics = diagnostics
			if len(diagnostics) > 0 {
				ch.msg("    diagnostics = %v", diagnostics)
			}

			atts := info.Info.Attributes()
			if len(atts) > 0 {
				ch.msg("    attributes = %v", atts)
			}

			panic("not implemented yet")

			if !info.HasDeletions() {
				ch.msg("    no deletions")
				segInfoStat.hasDeletions = false
			} else {
				ch.msg("     has deletions [delGen = %v]", info.DelGen())
				segInfoStat.hasDeletions = true
				segInfoStat.deletionsGen = info.DelGen()
			}

			ch.msg("    test: open reader.........")
			reader, err := NewSegmentReader(info, DEFAULT_TERMS_INDEX_DIVISOR, store.IO_CONTEXT_DEFAULT)
			if err != nil {
				return err
			}
			defer reader.Close()

			segInfoStat.openReaderPassed = true

			numDocs := reader.NumDocs()
			toLoseDocCount = numDocs
			if reader.hasDeletions() {
				if n := infoDocCount - info.DelCount(); n != reader.NumDocs() {
					return errors.New(fmt.Sprintf(
						"delete count mismatch: info=%v vs reader=%v",
						n, reader.NumDocs()))
				}
				if n := infoDocCount - reader.NumDocs(); n > reader.MaxDoc() {
					return errors.New(fmt.Sprintf(
						"too many deleted docs: maxDoc()=%v vs del count=%v",
						reader.MaxDoc(), n))
				}
				if n := infoDocCount - numDocs; n != info.DelCount() {
					return errors.New(fmt.Sprintf(
						"delete count mismatch: info=%v vs reader=%v",
						info.DelCount(), n))
				}
				liveDocs := reader.LiveDocs()
				if liveDocs == nil {
					return errors.New("segment should have deletions, but liveDocs is nil")
				} else {
					var numLive = 0
					for j := 0; j < liveDocs.Length(); j++ {
						if liveDocs.At(j) {
							numLive++
						}
					}
					if numLive != numDocs {
						return errors.New(fmt.Sprintf(
							"liveDocs count mismatch: info=%v, vs bits=%v",
							numDocs, numLive))
					}
				}

				segInfoStat.numDeleted = infoDocCount - numDocs
				ch.msg("OK [%v deleted docs]", segInfoStat.numDeleted)
			} else {
				if info.DelCount() != 0 {
					return errors.New(fmt.Sprintf(
						"delete count mismatch: info=%v vs reader=%v",
						info.DelCount(), infoDocCount-numDocs))
				}
				liveDocs := reader.LiveDocs()
				if liveDocs != nil {
					// it's ok for it to be non-nil here, as long as none are set right?
					for j := 0; j < liveDocs.Length(); j++ {
						if !liveDocs.At(j) {
							return errors.New(fmt.Sprintf(
								"liveDocs mismatch: info says no deletions but doc %v is deleted.", j))
						}
					}
				}
				ch.msg("OK")
			}
			if reader.MaxDoc() != infoDocCount {
				return errors.New(fmt.Sprintf(
					"SegmentReader.maxDoc() %v != SegmentInfos.docCount %v",
					reader.MaxDoc(), infoDocCount))
			}

			// Test getFieldInfos()
			ch.msg("    test: fields..............")
			fieldInfos := reader.FieldInfos()
			ch.msg("OK [%v fields]", fieldInfos.Size())
			segInfoStat.numFields = fieldInfos.Size()

			segInfoStat.fieldNormStatus = ch.testFieldNorms(reader)
			segInfoStat.termIndexStatus = ch.testPostings(reader)
			segInfoStat.storedFieldStatus = ch.testStoredFields(reader)
			segInfoStat.termVectorStatus = ch.testTermVectors(reader)
			segInfoStat.docValuesStatus = ch.testDocValues(reader)

			// Rethrow the first error we encountered
			// This will cause stats for failed segments to be incremented properly
			if segInfoStat.fieldNormStatus.err != nil {
				return errors.New("Field Norm test failed")
			} else if segInfoStat.termIndexStatus.err != nil {
				return errors.New("Term Index test failed")
			} else if segInfoStat.storedFieldStatus.err != nil {
				return errors.New("Stored Field test failed")
			} else if segInfoStat.termVectorStatus.err != nil {
				return errors.New("Term Vector test failed")
			} else if segInfoStat.docValuesStatus.err != nil {
				return errors.New("DocValues test failed")
			}

			ch.msg("")
			return nil
		}()
		if err != nil {
			if ch.failFast {
				panic("niy")
			}
			ch.msg("FAILED")
			comment := "fixIndex() would remove reference to this segment"
			ch.msg("    WARNING: %v; full error:", comment)
			ch.msg(string(debug.Stack()))
			ch.msg("")
			result.totLoseDocCount += toLoseDocCount
			result.numBadSegments++
		} else {
			// Keeper
			result.newSegments.Segments = append(result.newSegments.Segments, info.Clone())
		}
	}

	if result.numBadSegments == 0 {
		result.Clean = true
	} else {
		ch.msg(
			"WARNING: %v broken segments (containing %v documents) detected",
			result.numBadSegments, result.totLoseDocCount)
	}

	result.validCounter = result.maxSegmentName < sis.counter
	if !result.validCounter {
		result.Clean = false
		result.newSegments.counter = result.maxSegmentName + 1
		ch.msg(
			"ERROR: Next segment name counter %v is not greater than max segment name %v",
			sis.counter, result.maxSegmentName)
	}

	if result.Clean {
		ch.msg("No problems were detected with this index.\n")
	}

	return result
}

func (ch *CheckIndex) testFieldNorms(reader AtomicReader) *FieldNormStatus {
	panic("not implemented yet")
}

func (ch *CheckIndex) testPostings(reader AtomicReader) *TermIndexStatus {
	panic("not implemented yet")
}

func (ch *CheckIndex) testStoredFields(reader AtomicReader) *StoredFieldStatus {
	panic("not implemented yet")
}

func (ch *CheckIndex) testDocValues(reader AtomicReader) *DocValuesStatus {
	panic("not implemented yet")
}

func (ch *CheckIndex) testTermVectors(reader AtomicReader) *TermVectorStatus {
	panic("not implemented yet")
}
