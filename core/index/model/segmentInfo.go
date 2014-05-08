package model

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"regexp"
	"strconv"
)

type SegmentInfo struct {
	Dir            store.Directory
	version        string
	Name           string
	docCount       *util.SetOnce // number of docs in seg
	isCompoundFile bool
	codec          interface{}
	diagnostics    map[string]string
	files          map[string]bool // must use CheckFileNames()

	*AttributesMixin
}

func (info *SegmentInfo) SetDiagnostics(diagnostics map[string]string) {
	info.diagnostics = diagnostics
}

/* Returns diagnostics saved into the segment when it was written .*/
func (info *SegmentInfo) Diagnostics() map[string]string {
	return info.diagnostics
}

func NewSegmentInfo(dir store.Directory, version, name string, docCount int,
	isComoundFile bool, codec interface{}, diagnostics map[string]string, attributes map[string]string) *SegmentInfo {
	_, ok := dir.(*store.TrackingDirectoryWrapper)
	assert(!ok)
	return &SegmentInfo{
		Dir:             dir,
		version:         version,
		Name:            name,
		docCount:        util.NewSetOnceOf(docCount),
		isCompoundFile:  isComoundFile,
		codec:           codec,
		diagnostics:     diagnostics,
		AttributesMixin: &AttributesMixin{attributes},
	}
}

// seprate norms are not supported in >= 4.0
func (si *SegmentInfo) HasSeparateNorms() bool {
	return false
}

/* Returns true if this segment is stored as a compound file */
func (si *SegmentInfo) IsCompoundFile() bool {
	return si.isCompoundFile
}

/* Can only be called once. */
func (info *SegmentInfo) SetCodec(codec interface{}) {
	assert(info.codec == nil)
	assert2(codec != nil, "segmentCodecs must not be nil")
	info.codec = codec
}

/* Return Codec that wrote this segment. */
func (si *SegmentInfo) Codec() interface{} {
	return si.codec
}

func (si *SegmentInfo) DocCount() int {
	return si.docCount.Get().(int)
}

func (info *SegmentInfo) SetDocCount(docCount int) {
	info.docCount.Set(docCount)
}

/* Return all files referenced by this SegmentInfo. */
func (si *SegmentInfo) Files() map[string]bool {
	assert2(si.files != nil, "files were not computed yet")
	return si.files
}

func (si *SegmentInfo) String() string {
	return si.StringOf(si.Dir, 0)
}

func (si *SegmentInfo) StringOf(dir store.Directory, delCount int) string {
	var buf bytes.Buffer
	buf.WriteString(si.Name)
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

	if si.Dir != dir {
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

/* Returns the version of the code which wrote the segment. */
func (si *SegmentInfo) Version() string {
	return si.version
}

/* Sets the files written for this segment. */
func (si *SegmentInfo) SetFiles(files map[string]bool) {
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
