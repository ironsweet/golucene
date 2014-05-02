package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

// codecs/Codec.java

/*
Encodes/decodes an inverted index segment.

Note, when extending this class, the name is written into the index.
In order for the segment to be read, the name must resolve to your
implementation via forName(). This method use hard-coded map to
resolve codec names.

If you implement your own codec, make sure that it is included so
SPI can load it.
*/
type Codec interface {
	// Returns this codec's name
	Name() string
	// Encodes/decodes postings
	PostingsFormat() PostingsFormat
	// Encodes/decodes docvalues
	DocValuesFormat() DocValuesFormat
	// Encodes/decodes stored fields
	StoredFieldsFormat() StoredFieldsFormat
	// Encodes/decodes term vectors
	TermVectorsFormat() TermVectorsFormat
	// Encodes/decodes field infos file
	FieldInfosFormat() FieldInfosFormat
	// Encodes/decodes segment info file
	SegmentInfoFormat() SegmentInfoFormat
	// Encodes/decodes document normalization values
	NormsFormat() NormsFormat
	// Encodes/decodes live docs
	LiveDocsFormat() LiveDocsFormat
}

type CodecImpl struct {
	name             string
	fieldsFormat     StoredFieldsFormat
	vectorsFormat    TermVectorsFormat
	fieldInfosFormat FieldInfosFormat
	infosFormat      SegmentInfoFormat
	liveDocsFormat   LiveDocsFormat
	postingsFormat   PostingsFormat
	docValuesFormat  DocValuesFormat
	normsFormat      NormsFormat
}

func (codec *CodecImpl) Name() string {
	return codec.name
}

func (codec *CodecImpl) PostingsFormat() PostingsFormat {
	return codec.postingsFormat
}

func (codec *CodecImpl) DocValuesFormat() DocValuesFormat {
	return codec.docValuesFormat
}

func (codec *CodecImpl) StoredFieldsFormat() StoredFieldsFormat {
	return codec.fieldsFormat
}

func (codec *CodecImpl) TermVectorsFormat() TermVectorsFormat {
	return codec.vectorsFormat
}

func (codec *CodecImpl) FieldInfosFormat() FieldInfosFormat {
	return codec.fieldInfosFormat
}

func (codec *CodecImpl) SegmentInfoFormat() SegmentInfoFormat {
	return codec.infosFormat
}

func (codec *CodecImpl) NormsFormat() NormsFormat {
	return codec.normsFormat
}

func (codec *CodecImpl) LiveDocsFormat() LiveDocsFormat {
	return codec.liveDocsFormat
}

/*
returns the codec's name. Subclass can override to provide more
detail (such as parameters.)
*/
func (codec *CodecImpl) String() string {
	return codec.name
}

var allCodecs = map[string]Codec{
	"Lucene42": Lucene42Codec,
	"Lucene45": Lucene45Codec,
}

// workaround Lucene Java's SPI mechanism
func RegisterCodec(codecs ...Codec) {
	for _, codec := range codecs {
		allCodecs[codec.Name()] = codec
	}
}

// looks up a codec by name
func LoadCodec(name string) Codec {
	return allCodecs[name]
}

// returns a list of all available codec names
func AvailableCodecs() []string {
	ans := make([]string, 0, len(allCodecs))
	for name, _ := range allCodecs {
		ans = append(ans, name)
	}
	return ans
}

// Expert: returns the default codec used for newly created IndexWriterConfig(s).
var DefaultCodec = func() Codec { return LoadCodec("Lucene45") }

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
	FieldsConsumer(state SegmentWriteState) (FieldsConsumer, error)
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

var allPostingsFormats = map[string]PostingsFormat{
	"Lucene41": new(Lucene41PostingsFormat),
}

// workaround Lucene Java's SPI mechanism
func RegisterPostingsFormat(formats ...PostingsFormat) {
	for _, format := range formats {
		allPostingsFormats[format.Name()] = format
	}
}

/* looks up a format by name */
func LoadPostingsFormat(name string) PostingsFormat {
	v, ok := allPostingsFormats[name]
	assertn(ok, "Service '%v' not found.", name)
	return v
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
}

// codecs/DocValuesFormat.java

/*
Encodes/decodes per-document values.

Note, when extending this class, the name Name() may be written into
the index in certain configurations. In order for the segment to be
read, the name must resolve to your implemetation via LoadXYZ().
Since Go doesn't have Java's SPI locate mechanism, this method use
manual mappings to resolve format names.

If you implement your own format, make sure that it is manually
included.
*/
type DocValuesFormat interface {
	Name() string
	// Returns a DocValuesConsumer to write docvalues to the index.
	FieldsConsumer(state SegmentWriteState) (w DocValuesConsumer, err error)
	// Returns a DocValuesProducer to read docvalues from the index.
	//
	// NOTE: by the time this call returns, it must
	// hold open any files it will need to use; else, those files may
	// be deleted. Additionally, required fiels may be deleted during
	// the execution of this call before there is a chance to open them.
	// Under these circumstances an IO error should be returned by the
	// implementation. IO errors are expected and will automatically
	// cause a retry of the segment opening logic with the newly
	// revised segments.
	FieldsProducer(state SegmentReadState) (r DocValuesProducer, err error)
}

func LoadDocValuesProducer(name string, state SegmentReadState) (fp DocValuesProducer, err error) {
	switch name {
	case "Lucene42":
		return newLucene42DocValuesProducer(state, LUCENE42_DV_DATA_CODEC, LUCENE42_DV_DATA_EXTENSION,
			LUCENE42_DV_METADATA_CODEC, LUCENE42_DV_METADATA_EXTENSION)
	case "Lucene45":
		return newLucene45DocValuesProducer(state, LUCENE45_DV_DATA_CODEC, LUCENE45_DV_DATA_EXTENSION,
			LUCENE45_DV_META_CODEC, LUCENE45_DV_META_EXTENSION)
	}
	panic(fmt.Sprintf("Service '%v' not found.", name))
}

// codecs/DocValuesConsumer.java
/*
Abstract API that consumes numeric, binary and sorted docvalues.
Concret implementations of this actually do "something" with the
docvalues (write it into the index in a specific format).

The lifecycle is:

1. DocValuesConsumer is created by DocValuesFormat.FieldsConsumer()
or NormsFormat.NormsConsumer().
2. AddNumericField, AddBinaryField, or addSortedField are called for
each Numeric, Binary, or Sorted docvalues field. The API is a "pull"
rather than "push", and the implementation is free to iterate over
the values multiple times.
3. After all fields are added, the consumer is closed.
*/
type DocValuesConsumer interface {
	io.Closer
}

// codecs/StoredFieldsFormat.java

// Controls the format of stored fields
type StoredFieldsFormat interface {
	// Returns a StoredFieldsReader to load stored fields.
	FieldsReader(d store.Directory, si *model.SegmentInfo, fn model.FieldInfos, context store.IOContext) (r StoredFieldsReader, err error)
	// Returns a StoredFieldsWriter to write stored fields.
	FieldsWriter(d store.Directory, si *model.SegmentInfo, context store.IOContext) (w StoredFieldsWriter, err error)
}

// codecs/StoredFieldsReader.java

type StoredFieldsReader interface {
	io.Closer
	visitDocument(n int, visitor StoredFieldVisitor) error
	Clone() StoredFieldsReader
}

// codecs/StoredFieldsWriter.java

/*
Codec API for writing stored fields:

1. For every document, StartDocument() is called, informing the Codec
how many fields will be written.
2. WriteField() is called for each field in the document.
3. After all documents have been writen, Finish() is called for
verification/sanity-checks.
4. Finally the writer is closed.
*/
type StoredFieldsWriter interface {
	io.Closer
	// Called before writing the stored fields of te document.
	// WriteField() will be called numStoredFields times. Note that
	// this is called even if the document has no stored fields, in
	// this case numStoredFields will be zero.
	StartDocument(numStoredFields int) error
	// Called when a document and all its fields have been added.
	FinishDocument() error
	// Aborts writing entirely, implementation should remove any
	// partially-written files, etc.
	Abort()
	// Called before Close(), passing in the number of documents that
	// were written. Note that this is intentionally redundant
	// (equivalent to the number of calls to startDocument(int)), but a
	// Codec should check that this is the case to detect the JRE bug
	// described in LUCENE-1282.
	Finish(fis model.FieldInfos, numDocs int) error
}

// codecs/TermVectorsFormat.java

// Controls the format of term vectors
type TermVectorsFormat interface {
	// Returns a TermVectorsReader to read term vectors.
	VectorsReader(d store.Directory, si *model.SegmentInfo, fn model.FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error)
	// Returns a TermVectorsWriter to write term vectors.
	VectorsWriter(d store.Directory, si *model.SegmentInfo, ctx store.IOContext) (w TermVectorsWriter, err error)
}

// codecs/TermVectorsReader.java

type TermVectorsReader interface {
	io.Closer
	get(doc int) Fields
	clone() TermVectorsReader
}

// codecs/TermVectorsWriter.java

/*
Codec API for writing term vecrors:

1. For every document, StartDocument() is called, informing the Codec
how may fields will be written.
2. StartField() is called for each field in the document, informing
the codec how many terms will be written for that field, and whether
or not positions, offsets, or payloads are enabled.
3. Within each field, StartTerm() is called for each term.
4. If offsets and/or positions are enabled, then AddPosition() will
be called for each term occurrence.
5. After all documents have been written, Finish() is called for
verification/sanity-checks.
6. Finally the writer is closed.
*/
type TermVectorsWriter interface {
	io.Closer
	// Called before writing the term vectors of the document.
	// startField() will be called numVectorsFields times. Note that if
	// term vectors are enabled, this is called even if the document
	// has no vector fields, in this case numVectorFields will be zero.
	startDocument(int)
	// Called after a doc and all its fields have been added
	finishDocument() error
	// Aborts writing entirely, implementation should remove any
	// partially-written files, etc.
	abort()
	// Called before Close(), passing in the number of documents that
	// were written. Note that this is intentionally redendant
	// (equivalent to the number of calls to startDocument(int)), but a
	// Codec should check that this is the case to detect the JRE bug
	// described in LUCENE-1282.
	finish(model.FieldInfos, int) error
}

// codecs/FieldInfosFormat.java

// Encodes/decodes FieldInfos
type FieldInfosFormat interface {
	// Returns a FieldInfosReader to read field infos from the index
	FieldInfosReader() FieldInfosReader
	// Returns a FieldInfosWriter to write field infos to the index
	FieldInfosWriter() FieldInfosWriter
}

// codecs/FieldInfosReader.java

// Codec API for reading FieldInfos.
type FieldInfosReader func(d store.Directory, name string, ctx store.IOContext) (infos model.FieldInfos, err error)

// Codec API for writing FieldInfos.
type FieldInfosWriter func(d store.Directory, name string, infos model.FieldInfos, ctx store.IOContext) error

// codecs/SegmentInfoFormat.java

// Expert: Control the format of SegmentInfo (segment metadata file).
type SegmentInfoFormat interface {
	// Returns the SegmentInfoReader for reading SegmentInfo instances.
	SegmentInfoReader() SegmentInfoReader
	// Returns the SegmentInfoWriter for writing SegmentInfo instances.
	SegmentInfoWriter() SegmentInfoWriter
}

// codecs/SegmentInfoReader.java

// Read SegmentInfo data from a directory.
type SegmentInfoReader func(d store.Directory, name string, ctx store.IOContext) (info *model.SegmentInfo, err error)

// codecs/SegmentInfoWriter.java

// Write SegmentInfo data.
type SegmentInfoWriter func(d store.Directory, info *model.SegmentInfo, infos model.FieldInfos, ctx store.IOContext) error

// codecs/NormsFormat.java

// Encodes/decodes per-document score normalization values.
type NormsFormat interface {
	// Returns a DocValuesConsumer to write norms to the index.
	NormsConsumer(state SegmentWriteState) (w DocValuesConsumer, err error)
	// Returns a DocValuesProducer to read norms from the index.
	//
	// NOTE: by the time this call returns, it must
	// hold open any files it will need to use; else, those files may
	// be deleted. Additionally, required fiels may be deleted during
	// the execution of this call before there is a chance to open them.
	// Under these circumstances an IO error should be returned by the
	// implementation. IO errors are expected and will automatically
	// cause a retry of the segment opening logic with the newly
	// revised segments.
	NormsProducer(state SegmentReadState) (r DocValuesProducer, err error)
}

// codecs/DocvaluesProducer.java

// Abstract API that produces numeric, binary and sorted docvalues.
type DocValuesProducer interface {
	io.Closer
	Numeric(field model.FieldInfo) (v NumericDocValues, err error)
	Binary(field model.FieldInfo) (v BinaryDocValues, err error)
	Sorted(field model.FieldInfo) (v SortedDocValues, err error)
	SortedSet(field model.FieldInfo) (v SortedSetDocValues, err error)
}

// codecs/LiveDocsFormat.java

/* Format for live/deleted documents */
type LiveDocsFormat interface {
	// Creates a new MutableBits, with all bits set, for the specified size.
	NewLiveDocs(size int) util.MutableBits
	// Creates a new MutableBits of the same bits set and size of existing.
	// NewLiveDocs(existing util.Bits) (util.MutableBits, error)
	// Persist live docs bits. Use SegmentInfoPerCommit.nextDelGen() to
	// determine the generation of the deletes file you should write to.
	WriteLiveDocs(bits util.MutableBits, dir store.Directory,
		info *SegmentInfoPerCommit, newDelCount int, ctx store.IOContext) error
}
