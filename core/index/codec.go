package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"log"
)

// codec/Codec.java

/*
Encodes/decodes an inverted index segment.

Note, when extending this class, the name is written into the index.
In order for the segment to be read, the name must resolve to your
implementation via forName(). This method use hard-coded map to
resolve codec names.

If you implement your own codec, make sure that it is included so
SPI can load it.
*/
type Codec struct {
	name                      string
	ReadSegmentInfo           func(d store.Directory, segment string, ctx store.IOContext) (si SegmentInfo, err error)
	ReadFieldInfos            func(d store.Directory, segment string, ctx store.IOContext) (fi FieldInfos, err error)
	GetFieldsProducer         func(s SegmentReadState) (r FieldsProducer, err error)
	GetDocValuesProducer      func(s SegmentReadState) (r DocValuesProducer, err error)
	GetNormsDocValuesProducer func(s SegmentReadState) (r DocValuesProducer, err error)
	GetStoredFieldsReader     func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error)
	GetTermVectorsReader      func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error)
}

func (codec *Codec) Name() string {
	return codec.name
}

// looks up a codec by name
func LoadCodec(name string) *Codec {
	panic("not implemented yet")
}

// returns a list of all available codec names
func AvailableCodecs() []string {
	panic("not implemented yet")
}

// Expert: returns the default codec used for newly created IndexWriterConfig(s).
var DefaultCodec = func() *Codec { return LoadCodec("Lucene45") }

func LoadFieldsProducer(name string, state SegmentReadState) (fp FieldsProducer, err error) {
	switch name {
	case "Lucene41":
		postingsReader, err := NewLucene41PostingsReader(state.dir,
			state.fieldInfos,
			state.segmentInfo,
			state.context,
			state.segmentSuffix)
		if err != nil {
			return nil, err
		}
		success := false
		defer func() {
			if !success {
				log.Printf("Failed to load FieldsProducer for %v.", name)
				if err != nil {
					log.Print("DEBUG ", err)
				}
				util.CloseWhileSuppressingError(postingsReader)
			}
		}()

		fp, err := newBlockTreeTermsReader(state.dir,
			state.fieldInfos,
			state.segmentInfo,
			postingsReader,
			state.context,
			state.segmentSuffix,
			state.termsIndexDivisor)
		if err != nil {
			log.Print("DEBUG: ", err)
			return fp, err
		}
		success = true
		return fp, nil
	}
	panic(fmt.Sprintf("Service '%v' not found.", name))
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

// codec/StoredFieldsReader.java

type StoredFieldsReader interface {
	io.Closer
	visitDocument(n int, visitor StoredFieldVisitor) error
	Clone() StoredFieldsReader
}

// codec/TermVectorsReader.java

type TermVectorsReader interface {
	io.Closer
	get(doc int) Fields
	clone() TermVectorsReader
}
