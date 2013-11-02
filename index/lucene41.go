package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"log"
	"reflect"
)

// Lucene41PostingsReader.java

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

/*
Concrete class that reads docId (maybe frq,pos,offset,payload) list
with postings format.
*/
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
		_, err = codec.CheckHeader(posIn, LUCENE41_POS_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
		if err != nil {
			return r, err
		}

		if fis.hasPayloads || fis.hasOffsets {
			payIn, err = dir.OpenInput(util.SegmentFileName(si.name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			if err != nil {
				return r, err
			}
			_, err = codec.CheckHeader(payIn, LUCENE41_PAY_CODEC, LUCENE41_VERSION_CURRENT, LUCENE41_VERSION_CURRENT)
			if err != nil {
				return r, err
			}
		}
	}

	success = true
	return &Lucene41PostingsReader{docIn, posIn, payIn, forUtil}, nil
}

func (r *Lucene41PostingsReader) Init(termsIn store.IndexInput) error {
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

func (r *Lucene41PostingsReader) NewTermState() *BlockTermState {
	return newIntBlockTermState().BlockTermState
}

func (r *Lucene41PostingsReader) Close() error {
	return util.Close(r.docIn, r.posIn, r.payIn)
}

/* Reads but does not decode the byte[] blob holding
   metadata for the current terms block */
func (r *Lucene41PostingsReader) ReadTermsBlock(termsIn store.IndexInput, fieldInfo FieldInfo, _termState *BlockTermState) (err error) {
	termState := _termState.Self.(*intBlockTermState)
	numBytes, err := asInt(termsIn.ReadVInt())
	if err != nil {
		return err
	}

	if termState.bytes == nil {
		// TODO over-allocate
		termState.bytes = make([]byte, numBytes)
		termState.bytesReader = store.NewEmptyByteArrayDataInput()
	} else if len(termState.bytes) < numBytes {
		// TODO over-allocate
		termState.bytes = make([]byte, numBytes)
	}

	err = termsIn.ReadBytes(termState.bytes)
	if err != nil {
		return err
	}
	termState.bytesReader.Reset(termState.bytes)
	return nil
}

func (r *Lucene41PostingsReader) nextTerm(fieldInfo FieldInfo, _termState *BlockTermState) (err error) {
	termState := _termState.Self.(*intBlockTermState)
	isFirstTerm := termState.termBlockOrd == 0
	fieldHasPositions := fieldInfo.indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
	fieldHasOffsets := fieldInfo.indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	fieldHasPayloads := fieldInfo.storePayloads

	in := termState.bytesReader
	if isFirstTerm {
		if termState.docFreq == 1 {
			termState.singletonDocID, err = asInt(in.ReadVInt())
			if err != nil {
				return err
			}
			termState.docStartFP = 0
		} else {
			termState.singletonDocID = -1
			termState.docStartFP, err = in.ReadVLong()
			if err != nil {
				return err
			}
		}
		if fieldHasPositions {
			termState.posStartFP, err = in.ReadVLong()
			if err != nil {
				return err
			}
			if termState.totalTermFreq > LUCENE41_BLOCK_SIZE {
				termState.lastPosBlockOffset, err = in.ReadVLong()
				if err != nil {
					return err
				}
			} else {
				termState.lastPosBlockOffset = -1
			}
			if (fieldHasPayloads || fieldHasOffsets) && termState.totalTermFreq >= LUCENE41_BLOCK_SIZE {
				termState.payStartFP, err = in.ReadVLong()
				if err != nil {
					return err
				}
			} else {
				termState.payStartFP = -1
			}
		}
	} else {
		if termState.docFreq == 1 {
			termState.singletonDocID, err = asInt(in.ReadVInt())
			if err != nil {
				return err
			}
		} else {
			termState.singletonDocID = -1
			delta, err := in.ReadVLong()
			if err != nil {
				return err
			}
			termState.docStartFP += delta
		}
		if fieldHasPositions {
			delta, err := in.ReadVLong()
			if err != nil {
				return err
			}
			termState.posStartFP += delta
			if termState.totalTermFreq > LUCENE41_BLOCK_SIZE {
				termState.lastPosBlockOffset, err = in.ReadVLong()
				if err != nil {
					return err
				}
			} else {
				termState.lastPosBlockOffset = -1
			}
			if (fieldHasPayloads || fieldHasOffsets) && termState.totalTermFreq >= LUCENE41_BLOCK_SIZE {
				delta, err = in.ReadVLong()
				if err != nil {
					return err
				}
				if termState.payStartFP == -1 {
					termState.payStartFP = delta
				} else {
					termState.payStartFP += delta
				}
			}
		}
	}

	if termState.docFreq > LUCENE41_BLOCK_SIZE {
		termState.skipOffset, err = in.ReadVLong()
		if err != nil {
			return err
		}
	} else {
		termState.skipOffset = -1
	}

	return nil
}

type intBlockTermState struct {
	*BlockTermState
	docStartFP         int64
	posStartFP         int64
	payStartFP         int64
	skipOffset         int64
	lastPosBlockOffset int64
	// docid when there is a single pulsed posting, otherwise -1
	// freq is always implicitly totalTermFreq in this case.
	singletonDocID int

	// Only used by the "primary" TermState -- clones don't
	// copy this (basically they are "transient"):
	bytesReader *store.ByteArrayDataInput
	bytes       []byte
}

func newIntBlockTermState() *intBlockTermState {
	ts := &intBlockTermState{}
	parent := NewBlockTermState()
	ts.BlockTermState, parent.Self = parent, ts
	return ts
}

func (ts *intBlockTermState) Clone() TermState {
	clone := newIntBlockTermState()
	clone.CopyFrom(ts)
	return clone
}

func (ts *intBlockTermState) CopyFrom(other TermState) {
	assert(other != nil)
	if ots, ok := other.(*intBlockTermState); ok {
		ts.BlockTermState.CopyFrom(ots.BlockTermState)
		ts.docStartFP = ots.docStartFP
		ts.posStartFP = ots.posStartFP
		ts.payStartFP = ots.payStartFP
		ts.lastPosBlockOffset = ots.lastPosBlockOffset
		ts.skipOffset = ots.skipOffset
		ts.singletonDocID = ots.singletonDocID

		// Do not copy bytes, bytesReader (else TermState is
		// very heavy, ie drags around the entire block's
		// byte[]).  On seek back, if next() is in fact used
		// (rare!), they will be re-read from disk.
	} else {
		panic(fmt.Sprintf("Can not copy from %v", reflect.TypeOf(other).Name()))
	}
}

func (ts *intBlockTermState) String() string {
	return fmt.Sprintf("%v docStartFP=%v posStartFP=%v payStartFP=%v lastPosBlockOffset=%v skipOffset=%v singletonDocID=%v",
		ts.BlockTermState, ts.docStartFP, ts.posStartFP, ts.payStartFP, ts.lastPosBlockOffset, ts.skipOffset, ts.singletonDocID)
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
