package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
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
	Start(store.IndexOutput) error
	// Start a new term. Note that a matching call to finishTerm() is
	// done, only if the term has at least one document.
	StartTerm() error
	// Finishes the current term. The provided TermStats contains the
	// term's summary statistics.
	FinishTerm(stats *codec.TermStats) error
	// Called when the writing switches to another field.
	SetField(fieldInfo *model.FieldInfo)
	// Flush count terms starting at start "backwards", as a block.
	// start is a negative offset from the end of the terms stack, ie
	// bigger start means further back in the stack.
	flushTermsBlock(start, count int) error
}

// codec/BlockTreeTermsWriter.java

/* Suggested degault value for the minItemsInBlock parameter. */
const DEFAULT_MIN_BLOCK_SIZE = 25

/* Suggested default value for the maxItemsInBlock parameter. */
const DEFAULT_MAX_BLOCK_SIZE = 48

/* Extension of terms file */
const TERMS_EXTENSION = "tim"
const TERMS_CODEC_NAME = "BLOCK_TREE_TERMS_DICT"

/* Append-only */
const TERMS_VERSION_APPEND_ONLY = 1

/* Current terms format. */
const TERMS_VERSION_CURRENT = TERMS_VERSION_APPEND_ONLY

/* Extension of terms index file */
const TERMS_INDEX_EXTENSION = "tip"
const TERMS_INDEX_CODEC_NAME = "BLOCK_TREE_TERMS_INDEX"

/* Append-only */
const TERMS_INDEX_VERSION_APPEND_ONLY = 1

/* Current terms format. */
const TERMS_INDEX_VERSION_CURRENT = TERMS_INDEX_VERSION_APPEND_ONLY

type BlockTreeTermsWriterSPI interface {
	WriteHeader(store.IndexOutput) error
	WriteIndexHeader(store.IndexOutput) error
}

type FieldMetaData struct {
}

func newFieldMetaData(fieldInfo *model.FieldInfo,
	rootCode []byte, numTerms, indexStartFP, sumTotalTermFreq, sumDocFreq int64,
	docCount int) *FieldMetaData {
	panic("not implemented yet")
}

type BlockTreeTermsWriter struct {
	spi             BlockTreeTermsWriterSPI
	out             store.IndexOutput
	indexOut        store.IndexOutput
	minItemsInBlock int
	maxItemsInBlock int

	postingsWriter PostingsWriterBase
	fieldInfos     model.FieldInfos
	currentField   *model.FieldInfo

	fields []*FieldMetaData

	scratchBytes *store.RAMOutputStream

	bytesWriter  *store.RAMOutputStream
	bytesWriter2 *store.RAMOutputStream
}

/*
Create a new writer. The number of items (terms or sub-blocks) per
block will aim tobe between minItermsPerBlock and maxItemsPerBlock,
though in some cases, the blocks may be smaller than the min.
*/
func NewBlockTreeTermsWriter(state SegmentWriteState,
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
		fieldInfos:      state.fieldInfos,
		minItemsInBlock: minItemsInBlock,
		maxItemsInBlock: maxItemsInBlock,
		postingsWriter:  postingsWriter,
		scratchBytes:    store.NewRAMOutputStreamBuffer(),
		bytesWriter:     store.NewRAMOutputStreamBuffer(),
		bytesWriter2:    store.NewRAMOutputStreamBuffer(),
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
		termsFileName := util.SegmentFileName(state.segmentInfo.Name, state.segmentSuffix, TERMS_EXTENSION)
		if out, err = state.directory.CreateOutput(termsFileName, state.context); err != nil {
			return err
		}
		if err = ans.spi.WriteHeader(out); err != nil {
			return err
		}

		termsIndexFileName := util.SegmentFileName(state.segmentInfo.Name, state.segmentSuffix, TERMS_INDEX_EXTENSION)
		if indexOut, err = state.directory.CreateOutput(termsIndexFileName, state.context); err != nil {
			return err
		}
		if err = ans.spi.WriteIndexHeader(indexOut); err != nil {
			return err
		}

		// have consumer write its format/header
		if err = postingsWriter.Start(out); err != nil {
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
	return codec.WriteHeader(out, TERMS_INDEX_CODEC_NAME, TERMS_INDEX_VERSION_CURRENT)
}

func (w *BlockTreeTermsWriter) addField(field *model.FieldInfo) (TermsConsumer, error) {
	assert(w.currentField == nil || w.currentField.Name < field.Name)
	w.currentField = field
	return newTermsWriter(w, field), nil
}

func (w *BlockTreeTermsWriter) Close() error {
	panic("not implemented yet")
}

type PendingEntry interface {
	isTerm() bool
}

type PendingTerm struct {
	term  []byte
	stats *codec.TermStats
}

func newPendingTerm(term []byte, stats *codec.TermStats) *PendingTerm {
	return &PendingTerm{term, stats}
}

func (t *PendingTerm) isTerm() bool { return true }

type PendingBlock struct {
	prefix []byte
	index  *fst.FST
}

func newPendingBlock(prefix []byte, fp int64, hasTerms, isFloor bool,
	floorLeadByte int, subIndices []*fst.FST) *PendingBlock {
	panic("not implemented yet")
}

func (b *PendingBlock) isTerm() bool { return false }

func (b *PendingBlock) String() string {
	return fmt.Sprintf("BLOCK: %v", utf8ToString(b.prefix))
}

func (b *PendingBlock) compileIndex(floorBlocks []*PendingBlock,
	sctrachBytes *store.RAMOutputStream) error {
	panic("not implemented yet")
}

type TermsWriter struct {
	owner            *BlockTreeTermsWriter
	fieldInfo        *model.FieldInfo
	numTerms         int64
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

	scratchIntsRef *util.IntsRef
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
blocks. Returns how many blocks were written. If the entry count is
<= maxItemsPerBlock we just write a single blocks; else we break into
primary (initial) block and then one or more following floor blocks:
*/
func (w *TermsWriter) writeBlocks(prevTerm *util.IntsRef, prefixLength, count int) error {
	if prefixLength == 0 || count <= w.owner.maxItemsInBlock {
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
	if isLeafBlock {
		panic("not implemented yet")
	} else {
		panic("not implemented yet")
	}

	// TODO: we could block-write the term suffix pointer
	// this would take more space but would enable binary
	// search on lookup

	// write suffixes []byte blob to terms dict output:
	err = w.owner.out.WriteVInt(int32(w.owner.bytesWriter.FilePointer()<<1) | (map[bool]int32{true: 1, false: 0}[isLeafBlock]))
	if err == nil {
		err = w.owner.bytesWriter.WriteTo(w.owner.out)
		if err != nil {
			w.owner.bytesWriter.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

	// write term stats []byte blob
	err = w.owner.out.WriteVInt(int32(w.owner.bytesWriter2.FilePointer()))
	if err == nil {
		err = w.owner.bytesWriter2.WriteTo(w.owner.out)
		if err != nil {
			w.owner.bytesWriter2.Reset()
		}
	}
	if err != nil {
		return nil, err
	}

	// have postings writer write block
	err = w.owner.postingsWriter.flushTermsBlock(futureTermCount+termCount, termCount)
	if err != nil {
		return nil, err
	}

	// remove slice replaced by block:
	slice = nil

	if w.lastBlockIndex >= start {
		if w.lastBlockIndex < start+length {
			w.lastBlockIndex = start
		} else {
			w.lastBlockIndex -= length
		}
	}

	return newPendingBlock(prefix, startFP, termCount != 0, isFloor, floorLeadByte, subIndices), nil
}

func (w *TermsWriter) comparator() func(a, b []byte) bool {
	return util.UTF8SortedAsUnicodeLess
}

func (w *TermsWriter) startTerm(text []byte) (codec.PostingsConsumer, error) {
	assert(w.owner != nil)
	assert(w.owner.postingsWriter != nil)
	err := w.owner.postingsWriter.StartTerm()
	return w.owner.postingsWriter, err
}

func (w *TermsWriter) finishTerm(text []byte, stats *codec.TermStats) error {
	assert(stats.DocFreq > 0)

	err := w.blockBuilder.Add(fst.ToIntsRef(text, w.scratchIntsRef), w.noOutputs.NoOutput())
	if err != nil {
		return err
	}
	w.pending = append(w.pending, newPendingTerm(util.DeepCopyOf(util.NewBytesRef(text)).Value, stats))
	err = w.owner.postingsWriter.FinishTerm(stats)
	w.numTerms++
	return err
}

func (w *TermsWriter) finish(sumTotalTermFreq, sumDocFreq int64, docCount int) error {
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
			w.pending[0].(*PendingBlock).index.EmptyOutput().([]byte),
			w.numTerms,
			w.indexStartFP,
			sumTotalTermFreq,
			sumDocFreq,
			docCount))
	} else {
		assert(sumTotalTermFreq == 0 || w.fieldInfo.IndexOptions() == model.INDEX_OPT_DOCS_ONLY && sumTotalTermFreq == -1)
		assert(sumDocFreq == 0)
		assert(docCount == 0)
	}
	return nil
}
