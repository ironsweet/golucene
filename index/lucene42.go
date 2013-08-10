package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
	"log"
	"sync"
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
	LUCENE42_FI_OMIT_POSITIONS               = 0x80
)

var (
	Lucene42FieldInfosReader = func(dir store.Directory, segment string, context store.IOContext) (fi FieldInfos, err error) {
		log.Printf("Reading FieldInfos from %v...", dir)
		fi = FieldInfos{}
		fileName := util.SegmentFileName(segment, "", LUCENE42_FI_EXTENSION)
		input, err := dir.OpenInput(fileName, context)
		if err != nil {
			return fi, err
		}
		log.Print("DEBUG ", input)

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

		infos := make([]FieldInfo, size)
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

func getDocValuesType(input store.IndexInput, b byte) (t DocValuesType, err error) {
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
	ReadSegmentInfo           func(d store.Directory, segment string, ctx store.IOContext) (si SegmentInfo, err error)
	ReadFieldInfos            func(d store.Directory, segment string, ctx store.IOContext) (fi FieldInfos, err error)
	GetFieldsProducer         func(s SegmentReadState) (r FieldsProducer, err error)
	GetDocValuesProducer      func(s SegmentReadState) (r DocValuesProducer, err error)
	GetNormsDocValuesProducer func(s SegmentReadState) (r DocValuesProducer, err error)
	GetStoredFieldsReader     func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error)
	GetTermVectorsReader      func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error)
}

func LoadFieldsProducer(name string, state SegmentReadState) (fp FieldsProducer, err error) {
	switch name {
	case "Lucene41":
		postingsReader, err := NewLucene41PostingReader(state.dir,
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
	}
	panic(fmt.Sprintf("Service '%v' not found.", name))
}

const (
	LUCENE42_DV_DATA_CODEC         = "Lucene42DocValuesData"
	LUCENE42_DV_DATA_EXTENSION     = "dvd"
	LUCENE42_DV_METADATA_CODEC     = "Lucene42DocValuesMetadata"
	LUCENE42_DV_METADATA_EXTENSION = "dvm"

	LUCENE42_DV_VERSION_START           = 0
	LUCENE42_DV_VERSION_GCD_COMPRESSION = 1
	LUCENE42_DV_VERSION_CURRENT         = LUCENE42_DV_VERSION_GCD_COMPRESSION

	LUCENE42_DV_NUMBER = 0
	LUCENE42_DV_BYTES  = 1
	LUCENE42_DV_FST    = 2
)

type Lucene42DocValuesProducer struct {
	lock             sync.Mutex
	numerics         map[int]NumericEntry
	binaries         map[int]BinaryEntry
	fsts             map[int]FSTEntry
	data             store.IndexInput
	numericInstances map[int]NumericDocValues
	maxDoc           int
}

func newLucene42DocValuesProducer(state SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (dvp *Lucene42DocValuesProducer, err error) {
	dvp = &Lucene42DocValuesProducer{}
	dvp.maxDoc = int(state.segmentInfo.docCount)
	metaName := util.SegmentFileName(state.segmentInfo.name, state.segmentSuffix, metaExtension)
	// read in the entries from the metadata file.
	in, err := state.dir.OpenInput(metaName, state.context)
	if err != nil {
		return dvp, err
	}
	success := false
	defer func() {
		if success {
			err = util.Close(in)
		} else {
			util.CloseWhileSuppressingError(in)
		}
	}()

	version, err := codec.CheckHeader(in, metaCodec, LUCENE42_DV_VERSION_START, LUCENE42_DV_VERSION_CURRENT)
	if err != nil {
		return dvp, err
	}
	dvp.numerics = make(map[int]NumericEntry)
	dvp.binaries = make(map[int]BinaryEntry)
	dvp.fsts = make(map[int]FSTEntry)
	err = dvp.readFields(in, state.fieldInfos)
	if err != nil {
		return dvp, err
	}
	success = true

	success = false
	dataName := util.SegmentFileName(state.segmentInfo.name, state.segmentSuffix, dataExtension)
	dvp.data, err = state.dir.OpenInput(dataName, state.context)
	if err != nil {
		return dvp, err
	}
	version2, err := codec.CheckHeader(dvp.data, dataCodec, LUCENE42_DV_VERSION_START, LUCENE42_DV_VERSION_CURRENT)
	if err != nil {
		return dvp, err
	}

	if version != version2 {
		return dvp, errors.New("Format versions mismatch")
	}
	return dvp, nil
}

func (dvp *Lucene42DocValuesProducer) readFields(meta store.IndexInput, infos FieldInfos) (err error) {
	fieldNumber, err := meta.ReadVInt()
	for fieldNumber != -1 && err == nil {
		fieldType, err := meta.ReadByte()
		if err != nil {
			break
		}
		switch fieldType {
		case LUCENE42_DV_NUMBER:
			panic("not implemented yet")
		case LUCENE42_DV_BYTES:
			panic("not implemented yet")
		case LUCENE42_DV_FST:
			panic("not implemented yet")
		default:
			return errors.New(fmt.Sprintf("invalid entry type: %v, input=%v", fieldType, meta))
		}
		fieldNumber, err = meta.ReadVInt()
	}
	return err
}

func (dvp *Lucene42DocValuesProducer) Numeric(field FieldInfo) (v NumericDocValues, err error) {
	dvp.lock.Lock()
	defer dvp.lock.Unlock()

	if v, ok := dvp.numericInstances[int(field.number)]; ok {
		return v, nil
	}
	if v, err = dvp.loadNumeric(field); err == nil {
		dvp.numericInstances[int(field.number)] = v
	}
	return v, err
}

func (dvp *Lucene42DocValuesProducer) loadNumeric(field FieldInfo) (v NumericDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Binary(field FieldInfo) (v BinaryDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Sorted(field FieldInfo) (v SortedDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) SortedSet(field FieldInfo) (v SortedSetDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Close() error {
	return dvp.data.Close()
}

type NumericEntry struct {
	offset            int64
	format            byte
	packedIntsVersion int
}

type BinaryEntry struct {
	offset            int64
	numBytes          int64
	minLength         int
	maxLength         int
	packedIntsVersion int
	blockSize         int
}

type FSTEntry struct {
	offset  int64
	numOrds int64
}

const (
	PER_FIELD_FORMAT_KEY = "PerFieldPostingsFormat.format"
	PER_FIELD_SUFFIX_KEY = "PerFieldPostingsFormat.suffix"
)

func NewLucene42Codec() Codec {
	return Codec{ReadSegmentInfo: Lucene40SegmentInfoReader,
		ReadFieldInfos: Lucene42FieldInfosReader,
		GetFieldsProducer: func(readState SegmentReadState) (fp FieldsProducer, err error) {
			return newPerFieldPostingsReader(readState)
		},
		GetDocValuesProducer: func(s SegmentReadState) (dvp DocValuesProducer, err error) {
			return newPerFieldDocValuesReader(s)
		},
		GetNormsDocValuesProducer: func(s SegmentReadState) (dvp DocValuesProducer, err error) {
			return newLucene42DocValuesProducer(s, "Lucene41NormsData", "nvd", "Lucene41NormsMetadata", "nvm")
		},
		GetStoredFieldsReader: func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error) {
			return newLucene41StoredFieldsReader(d, si, fn, ctx)
		},
		GetTermVectorsReader: func(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error) {
			return newLucene42TermVectorsReader(d, si, fn, ctx)
		},
	}
}

type PerFieldPostingsReader struct {
	fields  map[string]FieldsProducer
	formats map[string]FieldsProducer
}

func newPerFieldPostingsReader(state SegmentReadState) (fp FieldsProducer, err error) {
	ans := PerFieldPostingsReader{
		make(map[string]FieldsProducer),
		make(map[string]FieldsProducer),
	}
	// Read _X.per and init each format:
	success := false
	defer func() {
		if !success {
			log.Printf("Failed to initialize PerFieldPostingsReader.")
			fps := make([]FieldsProducer, 0)
			for _, v := range ans.formats {
				fps = append(fps, v)
			}
			items := make([]io.Closer, len(fps))
			for i, v := range fps {
				items[i] = v
			}
			util.CloseWhileSuppressingError(items...)
		}
	}()
	// Read field name -> format name
	for _, fi := range state.fieldInfos.values {
		log.Printf("Processing %v...", fi)
		if fi.indexed {
			fieldName := fi.name
			log.Printf("Name: %v", fieldName)
			if formatName, ok := fi.attributes[PER_FIELD_FORMAT_KEY]; ok {
				log.Printf("Format: %v", formatName)
				// null formatName means the field is in fieldInfos, but has no postings!
				suffix := fi.attributes[PER_FIELD_SUFFIX_KEY]
				log.Printf("Suffix: %v", suffix)
				// assert suffix != nil
				segmentSuffix := formatName + "_" + suffix
				log.Printf("Segment suffix: %v", segmentSuffix)
				if _, ok := ans.formats[segmentSuffix]; !ok {
					log.Printf("Loading fields producer: %v", segmentSuffix)
					newReadState := state // clone
					newReadState.segmentSuffix = formatName + "_" + suffix
					fp, err = LoadFieldsProducer(formatName, newReadState)
					if err != nil {
						return fp, err
					}
					ans.formats[segmentSuffix] = fp
				}
				ans.fields[fieldName] = ans.formats[segmentSuffix]
			}
		}
	}
	success = true
	return &ans, nil
}

func (r *PerFieldPostingsReader) Terms(field string) Terms {
	if p, ok := r.fields[field]; ok {
		return p.Terms(field)
	}
	return nil
}

func (r *PerFieldPostingsReader) Close() error {
	fps := make([]FieldsProducer, 0)
	for _, v := range r.formats {
		fps = append(fps, v)
	}
	items := make([]io.Closer, len(fps))
	for i, v := range fps {
		items[i] = v
	}
	return util.Close(items...)
}

type PerFieldDocValuesReader struct {
	fields  map[string]DocValuesProducer
	formats map[string]DocValuesProducer
}

func newPerFieldDocValuesReader(state SegmentReadState) (dvp DocValuesProducer, err error) {
	ans := PerFieldDocValuesReader{
		make(map[string]DocValuesProducer), make(map[string]DocValuesProducer)}
	// Read _X.per and init each format:
	success := false
	defer func() {
		if !success {
			fps := make([]DocValuesProducer, 0)
			for _, v := range ans.formats {
				fps = append(fps, v)
			}
			items := make([]io.Closer, len(fps))
			for i, v := range fps {
				items[i] = v
			}
			util.CloseWhileSuppressingError(items...)
		}
	}()
	// Read field name -> format name
	for _, fi := range state.fieldInfos.values {
		if fi.docValueType != 0 {
			fieldName := fi.name
			if formatName, ok := fi.attributes[PER_FIELD_FORMAT_KEY]; ok {
				// null formatName means the field is in fieldInfos, but has no docvalues!
				suffix := fi.attributes[PER_FIELD_SUFFIX_KEY]
				// assert suffix != nil
				segmentSuffix := formatName + "_" + suffix
				if _, ok := ans.formats[segmentSuffix]; !ok {
					newReadState := state // clone
					newReadState.segmentSuffix = formatName + "_" + suffix
					if p, err := LoadDocValuesProducer(formatName, newReadState); err == nil {
						ans.formats[segmentSuffix] = p
					}
				}
				ans.fields[fieldName] = ans.formats[segmentSuffix]
			}
		}
	}
	success = true
	return &ans, nil
}

func (dvp *PerFieldDocValuesReader) Numeric(field FieldInfo) (v NumericDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.Numeric(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Binary(field FieldInfo) (v BinaryDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.Binary(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Sorted(field FieldInfo) (v SortedDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.Sorted(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) SortedSet(field FieldInfo) (v SortedSetDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.SortedSet(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Close() error {
	fps := make([]DocValuesProducer, 0)
	for _, v := range dvp.formats {
		fps = append(fps, v)
	}
	items := make([]io.Closer, len(fps))
	for i, v := range fps {
		items[i] = v
	}
	return util.Close(items...)
}

type Lucene42TermVectorsReader struct {
	*CompressingTermVectorsReader
}

func newLucene42TermVectorsReader(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r TermVectorsReader, err error) {
	formatName := "Lucene41StoredFields"
	compressionMode := codec.COMPRESSION_MODE_FAST
	// chunkSize := 1 << 12
	p, err := newCompressingTermVectorsReader(d, si, "", fn, ctx, formatName, compressionMode)
	if err == nil {
		r = &Lucene42TermVectorsReader{p}
	}
	return r, nil
}

type CompressingTermVectorsReader struct {
	vectorsStream store.IndexInput
	closed        bool
}

func newCompressingTermVectorsReader(d store.Directory, si SegmentInfo, segmentSuffix string, fn FieldInfos,
	ctx store.IOContext, formatName string, compressionMode codec.CompressionMode) (r *CompressingTermVectorsReader, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (r *CompressingTermVectorsReader) Close() (err error) {
	if !r.closed {
		err = util.Close(r.vectorsStream)
		r.closed = true
	}
	return err
}

func (r *CompressingTermVectorsReader) get(doc int) Fields {
	panic("not implemented yet")
}

func (r *CompressingTermVectorsReader) clone() TermVectorsReader {
	panic("not implemented yet")
}
