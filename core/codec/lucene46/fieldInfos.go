package lucene46

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// lucene46/Lucene46FieldInfosFormat.java

const (
	/* Extension of field info */
	FI_EXTENSION = "fnm"

	/* Codec header */
	FI_CODEC_NAME            = "Lucene46FieldInfos"
	FI_FORMAT_START          = 0
	FI_FORMAT_CHECKSUM       = 1
	FI_FORMAT_SORTED_NUMERIC = 2
	FI_FORMAT_CURRENT        = FI_FORMAT_SORTED_NUMERIC

	// Field flags
	FI_IS_INDEXED                   = 0x1
	FI_STORE_TERMVECTOR             = 0x2
	FI_STORE_OFFSETS_IN_POSTINGS    = 0x4
	FI_OMIT_NORMS                   = 0x10
	FI_STORE_PAYLOADS               = 0x20
	FI_OMIT_TERM_FREQ_AND_POSITIONS = 0x40
	FI_OMIT_POSITIONS               = 0x80
)

type Lucene46FieldInfosFormat struct {
	r FieldInfosReader
	w FieldInfosWriter
}

func NewLucene46FieldInfosFormat() *Lucene46FieldInfosFormat {
	return &Lucene46FieldInfosFormat{
		r: Lucene46FieldInfosReader,
		w: Lucene46FieldInfosWriter,
	}
}

func (f *Lucene46FieldInfosFormat) FieldInfosReader() FieldInfosReader {
	return f.r
}

func (f *Lucene46FieldInfosFormat) FieldInfosWriter() FieldInfosWriter {
	return f.w
}

var Lucene46FieldInfosReader = func(dir store.Directory,
	segment, suffix string, ctx store.IOContext) (fis FieldInfos, err error) {

	filename := util.SegmentFileName(segment, suffix, FI_EXTENSION)
	var input store.ChecksumIndexInput
	if input, err = dir.OpenChecksumInput(filename, ctx); err != nil {
		return
	}

	var success = false
	defer func() {
		if success {
			err = input.Close()
		} else {
			util.CloseWhileSuppressingError(input)
		}
	}()

	var codecVersion int
	if codecVersion, err = asInt(codec.CheckHeader(input, FI_CODEC_NAME, FI_FORMAT_START, FI_FORMAT_CURRENT)); err != nil {
		return
	}

	var size int
	if size, err = asInt(input.ReadVInt()); err != nil {
		return
	}

	var infos []*FieldInfo
	var name string
	var fieldNumber int32
	var bits, val byte
	var isIndexed, storeTermVector, omitNorms, storePayloads bool
	var indexOptions IndexOptions
	var docValuesType, normsType DocValuesType
	var dvGen int64
	var attributes map[string]string
	for i := 0; i < size; i++ {
		if name, err = input.ReadString(); err != nil {
			return
		}
		if fieldNumber, err = input.ReadVInt(); err != nil {
			return
		}
		assert2(fieldNumber >= 0,
			"invalid field number for field: %v, fieldNumber=%v (resource=%v)",
			name, fieldNumber, input)
		if bits, err = input.ReadByte(); err != nil {
			return
		}
		isIndexed = (bits & FI_IS_INDEXED) != 0
		storeTermVector = (bits & FI_STORE_TERMVECTOR) != 0
		omitNorms = (bits & FI_OMIT_NORMS) != 0
		storePayloads = (bits & FI_STORE_PAYLOADS) != 0
		switch {
		case !isIndexed:
			//
		case (bits & FI_OMIT_TERM_FREQ_AND_POSITIONS) != 0:
			indexOptions = INDEX_OPT_DOCS_ONLY
		case (bits & FI_OMIT_POSITIONS) != 0:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS
		case (bits & FI_STORE_OFFSETS_IN_POSTINGS) != 0:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		default:
			indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
		}

		// DV types are packed in one byte
		if val, err = input.ReadByte(); err != nil {
			return
		}
		if docValuesType, err = getDocValuesType(input, val&0x0F); err != nil {
			return
		}
		if normsType, err = getDocValuesType(input, (val>>4)&0x0F); err != nil {
			return
		}
		if dvGen, err = input.ReadLong(); err != nil {
			return
		}
		if attributes, err = input.ReadStringStringMap(); err != nil {
			return
		}
		infos = append(infos, NewFieldInfo(name, isIndexed, fieldNumber,
			storeTermVector, omitNorms, storePayloads, indexOptions,
			docValuesType, normsType, dvGen, attributes))
	}

	if codecVersion >= FI_FORMAT_CHECKSUM {
		if _, err = codec.CheckFooter(input); err != nil {
			return
		}
	} else {
		if err = codec.CheckEOF(input); err != nil {
			return
		}
	}
	fis = NewFieldInfos(infos)
	success = true
	return fis, nil
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

var Lucene46FieldInfosWriter = func(dir store.Directory,
	segName, suffix string, infos FieldInfos, ctx store.IOContext) (err error) {

	filename := util.SegmentFileName(segName, suffix, FI_EXTENSION)
	var output store.IndexOutput
	if output, err = dir.CreateOutput(filename, ctx); err != nil {
		return
	}

	var success = false
	defer func() {
		if success {
			err = output.Close()
		} else {
			util.CloseWhileSuppressingError(output)
		}
	}()

	if err = codec.WriteHeader(output, FI_CODEC_NAME, FI_FORMAT_CURRENT); err != nil {
		return
	}
	if err = output.WriteVInt(int32(infos.Size())); err != nil {
		return
	}
	for _, fi := range infos.Values {
		indexOptions := fi.IndexOptions()
		bits := byte(0)
		if fi.HasVectors() {
			bits |= FI_STORE_TERMVECTOR
		}
		if fi.OmitsNorms() {
			bits |= FI_OMIT_NORMS
		}
		if fi.HasPayloads() {
			bits |= FI_STORE_PAYLOADS
		}
		if fi.IsIndexed() {
			bits |= FI_IS_INDEXED
			assert(indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS || !fi.HasPayloads())
			switch indexOptions {
			case INDEX_OPT_DOCS_ONLY:
				bits |= FI_OMIT_TERM_FREQ_AND_POSITIONS
			case INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS:
				bits |= FI_STORE_OFFSETS_IN_POSTINGS
			case INDEX_OPT_DOCS_AND_FREQS:
				bits |= FI_OMIT_POSITIONS
			}
		}
		if err = output.WriteString(fi.Name); err == nil {
			if err = output.WriteVInt(fi.Number); err == nil {
				err = output.WriteByte(bits)
			}
		}
		if err != nil {
			return
		}

		// pack the DV types in one byte
		dv := byte(fi.DocValuesType())
		nrm := byte(fi.NormType())
		assert((dv&(^byte(0xF))) == 0 && (nrm&(^byte(0x0F))) == 0)
		val := (0xff & ((nrm << 4) | dv))
		if err = output.WriteByte(val); err == nil {
			if err = output.WriteLong(fi.DocValuesGen()); err == nil {
				err = output.WriteStringStringMap(fi.Attributes())
			}
		}
		if err != nil {
			return
		}
	}
	if err = codec.WriteFooter(output); err != nil {
		return
	}
	success = true
	return nil
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
