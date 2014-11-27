package lucene41

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/codec/blocktree"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

func init() {
	RegisterPostingsFormat(NewLucene41PostingsFormat())
}

// codecs/lucene41/Lucene41PostingsFormat.java

const (
	LUCENE41_DOC_EXTENSION = "doc"
	LUCENE41_POS_EXTENSION = "pos"
	LUCENE41_PAY_EXTENSION = "pay"

	LUCENE41_BLOCK_SIZE = 128
)

type Lucene41PostingsFormat struct {
	minTermBlockSize int
	maxTermBlockSize int
}

/* Creates Lucene41PostingsFormat wit hdefault settings. */
func NewLucene41PostingsFormat() *Lucene41PostingsFormat {
	return NewLucene41PostingsFormatWith(blocktree.DEFAULT_MIN_BLOCK_SIZE, blocktree.DEFAULT_MAX_BLOCK_SIZE)
}

/*
Creates Lucene41PostingsFormat with custom values for minBlockSize
and maxBlockSize passed to block terms directory.
*/
func NewLucene41PostingsFormatWith(minTermBlockSize, maxTermBlockSize int) *Lucene41PostingsFormat {
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
	ret, err := blocktree.NewBlockTreeTermsWriter(state, postingsWriter, f.minTermBlockSize, f.maxTermBlockSize)
	if err != nil {
		return nil, err
	}
	success = true
	return ret, nil
}

func (f *Lucene41PostingsFormat) FieldsProducer(state SegmentReadState) (FieldsProducer, error) {
	postingsReader, err := NewLucene41PostingsReader(state.Dir,
		state.FieldInfos,
		state.SegmentInfo,
		state.Context,
		state.SegmentSuffix)
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			fmt.Printf("Failed to load FieldsProducer for %v.\n", f.Name())
			util.CloseWhileSuppressingError(postingsReader)
		}
	}()

	fp, err := blocktree.NewBlockTreeTermsReader(state.Dir,
		state.FieldInfos,
		state.SegmentInfo,
		postingsReader,
		state.Context,
		state.SegmentSuffix,
		state.TermsIndexDivisor)
	if err != nil {
		return fp, err
	}
	success = true
	return fp, nil
}
