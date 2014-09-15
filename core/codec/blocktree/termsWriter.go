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

	scratchBytes *store.RAMOutputStream

	// bytesWriter  *store.RAMOutputStream
	// bytesWriter2 *store.RAMOutputStream
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
	scratchBytes *store.RAMOutputStream) (err error) {

	assert2(b.isFloor && len(floorBlocks) > 0 || (!b.isFloor && len(floorBlocks) == 0),
		"isFloor=%v floorBlocks=%v", b.isFloor, floorBlocks)

	assert(scratchBytes.FilePointer() == 0)

	// TODO: try writing the leading vLong in MSB order
	// (opposite of what Lucene does today), for better
	// outputs sharing in the FST
	if err = scratchBytes.WriteVLong(encodeOutput(b.fp, b.hasTerms, b.isFloor)); err != nil {
		return
	}
	if b.isFloor {
		if err = scratchBytes.WriteVInt(int32(len(floorBlocks))); err != nil {
			return
		}
		for _, sub := range floorBlocks {
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
		outputs, nil, false,
		packed.PackedInts.COMPACT, true, 15)

	// fmt.Printf("  compile index for prefix=%v\n", b.prefix)

	bytes := make([]byte, scratchBytes.FilePointer())
	assert(len(bytes) > 0)
	err = scratchBytes.WriteToBytes(bytes)
	if err != nil {
		return err
	}
	err = indexBuilder.Add(fst.ToIntsRef(b.prefix, b.sctrachIntsRef), bytes)
	if err != nil {
		return err
	}
	scratchBytes.Reset()

	// copy over index for all sub-blocks

	if b.subIndeces != nil {
		for _, subIndex := range b.subIndeces {
			if err = b.append(indexBuilder, subIndex); err != nil {
				return err
			}
		}
	}

	if floorBlocks != nil {
		for _, sub := range floorBlocks {
			if sub.subIndeces != nil {
				for _, subIndex := range sub.subIndeces {
					if err = b.append(indexBuilder, subIndex); err != nil {
						return err
					}
				}
			}
			sub.subIndeces = nil
		}
	}

	b.index, err = indexBuilder.Finish()
	if err != nil {
		return err
	}
	b.subIndeces = nil
	return nil
}

func (b *PendingBlock) append(builder *fst.Builder, subIndex *fst.FST) error {
	subIndexEnum := fst.NewBytesRefFSTEnum(subIndex)
	indexEnt, err := subIndexEnum.Next()
	for err == nil && indexEnt != nil {
		// fmt.Printf("      add sub=%v output=%v\n", indexEnt.Input, indexEnt.Output)
		err = builder.Add(fst.ToIntsRef(indexEnt.Input, b.sctrachIntsRef), indexEnt.Output)
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

	// Used only to partition terms into the block tree; we don't pull
	// an FST from this builder:
	noOutputs    *fst.NoOutputs
	blockBuilder *fst.Builder

	// PendingTerm or PendingBlock:
	pending []PendingEntry

	// Index into pending of most recently written block
	lastBlockIndex int

	// Re-used when segmenting a too-large block into floor blocks:
	subBytes         []int
	subTermCounts    []int
	subTermCountSums []int
	subSubCounts     []int

	minTerm []byte
	maxTerm []byte

	scratchIntsRef *util.IntsRef

	suffixWriter *store.RAMOutputStream
	statsWriter  *store.RAMOutputStream
	metaWriter   *store.RAMOutputStream
	bytesWriter  *store.RAMOutputStream
}

func newTermsWriter(owner *BlockTreeTermsWriter,
	fieldInfo *FieldInfo) *TermsWriter {
	owner.postingsWriter.SetField(fieldInfo)
	ans := &TermsWriter{
		owner:            owner,
		fieldInfo:        fieldInfo,
		noOutputs:        fst.NO_OUTPUT,
		lastBlockIndex:   -1,
		subBytes:         make([]int, 10),
		subTermCounts:    make([]int, 10),
		subTermCountSums: make([]int, 10),
		subSubCounts:     make([]int, 10),
		scratchIntsRef:   util.NewEmptyIntsRef(),
		suffixWriter:     store.NewRAMOutputStreamBuffer(),
		statsWriter:      store.NewRAMOutputStreamBuffer(),
		metaWriter:       store.NewRAMOutputStreamBuffer(),
		bytesWriter:      store.NewRAMOutputStreamBuffer(),
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
	ans.longsSize = owner.postingsWriter.SetField(fieldInfo)
	return ans
}

/*
Write the top count entries on the pending stack as one or more
blocks. If the entry count is <= maxItemsPerBlock we just write a
single blocks; else we break into primary (initial) block and then
one or more following floor blocks:
*/
func (w *TermsWriter) writeBlocks(prevTerm *util.IntsRef, prefixLength, count int) (err error) {
	// fmt.Printf("writeBlocks count=%v\n", count)
	if count <= w.owner.maxItemsInBlock {
		// Easy case: not floor block. Eg, prefix is "foo", and we found
		// 30 terms/sub-blocks starting w/ that prefix, and
		// minItemsInBlock <= 30 <= maxItemsInBlock.
		var nonFloorBlock *PendingBlock
		if nonFloorBlock, err = w.writeBlock(prevTerm, prefixLength,
			prefixLength, count, count, 0, false, -1, true); err != nil {
			return
		}
		if err = nonFloorBlock.compileIndex(nil, w.owner.scratchBytes); err != nil {
			return
		}
		w.pending = append(w.pending, nonFloorBlock)
		// fmt.Println("  1 block")

	} else {
		// Floor block case. E.g., prefix is "foo" but we have 100
		// terms/sub-blocks starting w/ that prefix. We segment the
		// entries into a primary block and following floor blocks using
		// the first label in the suffix to assign to floor blocks.

		// fmt.Printf("\nwbs count=%v\n", count)

		savLabel := prevTerm.Ints[prevTerm.Offset+prefixLength]

		// count up how many items fall under each unique label after the prefix.

		lastSuffixLeadLabel := -1
		termCount := 0
		subCount := 0
		numSubs := 0

		for _, ent := range w.pending[len(w.pending)-count:] {
			// first byte in the suffix of this term
			var suffixLeadLabel int
			if ent.isTerm() {
				term := ent.(*PendingTerm)
				if len(term.term) == prefixLength {
					// suffix is 0, ie prefix 'foo' and term is 'foo' so the
					// term has empty string suffix in this block
					assert(lastSuffixLeadLabel == -1)
					assert(numSubs == 0)
					suffixLeadLabel = -1
				} else {
					suffixLeadLabel = int(term.term[prefixLength])
				}
			} else {
				block := ent.(*PendingBlock)
				assert(len(block.prefix) > prefixLength)
				suffixLeadLabel = int(block.prefix[prefixLength])
			}

			if suffixLeadLabel != lastSuffixLeadLabel && termCount+subCount != 0 {
				if len(w.subBytes) == numSubs {
					w.subBytes = append(w.subBytes, lastSuffixLeadLabel)
					w.subTermCounts = append(w.subTermCounts, termCount)
					w.subSubCounts = append(w.subSubCounts, subCount)
				} else {
					w.subBytes[numSubs] = lastSuffixLeadLabel
					w.subTermCounts[numSubs] = termCount
					w.subSubCounts[numSubs] = subCount
				}
				lastSuffixLeadLabel = suffixLeadLabel
				termCount, subCount = 0, 0
				numSubs++
			}

			if ent.isTerm() {
				termCount++
			} else {
				subCount++
			}
		}

		if len(w.subBytes) == numSubs {
			w.subBytes = append(w.subBytes, lastSuffixLeadLabel)
			w.subTermCounts = append(w.subTermCounts, termCount)
			w.subSubCounts = append(w.subSubCounts, subCount)
		} else {
			w.subBytes[numSubs] = lastSuffixLeadLabel
			w.subTermCounts[numSubs] = termCount
			w.subSubCounts[numSubs] = subCount
		}
		numSubs++

		for len(w.subTermCountSums) < numSubs {
			w.subTermCountSums = append(w.subTermCountSums, 0)
		}

		// Roll up (backwards) the termCounts; postings impl needs this
		// to know where to pull the term slice from its pending term
		// stack:
		sum := 0
		for idx := numSubs - 1; idx >= 0; idx-- {
			sum += w.subTermCounts[idx]
			w.subTermCountSums[idx] = sum
		}

		// Naive greedy segmentation; this is not always best (it can
		// produce a too-small block as the last block):
		pendingCount := 0
		startLabel := w.subBytes[0]
		curStart := count
		subCount = 0

		var floorBlocks []*PendingBlock
		var firstBlock *PendingBlock

		for sub := 0; sub < numSubs; sub++ {
			pendingCount += w.subTermCounts[sub] + w.subSubCounts[sub]
			// fmt.Printf("  %v\n", w.subTermCounts[sub]+w.subSubCounts[sub])
			subCount++

			// Greedily make a floor block as soon as we've crossed the min count
			if pendingCount >= w.owner.minItemsInBlock {
				var curPrefixLength int
				if startLabel == -1 {
					curPrefixLength = prefixLength
				} else {
					curPrefixLength = prefixLength + 1
					// floor term:
					prevTerm.Ints[prevTerm.Offset+prefixLength] = startLabel
				}
				// fmt.Printf("  %v subs\n", subCount)
				var floorBlock *PendingBlock
				if floorBlock, err = w.writeBlock(prevTerm, prefixLength,
					curPrefixLength, curStart, pendingCount,
					w.subTermCountSums[sub+1], true, startLabel,
					curStart == pendingCount); err != nil {
					return err
				}
				if firstBlock == nil {
					firstBlock = floorBlock
				} else {
					floorBlocks = append(floorBlocks, floorBlock)
				}
				curStart -= pendingCount
				// fmt.Printf("  floor=%v\n", pendingCount)
				pendingCount = 0

				assert2(w.owner.minItemsInBlock == 1 || subCount > 1,
					"minItemsInBlock=%v subCount=%v sub=%v of %v subTermCount=%v subSubCount=%v depth=%v",
					w.owner.minItemsInBlock, subCount, sub, numSubs,
					w.subTermCountSums[sub], w.subSubCounts[sub], prefixLength)
				subCount = 0
				startLabel = w.subBytes[sub+1]

				if curStart == 0 {
					break
				}

				if curStart <= w.owner.maxItemsInBlock {
					// remainder is small enough to fit into a block. NOTE that
					// this may be too small (< minItemsInBlock); need a true
					// segmenter here
					assert(startLabel != -1)
					assert(firstBlock != nil)
					prevTerm.Ints[prevTerm.Offset+prefixLength] = startLabel
					// fmt.Printf("  final %v subs\n", numSubs-sub-1)
					var b *PendingBlock
					if b, err = w.writeBlock(prevTerm, prefixLength, prefixLength+1,
						curStart, curStart, 0, true, startLabel, true); err != nil {
						return err
					}
					floorBlocks = append(floorBlocks, b)
					break
				}
			}
		}

		prevTerm.Ints[prevTerm.Offset+prefixLength] = savLabel

		assert(firstBlock != nil)
		if err = firstBlock.compileIndex(floorBlocks, w.owner.scratchBytes); err != nil {
			return err
		}

		w.pending = append(w.pending, firstBlock)
		// fmt.Printf("  done len(pending)=%v\n", len(w.pending))
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

	assert2(start >= 0 && start+length <= len(w.pending),
		"len(pending)=%v startBackward=%v length=%v",
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

	var isLeafBlock bool
	if w.lastBlockIndex < start {
		// this block defintely does not contain sub-blocks:
		isLeafBlock = true
		// fmt.Printf("no scan true isFloor=%v\n", isFloor)
	} else if !isFloor {
		// this block definitely does contain at least one sub-block:
		isLeafBlock = false
		// fmt.Printf("no scan false %v vs start=%v len=%v\n", w.lastBlockIndex, start, length)
	} else {
		// must scan up-front to see if there is a sub-block
		v := true
		// fmt.Printf("scan %v vs start=%v len=%v\n", w.lastBlockIndex, start, length)
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

	assert(w.longsSize > 0)
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
			if err := w.statsWriter.WriteVInt(int32(state.DocFreq)); err != nil {
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
		subIndices = nil
		termCount = 0
		for _, ent := range slice {
			if ent.isTerm() {
				term := ent.(*PendingTerm)
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
				if err = w.owner.postingsWriter.EncodeTerm(longs, w.bytesWriter, w.fieldInfo, state, absolute); err != nil {
					return nil, err
				}
				for pos := 0; pos < w.longsSize; pos++ {
					assert(longs[pos] >= 0)
					if err = w.metaWriter.WriteVLong(longs[pos]); err != nil {
						return nil, err
					}
				}
				if err = w.bytesWriter.WriteTo(w.metaWriter); err != nil {
					return nil, err
				}
				w.bytesWriter.Reset()
				absolute = false

				termCount++

			} else {
				block := ent.(*PendingBlock)
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
			(map[bool]int32{true: 1, false: 0}[isLeafBlock])); err == nil {
		if err = w.suffixWriter.WriteTo(w.owner.out); err == nil {
			w.suffixWriter.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

	// write term stats []byte blob
	if err = w.owner.out.WriteVInt(int32(w.statsWriter.FilePointer())); err == nil {
		if err = w.statsWriter.WriteTo(w.owner.out); err == nil {
			w.statsWriter.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

	// Write term meta data []byte blob
	if err = w.owner.out.WriteVInt(int32(w.metaWriter.FilePointer())); err == nil {
		if err = w.metaWriter.WriteTo(w.owner.out); err == nil {
			w.metaWriter.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

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
	// fmt.Printf("BTTW.finishTerm term=%v:%v seg=%v df=%v\n",
	// w.fieldInfo.Name, utf8ToString(text), w.owner.segment, stats.DocFreq)

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
		// fmt.Printf("  write FST %v field=%v\n", w.indexStartFP, w.fieldInfo.Name)

		w.owner.fields = append(w.owner.fields, newFieldMetaData(
			w.fieldInfo,
			w.pending[0].(*PendingBlock).index.EmptyOutput().([]byte),
			w.numTerms,
			w.indexStartFP,
			sumTotalTermFreq,
			sumDocFreq,
			docCount,
			w.longsSize,
			w.minTerm, w.maxTerm))
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
