package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/codec/compressing"
	// "github.com/balzaczyy/golucene/core/codec/lucene40"
	"github.com/balzaczyy/golucene/core/codec/lucene41"
	// docu "github.com/balzaczyy/golucene/core/document"
	. "github.com/balzaczyy/golucene/core/index/model"
	. "github.com/balzaczyy/golucene/core/search/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"log"
	"reflect"
)

// codecs/lucene41/Lucene41PostingsFormat.java

type Lucene41PostingsFormat struct {
	minTermBlockSize int
	maxTermBlockSize int
}

/* Creates Lucene41PostingsFormat wit hdefault settings. */
func newLucene41PostingsFormat() *Lucene41PostingsFormat {
	return newLucene41PostingsFormatWith(DEFAULT_MIN_BLOCK_SIZE, DEFAULT_MAX_BLOCK_SIZE)
}

/*
Creates Lucene41PostingsFormat with custom values for minBlockSize
and maxBlockSize passed to block terms directory.
*/
func newLucene41PostingsFormatWith(minTermBlockSize, maxTermBlockSize int) *Lucene41PostingsFormat {
	assert(minTermBlockSize > 1)
	assert(minTermBlockSize <= maxTermBlockSize)
	return &Lucene41PostingsFormat{
		minTermBlockSize: minTermBlockSize,
		maxTermBlockSize: maxTermBlockSize,
	}
}

func (f *Lucene41PostingsFormat) Name() string {
	return "Lucene41"
}

func (f *Lucene41PostingsFormat) String() {
	panic("not implemented yet")
}

func (f *Lucene41PostingsFormat) FieldsConsumer(state *SegmentWriteState) (FieldsConsumer, error) {
	postingsWriter, err := newLucene41PostingsWriterCompact(state)
	if err != nil {
		return nil, err
	}
	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(postingsWriter)
		}
	}()
	ret, err := NewBlockTreeTermsWriter(state, postingsWriter, f.minTermBlockSize, f.maxTermBlockSize)
	if err != nil {
		return nil, err
	}
	success = true
	return ret, nil
}

func (f *Lucene41PostingsFormat) FieldsProducer(state SegmentReadState) (FieldsProducer, error) {
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
			log.Printf("Failed to load FieldsProducer for %v.", f.Name())
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

// Lucene41PostingsWriter.java

/*
Expert: the maximum number of skip levels. Smaller values result in
slightly smaller indexes, but slower skipping in big posting lists.
*/

const maxSkipLevels = 10

/*
Concrete class that writes docId (maybe frq,pos,offset,payloads) list
with postings format.

Postings list for each term will be stored separately.
*/
type Lucene41PostingsWriter struct {
	docOut store.IndexOutput
	posOut store.IndexOutput
	payOut store.IndexOutput

	lastState *intBlockTermState

	fieldHasFreqs     bool
	fieldHasPositions bool
	fieldHasOffsets   bool
	fieldHasPayloads  bool

	// Holds starting file pointers for current term:
	docStartFP int64
	posStartFP int64
	payStartFP int64

	docDeltaBuffer []int
	freqBuffer     []int
	docBufferUpto  int

	posDeltaBuffer         []int
	payloadLengthBuffer    []int
	offsetStartDeltaBuffer []int
	offsetLengthBuffer     []int
	posBufferUpto          int

	payloadBytes    []byte
	payloadByteUpto int

	lastBlockDocId           int
	lastBlockPosFP           int64
	lastBlockPayFP           int64
	lastBlockPosBufferUpto   int
	lastBlockPayloadByteUpto int

	lastDocId       int
	lastPosition    int
	lastStartOffset int
	docCount        int

	encoded []byte

	forUtil    ForUtil
	skipWriter *lucene41.SkipWriter
}

/* Creates a postings writer with the specified PackedInts overhead ratio */
func newLucene41PostingsWriter(state *SegmentWriteState,
	accetableOverheadRatio float32) (*Lucene41PostingsWriter, error) {
	docOut, err := state.Directory.CreateOutput(
		util.SegmentFileName(state.SegmentInfo.Name,
			state.SegmentSuffix,
			LUCENE41_DOC_EXTENSION),
		state.Context)
	if err != nil {
		return nil, err
	}

	ans := new(Lucene41PostingsWriter)
	if err = func() error {
		var posOut store.IndexOutput
		var payOut store.IndexOutput
		var success = false
		defer func() {
			if !success {
				util.CloseWhileSuppressingError(docOut, posOut, payOut)
			}
		}()

		err := codec.WriteHeader(docOut, LUCENE41_DOC_CODEC, LUCENE41_VERSION_CURRENT)
		if err != nil {
			return err
		}
		ans.forUtil, err = NewForUtilInto(accetableOverheadRatio, docOut)
		if err != nil {
			return err
		}
		if state.FieldInfos.HasProx {
			ans.posDeltaBuffer = make([]int, MAX_DATA_SIZE)
			posOut, err = state.Directory.CreateOutput(util.SegmentFileName(
				state.SegmentInfo.Name, state.SegmentSuffix, LUCENE41_POS_EXTENSION),
				state.Context)
			if err != nil {
				return err
			}

			err = codec.WriteHeader(posOut, LUCENE41_POS_CODEC, LUCENE41_VERSION_CURRENT)
			if err != nil {
				return err
			}

			if state.FieldInfos.HasPayloads {
				ans.payloadBytes = make([]byte, 128)
				ans.payloadLengthBuffer = make([]int, MAX_DATA_SIZE)
			}

			if state.FieldInfos.HasOffsets {
				ans.offsetStartDeltaBuffer = make([]int, MAX_DATA_SIZE)
				ans.offsetLengthBuffer = make([]int, MAX_DATA_SIZE)
			}

			if state.FieldInfos.HasPayloads || state.FieldInfos.HasOffsets {
				payOut, err = state.Directory.CreateOutput(util.SegmentFileName(
					state.SegmentInfo.Name, state.SegmentSuffix, LUCENE41_PAY_EXTENSION),
					state.Context)
				if err != nil {
					return err
				}
				err = codec.WriteHeader(payOut, LUCENE41_PAY_CODEC, LUCENE41_VERSION_CURRENT)
			}
		}
		ans.payOut, ans.posOut = payOut, posOut
		ans.docOut = docOut
		success = true
		return nil
	}(); err != nil {
		return nil, err
	}

	ans.docDeltaBuffer = make([]int, MAX_DATA_SIZE)
	ans.freqBuffer = make([]int, MAX_DATA_SIZE)
	ans.encoded = make([]byte, MAX_ENCODED_SIZE)

	// TODO: should we try skipping every 2/4 blocks...?
	ans.skipWriter = lucene41.NewSkipWriter(
		maxSkipLevels,
		LUCENE41_BLOCK_SIZE,
		state.SegmentInfo.DocCount(),
		ans.docOut,
		ans.posOut,
		ans.payOut)

	return ans, nil
}

/* Creates a postings writer with PackedInts.COMPACT */
func newLucene41PostingsWriterCompact(state *SegmentWriteState) (*Lucene41PostingsWriter, error) {
	return newLucene41PostingsWriter(state, packed.PackedInts.COMPACT)
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
}

func newIntBlockTermState() *intBlockTermState {
	ts := &intBlockTermState{
		skipOffset:         -1,
		lastPosBlockOffset: -1,
		singletonDocID:     -1,
	}
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
		ts.BlockTermState.internalCopyFrom(ots.BlockTermState)
		ts.docStartFP = ots.docStartFP
		ts.posStartFP = ots.posStartFP
		ts.payStartFP = ots.payStartFP
		ts.lastPosBlockOffset = ots.lastPosBlockOffset
		ts.skipOffset = ots.skipOffset
		ts.singletonDocID = ots.singletonDocID
	} else if ots, ok := other.(*BlockTermState); ok && ots.Self != nil {
		// try copy from other's sub class
		ts.CopyFrom(ots.Self)
	} else {
		panic(fmt.Sprintf("Can not copy from %v", reflect.TypeOf(other).Name()))
	}
}

func (ts *intBlockTermState) String() string {
	return fmt.Sprintf("%v docStartFP=%v posStartFP=%v payStartFP=%v lastPosBlockOffset=%v skipOffset=%v singletonDocID=%v",
		ts.BlockTermState.toString(), ts.docStartFP, ts.posStartFP, ts.payStartFP, ts.lastPosBlockOffset, ts.skipOffset, ts.singletonDocID)
}

func (w *Lucene41PostingsWriter) init(termsOut store.IndexOutput) error {
	err := codec.WriteHeader(termsOut, LUCENE41_TERMS_CODEC, LUCENE41_VERSION_CURRENT)
	if err == nil {
		err = termsOut.WriteVInt(LUCENE41_BLOCK_SIZE)
	}
	return err
}

func (w *Lucene41PostingsWriter) SetField(fieldInfo *FieldInfo) int {
	n := int(fieldInfo.IndexOptions())
	w.fieldHasFreqs = n >= int(INDEX_OPT_DOCS_AND_FREQS)
	w.fieldHasPositions = n >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS)
	w.fieldHasOffsets = n >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)
	w.fieldHasPayloads = fieldInfo.HasPayloads()
	w.skipWriter.SetField(w.fieldHasPositions, w.fieldHasOffsets, w.fieldHasPayloads)
	panic("not implemented yet")
}

func (w *Lucene41PostingsWriter) StartTerm() error {
	w.docStartFP = w.docOut.FilePointer()
	if w.fieldHasPositions {
		w.posStartFP = w.posOut.FilePointer()
		if w.fieldHasPayloads || w.fieldHasOffsets {
			w.payStartFP = w.payOut.FilePointer()
		}
	}
	w.lastDocId = 0
	w.lastBlockDocId = -1
	w.skipWriter.ResetSkip()
	return nil
}

func (w *Lucene41PostingsWriter) StartDoc(docId, termDocFreq int) error {
	// Have collected a block of docs, and get a new doc. Should write
	// skip data as well as postings list for current block.
	if w.lastBlockDocId != -1 && w.docBufferUpto == 0 {
		if err := w.skipWriter.BufferSkip(w.lastBlockDocId, w.docCount,
			w.lastBlockPosFP, w.lastBlockPayFP, w.lastBlockPosBufferUpto,
			w.lastBlockPayloadByteUpto); err != nil {
			return err
		}
	}

	docDelta := docId - w.lastDocId
	if docId < 0 || (w.docCount > 0 && docDelta <= 0) {
		return errors.New(fmt.Sprintf(
			"docs out of order (%v <= %v) (docOut : %v)",
			docId, w.lastDocId, w.docOut))
	}
	w.docDeltaBuffer[w.docBufferUpto] = docDelta
	if w.fieldHasFreqs {
		w.freqBuffer[w.docBufferUpto] = termDocFreq
	}
	w.docBufferUpto++
	w.docCount++

	if w.docBufferUpto == LUCENE41_BLOCK_SIZE {
		panic("not implemented yet")
	}

	w.lastDocId = docId
	w.lastPosition = 0
	w.lastStartOffset = 0
	return nil
}

/* Add a new opsition & payload */
func (w *Lucene41PostingsWriter) AddPosition(position int, payload []byte, startOffset, endOffset int) error {
	w.posDeltaBuffer[w.posBufferUpto] = position - w.lastPosition
	if w.fieldHasPayloads {
		if len(payload) == 0 {
			// no paylaod
			w.payloadLengthBuffer[w.posBufferUpto] = 0
		} else {
			panic("not implemented yet")
		}
	}

	if w.fieldHasOffsets {
		panic("not implemented yet")
	}

	w.posBufferUpto++
	w.lastPosition = position
	if w.posBufferUpto == LUCENE41_BLOCK_SIZE {
		panic("not implemented yet")
	}
	return nil
}

func (w *Lucene41PostingsWriter) FinishDoc() error {
	// since we don't know df for current term, we had to buffer those
	// skip data for each block, and when a new doc comes, write them
	// to skip file.
	if w.docBufferUpto == LUCENE41_BLOCK_SIZE {
		w.lastBlockDocId = w.lastDocId
		if w.posOut != nil {
			if w.payOut != nil {
				w.lastBlockPayFP = w.payOut.FilePointer()
			}
			w.lastBlockPosFP = w.posOut.FilePointer()
			w.lastBlockPosBufferUpto = w.posBufferUpto
			w.lastBlockPayloadByteUpto = w.payloadByteUpto
		}
		w.docBufferUpto = 0
	}
	return nil
}

/* Called when we are done adding docs to this term */
func (w *Lucene41PostingsWriter) FinishTerm(_state *BlockTermState) error {
	state := _state.Self.(*intBlockTermState)
	assert(state.docFreq > 0)

	// TODO: wasteful we are couting this (counting # docs for this term) in two places?
	assert2(state.docFreq == w.docCount, "%v vs %v", state.docFreq, w.docCount)

	// docFreq == 1, don't write the single docId/freq to a separate
	// file along with a pointer to it.
	var singletonDocId int
	if state.docFreq == 1 {
		// pulse the singleton docId into the term dictionary, freq is implicitly totalTermFreq
		singletonDocId = w.docDeltaBuffer[0]
	} else {
		panic("not implemented yet")
	}

	var lastPosBlockOffset int64
	if w.fieldHasPositions {
		// totalTermFreq is just total number of positions (or payloads,
		// or offsets) associated with current term.
		assert(state.totalTermFreq != -1)
		if state.totalTermFreq > LUCENE41_BLOCK_SIZE {
			// record file offset for last pos in last block
			lastPosBlockOffset = w.posOut.FilePointer() - w.posStartFP
		} else {
			lastPosBlockOffset = -1
		}
		if w.posBufferUpto > 0 {
			// TODO: should we send offsets/payloads to .pay...? seems
			// wasteful (have to store extra vlong for low (< BLOCK_SIZE)
			// DF terms = vast vast majority)

			// vInt encode the remaining positions/payloads/offsets:
			// lastPayloadLength := -1 // force first payload length to be written
			// lastOffsetLength := -1  // force first offset length to be written
			payloadBytesReadUpto := 0
			for i := 0; i < w.posBufferUpto; i++ {
				posDelta := w.posDeltaBuffer[i]
				if w.fieldHasPayloads {
					panic("not implemented yet")
				} else {
					err := w.posOut.WriteVInt(int32(posDelta))
					if err != nil {
						return err
					}
				}

				if w.fieldHasOffsets {
					panic("not implemented yet")
				}
			}

			if w.fieldHasPayloads {
				assert(payloadBytesReadUpto == w.payloadByteUpto)
				w.payloadByteUpto = 0
			}
		}
	} else {
		lastPosBlockOffset = -1
	}

	var skipOffset int64
	if w.docCount > LUCENE41_BLOCK_SIZE {
		n, err := w.skipWriter.WriteSkip(w.docOut)
		if err != nil {
			return err
		}
		skipOffset = n - w.docStartFP
	} else {
		skipOffset = -1
	}

	state.docStartFP = w.docStartFP
	state.posStartFP = w.posStartFP
	state.payStartFP = w.payStartFP
	state.singletonDocID = singletonDocId
	state.skipOffset = skipOffset
	state.lastPosBlockOffset = lastPosBlockOffset
	w.docBufferUpto = 0
	w.posBufferUpto = 0
	w.lastDocId = 0
	w.docCount = 0
	return nil
}

func (w *Lucene41PostingsWriter) encodeTerm(longs []int64,
	out DataOutput, fieldInfo *FieldInfo, _state *BlockTermState,
	absolute bool) error {
	panic("not implemented yet")
	// func (w *Lucene41PostingsWriter) flushTermsBlock(start, count int) error {
	// if count == 0 {
	// 	return w.termsOut.WriteByte(0)
	// }

	// assert(start <= len(w.pendingTerms))
	// assert(count <= start)

	// limit := len(w.pendingTerms) - start + count

	// lastDocStartFP := int64(0)
	// lastPosStartFP := int64(0)
	// lastPayStartFP := int64(0)
	// for _, term := range w.pendingTerms[limit-count : limit] {
	// 	if term.singletonDocId == -1 {
	// 		err := w.bytesWriter.WriteVLong(term.docStartFP - lastDocStartFP)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		lastDocStartFP = term.docStartFP
	// 	} else {
	// 		err := w.bytesWriter.WriteVInt(int32(term.singletonDocId))
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	if w.fieldHasPositions {
	// 		err := w.bytesWriter.WriteVLong(term.posStartFP - lastPosStartFP)
	// 		if err == nil {
	// 			lastPosStartFP = term.posStartFP
	// 			if term.lastPosBlockOffset != -1 {
	// 				err = w.bytesWriter.WriteVLong(term.lastPosBlockOffset)
	// 			}
	// 		}
	// 		if err == nil && (w.fieldHasPayloads || w.fieldHasOffsets) && term.payStartFP != -1 {
	// 			err = w.bytesWriter.WriteVLong(term.payStartFP - lastPayStartFP)
	// 			if err == nil {
	// 				lastPayStartFP = term.payStartFP
	// 			}
	// 		}
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	if term.skipOffset != -1 {
	// 		err := w.bytesWriter.WriteVLong(term.skipOffset)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// err := w.termsOut.WriteVInt(int32(w.bytesWriter.FilePointer()))
	// if err == nil {
	// 	err = w.bytesWriter.WriteTo(w.termsOut)
	// 	if err == nil {
	// 		w.bytesWriter.Reset()
	// 	}
	// }
	// if err != nil {
	// 	return err
	// }

	// // remove the terms we just wrote:
	// w.pendingTerms = append(w.pendingTerms[:limit-count], w.pendingTerms[limit:]...)
	// return nil
}

func (w *Lucene41PostingsWriter) Close() error {
	panic("not implemented yet")
	// return util.Close(w.docOut, w.posOut, w.payOut)
}

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

	LUCENE41_VERSION_START         = 0
	LUCENE41_VERSION_META_ARRAY    = 1
	LUCENE41_VERSION_META_CHECKSUM = 2
	LUCENE41_VERSION_CURRENT       = LUCENE41_VERSION_META_CHECKSUM
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
	version int
}

func NewLucene41PostingsReader(dir store.Directory,
	fis FieldInfos, si *SegmentInfo,
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

	docIn, err = dir.OpenInput(util.SegmentFileName(si.Name, segmentSuffix, LUCENE41_DOC_EXTENSION), ctx)
	if err != nil {
		return nil, err
	}
	var version int32
	version, err = codec.CheckHeader(docIn, LUCENE41_DOC_CODEC, LUCENE41_VERSION_START, LUCENE41_VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	forUtil, err := NewForUtilFrom(docIn)
	if err != nil {
		return nil, err
	}

	if fis.HasProx {
		posIn, err = dir.OpenInput(util.SegmentFileName(si.Name, segmentSuffix, LUCENE41_POS_EXTENSION), ctx)
		if err != nil {
			return nil, err
		}
		_, err = codec.CheckHeader(posIn, LUCENE41_POS_CODEC, version, version)
		if err != nil {
			return nil, err
		}

		if fis.HasPayloads || fis.HasOffsets {
			payIn, err = dir.OpenInput(util.SegmentFileName(si.Name, segmentSuffix, LUCENE41_PAY_EXTENSION), ctx)
			if err != nil {
				return nil, err
			}
			_, err = codec.CheckHeader(payIn, LUCENE41_PAY_CODEC, version, version)
			if err != nil {
				return nil, err
			}
		}
	}

	success = true
	return &Lucene41PostingsReader{docIn, posIn, payIn, forUtil, int(version)}, nil
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

/**
 * Read values that have been written using variable-length encoding instead of bit-packing.
 */
func readVIntBlock(docIn store.IndexInput, docBuffer []int,
	freqBuffer []int, num int, indexHasFreq bool) (err error) {
	if indexHasFreq {
		for i := 0; i < num; i++ {
			code, err := asInt(docIn.ReadVInt())
			if err != nil {
				return err
			}
			docBuffer[i] = int(uint(code) >> 1)
			if (code & 1) != 0 {
				freqBuffer[i] = 1
			} else {
				freqBuffer[i], err = asInt(docIn.ReadVInt())
				if err != nil {
					return err
				}
			}
		}
	} else {
		for i := 0; i < num; i++ {
			docBuffer[i], err = asInt(docIn.ReadVInt())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Lucene41PostingsReader) NewTermState() *BlockTermState {
	return newIntBlockTermState().BlockTermState
}

func (r *Lucene41PostingsReader) Close() error {
	return util.Close(r.docIn, r.posIn, r.payIn)
}

func (r *Lucene41PostingsReader) decodeTerm(longs []int64,
	in util.DataInput, fieldInfo *FieldInfo,
	_termState *BlockTermState, absolute bool) error {
	panic("not implemented yet")
}

func (r *Lucene41PostingsReader) docs(fieldInfo *FieldInfo,
	termState *BlockTermState, liveDocs util.Bits,
	reuse DocsEnum, flags int) (de DocsEnum, err error) {

	var docsEnum *blockDocsEnum
	if v, ok := reuse.(*blockDocsEnum); ok {
		docsEnum = v
		if !docsEnum.canReuse(r.docIn, fieldInfo) {
			docsEnum = newBlockDocsEnum(r, fieldInfo)
		}
	} else {
		docsEnum = newBlockDocsEnum(r, fieldInfo)
	}
	return docsEnum.reset(liveDocs, termState.Self.(*intBlockTermState), flags)
}

type blockDocsEnum struct {
	*Lucene41PostingsReader // embedded struct

	encoded []byte

	docDeltaBuffer []int
	freqBuffer     []int

	docBufferUpto int

	// skipper Lucene41SkipReader
	skipped bool

	startDocIn store.IndexInput

	docIn            store.IndexInput
	indexHasFreq     bool
	indexHasPos      bool
	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int
	totalTermFreq int64
	docUpto       int
	doc           int
	accum         int
	freq          int

	// Where this term's postings start in the .doc file:
	docTermStartFP int64

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int64

	// docID for next skip point, we won't use skipper if
	// target docID is not larger than this
	nextSkipDoc int

	liveDocs util.Bits

	needsFreq      bool
	singletonDocID int
}

func newBlockDocsEnum(owner *Lucene41PostingsReader,
	fieldInfo *FieldInfo) *blockDocsEnum {

	return &blockDocsEnum{
		Lucene41PostingsReader: owner,
		docDeltaBuffer:         make([]int, MAX_DATA_SIZE),
		freqBuffer:             make([]int, MAX_DATA_SIZE),
		startDocIn:             owner.docIn,
		docIn:                  nil,
		indexHasFreq:           fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS,
		indexHasPos:            fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS,
		indexHasOffsets:        fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS,
		indexHasPayloads:       fieldInfo.HasPayloads(),
		encoded:                make([]byte, MAX_ENCODED_SIZE),
	}
}

func (de *blockDocsEnum) canReuse(docIn store.IndexInput, fieldInfo *FieldInfo) bool {
	return docIn == de.startDocIn &&
		de.indexHasFreq == (fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS) &&
		de.indexHasPos == (fieldInfo.IndexOptions() >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS) &&
		de.indexHasPayloads == fieldInfo.HasPayloads()
}

func (de *blockDocsEnum) reset(liveDocs util.Bits, termState *intBlockTermState, flags int) (ret DocsEnum, err error) {
	de.liveDocs = liveDocs
	log.Printf("  FPR.reset: termState=%v", termState)
	de.docFreq = termState.docFreq
	if de.indexHasFreq {
		de.totalTermFreq = termState.totalTermFreq
	} else {
		de.totalTermFreq = int64(de.docFreq)
	}
	de.docTermStartFP = termState.docStartFP // <---- docTermStartFP should be 178 instead of 0
	de.skipOffset = termState.skipOffset
	de.singletonDocID = termState.singletonDocID
	if de.docFreq > 1 {
		if de.docIn == nil {
			// lazy init
			de.docIn = de.startDocIn.Clone()
		}
		err = de.docIn.Seek(de.docTermStartFP)
		if err != nil {
			return nil, err
		}
	}

	de.doc = -1
	de.needsFreq = (flags & DOCS_ENUM_FLAG_FREQS) != 0
	if !de.indexHasFreq {
		for i, _ := range de.freqBuffer {
			de.freqBuffer[i] = 1
		}
	}
	de.accum = 0
	de.docUpto = 0
	de.nextSkipDoc = LUCENE41_BLOCK_SIZE - 1 // we won't skip if target is found in first block
	de.docBufferUpto = LUCENE41_BLOCK_SIZE
	de.skipped = false
	return de, nil
}

func (de *blockDocsEnum) Freq() (n int, err error) {
	return de.freq, nil
}

func (de *blockDocsEnum) DocId() int {
	return de.doc
}

func (de *blockDocsEnum) refillDocs() (err error) {
	left := de.docFreq - de.docUpto
	assert(left > 0)

	if left >= LUCENE41_BLOCK_SIZE {
		log.Printf("    fill doc block from fp=%v", de.docIn.FilePointer())
		panic("not implemented yet")
	} else if de.docFreq == 1 {
		de.docDeltaBuffer[0] = de.singletonDocID
		de.freqBuffer[0] = int(de.totalTermFreq)
	} else {
		// Read vInts:
		log.Printf("    fill last vInt block from fp=%v", de.docIn.FilePointer())
		err = readVIntBlock(de.docIn, de.docDeltaBuffer, de.freqBuffer, left, de.indexHasFreq)
	}
	de.docBufferUpto = 0
	return
}

func (de *blockDocsEnum) NextDoc() (n int, err error) {
	log.Println("FPR.nextDoc")
	for {
		log.Printf("  docUpto=%v (of df=%v) docBufferUpto=%v", de.docUpto, de.docFreq, de.docBufferUpto)

		if de.docUpto == de.docFreq {
			log.Println("  return doc=END")
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}

		if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
			err = de.refillDocs()
			if err != nil {
				return 0, err
			}
		}

		log.Printf("    accum=%v docDeltaBuffer[%v]=%v", de.accum, de.docBufferUpto, de.docDeltaBuffer[de.docBufferUpto])
		de.accum += de.docDeltaBuffer[de.docBufferUpto]
		de.docUpto++

		if de.liveDocs == nil || de.liveDocs.At(de.accum) {
			de.doc = de.accum
			de.freq = de.freqBuffer[de.docBufferUpto]
			de.docBufferUpto++
			log.Printf("  return doc=%v freq=%v", de.doc, de.freq)
			return de.doc, nil
		}
		log.Printf("  doc=%v is deleted; try next doc", de.accum)
		de.docBufferUpto++
	}
}

func (de *blockDocsEnum) Advance(target int) (int, error) {
	// TODO: make frq block load lazy/skippable
	fmt.Printf("  FPR.advance target=%v\n", target)

	// current skip docID < docIDs generated from current buffer <= next
	// skip docID, we don't need to skip if target is buffered already
	if de.docFreq > LUCENE41_BLOCK_SIZE && target > de.nextSkipDoc {
		fmt.Println("load skipper")

		panic("not implemented yet")
	}
	if de.docUpto == de.docFreq {
		de.doc = NO_MORE_DOCS
		return de.doc, nil
	}
	if de.docBufferUpto == LUCENE41_BLOCK_SIZE {
		err := de.refillDocs()
		if err != nil {
			return 0, nil
		}
	}

	// Now scan.. this is an inlined/pared down version of nextDoc():
	for {
		fmt.Printf("  scan doc=%v docBufferUpto=%v\n", de.accum, de.docBufferUpto)
		de.accum += de.docDeltaBuffer[de.docBufferUpto]
		de.docUpto++

		if de.accum >= target {
			break
		}
		de.docBufferUpto++
		if de.docUpto == de.docFreq {
			de.doc = NO_MORE_DOCS
			return de.doc, nil
		}
	}

	if de.liveDocs == nil || de.liveDocs.At(de.accum) {
		fmt.Printf("  return doc=%v\n", de.accum)
		de.freq = de.freqBuffer[de.docBufferUpto]
		de.docBufferUpto++
		de.doc = de.accum
		return de.doc, nil
	} else {
		fmt.Println("  now do nextDoc()")
		de.docBufferUpto++
		return de.NextDoc()
	}
}

// lucene41/Lucene41StoredFieldsFormat.java

/*
Lucene 4.1 stored fields format.

Principle

This StoredFieldsFormat compresses blocks of 16KB of documents in
order to improve the compression ratio compared to document-level
compression. It uses the LZ4 compression algorithm, which is fast to
compress and very fast to decompress dta. Although the compression
method that is used focuses more on speed than on compression ratio,
it should provide interesting compression ratios for redundant inputs
(such as log files, HTML or plain text).

File formats

Stored fields are represented by two files:

1. field_data

A fields data file (extension .fdt). This file stores a compact
representation of documents in compressed blocks of 16KB or more.
When writing a segment, documents are appended to an in-memory []byte
buffer. When its size reaches 16KB or more, some metadata about the
documents is flushed to disk, immediately followed by a compressed
representation of the buffer using the [LZ4](http://codec.google.com/p/lz4/)
[compression format](http://fastcompression.blogspot.ru/2011/05/lz4-explained.html)

Here is a more detailed description of the field data fiel format:

- FieldData (.dft) --> <Header>, packedIntsVersion, <Chunk>^ChunkCount
- Header --> CodecHeader
- PackedIntsVersion --> PackedInts.VERSION_CURRENT as a VInt
- ChunkCount is not known in advance and is the number of chunks
nucessary to store all document of the segment
- Chunk --> DocBase, ChunkDocs, DocFieldCounts, DocLengths, <CompressedDoc>
- DocBase --> the ID of the first document of the chunk as a VInt
- ChunkDocs --> the number of the documents in the chunk as a VInt
- DocFieldCounts --> the number of stored fields or every document
in the chunk,  encoded as followed:
  - if hunkDocs=1, the unique value is encoded as a VInt
  - else read VInt (let's call it bitsRequired)
    - if bitsRequired is 0 then all values are equal, and the common
    value is the following VInt
    - else bitsRequired is the number of bits required to store any
    value, and values are stored in a packed array where every value
    is stored on exactly bitsRequired bits
- DocLenghts --> the lengths of all documents in the chunk, encodedwith the same method as DocFieldCounts
- CompressedDocs --> a compressed representation of <Docs> using
the LZ4 compression format
- Docs --> <Doc>^ChunkDocs
- Doc --> <FieldNumAndType, Value>^DocFieldCount
- FieldNumAndType --> a VLong, whose 3 last bits are Type and other
bits are FieldNum
- Type -->
  - 0: Value is string
  - 1: Value is BinaryValue
  - 2: Value is int
  - 3: Value is float32
  - 4: Value is int64
  - 5: Value is float64
  - 6, 7: unused
- FieldNum --> an ID of the field
- Value --> string | BinaryValue | int | float32 | int64 | float64
dpending on Type
- BinaryValue --> ValueLength <Byte>&ValueLength

Notes

- If documents are larger than 16KB then chunks will likely contain
only one document. However, documents can never spread across several
chunks (all fields of a single document are in the same chunk).
- When at least one document in a chunk is large enough so that the
chunk is larger than 32KB, then chunk will actually be compressed in
several LZ4 blocks of 16KB. This allows StoredFieldsVisitors which
are only interested in the first fields of a document to not have to
decompress 10MB of data if the document is 10MB, but only 16KB.
- Given that the original lengths are written in the metadata of the
chunk, the decompressorcan leverage this information to stop decoding
as soon as enough data has been decompressed.
- In case documents are incompressible, CompressedDocs will be less
than 0.5% larger than Docs.

2. field_index

A fields index file (extension .fdx).

- FieldsIndex (.fdx) --> <Header>, <ChunkINdex>
- Header --> CodecHeader
- ChunkIndex: See CompressingStoredFieldsInexWriter

Known limitations

This StoredFieldsFormat does not support individual documents larger
than (2^32 - 2^14) bytes. In case this is a problem, you should use
another format, such as Lucene40StoredFieldsFormat.
*/
type Lucene41StoredFieldsFormat struct {
	*compressing.CompressingStoredFieldsFormat
}

func newLucene41StoredFieldsFormat() *Lucene41StoredFieldsFormat {
	return &Lucene41StoredFieldsFormat{
		compressing.NewCompressingStoredFieldsFormat("Lucene41StoredFields", "", compressing.COMPRESSION_MODE_FAST, 1<<14),
	}
}
