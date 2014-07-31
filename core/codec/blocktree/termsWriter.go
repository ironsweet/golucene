package blocktree

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/fst"
	"github.com/balzaczyy/golucene/core/util/packed"
	"io"
	"math"
)

// codec/PostingsWriterBase.java

/*
Extension of PostingsConsumer to support pluggable term dictionaries.

This class contains additional hooks to interact with the provided
term dictionaries such as BlockTreeTermsWriter. If you want to re-use
an existing implementation and are only interested in customizing the
format of the postings list, extend this class instead.
*/
type PostingsWriterBase interface {
	codec.PostingsConsumer
	io.Closer

	// Called once after startup, before any terms have been added.
	// Implementations typically write a header to the provided termsOut.
	Init(store.IndexOutput) error
	NewTermState() *BlockTermState
	// Start a new term. Note that a matching call to finishTerm() is
	// done, only if the term has at least one document.
	StartTerm() error
	// Finishes the current term. The provided TermStats contains the
	// term's summary statistics.
	FinishTerm(*BlockTermState) error
	EncodeTerm([]int64, util.DataOutput, *model.FieldInfo, *BlockTermState, bool) error
	// Called when the writing switches to another field.
	SetField(fieldInfo *model.FieldInfo) int
}

// codec/BlockTreeTermsWriter.java

/* Suggested degault value for the minItemsInBlock parameter. */
const DEFAULT_MIN_BLOCK_SIZE = 25

/* Suggested default value for the maxItemsInBlock parameter. */
const DEFAULT_MAX_BLOCK_SIZE = 48

/* Extension of terms file */
const TERMS_EXTENSION = "tim"
const TERMS_CODEC_NAME = "BLOCK_TREE_TERMS_DICT"

const (
	/* Append-only */
	TERMS_VERSION_APPEND_ONLY   = 1
	TERMS_VERSION_META_ARRAY    = 2
	TERMS_VERSION_CHECKSUM      = 3
	TERMS_VERSION_MIN_MAX_TERMS = 4
)

/* Current terms format. */
const TERMS_VERSION_CURRENT = TERMS_VERSION_MIN_MAX_TERMS

/* Extension of terms index file */
const TERMS_INDEX_EXTENSION = "tip"
const TERMS_INDEX_CODEC_NAME = "BLOCK_TREE_TERMS_INDEX"

type BlockTreeTermsWriterSPI interface {
	WriteHeader(store.IndexOutput) error
	WriteIndexHeader(store.IndexOutput) error
}

type FieldMetaData struct {
	fieldInfo        *model.FieldInfo
	rootCode         []byte
	numTerms         int64
	indexStartFP     int64
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	longsSize        int
	minTerm          []byte
	maxTerm          []byte
}

func newFieldMetaData(fieldInfo *model.FieldInfo,
	rootCode []byte, numTerms, indexStartFP, sumTotalTermFreq, sumDocFreq int64,
	docCount, longsSize int, minTerm, maxTerm []byte) *FieldMetaData {
	assert(numTerms > 0)
	assert2(rootCode != nil, "field=%v numTerms=%v", fieldInfo.Name, numTerms)
	return &FieldMetaData{
		fieldInfo,
		rootCode,
		numTerms,
		indexStartFP,
		sumTotalTermFreq,
		sumDocFreq,
		docCount,
		longsSize,
		minTerm,
		maxTerm,
	}
}

type BlockTreeTermsWriter struct {
	spi BlockTreeTermsWriterSPI

	out             store.IndexOutput
	indexOut        store.IndexOutput
	maxDoc          int
	minItemsInBlock int
	maxItemsInBlock int

	postingsWriter PostingsWriterBase
	fieldInfos     model.FieldInfos
	currentField   *model.FieldInfo

	fields  []*FieldMetaData
	segment string

	scratchBytes *store.RAMOutputStream

	// bytesWriter  *store.RAMOutputStream
	// bytesWriter2 *store.RAMOutputStream
}

/*
Create a new writer. The number of items (terms or sub-blocks) per
block will aim tobe between minItermsPerBlock and maxItemsPerBlock,
though in some cases, the blocks may be smaller than the min.
*/
func NewBlockTreeTermsWriter(state *model.SegmentWriteState,
	postingsWriter PostingsWriterBase,
	minItemsInBlock, maxItemsInBlock int) (*BlockTreeTermsWriter, error) {
	assert2(minItemsInBlock >= 2, "minItemsInBlock must be >= 2; got %v", minItemsInBlock)
	assert2(maxItemsInBlock >= 1, "maxItemsInBlock must be >= 1; got %v", maxItemsInBlock)
	assert2(minItemsInBlock <= maxItemsInBlock,
		"maxItemsInBlock must be >= minItemsInBlock; got maxItemsInBlock=%v minItemsInBlock=%v",
		maxItemsInBlock, minItemsInBlock)
	assert2(2*(minItemsInBlock-1) <= maxItemsInBlock,
		"maxItemsInBlock must be at least 2*(minItemsInBlock-1; got maxItemsInBlock=%v minItemsInBlock=%v",
		maxItemsInBlock, minItemsInBlock)

	ans := &BlockTreeTermsWriter{
		maxDoc:          state.SegmentInfo.DocCount(),
		fieldInfos:      state.FieldInfos,
		minItemsInBlock: minItemsInBlock,
		maxItemsInBlock: maxItemsInBlock,
		postingsWriter:  postingsWriter,
		segment:         state.SegmentInfo.Name,
		scratchBytes:    store.NewRAMOutputStreamBuffer(),
		// bytesWriter:     store.NewRAMOutputStreamBuffer(),
		// bytesWriter2:    store.NewRAMOutputStreamBuffer(),
	}
	ans.spi = ans
	var out, indexOut store.IndexOutput
	if err := func() error {
		var success = false
		defer func() {
			if !success {
				util.CloseWhileSuppressingError(out, indexOut)
			}
		}()

		var err error
		termsFileName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, TERMS_EXTENSION)
		if out, err = state.Directory.CreateOutput(termsFileName, state.Context); err != nil {
			return err
		}
		if err = ans.spi.WriteHeader(out); err != nil {
			return err
		}

		termsIndexFileName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, TERMS_INDEX_EXTENSION)
		if indexOut, err = state.Directory.CreateOutput(termsIndexFileName, state.Context); err != nil {
			return err
		}
		if err = ans.spi.WriteIndexHeader(indexOut); err != nil {
			return err
		}

		// have consumer write its format/header
		if err = postingsWriter.Init(out); err != nil {
			return err
		}
		success = true
		return nil
	}(); err != nil {
		return nil, err
	}
	ans.out = out
	ans.indexOut = indexOut
	return ans, nil
}

func (w *BlockTreeTermsWriter) WriteHeader(out store.IndexOutput) error {
	return codec.WriteHeader(out, TERMS_CODEC_NAME, TERMS_VERSION_CURRENT)
}

func (w *BlockTreeTermsWriter) WriteIndexHeader(out store.IndexOutput) error {
	return codec.WriteHeader(out, TERMS_CODEC_NAME, TERMS_VERSION_CURRENT)
}

/* Writes the terms file trailer. */
func (w *BlockTreeTermsWriter) writeTrailer(out store.IndexOutput, dirStart int64) error {
	return out.WriteLong(dirStart)
}

/* Writes the index file trailer. */
func (w *BlockTreeTermsWriter) writeIndexTrailer(indexOut store.IndexOutput, dirStart int64) error {
	return indexOut.WriteLong(dirStart)
}

func (w *BlockTreeTermsWriter) AddField(field *model.FieldInfo) (TermsConsumer, error) {
	assert(w.currentField == nil || w.currentField.Name < field.Name)
	w.currentField = field
	return newTermsWriter(w, field), nil
}

func encodeOutput(fp int64, hasTerms bool, isFloor bool) int64 {
	assert(fp < (1 << 62))
	ans := (fp << 2)
	if hasTerms {
		ans |= BTT_OUTPUT_FLAG_HAS_TERMS
	}
	if isFloor {
		ans |= BTT_OUTPUT_FLAG_IS_FLOOR
	}
	return ans
}

type PendingEntry interface {
	isTerm() bool
}

type PendingTerm struct {
	term []byte
	// stats + metadata
	state *BlockTermState
}

func newPendingTerm(term []byte, state *BlockTermState) *PendingTerm {
	return &PendingTerm{term, state}
}

func (t *PendingTerm) isTerm() bool { return true }

func (t *PendingTerm) String() string { panic("not implemented yet") }

type PendingBlock struct {
	prefix         []byte
	fp             int64
	index          *fst.FST
	subIndeces     []*fst.FST
	hasTerms       bool
	isFloor        bool
	floorLeadByte  int
	sctrachIntsRef *util.IntsRef
}

func newPendingBlock(prefix []byte, fp int64, hasTerms, isFloor bool,
	floorLeadByte int, subIndices []*fst.FST) *PendingBlock {
	return &PendingBlock{
		prefix, fp, nil, subIndices, hasTerms, isFloor, floorLeadByte, util.NewEmptyIntsRef(),
	}
}

func (b *PendingBlock) isTerm() bool { return false }

func (b *PendingBlock) String() string {
	return fmt.Sprintf("BLOCK: %v", utf8ToString(b.prefix))
}

func (b *PendingBlock) compileIndex(floorBlocks []*PendingBlock,
	scratchBytes *store.RAMOutputStream) error {
	assert2(b.isFloor && len(floorBlocks) > 0 || (!b.isFloor && len(floorBlocks) == 0),
		"isFloor=%v floorBlocks=%v", b.isFloor, floorBlocks)

	assert(scratchBytes.FilePointer() == 0)

	// TODO: try writing the leading vLong in MSB order
	// (opposite of what Lucene does today), for better
	// outputs sharing in the FST
	err := scratchBytes.WriteVLong(encodeOutput(b.fp, b.hasTerms, b.isFloor))
	if err != nil {
		return err
	}
	if b.isFloor {
		panic("not implemented yet")
	}

	outputs := fst.ByteSequenceOutputsSingleton()
	indexBuilder := fst.NewBuilder(fst.INPUT_TYPE_BYTE1,
		0, 0, true, false, int(math.MaxInt32),
		outputs, nil, false,
		packed.PackedInts.COMPACT, true, 15)

	bytes := make([]byte, scratchBytes.FilePointer())
	assert(len(bytes) > 0)
	err = scratchBytes.WriteToBytes(bytes)
	if err != nil {
		return err
	}
	err = indexBuilder.Add(fst.ToIntsRef(b.prefix, b.sctrachIntsRef), util.NewBytesRef(bytes))
	if err != nil {
		return err
	}
	scratchBytes.Reset()

	// copy over index for all sub-blocks
	if b.subIndeces != nil {
		panic("not implemented yet")
	}

	if floorBlocks != nil {
		panic("not implemented yet")
	}

	b.index, err = indexBuilder.Finish()
	if err != nil {
		return err
	}
	b.subIndeces = nil
	return nil
}

type TermsWriter struct {
	owner            *BlockTreeTermsWriter
	fieldInfo        *model.FieldInfo
	longsSize        int
	numTerms         int64
	docsSeen         *util.FixedBitSet
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	indexStartFP     int64

	// Used only to partition terms into the block tree; we don't pull
	// an FST from this builder:
	noOutputs    *fst.NoOutputs
	blockBuilder *fst.Builder

	// PendingTerm or PendingBlock:
	pending []PendingEntry

	// Index into pending of most recently written block
	lastBlockIndex int

	minTerm []byte
	maxTerm []byte

	scratchIntsRef *util.IntsRef

	suffixWriter *store.RAMOutputStream
	statsWriter  *store.RAMOutputStream
	metaWriter   *store.RAMOutputStream
	bytesWriter  *store.RAMOutputStream
}

func newTermsWriter(owner *BlockTreeTermsWriter,
	fieldInfo *model.FieldInfo) *TermsWriter {
	owner.postingsWriter.SetField(fieldInfo)
	ans := &TermsWriter{
		owner:          owner,
		fieldInfo:      fieldInfo,
		noOutputs:      fst.NO_OUTPUT,
		lastBlockIndex: -1,
		scratchIntsRef: util.NewEmptyIntsRef(),
		suffixWriter:   store.NewRAMOutputStreamBuffer(),
		statsWriter:    store.NewRAMOutputStreamBuffer(),
		metaWriter:     store.NewRAMOutputStreamBuffer(),
		bytesWriter:    store.NewRAMOutputStreamBuffer(),
	}
	// This builder is just used transiently to fragment terms into
	// "good" blocks; we don't save the resulting FST:
	ans.blockBuilder = fst.NewBuilder(
		fst.INPUT_TYPE_BYTE1, 0, 0, true, true,
		int(math.MaxInt32), fst.NO_OUTPUT,
		//Assign terms to blocks "naturally", ie, according to the number of
		//terms under a given prefix that we encounter:
		func(frontier []*fst.UnCompiledNode, prefixLenPlus1 int, lastInput *util.IntsRef) error {
			for idx := lastInput.Length; idx >= prefixLenPlus1; idx-- {
				node := frontier[idx]

				totCount := int64(0)

				if node.IsFinal {
					totCount++
				}

				for arcIdx := 0; arcIdx < node.NumArcs; arcIdx++ {
					target := node.Arcs[arcIdx].Target.(*fst.UnCompiledNode)
					totCount += target.InputCount
					target.Clear()
					node.Arcs[arcIdx].Target = nil
				}
				node.NumArcs = 0

				if totCount >= int64(ans.owner.minItemsInBlock) || idx == 0 {
					// we are on a prefix node that has enough entries (terms
					// or sub-blocks) under it to let us write a new block or
					// multiple blocks (main block + follow on floor blocks):
					err := ans.writeBlocks(lastInput, idx, int(totCount))
					if err != nil {
						return err
					}
					node.InputCount = 1
				} else {
					// stragglers! carry count upwards
					node.InputCount = totCount
				}
				frontier[idx] = fst.NewUnCompiledNode(ans.blockBuilder, idx)
			}
			return nil
		}, false, packed.PackedInts.COMPACT,
		true, 15)
	return ans
}

/*
Write the top count entries on the pending stack as one or more
blocks. If the entry count is <= maxItemsPerBlock we just write a
single blocks; else we break into primary (initial) block and then
one or more following floor blocks:
*/
func (w *TermsWriter) writeBlocks(prevTerm *util.IntsRef, prefixLength, count int) error {
	fmt.Printf("writeBlocks count=%v\n", count)
	if count <= w.owner.maxItemsInBlock {
		// Easy case: not floor block. Eg, prefix is "foo", and we found
		// 30 terms/sub-blocks starting w/ that prefix, and
		// minItemsInBlock <= 30 <= maxItemsInBlock.
		nonFloorBlock, err := w.writeBlock(prevTerm, prefixLength, prefixLength, count, count, 0, false, -1, true)
		if err != nil {
			return err
		}
		err = nonFloorBlock.compileIndex(nil, w.owner.scratchBytes)
		if err != nil {
			return err
		}
		w.pending = append(w.pending, nonFloorBlock)
		fmt.Println("  1 block")
	} else {
		panic("not implemented yet")
	}
	w.lastBlockIndex = len(w.pending) - 1
	return nil
}

/* Write all entries in the pending slice as a single block: */
func (w *TermsWriter) writeBlock(prevTerm *util.IntsRef, prefixLength,
	indexPrefixLength, startBackwards, length, futureTermCount int,
	isFloor bool, floorLeadByte int, isLastInFloor bool) (*PendingBlock, error) {

	assert(length > 0)

	start := len(w.pending) - startBackwards

	assert2(start >= 0, "len(pending)=%v startBackward=%v length=%v",
		len(w.pending), startBackwards, length)

	slice := w.pending[start : start+length]

	startFP := w.owner.out.FilePointer()

	prefix := make([]byte, indexPrefixLength)
	for m, _ := range prefix {
		prefix[m] = byte(prevTerm.At(m))
	}

	// write block header:
	err := w.owner.out.WriteVInt(int32(length<<1) | (map[bool]int32{true: 1, false: 0}[isLastInFloor]))
	if err != nil {
		return nil, err
	}

	// 1st pass: pack term suffix bytes into []byte blob
	// TODO: cutover to bulk int codec... simple64?

	var isLeafBlock bool
	if w.lastBlockIndex < start {
		// this block defintely does not contain sub-blocks:
		isLeafBlock = true
		fmt.Printf("no scan true isFloor=%v\n", isFloor)
	} else if !isFloor {
		// this block definitely does contain at least one sub-block:
		isLeafBlock = false
		fmt.Printf("no scan false %v vs start=%v len=%v", w.lastBlockIndex, start, length)
	} else {
		// must scan up-front to see if there is a sub-block
		v := true
		fmt.Printf("scan %v vs start=%v len=%v", w.lastBlockIndex, start, length)
		for _, ent := range slice {
			if !ent.isTerm() {
				v = false
				break
			}
		}
		isLeafBlock = v
	}

	var subIndices []*fst.FST

	var termCount int

	longs := make([]int64, w.longsSize)
	var absolute = true

	if isLeafBlock {
		subIndices = nil
		for _, ent := range slice {
			assert(ent.isTerm())
			term := ent.(*PendingTerm)
			state := term.state
			suffix := len(term.term) - prefixLength
			// for leaf block we write suffix straight
			if err = w.suffixWriter.WriteVInt(int32(suffix)); err != nil {
				return nil, err
			}
			if err = w.suffixWriter.WriteBytes(term.term[prefixLength : prefixLength+suffix]); err != nil {
				return nil, err
			}

			// write term stats, to separate []byte blob:
			if err := w.suffixWriter.WriteVInt(int32(state.DocFreq)); err != nil {
				return nil, err
			}
			if w.fieldInfo.IndexOptions() != model.INDEX_OPT_DOCS_ONLY {
				assert2(state.TotalTermFreq >= int64(state.DocFreq),
					"%v vs %v", state.TotalTermFreq, state.DocFreq)
				if err := w.suffixWriter.WriteVLong(state.TotalTermFreq - int64(state.DocFreq)); err != nil {
					return nil, err
				}
			}

			// Write term meta data
			if err = w.owner.postingsWriter.EncodeTerm(longs, w.bytesWriter, w.fieldInfo, state, absolute); err != nil {
				return nil, err
			}
			for _, v := range longs {
				assert(v >= 0)
				if err = w.metaWriter.WriteVLong(v); err != nil {
					return nil, err
				}
			}
			if err = w.bytesWriter.WriteTo(w.metaWriter); err != nil {
				return nil, err
			}
			w.bytesWriter.Reset()
			absolute = false
		}
		termCount = length
	} else {
		panic("not implemented yet")
	}

	// TODO: we could block-write the term suffix pointer
	// this would take more space but would enable binary
	// search on lookup

	// write suffixes []byte blob to terms dict output:
	if err = w.owner.out.WriteVInt(
		int32(w.suffixWriter.FilePointer()<<1) |
			(map[bool]int32{true: 1, false: 0}[isLeafBlock])); err == nil {
		if err = w.suffixWriter.WriteTo(w.owner.out); err == nil {
			w.suffixWriter.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

	// write term stats []byte blob
	if err = w.owner.out.WriteVInt(int32(w.suffixWriter.FilePointer())); err == nil {
		if err = w.suffixWriter.WriteTo(w.owner.out); err == nil {
			w.suffixWriter.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

	// Write term meta data []byte blob
	panic("not implemented yet")

	// remove slice replaced by block:
	w.pending = append(w.pending[:start], w.pending[start+length:]...)

	if w.lastBlockIndex >= start {
		if w.lastBlockIndex < start+length {
			w.lastBlockIndex = start
		} else {
			w.lastBlockIndex -= length
		}
	}

	return newPendingBlock(prefix, startFP, termCount != 0, isFloor, floorLeadByte, subIndices), nil
}

func (w *TermsWriter) Comparator() func(a, b []byte) bool {
	return util.UTF8SortedAsUnicodeLess
}

func (w *TermsWriter) StartTerm(text []byte) (codec.PostingsConsumer, error) {
	assert(w.owner != nil)
	assert(w.owner.postingsWriter != nil)
	err := w.owner.postingsWriter.StartTerm()
	return w.owner.postingsWriter, err
}

func (w *TermsWriter) FinishTerm(text []byte, stats *codec.TermStats) (err error) {
	assert(stats.DocFreq > 0)
	fmt.Printf("BTTW.finishTerm term=%v:%v seg=%v df=%v",
		w.fieldInfo.Name, utf8ToString(text), w.owner.segment, stats.DocFreq)

	if err = w.blockBuilder.Add(fst.ToIntsRef(text, w.scratchIntsRef), w.noOutputs.NoOutput()); err != nil {
		return
	}
	state := w.owner.postingsWriter.NewTermState()
	state.DocFreq = stats.DocFreq
	state.TotalTermFreq = stats.TotalTermFreq
	if err = w.owner.postingsWriter.FinishTerm(state); err != nil {
		return
	}

	term := newPendingTerm(util.DeepCopyOf(util.NewBytesRef(text)).Value, state)
	w.pending = append(w.pending, term)
	w.numTerms++

	if w.minTerm == nil {
		w.minTerm = util.DeepCopyOf(util.NewBytesRef(text)).Value
	}
	w.maxTerm = util.DeepCopyOf(util.NewBytesRef(text)).Value
	return nil
}

func (w *TermsWriter) Finish(sumTotalTermFreq, sumDocFreq int64, docCount int) error {
	if w.numTerms > 0 {
		_, err := w.blockBuilder.Finish()
		if err != nil {
			return err
		}

		// we better have one final "root" block:
		assert2(len(w.pending) == 1 && !w.pending[0].isTerm(),
			"len(pending) = %v pending=%v", len(w.pending), w.pending)
		root := w.pending[0].(*PendingBlock)
		assert(len(root.prefix) == 0)
		assert(root.index.EmptyOutput() != nil)

		w.sumTotalTermFreq = sumTotalTermFreq
		w.sumDocFreq = sumDocFreq
		w.docCount = docCount

		// Write FST to index
		w.indexStartFP = w.owner.indexOut.FilePointer()
		err = root.index.Save(w.owner.indexOut)
		if err != nil {
			return err
		}
		fmt.Printf("  write FST %v field=%v\n", w.indexStartFP, w.fieldInfo.Name)

		w.owner.fields = append(w.owner.fields, newFieldMetaData(
			w.fieldInfo,
			w.pending[0].(*PendingBlock).index.EmptyOutput().(*util.BytesRef).Value,
			w.numTerms,
			w.indexStartFP,
			sumTotalTermFreq,
			sumDocFreq,
			docCount,
			w.longsSize,
			w.minTerm, w.maxTerm))
	} else {
		assert(sumTotalTermFreq == 0 || w.fieldInfo.IndexOptions() == model.INDEX_OPT_DOCS_ONLY && sumTotalTermFreq == -1)
		assert(sumDocFreq == 0)
		assert(docCount == 0)
	}
	return nil
}

func (w *BlockTreeTermsWriter) Close() (err error) {
	var success = false
	defer func() {
		if success {
			util.Close(w.out, w.indexOut, w.postingsWriter)
		} else {
			util.CloseWhileSuppressingError(w.out, w.indexOut, w.postingsWriter)
		}
	}()

	dirStart := w.out.FilePointer()
	indexDirStart := w.indexOut.FilePointer()

	if err = w.out.WriteVInt(int32(len(w.fields))); err != nil {
		return
	}

	for _, field := range w.fields {
		fmt.Printf("  field %v %v terms\n", field.fieldInfo.Name, field.numTerms)
		if err = w.out.WriteVInt(field.fieldInfo.Number); err == nil {
			assert(field.numTerms > 0)
			if err = w.out.WriteVLong(field.numTerms); err == nil {
				if err = w.out.WriteVInt(int32(len(field.rootCode))); err == nil {
					err = w.out.WriteBytes(field.rootCode)
					if err == nil && field.fieldInfo.IndexOptions() != model.INDEX_OPT_DOCS_ONLY {
						err = w.out.WriteVLong(field.sumTotalTermFreq)
					}
					if err == nil {
						if err = w.out.WriteVLong(field.sumDocFreq); err == nil {
							if err = w.out.WriteVInt(int32(field.docCount)); err == nil {
								if err = w.out.WriteVInt(int32(field.longsSize)); err == nil {
									if err = w.indexOut.WriteVLong(field.indexStartFP); err == nil {
										if err = writeBytesRef(w.out, field.minTerm); err == nil {
											err = writeBytesRef(w.out, field.maxTerm)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if err == nil {
		if err = w.writeTrailer(w.out, dirStart); err == nil {
			if err = codec.WriteFooter(w.out); err == nil {
				if err = w.writeIndexTrailer(w.indexOut, indexDirStart); err == nil {
					if err = codec.WriteFooter(w.indexOut); err == nil {
						success = true
					}
				}
			}
		}
	}
	return
}

func writeBytesRef(out store.IndexOutput, bytes []byte) error {
	panic("not implemented yet")
}
