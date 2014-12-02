package spi

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
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

func NewCodec(name string,
	fieldsFormat StoredFieldsFormat,
	vectorsFormat TermVectorsFormat,
	fieldInfosFormat FieldInfosFormat,
	infosFormat SegmentInfoFormat,
	liveDocsFormat LiveDocsFormat,
	postingsFormat PostingsFormat,
	docValuesFormat DocValuesFormat,
	normsFormat NormsFormat) *CodecImpl {
	return &CodecImpl{name, fieldsFormat, vectorsFormat, fieldInfosFormat,
		infosFormat, liveDocsFormat, postingsFormat, docValuesFormat, normsFormat}
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

var allCodecs = make(map[string]Codec)

// workaround Lucene Java's SPI mechanism
func RegisterCodec(codecs ...Codec) {
	for _, codec := range codecs {
		fmt.Printf("Found codec: %v\n", codec.Name())
		allCodecs[codec.Name()] = codec
	}
}

// looks up a codec by name
func LoadCodec(name string) Codec {
	c, ok := allCodecs[name]
	if !ok {
		fmt.Println("Unknown codec:", name)
		fmt.Println("Available codecs:", allCodecs)
		assert(ok)
	}
	return c
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
var DefaultCodec = func() Codec {
	ans := LoadCodec("Lucene410")
	assert(ans != nil)
	return ans
}

// codecs/StoredFieldsFormat.java

// Controls the format of stored fields
type StoredFieldsFormat interface {
	// Returns a StoredFieldsReader to load stored fields.
	FieldsReader(d store.Directory, si *SegmentInfo, fn FieldInfos, context store.IOContext) (r StoredFieldsReader, err error)
	// Returns a StoredFieldsWriter to write stored fields.
	FieldsWriter(d store.Directory, si *SegmentInfo, context store.IOContext) (w StoredFieldsWriter, err error)
}

// codecs/TermVectorsFormat.java

// Controls the format of term vectors
type TermVectorsFormat interface {
	// Returns a TermVectorsReader to read term vectors.
	VectorsReader(d store.Directory, si *SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error)
	// Returns a TermVectorsWriter to write term vectors.
	VectorsWriter(d store.Directory, si *SegmentInfo, ctx store.IOContext) (w TermVectorsWriter, err error)
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
type FieldInfosReader func(d store.Directory, name, suffix string, ctx store.IOContext) (infos FieldInfos, err error)

// Codec API for writing FieldInfos.
type FieldInfosWriter func(d store.Directory, name, suffix string, infos FieldInfos, ctx store.IOContext) error

// codecs/NormsFormat.java

// Encodes/decodes per-document score normalization values.
type NormsFormat interface {
	// Returns a DocValuesConsumer to write norms to the index.
	NormsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error)
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

// codecs/LiveDocsFormat.java

/* Format for live/deleted documents */
type LiveDocsFormat interface {
	// Creates a new MutableBits, with all bits set, for the specified size.
	NewLiveDocs(size int) util.MutableBits
	// Creates a new MutableBits of the same bits set and size of existing.
	// NewLiveDocs(existing util.Bits) (util.MutableBits, error)
	// Persist live docs bits. Use SegmentCommitInfo.nextDelGen() to
	// determine the generation of the deletes file you should write to.
	WriteLiveDocs(bits util.MutableBits, dir store.Directory,
		info *SegmentCommitInfo, newDelCount int, ctx store.IOContext) error
	// Records all files in use by this SegmentCommitInfo
	Files(*SegmentCommitInfo) []string
}
