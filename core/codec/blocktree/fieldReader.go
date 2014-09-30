package blocktree

import (
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util/fst"
)

type FieldReader struct {
	numTerms         int64
	fieldInfo        *FieldInfo
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	indexStartFP     int64
	rootBlockFP      int64
	rootCode         []byte
	minTerm          []byte
	maxTerm          []byte
	longsSize        int
	parent           *BlockTreeTermsReader

	index *fst.FST
}

func newFieldReader(parent *BlockTreeTermsReader,
	fieldInfo *FieldInfo, numTerms int64, rootCode []byte,
	sumTotalTermFreq, sumDocFreq int64, docCount int,
	indexStartFP int64, longsSize int, indexIn store.IndexInput,
	minTerm, maxTerm []byte) (r FieldReader, err error) {

	// log.Print("Initializing FieldReader...")
	assert(numTerms > 0)
	r = FieldReader{
		parent:           parent,
		fieldInfo:        fieldInfo,
		numTerms:         numTerms,
		sumTotalTermFreq: sumTotalTermFreq,
		sumDocFreq:       sumDocFreq,
		docCount:         docCount,
		indexStartFP:     indexStartFP,
		rootCode:         rootCode,
		longsSize:        longsSize,
		minTerm:          minTerm,
		maxTerm:          maxTerm,
	}
	// log.Printf("BTTR: seg=%v field=%v rootBlockCode=%v divisor=",
	// 	parent.segment, fieldInfo.Name, rootCode)

	in := store.NewByteArrayDataInput(rootCode)
	n, err := in.ReadVLong()
	if err != nil {
		return r, err
	}
	r.rootBlockFP = int64(uint64(n) >> BTT_OUTPUT_FLAGS_NUM_BITS)

	if indexIn != nil {
		clone := indexIn.Clone()
		// log.Printf("start=%v field=%v", indexStartFP, fieldInfo.Name)
		clone.Seek(indexStartFP)
		r.index, err = fst.LoadFST(clone, fst.ByteSequenceOutputsSingleton())
	}

	return r, err
}

func (r *FieldReader) Iterator(reuse TermsEnum) TermsEnum {
	return newSegmentTermsEnum(r)
}

func (r *FieldReader) SumTotalTermFreq() int64 {
	return r.sumTotalTermFreq
}

func (r *FieldReader) SumDocFreq() int64 {
	return r.sumDocFreq
}

func (r *FieldReader) DocCount() int {
	return int(r.docCount)
}
