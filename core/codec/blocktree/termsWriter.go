package blocktree

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/fst"
	"github.com/balzaczyy/golucene/core/util/packed"
	"io"
	"math"
	"strings"
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
	EncodeTerm([]int64, util.DataOutput, *FieldInfo, *BlockTermState, bool) error
	// Called when the writing switches to another field.
	SetField(fieldInfo *FieldInfo) int
}

// codec/BlockTreeTermsWriter.java
const (
	/* Suggested degault value for the minItemsInBlock parameter. */
	DEFAULT_MIN_BLOCK_SIZE = 25

	/* Suggested default value for the maxItemsInBlock parameter. */
	DEFAULT_MAX_BLOCK_SIZE = 48

	/* Extension of terms file */
	TERMS_EXTENSION  = "tim"
	TERMS_CODEC_NAME = "BLOCK_TREE_TERMS_DICT"

	TERMS_VERSION_START = 0
	/* Append-only */
	TERMS_VERSION_APPEND_ONLY   = 1
	TERMS_VERSION_META_ARRAY    = 2
	TERMS_VERSION_CHECKSUM      = 3
	TERMS_VERSION_MIN_MAX_TERMS = 4
	/* Current terms format. */
	TERMS_VERSION_CURRENT = TERMS_VERSION_MIN_MAX_TERMS

	/* Extension of terms index file */
	TERMS_INDEX_EXTENSION  = "tip"
	TERMS_INDEX_CODEC_NAME = "BLOCK_TREE_TERMS_INDEX"
)

type BlockTreeTermsWriterSPI interface {
	WriteHeader(store.IndexOutput) error
	WriteIndexHeader(store.IndexOutput) error
}

type FieldMetaData struct {
	fieldInfo        *FieldInfo
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

func newFieldMetaData(fieldInfo *FieldInfo,
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
	fieldInfos     FieldInfos
	currentField   *FieldInfo

	fields  []*FieldMetaData
	segment string

	scratchBytes   *store.RAMOutputStream
	scratchIntsRef *util.IntsRefBuilder
}

/*
Create a new writer. The number of items (terms or sub-blocks) per
block will aim tobe between minItermsPerBlock and maxItemsPerBlock,
though in some cases, the blocks may be smaller than the min.
*/
func NewBlockTreeTermsWriter(state *SegmentWriteState,
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
		scratchIntsRef:  util.NewIntsRefBuilder(),
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
	return codec.WriteHeader(out, TERMS_INDEX_CODEC_NAME, TERMS_VERSION_CURRENT)
}

/* Writes the terms file trailer. */
func (w *BlockTreeTermsWriter) writeTrailer(out store.IndexOutput, dirStart int64) error {
	return out.WriteLong(dirStart)
}

/* Writes the index file trailer. */
func (w *BlockTreeTermsWriter) writeIndexTrailer(indexOut store.IndexOutput, dirStart int64) error {
	return indexOut.WriteLong(dirStart)
}

func (w *BlockTreeTermsWriter) AddField(field *FieldInfo) (TermsConsumer, error) {
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
	clone := make([]byte, len(term))
	copy(clone, term)
	return &PendingTerm{clone, state}
}

func (t *PendingTerm) isTerm() bool { return true }

func (t *PendingTerm) String() string { panic("not implemented yet") }

type PendingBlock struct {
	prefix        []byte
	fp            int64
	index         *fst.FST
	subIndices    []*fst.FST
	hasTerms      bool
	isFloor       bool
	floorLeadByte int
}

func newPendingBlock(prefix []byte, fp int64, hasTerms, isFloor bool,
	floorLeadByte int, subIndices []*fst.FST) *PendingBlock {
	return &PendingBlock{
		prefix:        prefix,
		fp:            fp,
		index:         nil,
		subIndices:    subIndices,
		hasTerms:      hasTerms,
		isFloor:       isFloor,
		floorLeadByte: floorLeadByte,
	}
}

func (b *PendingBlock) isTerm() bool { return false }

func (b *PendingBlock) String() string {
	return fmt.Sprintf("BLOCK: %v", utf8ToString(b.prefix))
}

func (b *PendingBlock) compileIndex(blocks []*PendingBlock,
	scratchBytes *store.RAMOutputStream,
	scratchIntsRef *util.IntsRefBuilder) (err error) {

	assert2(b.isFloor && len(blocks) > 1 || (!b.isFloor && len(blocks) == 1),
		"isFloor=%v blocks=%v", b.isFloor, blocks)
	assert(blocks[0] == b)

	assert(scratchBytes.FilePointer() == 0)

	// TODO: try writing the leading vLong in MSB order
	// (opposite of what Lucene does today), for better
	// outputs sharing in the FST
	if err = scratchBytes.WriteVLong(encodeOutput(b.fp, b.hasTerms, b.isFloor)); err != nil {
		return
	}
	if b.isFloor {
		if err = scratchBytes.WriteVInt(int32(len(blocks) - 1)); err != nil {
			return
		}
		for _, sub := range blocks[1:] {
			assert(sub.floorLeadByte != -1)
			// fmt.Printf("    write floorLeadByte=%v\n", util.ItoHex(int64(sub.floorLeadByte)))
			if err = scratchBytes.WriteByte(byte(sub.floorLeadByte)); err != nil {
				return
			}
			assert(sub.fp > b.fp)
			if err = scratchBytes.WriteVLong((sub.fp-b.fp)<<1 |
				int64(map[bool]int{true: 1, false: 0}[sub.hasTerms])); err != nil {
				return
			}
		}
	}

	outputs := fst.ByteSequenceOutputsSingleton()
	indexBuilder := fst.NewBuilder(fst.INPUT_TYPE_BYTE1,
		0, 0, true, false, int(math.MaxInt32),
		outputs, false,
		packed.PackedInts.COMPACT, true, 15)

	// fmt.Printf("  compile index for prefix=%v\n", b.prefix)

	bytes := make([]byte, scratchBytes.FilePointer())
	assert(len(bytes) > 0)
	err = scratchBytes.WriteToBytes(bytes)
	if err != nil {
		return err
	}
	err = indexBuilder.Add(fst.ToIntsRef(b.prefix, scratchIntsRef), bytes)
	if err != nil {
		return err
	}
	scratchBytes.Reset()

	// copy over index for all sub-blocks
	for _, block := range blocks {
		if block.subIndices != nil {
			for _, subIndex := range block.subIndices {
				if err = b.append(indexBuilder, subIndex, scratchIntsRef); err != nil {
					return err
				}
			}
		}
		block.subIndices = nil
	}

	if b.index, err = indexBuilder.Finish(); err != nil {
		return err
	}
	assert(b.subIndices == nil)
	return nil
}

func (b *PendingBlock) append(
	builder *fst.Builder,
	subIndex *fst.FST,
	scratchIntsRef *util.IntsRefBuilder) error {

	subIndexEnum := fst.NewBytesRefFSTEnum(subIndex)
	indexEnt, err := subIndexEnum.Next()
	for err == nil && indexEnt != nil {
		// fmt.Printf("      add sub=%v output=%v\n", indexEnt.Input, indexEnt.Output)
		err = builder.Add(fst.ToIntsRef(indexEnt.Input.ToBytes(), scratchIntsRef), indexEnt.Output)
		if err == nil {
			indexEnt, err = subIndexEnum.Next()
		}
	}
	return err
}

type TermsWriter struct {
	owner            *BlockTreeTermsWriter
	fieldInfo        *FieldInfo
	longsSize        int
	numTerms         int64
	docsSeen         *util.FixedBitSet
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	indexStartFP     int64

	// Records index into pending where the current prefix at that
	// length "started"; for example, if current term starts with 't',
	// startsByPrefix[0] is the index into pending for the first
	// term/sub-block starting with 't'. We use this to figure out when
	// to write a new block:
	lastTerm     *util.BytesRefBuilder
	prefixStarts []int

	longs []int64

	// Pending stack of terms and blocks. As terms arrive (in sorted
	// order) we append to this stack, and once the top of the stak has
	// enough terms starting with a common prefix, we write a new block
	// with those terms and replace those terms in the stack with a new
	// block:
	pending []PendingEntry

	// Reused in writeBlocks:
	newBlocks []*PendingBlock

	firstPendingTerm *PendingTerm
	lastPendingTerm  *PendingTerm

	suffixWriter *store.RAMOutputStream
	statsWriter  *store.RAMOutputStream
	metaWriter   *store.RAMOutputStream
	bytesWriter  *store.RAMOutputStream
}

func newTermsWriter(owner *BlockTreeTermsWriter,
	fieldInfo *FieldInfo) *TermsWriter {
	owner.postingsWriter.SetField(fieldInfo)
	ans := &TermsWriter{
		owner:        owner,
		fieldInfo:    fieldInfo,
		lastTerm:     util.NewBytesRefBuilder(),
		prefixStarts: make([]int, 8),
		suffixWriter: store.NewRAMOutputStreamBuffer(),
		statsWriter:  store.NewRAMOutputStreamBuffer(),
		metaWriter:   store.NewRAMOutputStreamBuffer(),
		bytesWriter:  store.NewRAMOutputStreamBuffer(),
	}
	ans.longsSize = owner.postingsWriter.SetField(fieldInfo)
	ans.longs = make([]int64, ans.longsSize)
	return ans
}

/* Writes the top count entries in pending, using prevTerm to compute the prefix. */
func (w *TermsWriter) writeBlocks(prefixLength, count int) (err error) {
	assert(count > 0)

	// Root block better writes all remaining pending entries:
	assert(prefixLength > 0 || count == len(w.pending))

	lastSuffixLeadLabel := -1

	// True if we saw at least one term in this block (we record if a
	// block only points to sub-blocks in the terms index so we can
	// avoid seeking to it when we are looking for a term):
	hasTerms := false
	hasSubBlocks := false

	end := len(w.pending)
	start := end - count
	nextBlockStart := start
	nextFloorLeadLabel := -1

	for i, ent := range w.pending[start:] {
		var suffixLeadLabel int
		if ent.isTerm() {
			term := ent.(*PendingTerm)
			if len(term.term) == prefixLength {
				// suffix is 0, ie prefix 'foo' and term is 'foo' so the
				// term has empty string suffix in this block
				assert(lastSuffixLeadLabel == -1)
				suffixLeadLabel = -1
			} else {
				suffixLeadLabel = int(term.term[prefixLength])
			}
		} else {
			block := ent.(*PendingBlock)
			assert(len(block.prefix) > prefixLength)
			suffixLeadLabel = int(block.prefix[prefixLength])
		}

		if suffixLeadLabel != lastSuffixLeadLabel {
			if itemsInBlock := i + count - nextBlockStart; itemsInBlock >= w.owner.minItemsInBlock &&
				end-nextBlockStart > w.owner.maxItemsInBlock {
				// The count is too large for one block, so we must break
				// it into "floor" blocks, where we record the leading
				// label of the suffix of the first term in each floor
				// block, so at search time we can jump to the right floor
				// block. We just use a naive greedy segmenter here: make a
				// new floor block as soon as we have at least
				// minItemsInBlock. This is not always best: it often
				// produces a too-small block as the final block:
				isFloor := itemsInBlock < count
				var block *PendingBlock
				if block, err = w.writeBlock(prefixLength, isFloor,
					nextFloorLeadLabel, nextBlockStart, i+count, hasTerms,
					hasSubBlocks); err != nil {
					return
				}
				w.newBlocks = append(w.newBlocks, block)

				hasTerms = false
				hasSubBlocks = false
				nextFloorLeadLabel = suffixLeadLabel
				nextBlockStart = i + count
			}

			lastSuffixLeadLabel = suffixLeadLabel
		}

		if ent.isTerm() {
			hasTerms = true
		} else {
			hasSubBlocks = true
		}
	}

	// Write last block, if any:
	if nextBlockStart < end {
		itemsInBlock := end - nextBlockStart
		isFloor := itemsInBlock < count
		var block *PendingBlock
		if block, err = w.writeBlock(prefixLength, isFloor,
			nextFloorLeadLabel, nextBlockStart, end, hasTerms,
			hasSubBlocks); err != nil {
			return
		}
		w.newBlocks = append(w.newBlocks, block)
	}

	assert(len(w.newBlocks) > 0)

	firstBlock := w.newBlocks[0]

	assert(firstBlock.isFloor || len(w.newBlocks) == 1)

	if err = firstBlock.compileIndex(w.newBlocks,
		w.owner.scratchBytes, w.owner.scratchIntsRef); err != nil {
		return
	}

	// Remove slice from the top of the pending stack, that we just wrote:
	w.pending = w.pending[:start]

	// Append new block
	w.pending = append(w.pending, firstBlock)

	w.newBlocks = nil
	return nil
}

/*
Writes the specified slice (start is inclusive, end is exclusive)
from pending stack as a new block. If isFloor is true, there were too
many (more than maxItemsInBlock) entries sharing the same prefix, and
so we broke it into multiple floor blocks where we record the
starting lable of the suffix of each floor block.
*/
func (w *TermsWriter) writeBlock(
	prefixLength int,
	isFloor bool,
	floorLeadLabel, start, end int,
	hasTerms, hasSubBlocks bool) (*PendingBlock, error) {

	assert(end > start)

	startFP := w.owner.out.FilePointer()

	hasFloorLeadLabel := isFloor && floorLeadLabel != -1

	prefix := make([]byte, prefixLength)
	copy(prefix, w.lastTerm.Bytes()[:prefixLength])

	// write block header:
	numEntries := end - start
	code := numEntries << 1
	if end == len(w.pending) { // last block
		code |= 1
	}
	var err error
	if err = w.owner.out.WriteVInt(int32(code)); err != nil {
		return nil, err
	}

	// fmt.Printf("  writeBlock %vseg=%v len(pending)=%v prefixLength=%v "+
	// 	"indexPrefix=%v entCount=%v startFP=%v futureTermCount=%v%v "+
	// 	"isLastInFloor=%v\n",
	// 	map[bool]string{true: "(floor) "}[isFloor],
	// 	w.owner.segment,
	// 	len(w.pending),
	// 	prefixLength,
	// 	prefix,
	// 	length,
	// 	startFP,
	// 	futureTermCount,
	// 	map[bool]string{true: fmt.Sprintf(" floorLeadByte=%v", strconv.FormatInt(int64(floorLeadByte&0xff), 16))}[isFloor],
	// 	isLastInFloor,
	// )

	// 1st pass: pack term suffix bytes into []byte blob
	// TODO: cutover to bulk int codec... simple64?

	// We optimize the leaf block case (block has only terms), writing
	// a more compact format in this case:
	isLeafBlock := !hasSubBlocks

	var subIndices []*fst.FST

	var absolute = true

	if isLeafBlock { // only terms
		subIndices = nil
		for i, ent := range w.pending[start:end] {
			assert2(ent.isTerm(), "i=%v", i+start)

			term := ent.(*PendingTerm)
			assert2(strings.HasPrefix(string(term.term), string(prefix)), "term.term=%v prefix=%v", term.term, prefix)
			state := term.state
			suffix := len(term.term) - prefixLength
			// for leaf block we write suffix straight
			if err = w.suffixWriter.WriteVInt(int32(suffix)); err != nil {
				return nil, err
			}
			if err = w.suffixWriter.WriteBytes(term.term[prefixLength : prefixLength+suffix]); err != nil {
				return nil, err
			}
			assert(floorLeadLabel == -1 || int(term.term[prefixLength]) >= floorLeadLabel)

			// write term stats, to separate []byte blob:
			if err = w.statsWriter.WriteVInt(int32(state.DocFreq)); err != nil {
				return nil, err
			}
			if w.fieldInfo.IndexOptions() != INDEX_OPT_DOCS_ONLY {
				assert2(state.TotalTermFreq >= int64(state.DocFreq),
					"%v vs %v", state.TotalTermFreq, state.DocFreq)
				if err := w.statsWriter.WriteVLong(state.TotalTermFreq - int64(state.DocFreq)); err != nil {
					return nil, err
				}
			}

			// Write term meta data
			if err = w.owner.postingsWriter.EncodeTerm(w.longs, w.bytesWriter, w.fieldInfo, state, absolute); err != nil {
				return nil, err
			}
			for _, v := range w.longs[:w.longsSize] {
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

	} else { // mixed terms and sub-blocks
		subIndices = nil
		for _, ent := range w.pending[start:end] {
			if ent.isTerm() {
				term := ent.(*PendingTerm)
				assert2(strings.HasPrefix(string(term.term), string(prefix)), "term.term=%v prefix=%v", term.term, prefix)
				state := term.state
				suffix := len(term.term) - prefixLength
				// for non-leaf block we borrow 1 bit to record
				// if entr is term or sub-block
				if err = w.suffixWriter.WriteVInt(int32(suffix << 1)); err != nil {
					return nil, err
				}
				if err = w.suffixWriter.WriteBytes(term.term[prefixLength : prefixLength+suffix]); err != nil {
					return nil, err
				}
				assert(floorLeadLabel == -1 || int(term.term[prefixLength]) >= floorLeadLabel)

				// write term stats, to separate []byte block:
				if err = w.statsWriter.WriteVInt(int32(state.DocFreq)); err != nil {
					return nil, err
				}
				if w.fieldInfo.IndexOptions() != INDEX_OPT_DOCS_ONLY {
					assert(state.TotalTermFreq >= int64(state.DocFreq))
					if err = w.statsWriter.WriteVLong(state.TotalTermFreq - int64(state.DocFreq)); err != nil {
						return nil, err
					}
				}

				// write term meta data
				if err = w.owner.postingsWriter.EncodeTerm(w.longs, w.bytesWriter, w.fieldInfo, state, absolute); err != nil {
					return nil, err
				}
				for _, v := range w.longs[:w.longsSize] {
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

			} else {
				block := ent.(*PendingBlock)
				assert(strings.HasPrefix(string(block.prefix), string(prefix)))
				suffix := len(block.prefix) - prefixLength

				assert(suffix > 0)

				// for non-leaf block we borrow 1 bit to record if entry is
				// term or sub-block
				if err = w.suffixWriter.WriteVInt(int32((suffix << 1) | 1)); err != nil {
					return nil, err
				}
				if err = w.suffixWriter.WriteBytes(block.prefix[prefixLength : prefixLength+suffix]); err != nil {
					return nil, err
				}

				assert(floorLeadLabel == -1 || int(block.prefix[prefixLength]) >= floorLeadLabel)

				assert(block.fp < startFP)

				if err = w.suffixWriter.WriteVLong(startFP - block.fp); err != nil {
					return nil, err
				}
				subIndices = append(subIndices, block.index)
			}
		}

		assert(len(subIndices) != 0)
	}

	// TODO: we could block-write the term suffix pointer
	// this would take more space but would enable binary
	// search on lookup

	// write suffixes []byte blob to terms dict output:
	if err = w.owner.out.WriteVInt(
		int32(w.suffixWriter.FilePointer()<<1) |
			(map[bool]int32{true: 1, false: 0}[isLeafBlock])); err != nil {
		return nil, err
	}
	if err = w.suffixWriter.WriteTo(w.owner.out); err != nil {
		return nil, err
	}
	w.suffixWriter.Reset()

	// write term stats []byte blob
	if err = w.owner.out.WriteVInt(int32(w.statsWriter.FilePointer())); err != nil {
		return nil, err
	}
	if err = w.statsWriter.WriteTo(w.owner.out); err != nil {
		return nil, err
	}
	w.statsWriter.Reset()

	// Write term meta data []byte blob
	if err = w.owner.out.WriteVInt(int32(w.metaWriter.FilePointer())); err != nil {
		return nil, err
	}
	if err = w.metaWriter.WriteTo(w.owner.out); err != nil {
		return nil, err
	}
	w.metaWriter.Reset()

	if hasFloorLeadLabel {
		prefix = append(prefix, byte(floorLeadLabel))
	}

	return newPendingBlock(prefix, startFP, hasTerms, isFloor, floorLeadLabel, subIndices), nil
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
	// fmt.Printf("BTTW.finishTerm term=%v:%v seg=%v df=%v\n",
	// w.fieldInfo.Name, utf8ToString(text), w.owner.segment, stats.DocFreq)

	assert2(w.fieldInfo.IndexOptions() == INDEX_OPT_DOCS_ONLY ||
		stats.TotalTermFreq >= int64(stats.DocFreq),
		"postingsWriter=%v", w.owner.postingsWriter)
	state := w.owner.postingsWriter.NewTermState()
	state.DocFreq = stats.DocFreq
	state.TotalTermFreq = stats.TotalTermFreq
	if err = w.owner.postingsWriter.FinishTerm(state); err != nil {
		return
	}

	w.sumDocFreq += int64(state.DocFreq)
	w.sumTotalTermFreq += state.TotalTermFreq
	if err = w.pushTerm(text); err != nil {
		return
	}

	term := newPendingTerm(text, state)
	w.pending = append(w.pending, term)
	w.numTerms++
	if w.firstPendingTerm == nil {
		w.firstPendingTerm = term
	}
	w.lastPendingTerm = term
	return nil
}

/* Pushes the new term to the top of the stack, and writes new blocks. */
func (w *TermsWriter) pushTerm(text []byte) error {
	limit := w.lastTerm.Length()
	if len(text) < limit {
		limit = len(text)
	}

	// Find common prefix between last term and current term:
	pos := 0
	for pos < limit && w.lastTerm.At(pos) == text[pos] {
		pos++
	}

	// Close the "abandoned" suffix now:
	for i := w.lastTerm.Length() - 1; i >= pos; i-- {
		// How many items on top of the stack share the current suffix
		// we are closing:
		if prefixTopSize := len(w.pending) - w.prefixStarts[i]; prefixTopSize >= w.owner.minItemsInBlock {
			if err := w.writeBlocks(i+1, prefixTopSize); err != nil {
				return err
			}
			w.prefixStarts[i] -= prefixTopSize - 1
		}
	}

	if len(w.prefixStarts) < len(text) {
		w.prefixStarts = util.GrowIntSlice(w.prefixStarts, len(text))
	}

	// Init new tail:
	for i := pos; i < len(text); i++ {
		w.prefixStarts[i] = len(w.pending)
	}

	w.lastTerm.Copy(text)
	return nil
}

func (w *TermsWriter) Finish(sumTotalTermFreq, sumDocFreq int64, docCount int) (err error) {
	if w.numTerms > 0 {
		// Add empty term to force closing of all final blocks:
		w.pushTerm(nil)

		// TODO: if len(pending) is already 1 with a non-zero prefix length
		// we can save writing a "degenerate" root block, but we have to
		// fix all the palces that assume the root blocks' prefix is the empty string:
		if err = w.writeBlocks(0, len(w.pending)); err != nil {
			return err
		}

		// we better have one final "root" block:
		assert2(len(w.pending) == 1 && !w.pending[0].isTerm(),
			"len(pending) = %v pending=%v", len(w.pending), w.pending)
		root := w.pending[0].(*PendingBlock)
		assert2(len(root.prefix) == 0, "%v", root.prefix)
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
		// fmt.Printf("  write FST %v field=%v\n", w.indexStartFP, w.fieldInfo.Name)

		assert(w.firstPendingTerm != nil)
		minTerm := w.firstPendingTerm.term
		assert(w.lastPendingTerm != nil)
		maxTerm := w.lastPendingTerm.term

		w.owner.fields = append(w.owner.fields, newFieldMetaData(
			w.fieldInfo,
			w.pending[0].(*PendingBlock).index.EmptyOutput().([]byte),
			w.numTerms,
			w.indexStartFP,
			sumTotalTermFreq,
			sumDocFreq,
			docCount,
			w.longsSize,
			minTerm, maxTerm))
	} else {
		assert(sumTotalTermFreq == 0 || w.fieldInfo.IndexOptions() == INDEX_OPT_DOCS_ONLY && sumTotalTermFreq == -1)
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
		// fmt.Printf("  field %v %v terms\n", field.fieldInfo.Name, field.numTerms)
		if err = w.out.WriteVInt(field.fieldInfo.Number); err == nil {
			assert(field.numTerms > 0)
			if err = w.out.WriteVLong(field.numTerms); err == nil {
				if err = w.out.WriteVInt(int32(len(field.rootCode))); err == nil {
					err = w.out.WriteBytes(field.rootCode)
					if err == nil && field.fieldInfo.IndexOptions() != INDEX_OPT_DOCS_ONLY {
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

func writeBytesRef(out store.IndexOutput, bytes []byte) (err error) {
	if err = out.WriteVInt(int32(len(bytes))); err == nil {
		err = out.WriteBytes(bytes)
	}
	return
}
