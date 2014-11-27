package lucene41

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"reflect"
)

// Lucene41PostingsWriter.java

/*
Expert: the maximum number of skip levels. Smaller values result in
slightly smaller indexes, but slower skipping in big posting lists.
*/
const maxSkipLevels = 10

const (
	LUCENE41_TERMS_CODEC = "Lucene41PostingsWriterTerms"
	LUCENE41_DOC_CODEC   = "Lucene41PostingsWriterDoc"
	LUCENE41_POS_CODEC   = "Lucene41PostingsWriterPos"
	LUCENE41_PAY_CODEC   = "Lucene41PostingsWriterPay"

	LUCENE41_VERSION_START      = 0
	LUCENE41_VERSION_META_ARRAY = 1
	LUCENE41_VERSION_CHECKSUM   = 2
	LUCENE41_VERSION_CURRENT    = LUCENE41_VERSION_CHECKSUM
)

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

	forUtil    *ForUtil
	skipWriter *SkipWriter
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
	ans.skipWriter = NewSkipWriter(
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

var emptyState = newIntBlockTermState()

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
		ts.BlockTermState.CopyFrom_(ots.BlockTermState)
		ts.docStartFP = ots.docStartFP
		ts.posStartFP = ots.posStartFP
		ts.payStartFP = ots.payStartFP
		ts.lastPosBlockOffset = ots.lastPosBlockOffset
		ts.skipOffset = ots.skipOffset
		ts.singletonDocID = ots.singletonDocID
	} else {
		panic(fmt.Sprintf("Can not copy from %v", reflect.TypeOf(other).Name()))
	}
}

func (ts *intBlockTermState) String() string {
	return fmt.Sprintf("%v docStartFP=%v posStartFP=%v payStartFP=%v lastPosBlockOffset=%v skipOffset=%v singletonDocID=%v",
		ts.BlockTermState, ts.docStartFP, ts.posStartFP, ts.payStartFP, ts.lastPosBlockOffset, ts.skipOffset, ts.singletonDocID)
}

func (w *Lucene41PostingsWriter) NewTermState() *BlockTermState {
	return newIntBlockTermState().BlockTermState
}

func (w *Lucene41PostingsWriter) Init(termsOut store.IndexOutput) error {
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
	w.lastState = emptyState
	if w.fieldHasPositions {
		if w.fieldHasPayloads || w.fieldHasOffsets {
			return 3 // doc + pos + pay FP
		} else {
			return 2 // doc + pos FP
		}
	} else {
		return 1 // doc FP
	}
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
		if err := w.forUtil.writeBlock(w.docDeltaBuffer, w.encoded, w.docOut); err != nil {
			return err
		}
		if w.fieldHasFreqs {
			if err := w.forUtil.writeBlock(w.freqBuffer, w.encoded, w.docOut); err != nil {
				return err
			}
		}
		// NOTE: don't set docBufferUpto back to 0 here; finishDoc will
		// do so (because it needs to see that the block was filled so it
		// can save skip data)
	}

	w.lastDocId = docId
	w.lastPosition = 0
	w.lastStartOffset = 0
	return nil
}

/* Add a new opsition & payload */
func (w *Lucene41PostingsWriter) AddPosition(position int,
	payload []byte, startOffset, endOffset int) error {

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
		var err error
		if err = w.forUtil.writeBlock(w.posDeltaBuffer, w.encoded, w.posOut); err != nil {
			return err
		}

		if w.fieldHasPayloads {
			panic("niy")
		}
		if w.fieldHasOffsets {
			panic("niy")
		}
		w.posBufferUpto = 0
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
	assert(state.DocFreq > 0)

	// TODO: wasteful we are couting this (counting # docs for this term) in two places?
	assert2(state.DocFreq == w.docCount, "%v vs %v", state.DocFreq, w.docCount)

	// docFreq == 1, don't write the single docId/freq to a separate
	// file along with a pointer to it.
	var singletonDocId int
	if state.DocFreq == 1 {
		// pulse the singleton docId into the term dictionary, freq is implicitly totalTermFreq
		singletonDocId = w.docDeltaBuffer[0]
	} else {
		singletonDocId = -1
		// vInt encode the remaining doc dealtas and freqs;
		var err error
		for i := 0; i < w.docBufferUpto; i++ {
			docDelta := w.docDeltaBuffer[i]
			freq := w.freqBuffer[i]
			if !w.fieldHasFreqs {
				if err = w.docOut.WriteVInt(int32(docDelta)); err != nil {
					return err
				}
			} else if w.freqBuffer[i] == 1 {
				if err = w.docOut.WriteVInt(int32((docDelta << 1) | 1)); err != nil {
					return err
				}
			} else {
				if err = w.docOut.WriteVInt(int32(docDelta << 1)); err != nil {
					return err
				}
				if err = w.docOut.WriteVInt(int32(freq)); err != nil {
					return err
				}
			}
		}
	}

	var lastPosBlockOffset int64
	if w.fieldHasPositions {
		// totalTermFreq is just total number of positions (or payloads,
		// or offsets) associated with current term.
		assert(state.TotalTermFreq != -1)
		if state.TotalTermFreq > LUCENE41_BLOCK_SIZE {
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

func (w *Lucene41PostingsWriter) EncodeTerm(longs []int64,
	out util.DataOutput, fieldInfo *FieldInfo, _state *BlockTermState,
	absolute bool) (err error) {

	assert(longs != nil)
	assert(len(longs) > 0)
	state := _state.Self.(*intBlockTermState)
	if absolute {
		w.lastState = emptyState
	}
	longs[0] = state.docStartFP - w.lastState.docStartFP
	if w.fieldHasPositions {
		longs[1] = state.posStartFP - w.lastState.posStartFP
		if w.fieldHasPayloads || w.fieldHasOffsets {
			longs[2] = state.payStartFP - w.lastState.payStartFP
		}
	}
	if state.singletonDocID != -1 {
		if err = out.WriteVInt(int32(state.singletonDocID)); err != nil {
			return
		}
	}
	if w.fieldHasPositions && state.lastPosBlockOffset != -1 {
		if err = out.WriteVLong(state.lastPosBlockOffset); err != nil {
			return
		}
	}
	if state.skipOffset != -1 {
		if err = out.WriteVLong(state.skipOffset); err != nil {
			return
		}
	}
	w.lastState = state
	return nil
}

func (w *Lucene41PostingsWriter) Close() (err error) {
	var success = false
	defer func() {
		if success {
			err = util.Close(w.docOut, w.posOut, w.payOut)
		} else {
			util.CloseWhileSuppressingError(w.docOut, w.posOut, w.payOut)
		}
		w.docOut = nil
		w.posOut = nil
		w.payOut = nil
	}()

	if err == nil && w.docOut != nil {
		err = codec.WriteFooter(w.docOut)
	}
	if err == nil && w.posOut != nil {
		err = codec.WriteFooter(w.posOut)
	}
	if err == nil && w.payOut != nil {
		err = codec.WriteFooter(w.payOut)
	}
	if err != nil {
		return
	}
	success = true
	return nil
}
