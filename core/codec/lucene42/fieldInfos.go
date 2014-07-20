package lucene42

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
)

// lucene42/Lucene42FieldInfosFormat.java

/*
Lucene 4.2 Field Infos format.

Field names are stored in the field info file, with suffix .fnm.

FieldInfos (.fnm) --> Header, HeaderCount, <FieldName, FieldNumber,
                      FieldBits, DocValuesBits, Attribute>^FieldsCount

Data types:
- Header --> CodecHeader
- FieldsCount --> VInt
- FieldName --> string
- FieldBits, DocValuesBit --> byte
- FieldNumber --> VInt
- Attributes --> map[string]string

Field Description:
- FieldsCount: the number of fields in this file.
- FieldName: name of the field as a UTF-8 string.
- FieldNumber: the field's number. NOte that unlike previous versions
  of Lucene, the fields are not numbered implicitly by their order in
  the file, instead explicitly.
- FieldBits: a byte containing field options.
  - The low-order bit is one for indexed fields, and zero for non-indexed
    fields.
  - The second lowest-order bit is one for fields that have term vectors
    stored, and zero for fields without term vectors.
  - If the third lowest order-bit is set (0x4), offsets are stored into
    the postings list in addition to positions.
  - Fourth bit is unsed.
  - If the fifth lowest-order bit is set (0x10), norms are omitted for
    the indexed field.
  - If the sixth lowest-order bit is set (0x20), payloads are stored
    for the indexed field.
  - If the seventh lowest-order bit is set (0x40), term frequencies a
    and ositions omitted for the indexed field.
  - If the eighth lowest-order bit is set (0x80), positions are omitted
    for the indexed field.
- DocValuesBits: a byte containing per-document value types. The type
  recorded as two four-bit intergers, with the high-order bits
  representing norms options, and low-order bits representing DocVlaues
  options. Each four-bit interger can be decoded as such:
  - 0: no DocValues for this field.
  - 1: NumericDocValues.
  - 2: BinaryDocvalues.
  - 3: SortedDocValues.
- Attributes: a key-value map of codec-private attributes.
*/
type Lucene42FieldInfosFormat struct {
	reader FieldInfosReader
	// writer FieldInfosWriter
}

func NewLucene42FieldInfosFormat() *Lucene42FieldInfosFormat {
	return &Lucene42FieldInfosFormat{
		reader: Lucene42FieldInfosReader,
		// writer: Lucene42FieldInfosWriter,
	}
}

func (f *Lucene42FieldInfosFormat) FieldInfosReader() FieldInfosReader {
	return f.reader
}

func (f *Lucene42FieldInfosFormat) FieldInfosWriter() FieldInfosWriter {
	panic("this codec can only be used for reading")
	// return f.writer
}

const (
	// Extension of field infos
	LUCENE42_FI_EXTENSION = "fnm"

	// Codec header
	LUCENE42_FI_CODEC_NAME     = "Lucene42FieldInfos"
	LUCENE42_FI_FORMAT_START   = 0
	LUCENE42_FI_FORMAT_CURRENT = LUCENE42_FI_FORMAT_START

	// Field flags
	LUCENE42_FI_IS_INDEXED                   = 0x1
	LUCENE42_FI_STORE_TERMVECTOR             = 0x2
	LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS    = 0x4
	LUCENE42_FI_OMIT_NORMS                   = 0x10
	LUCENE42_FI_STORE_PAYLOADS               = 0x20
	LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS = 0x40
	LUCENE42_FI_OMIT_POSITIONS               = 0x80
)

var Lucene42FieldInfosReader = func(dir store.Directory,
	segment, suffix string, context store.IOContext) (fi FieldInfos, err error) {

	log.Printf("Reading FieldInfos from %v...", dir)
	fi = FieldInfos{}
	fileName := util.SegmentFileName(segment, "", LUCENE42_FI_EXTENSION)
	log.Printf("Segment: %v", fileName)
	input, err := dir.OpenInput(fileName, context)
	if err != nil {
		return fi, err
	}
	log.Printf("Reading %v", input)

	success := false
	defer func() {
		if success {
			input.Close()
		} else {
			util.CloseWhileHandlingError(err, input)
		}
	}()

	_, err = codec.CheckHeader(input,
		LUCENE42_FI_CODEC_NAME,
		LUCENE42_FI_FORMAT_START,
		LUCENE42_FI_FORMAT_CURRENT)
	if err != nil {
		return fi, err
	}

	size, err := input.ReadVInt() //read in the size
	if err != nil {
		return fi, err
	}
	log.Printf("Found %v FieldInfos.", size)

	infos := make([]*FieldInfo, size)
	for i, _ := range infos {
		name, err := input.ReadString()
		if err != nil {
			return fi, err
		}
		fieldNumber, err := input.ReadVInt()
		if err != nil {
			return fi, err
		}
		bits, err := input.ReadByte()
		if err != nil {
			return fi, err
		}
		isIndexed := (bits & LUCENE42_FI_IS_INDEXED) != 0
		storeTermVector := (bits & LUCENE42_FI_STORE_TERMVECTOR) != 0
		omitNorms := (bits & LUCENE42_FI_OMIT_NORMS) != 0
		storePayloads := (bits & LUCENE42_FI_STORE_PAYLOADS) != 0
		var indexOptions IndexOptions
		switch {
		case !isIndexed:
			indexOptions = IndexOptions(0)
		case (bits & LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS) != 0:
			indexOptions = INDEX_OPT_DOCS_ONLY
		case (bits & LUCENE42_FI_OMIT_POSITIONS) != 0:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS
		case (bits & LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS) != 0:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		default:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
		}

		// DV Types are packed in one byte
		val, err := input.ReadByte()
		if err != nil {
			return fi, err
		}
		docValuesType, err := getDocValuesType(input, (byte)(val&0x0F))
		if err != nil {
			return fi, err
		}
		normsType, err := getDocValuesType(input, (byte)((uint8(val)>>4)&0x0F))
		if err != nil {
			return fi, err
		}
		attributes, err := input.ReadStringStringMap()
		if err != nil {
			return fi, err
		}
		infos[i] = NewFieldInfo(name, isIndexed, fieldNumber, storeTermVector,
			omitNorms, storePayloads, indexOptions, docValuesType, normsType, -1, attributes)
	}

	if err = codec.CheckEOF(input); err != nil {
		return fi, err
	}
	fi = NewFieldInfos(infos)
	success = true
	return fi, nil
}

func getDocValuesType(input store.IndexInput, b byte) (t DocValuesType, err error) {
	switch b {
	case 0:
		return DocValuesType(0), nil
	case 1:
		return DOC_VALUES_TYPE_NUMERIC, nil
	case 2:
		return DOC_VALUES_TYPE_BINARY, nil
	case 3:
		return DOC_VALUES_TYPE_SORTED, nil
	case 4:
		return DOC_VALUES_TYPE_SORTED_SET, nil
	default:
		return DocValuesType(0), errors.New(
			fmt.Sprintf("invalid docvalues byte: %v (resource=%v)", b, input))
	}
}

// lucene42/Lucene42FieldInfosWriter.java
// var Lucene42FieldInfosWriter = func(dir store.Directory,
// 	segName string, infos FieldInfos, ctx store.IOContext) (err error) {

// 	fileName := util.SegmentFileName(segName, "", LUCENE42_FI_EXTENSION)
// 	var output store.IndexOutput
// 	output, err = dir.CreateOutput(fileName, ctx)
// 	if err != nil {
// 		return err
// 	}

// 	var success = false
// 	defer func() {
// 		if success {
// 			err = mergeError(err, output.Close())
// 		} else {
// 			util.CloseWhileSuppressingError(output)
// 		}
// 	}()

// 	err = codec.WriteHeader(output, LUCENE42_FI_CODEC_NAME, LUCENE42_FI_FORMAT_CURRENT)
// 	if err != nil {
// 		return err
// 	}
// 	err = output.WriteVInt(int32(len(infos.Values)))
// 	if err != nil {
// 		return err
// 	}
// 	for _, fi := range infos.Values {
// 		indexOptions := fi.IndexOptions()
// 		bits := byte(0x0)
// 		if fi.HasVectors() {
// 			bits |= LUCENE42_FI_STORE_TERMVECTOR
// 		}
// 		if fi.OmitsNorms() {
// 			bits |= LUCENE42_FI_OMIT_NORMS
// 		}
// 		if fi.HasPayloads() {
// 			bits |= LUCENE42_FI_STORE_PAYLOADS
// 		}
// 		if fi.IsIndexed() {
// 			bits |= LUCENE42_FI_IS_INDEXED
// 			assert(int(indexOptions) >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS) || !fi.HasPayloads())
// 			switch indexOptions {
// 			case INDEX_OPT_DOCS_ONLY:
// 				bits |= LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS
// 			case INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS:
// 				bits |= LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS
// 			case INDEX_OPT_DOCS_AND_FREQS:
// 				bits |= LUCENE42_FI_OMIT_POSITIONS
// 			}
// 		}
// 		err = output.WriteString(fi.Name)
// 		if err != nil {
// 			return err
// 		}
// 		err = output.WriteVInt(fi.Number)
// 		if err != nil {
// 			return err
// 		}
// 		err = output.WriteByte(bits)
// 		if err != nil {
// 			return err
// 		}

// 		// pack the DV types in one byte
// 		dv := docValuesByte(fi.DocValuesType())
// 		nrm := docValuesByte(fi.NormType())
// 		assert((int(dv)&(^0xF)) == 0 && (int(nrm)&(^0x0F)) == 0)
// 		val := byte(0xFF & ((nrm << 4) | dv))
// 		err = output.WriteByte(val)
// 		if err != nil {
// 			return err
// 		}
// 		err = output.WriteStringStringMap(fi.Attributes())
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	success = true
// 	return nil
// }

// func docValuesByte(typ DocValuesType) byte {
// 	n := byte(typ)
// 	assert(n >= 0 && n <= 4)
// 	return n
// }
