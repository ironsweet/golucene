package spi

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/index/model"
	"io"
)

// codecs/PostingsFormat.java

/*
Encodes/decodes terms, postings, and proximity data.

Note, when extending this class, the name Name() may be written into
the index in certain configurations. In order for the segment to be
read, the name must resolve to your implemetation via LoadPostingsFormat().
Since Go doesn't have Java's SPI locate mechanism, this method use
manual mappings to resolve format names.

If you implement your own format, make sure that it is manually
included.
*/
type PostingsFormat interface {
	// Returns this posting format's name
	Name() string
	// Writes a new segment
	FieldsConsumer(state *SegmentWriteState) (FieldsConsumer, error)
	// Reads a segment. NOTE: by the time this call returns, it must
	// hold open any files it will need to use; else, those files may
	// be deleted. Additionally, required fiels may be deleted during
	// the execution of this call before there is a chance to open them.
	// Under these circumstances an IO error should be returned by the
	// implementation. IO errors are expected and will automatically
	// cause a retry of the segment opening logic with the newly
	// revised segments.
	FieldsProducer(state SegmentReadState) (FieldsProducer, error)
}

type PostingsFormatImpl struct {
	name string
}

// Returns this posting format's name
func (pf *PostingsFormatImpl) Name() string {
	return pf.name
}

func (pf *PostingsFormatImpl) String() string {
	return fmt.Sprintf("PostingsFormat(name=%v)", pf.name)
}

var allPostingsFormats = map[string]PostingsFormat{}

// workaround Lucene Java's SPI mechanism
func RegisterPostingsFormat(formats ...PostingsFormat) {
	for _, format := range formats {
		fmt.Printf("Found postings format: %v\n", format.Name())
		allPostingsFormats[format.Name()] = format
	}
}

/* looks up a format by name */
func LoadPostingsFormat(name string) PostingsFormat {
	v, ok := allPostingsFormats[name]
	assert2(ok, "Service '%v' not found.", name)
	return v
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

/* Returns a list of all available format names. */
func AvailablePostingsFormats() []string {
	ans := make([]string, 0, len(allPostingsFormats))
	for name, _ := range allPostingsFormats {
		ans = append(ans, name)
	}
	return ans
}

// codecs/FieldsConsumer.java

/*
Abstract API that consumes terms, doc, freq, prox, offset and
payloads postings. Concrete implementations of this actually do
"something" with the postings (write it into the index in a specific
format).

The lifecycle is:

1. FieldsConsumer is created by PostingsFormat.FieldsConsumer().
2. For each field, AddField() is called, returning a TermsConsumer
for the field.
*/
type FieldsConsumer interface {
	io.Closer
	// Add a new field
	AddField(field *FieldInfo) (TermsConsumer, error)
}

type FieldsProducer interface {
	Fields
	io.Closer
}
