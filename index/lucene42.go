package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
)

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
	LUCENE42_FI_OMIT_POSITIONS               = -128
)

var (
	Lucene42FieldInfosReader = func(dir *store.Directory, segment string, context store.IOContext) (fi FieldInfos, err error) {
		fi = FieldInfos{}
		fileName := SegmentFileName(segment, "", LUCENE42_FI_EXTENSION)
		input, err := dir.OpenInput(fileName, iocontext)
		if err != nil {
			return fi, err
		}

		success := false
		defer func() {
			if success {
				input.Close()
			} else {
				util.CloseWhileHandlingError(input)
			}
		}()

		_, err = CheckHeader(input, LUCENE42_FI_CODEC_NAME,
			LUCENE42_FI_FORMAT_START,
			LUCENE42_FI_FORMAT_CURRENT)
		if err != nil {
			return fi, err
		}

		size, err := input.readVInt() //read in the size
		if err != nil {
			return fi, err
		}

		infos = make([]FieldInfo, size)
		for i, _ := range infos {
			name, err := input.readString()
			if err != nil {
				return fi, err
			}
			fieldNumber, err := input.readVInt()
			if err != nil {
				return fi, err
			}
			bits = input.readByte()
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
				indexOptions = null
			case (bits & LUCENE42_FI_OMIT_TERM_FREQ_AND_POSITIONS) != 0:
				indexOptions = INDEX_OPT_DOCS_ONLY
			case (bits & LUCENE42_FI_OMIT_POSITIONS) != 0:
				indexOptions = IINDEX_OPT_DOCS_AND_FREQS
			case (bits & LUCENE42_FI_STORE_OFFSETS_IN_POSTINGS) != 0:
				indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
			default:
				indexOptions = INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
			}

			// DV Types are packed in one byte
			val, err = input.readByte()
			if err != nil {
				return fi, err
			}
			docValuesType = getDocValuesType(input, (byte)(val&0x0F))
			normsType = getDocValuesType(input, (byte)((uint8(val)>>4)&0x0F))
			attributes, err := input.readStringStringMap()
			if err != nil {
				return fi, err
			}
			infos[i] = NewFieldInfo(name, isIndexed, fieldNumber, storeTermVector,
				omitNorms, storePayloads, indexOptions, docValuesType, normsType, attributes)
		}

		if input.FilePointer() != input.Length() {
			return fi, errors.New(fmt.Sprintf(
				"did not read all bytes from file '%v': read %v vs size %v (resource: %v)",
				fileName, input.FilePointer(), input.Length(), input))
		}
		fi = NewFieldInfos(infos)
		success = true
		return fi, nil
	}
)

func getDocValuesType(input *store.IndexInput, b byte) (t DocValuesType, err error) {
	switch b {
	case 0:
		return DOC_VALUES_TYPE_NUMERIC, nil
	case 1:
		return DOC_VALUES_TYPE_BINARY, nil
	case 2:
		return DOC_VALUES_TYPE_SORTED, nil
	case 3:
		return DOC_VALUES_TYPE_SORTED_SET, nil
	default:
		return DocValuesType(0), errors.New(
			fmt.Sprintf("invalid docvalues byte: %v (resource=%v)", b, input))
	}
}

type Codec struct {
	ReadSegmentInfo   func(dir *store.Directory, segment string, context store.IOContext) (si SegmentInfo, err error)
	ReadFieldInfos    func(dir *store.Directory, segment string, context store.IOContext) (fi FieldInfos, err error)
	GetFieldsProducer func(readState SegmentReadState) (fp FieldsProducer, err error)
}

func LoadFieldsProducer(name string, state SegmentReadState) FieldsProducer {
	switch name {
	case "Lucene41":
		postingsReader := NewLucene41PostingReader(state.dir,
			state.fieldInfos,
			state.segmentInfo,
			state.context,
			state.segmentSuffix)
		success := false
		defer func() {
			if !success {
				util.CloseWhileSupressingError(postingsReader)
			}
		}()

		ret := NewBlockTreeTermsReader(state.dir,
			state.fieldInfos,
			state.segmentInfo,
			postingsReader,
			state.context,
			state.segmentSuffix,
			state.termsIndexDivisor)
		success = true
		return ret
	}
	panic(fmt.Sprintf("Service '%v' not found.", name))
}

const (
	PER_FIELD_FORMAT_KEY = "PerFieldPostingsFormat.format"
	PER_FIELD_SUFFIX_KEY = "PerFieldPostingsFormat.suffix"
)

func NewLucene42Codec() Codec {
	fields := make(map[string]FieldsProducer)
	formats := make(map[string]FieldsProducer)
	return Codec{ReadSegmentInfo: Lucene40SegmentInfoReader,
		ReadFieldInfos: Lucene42FieldInfosReader,
		GetFieldsProducer: func(readState SegmentReadState) (fp FieldsProducer, err error) {
			// Read _X.per and init each format:
			success := false
			defer func() {
				if !success {
					fps := make([]FieldsProducer, 0)
					for _, v := range formats {
						fps = append(fps, v)
					}
					util.CloseWhileSuppressingError(fps)
				}
			}()
			// Read field name -> format name
			for _, fi := range readState.fieldInfos {
				if fi.indexed {
					fieldName := fi.name
					formatName = fi.attributes[PER_FIELD_FORMAT_KEY]
					if formatName != nil {
						// null formatName means the field is in fieldInfos, but has no postings!
						suffix := fi.attributes[PER_FIELD_SUFFIX_KEY]
						// assert suffix != nil
						fp = LoadFieldsProducer(formatName, readState)
						segmentSuffix := formatName + "_" + suffix
						if _, ok := formats[segmentSuffix]; !ok {
							formats[segmentSuffix] = format
						}
						fields[fieldName] = formats[segmentSuffix]
					}
				}
			}
			success = true
		}}
}

type FieldsProducer struct {
	*Fields
	close func() error
}

func (fp *FieldsProducer) Close() error {
	return fp.close()
}

type PostingsReaderBase interface {
	io.Closer
	// init(termsIn IndexInput)
	// newTermState() BlockTermState
	// nextTerm(fieldInfo FieldInfo, state BlockTermState)
	// docs(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits, reuse DocsEnum, flags int)
	// docsAndPositions(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits)
}
