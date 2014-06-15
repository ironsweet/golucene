package index

import (
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/index/model"
)

// codec/BlockTreeTermsWriter.java

/* Suggested degault value for the minItemsInBlock parameter. */
const DEFAULT_MIN_BLOCK_SIZE = 25

/* Suggested default value for the maxItemsInBlock parameter. */
const DEFAULT_MAX_BLOCK_SIZE = 48

type BlockTreeTermsWriter struct {
}

/*
Create a new writer. The number of items (terms or sub-blocks) per
block will aim tobe between minItermsPerBlock and maxItemsPerBlock,
though in some cases, the blocks may be smaller than the min.
*/
func NewBlockTreeTermsWriter(state SegmentWriteState,
	postingsWriter codec.PostingsWriterBase,
	minItemsInBlock, maxItemsInBlock int) (*BlockTreeTermsWriter, error) {
	panic("not implemented yet")
}

func (w *BlockTreeTermsWriter) addField(field *model.FieldInfo) (TermsConsumer, error) {
	panic("not implemented yet")
}

func (w *BlockTreeTermsWriter) Close() error {
	panic("not implemented yet")
}
