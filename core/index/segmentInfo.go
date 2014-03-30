package index

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"regexp"
	"strconv"
)

const (
	INDEX_FILENAME_SEGMENTS     = "segments"
	INDEX_FILENAME_SEGMENTS_GEN = "segments.gen"
	COMOPUND_FILE_EXTENSION     = "cfs"
	VERSION_40                  = 0
	FORMAT_SEGMENTS_GEN_CURRENT = -2
)

type SegmentInfo struct {
	dir            store.Directory
	version        string
	name           string
	docCount       *util.SetOnce // number of docs in seg
	isCompoundFile bool
	codec          Codec
	diagnostics    map[string]string
	attributes     map[string]string
	files          map[string]bool // must use CheckFileNames()
}

func newSegmentInfo(dir store.Directory, version, name string, docCount int,
	isComoundFile bool, codec Codec, diagnostics map[string]string, attributes map[string]string) *SegmentInfo {
	_, ok := dir.(*TrackingDirectoryWrapper)
	assert(!ok)
	return &SegmentInfo{
		dir:            dir,
		version:        version,
		name:           name,
		docCount:       util.NewSetOnceOf(docCount),
		isCompoundFile: isComoundFile,
		codec:          codec,
		diagnostics:    diagnostics,
		attributes:     attributes,
	}
}

// seprate norms are not supported in >= 4.0
func (si *SegmentInfo) hasSeparateNorms() bool {
	return false
}

/* Return all files referenced by this SegmentInfo. */
func (si *SegmentInfo) Files() map[string]bool {
	assert2(si.files != nil, "files were not computed yet")
	return si.files
}

func (si *SegmentInfo) String() string {
	return si.StringOf(si.dir, 0)
}

func (si *SegmentInfo) StringOf(dir store.Directory, delCount int) string {
	var buf bytes.Buffer
	buf.WriteString(si.name)
	buf.WriteString("(")
	if si.version == "" {
		buf.WriteString("?")
	} else {
		buf.WriteString(si.version)
	}
	buf.WriteString("):")
	if si.isCompoundFile {
		buf.WriteString("c")
	} else {
		buf.WriteString("C")
	}

	if si.dir != dir {
		buf.WriteString("x")
	}
	fmt.Fprintf(&buf, "%v", si.docCount)

	if delCount != 0 {
		buf.WriteString("/")
		buf.WriteString(strconv.Itoa(delCount))
	}

	// TODO: we could append toString of attributes() here?

	return buf.String()
}

/* Sets the files written for this segment. */
func (si *SegmentInfo) setFiles(files map[string]bool) {
	si.checkFileNames(files)
	si.files = files
}

var CODEC_FILE_PATTERN = regexp.MustCompile("_[a-z0-9]+(_.*)?\\..*")

func (si *SegmentInfo) checkFileNames(files map[string]bool) {
	for file, _ := range files {
		if !CODEC_FILE_PATTERN.MatchString(file) {
			panic(fmt.Sprintf("invalid codec filename '%v', must match: %v", file, CODEC_FILE_PATTERN))
		}
	}
}
