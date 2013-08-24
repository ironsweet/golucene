package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"log"
	"math"
)

const (
	LUCENE41_DOC_EXTENSION = "doc"
	LUCENE41_POS_EXTENSION = "pos"
	LUCENE41_PAY_EXTENSION = "pay"

	LUCENE41_BLOCK_SIZE = 128

	LUCENE41_TERMS_CODEC = "Lucene41PostingsWriterTerms"
	LUCENE41_DOC_CODEC   = "Lucene41PostingsWriterDoc"
	LUCENE41_POS_CODEC   = "Lucene41PostingsWriterPos"
	LUCENE41_PAY_CODEC   = "Lucene41PostingsWriterPay"

	LUCENE41_VERSION_START   = 0
	LUCENE41_VERSION_CURRENT = LUCENE41_VERSION_START
)

type Lucene41PostingsReader struct {
	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	forUtil ForUtil
}

func NewLucene41PostingsReader(dir store.Directory, fis FieldInfos, si SegmentInfo,
	ctx store.IOContext, segmentSuffix string) (r PostingsReaderBase, err error) {
	log.Print("Initializing Lucene41PostingsReader...")
	success := false
	var docIn, posIn, payIn store.IndexInput = nil, nil, nil
	defer func() {
		if !success {
			log.Print("Failed to initialize Lucene41PostingsReader.")
			if err != nil {
				log.Print("DEBUG ", err)
			}
			util.CloseWhileSuppressingError(docIn, posIn, payIn)
		}
	}()

	docIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_DOC_EXTENSION), ctx)
	if err != nil {
		return r, err
	}
	log.Print("DEBUG docIn: ", docIn)
	_, err = codec.CheckHeader(docIn, LUCENE41_DOC_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
	if err != nil {
		return r, err
	}
	forUtil, err := NewForUtil(docIn)
	if err != nil {
		return r, err
	}

	if fis.hasProx {
		posIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_POS_EXTENSION), ctx)
		if err != nil {
			return r, err
		}
		log.Print("DEBUG posIn: ", posIn)
		_, err = codec.CheckHeader(posIn, LUCENE41_POS_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
		if err != nil {
			return r, err
		}

		if fis.hasPayloads || fis.hasOffsets {
			payIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			if err != nil {
				return r, err
			}
			log.Print("DEBUG payIn: ", payIn)
			_, err = codec.CheckHeader(payIn, LUCENE41_PAY_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
			if err != nil {
				return r, err
			}
		}
	}

	success = true
	return &Lucene41PostingsReader{docIn, posIn, payIn, forUtil}, nil
}

func (r *Lucene41PostingsReader) init(termsIn store.IndexInput) error {
	log.Printf("Initializing from: %v", termsIn)
	// Make sure we are talking to the matching postings writer
	_, err := codec.CheckHeader(termsIn, LUCENE41_TERMS_CODEC, LUCENE41_VERSION_START, LUCENE41_VERSION_CURRENT)
	if err != nil {
		return err
	}
	indexBlockSize, err := termsIn.ReadVInt()
	if err != nil {
		return err
	}
	log.Printf("Index block size: %v", indexBlockSize)
	if indexBlockSize != LUCENE41_BLOCK_SIZE {
		panic(fmt.Sprintf("index-time BLOCK_SIZE (%v) != read-time BLOCK_SIZE (%v)", indexBlockSize, LUCENE41_BLOCK_SIZE))
	}
	return nil
}

func (r *Lucene41PostingsReader) Close() error {
	return util.Close(r.docIn, r.posIn, r.payIn)
}

type ForUtil struct {
	encodedSizes []int32
	encoders     []util.PackedIntsEncoder
	decoders     []util.PackedIntsDecoder
	iterations   []int32
}

type DataInput interface {
	ReadVInt() (n int32, err error)
}

func NewForUtil(in DataInput) (fu ForUtil, err error) {
	self := ForUtil{}
	packedIntsVersion, err := in.ReadVInt()
	if err != nil {
		return self, err
	}
	util.CheckVersion(packedIntsVersion)
	self.encodedSizes = make([]int32, 33)
	self.encoders = make([]util.PackedIntsEncoder, 33)
	self.decoders = make([]util.PackedIntsDecoder, 33)
	self.iterations = make([]int32, 33)

	for bpv := 1; bpv <= 32; bpv++ {
		code, err := in.ReadVInt()
		if err != nil {
			return self, err
		}
		formatId := uint32(code) >> 5
		bitsPerValue := (uint32(code) & 31) + 1

		format := util.PackedFormat(formatId)
		// assert format.isSupported(bitsPerValue)
		self.encodedSizes[bpv] = encodedSize(format, packedIntsVersion, bitsPerValue)
		self.encoders[bpv] = util.GetPackedIntsEncoder(format, packedIntsVersion, bitsPerValue)
		self.decoders[bpv] = util.GetPackedIntsDecoder(format, packedIntsVersion, bitsPerValue)
		self.iterations[bpv] = computeIterations(self.decoders[bpv])
	}
	return self, nil
}

func encodedSize(format util.PackedFormat, packedIntsVersion int32, bitsPerValue uint32) int32 {
	byteCount := format.ByteCount(packedIntsVersion, LUCENE41_BLOCK_SIZE, bitsPerValue)
	// assert byteCount >= 0 && byteCount <= math.MaxInt32()
	return int32(byteCount)
}

func computeIterations(decoder util.PackedIntsDecoder) int32 {
	return int32(math.Ceil(float64(LUCENE41_BLOCK_SIZE) / float64(decoder.ByteValueCount())))
}

type Lucene41StoredFieldsReader struct {
	*CompressingStoredFieldsReader
}

func newLucene41StoredFieldsReader(d store.Directory, si SegmentInfo, fn FieldInfos, ctx store.IOContext) (r StoredFieldsReader, err error) {
	formatName := "Lucene41StoredFields"
	compressionMode := codec.COMPRESSION_MODE_FAST
	// chunkSize := 1 << 14
	p, err := newCompressingStoredFieldsReader(d, si, "", fn, ctx, formatName, compressionMode)
	if err == nil {
		r = &Lucene41StoredFieldsReader{p}
	}
	return r, nil
}

const (
	CODEC_SFX_IDX             = "Index"
	CODEC_SFX_DAT             = "Data"
	CODEC_SFX_VERSION_START   = 0
	CODEC_SFX_VERSION_CURRENT = CODEC_SFX_VERSION_START
)

type CompressingStoredFieldsReader struct {
	fieldInfos        FieldInfos
	indexReader       *CompressingStoredFieldsIndexReader
	fieldsStream      store.IndexInput
	packedIntsVersion int
	compressionMode   codec.CompressionMode
	decompressor      codec.Decompressor
	bytes             []byte
	numDocs           int
	closed            bool
}

// CompressingStoredFieldsReader.java L90
func newCompressingStoredFieldsReader(d store.Directory, si SegmentInfo, segmentSuffix string, fn FieldInfos,
	ctx store.IOContext, formatName string, compressionMode codec.CompressionMode) (r *CompressingStoredFieldsReader, err error) {
	r = &CompressingStoredFieldsReader{}
	r.compressionMode = compressionMode
	segment := si.name
	r.fieldInfos = fn
	r.numDocs = int(si.docCount)

	var indexStream store.IndexInput
	success := false
	defer func() {
		if !success {
			log.Println("Failed to initialize CompressionStoredFieldsReader.")
			if err != nil {
				log.Print(err)
			}
			util.Close(r, indexStream)
		}
	}()

	// Load the index into memory
	indexStreamFN := util.SegmentFileName(segment, segmentSuffix, LUCENE40_SF_FIELDS_INDEX_EXTENSION)
	indexStream, err = d.OpenInput(indexStreamFN, ctx)
	if err != nil {
		return nil, err
	}
	codecNameIdx := formatName + CODEC_SFX_IDX
	codec.CheckHeader(indexStream, codecNameIdx, CODEC_SFX_VERSION_START, CODEC_SFX_VERSION_CURRENT)
	if int64(codec.HeaderLength(codecNameIdx)) != indexStream.FilePointer() {
		panic("assert fail")
	}
	r.indexReader, err = newCompressingStoredFieldsIndexReader(indexStream, si)
	if err != nil {
		return nil, err
	}
	err = indexStream.Close()
	if err != nil {
		return nil, err
	}
	indexStream = nil

	// Open the data file and read metadata
	fieldsStreamFN := util.SegmentFileName(segment, segmentSuffix, LUCENE40_SF_FIELDS_EXTENSION)
	r.fieldsStream, err = d.OpenInput(fieldsStreamFN, ctx)
	if err != nil {
		return nil, err
	}
	codecNameDat := formatName + CODEC_SFX_DAT
	codec.CheckHeader(r.fieldsStream, codecNameDat, CODEC_SFX_VERSION_START, CODEC_SFX_VERSION_CURRENT)
	if int64(codec.HeaderLength(codecNameDat)) != r.fieldsStream.FilePointer() {
		panic("assert fail")
	}

	n, err := r.fieldsStream.ReadVInt()
	if err != nil {
		return nil, err
	}
	r.packedIntsVersion = int(n)
	r.decompressor = compressionMode.NewDecompressor()
	r.bytes = make([]byte, 0)

	success = true
	return r, nil
}

func (r *CompressingStoredFieldsReader) ensureOpen() {
	if r.closed {
		panic("this FieldsReader is closed")
	}
}

func (r *CompressingStoredFieldsReader) Close() (err error) {
	if !r.closed {
		if err = util.Close(r.fieldsStream); err == nil {
			r.closed = true
		}
	}
	return err
}

func (r *CompressingStoredFieldsReader) visitDocument(n int, visitor StoredFieldVisitor) error {
	panic("not implemented yet")
	return nil
}

func (r *CompressingStoredFieldsReader) clone() StoredFieldsReader {
	r.ensureOpen()
	// return CompressingStoredFieldsProducer()
	panic("not implemented yet")
	return nil
}

type CompressingStoredFieldsIndexReader struct {
	maxDoc              int
	docBases            []int
	startPointers       []int64
	avgChunkDocs        []int
	avgChunkSizes       []int64
	docBasesDeltas      []util.PackedIntsReader
	startPointersDeltas []util.PackedIntsReader
}

func newCompressingStoredFieldsIndexReader(fieldsIndexIn store.IndexInput, si SegmentInfo) (r *CompressingStoredFieldsIndexReader, err error) {
	r = &CompressingStoredFieldsIndexReader{}
	r.maxDoc = int(si.docCount)
	r.docBases = make([]int, 0, 16)
	r.startPointers = make([]int64, 0, 16)
	r.avgChunkDocs = make([]int, 0, 16)
	r.avgChunkSizes = make([]int64, 0, 16)
	r.docBasesDeltas = make([]util.PackedIntsReader, 0, 16)
	r.startPointersDeltas = make([]util.PackedIntsReader, 0, 16)

	packedIntsVersion, err := fieldsIndexIn.ReadVInt()
	if err != nil {
		return nil, err
	}

	for blockCount := 0; ; blockCount++ {
		numChunks, err := fieldsIndexIn.ReadVInt()
		if err != nil {
			return nil, err
		}
		if numChunks == 0 {
			break
		}

		{ // doc bases
			n, err := fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			r.docBases = append(r.docBases, int(n))
			n, err = fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			r.avgChunkDocs = append(r.avgChunkDocs, int(n))
			bitsPerDocBase, err := fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			if bitsPerDocBase > 32 {
				return nil, errors.New(fmt.Sprintf("Corrupted bitsPerDocBase (resource=%v)", fieldsIndexIn))
			}
			pr, err := util.NewPackedReaderNoHeader(fieldsIndexIn, util.PACKED, packedIntsVersion, numChunks, uint32(bitsPerDocBase))
			if err != nil {
				return nil, err
			}
			r.docBasesDeltas = append(r.docBasesDeltas, pr)
		}

		{ // start pointers
			n, err := fieldsIndexIn.ReadVLong()
			if err != nil {
				return nil, err
			}
			r.startPointers = append(r.startPointers, n)
			n, err = fieldsIndexIn.ReadVLong()
			if err != nil {
				return nil, err
			}
			r.avgChunkSizes = append(r.avgChunkSizes, n)
			bitsPerStartPointer, err := fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			if bitsPerStartPointer > 64 {
				return nil, errors.New(fmt.Sprintf("Corrupted bitsPerStartPonter (resource=%v)", fieldsIndexIn))
			}
			pr, err := util.NewPackedReaderNoHeader(fieldsIndexIn, util.PACKED, packedIntsVersion, numChunks, uint32(bitsPerStartPointer))
			if err != nil {
				return nil, err
			}
			r.startPointersDeltas = append(r.startPointersDeltas, pr)
		}
	}

	return r, nil
}
