package index

import (
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
	// Called when the writing switches to another field.
	SetField(fieldInfo *model.FieldInfo)
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

type BlockTreeTermsWriter struct {
	spi             BlockTreeTermsWriterSPI
	out             store.IndexOutput
	indexOut        store.IndexOutput
	minItemsInBlock int
	maxItemsInBlock int

	postingsWriter PostingsWriterBase
	fieldInfos     model.FieldInfos
	currentField   *model.FieldInfo
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

type TermsWriter struct {
	owner     *BlockTreeTermsWriter
	fieldInfo *model.FieldInfo

	// Used only to partition terms into the block tree; we don't pull
	// an FST from this builder:
	noOutputs    *fst.NoOutputs
	blockBuilder *fst.Builder
}

func newTermsWriter(owner *BlockTreeTermsWriter,
	fieldInfo *model.FieldInfo) *TermsWriter {
	owner.postingsWriter.SetField(fieldInfo)
	return &TermsWriter{
		owner:     owner,
		fieldInfo: fieldInfo,
		noOutputs: fst.NO_OUTPUT,
		// This builder is just used transiently to fragment terms into
		// "good" blocks; we don't save the resulting FST:
		blockBuilder: fst.NewBuilder(
			fst.INPUT_TYPE_BYTE1, 0, 0, true, true,
			int(math.MaxInt32), fst.NO_OUTPUT,
			//Assign terms to blocks "naturally", ie, according to the number of
			//terms under a given prefix that we encounter:
			func(frontier []*fst.UnCompiledNode, prefixLenPlus1 int, lastInput []int) error {
				panic("not implemented yet")
			}, false, packed.PackedInts.COMPACT,
			true, 15),
	}
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
	panic("not implemented yet")
}

func (w *TermsWriter) finish(sumTotalTermFreq, sumDocFreq int64, docCount int) error {
	panic("not implemented yet")
}
