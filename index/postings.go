package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
	"sort"
)

type FieldsProducer interface {
	Fields
	io.Closer
}

// TODO use for BTTR only

const (
	BTT_OUTPUT_FLAGS_NUM_BITS = 2
	BTT_EXTENSION             = "tim"
	BTT_CODEC_NAME            = "BLOCK_TREE_TERMS_DICT"
	BTT_VERSION_START         = 0
	BTT_VERSION_APPEND_ONLY   = 1
	BTT_VERSION_CURRENT       = BTT_VERSION_APPEND_ONLY

	BTT_INDEX_EXTENSION           = "tip"
	BTT_INDEX_CODEC_NAME          = "BLOCK_TREE_TERMS_INDEX"
	BTT_INDEX_VERSION_START       = 0
	BTT_INDEX_VERSION_APPEND_ONLY = 1
	BTT_INDEX_VERSION_CURRENT     = BTT_INDEX_VERSION_APPEND_ONLY
)

type BlockTreeTermsReader struct {
	in             store.IndexInput
	postingsReader PostingsReaderBase
	fields         map[string]FieldReader
	dirOffset      int64
	indexDirOffset int64
	segment        string
	version        int
}

func newBlockTreeTermsReader(dir store.Directory, fieldInfos FieldInfos, info SegmentInfo,
	postingsReader PostingsReaderBase, ctx store.IOContext,
	segmentSuffix string, indexDivisor int) (p FieldsProducer, err error) {
	fp := &BlockTreeTermsReader{postingsReader: postingsReader, segment: info.name}
	fp.in, err = dir.OpenInput(util.SegmentFileName(info.name, segmentSuffix, BTT_EXTENSION), ctx)
	if err != nil {
		return fp, err
	}

	success := false
	var indexIn store.IndexInput
	defer func() {
		if !success {
			// this.close() will close in:
			util.CloseWhileSuppressingError(indexIn, fp)
		}
	}()

	fp.version, err = fp.readHeader(fp.in)
	if err != nil {
		return fp, err
	}

	if indexDivisor != -1 {
		indexIn, err = dir.OpenInput(util.SegmentFileName(info.name, segmentSuffix, BTT_INDEX_EXTENSION), ctx)
		if err != nil {
			return fp, err
		}

		indexVersion, err := fp.readIndexHeader(indexIn)
		if err != nil {
			return fp, err
		}
		if int(indexVersion) != fp.version {
			return fp, errors.New(fmt.Sprintf("mixmatched version files: %v=%v,%v=%v", fp.in, fp.version, indexIn, indexVersion))
		}

		// Have PostingsReader init itself
		postingsReader.init(fp.in)

		// Read per-field details
		fp.seekDir(fp.in, fp.dirOffset)
		if indexDivisor != -1 {
			fp.seekDir(indexIn, fp.indexDirOffset)
		}

		numFields, err := fp.in.ReadVInt()
		if err != nil {
			return fp, err
		}
		if numFields < 0 {
			return fp, errors.New(fmt.Sprintf("invalid numFields: %v (resource=%v)", numFields, fp.in))
		}

		for i := int32(0); i < numFields; i++ {
			if field, err := fp.in.ReadVInt(); err == nil {
				if numTerms, err := fp.in.ReadVLong(); err == nil {
					// assert numTerms >= 0
					if numBytes, err := fp.in.ReadVInt(); err == nil {
						rootCode := make([]byte, numBytes)
						if err = fp.in.ReadBytes(rootCode); err == nil {
							fieldInfo := fieldInfos.byNumber[field]
							// assert fieldInfo != nil
							var sumTotalTermFreq int64
							if fieldInfo.indexOptions == INDEX_OPT_DOCS_ONLY {
								sumTotalTermFreq = -1
							} else {
								sumTotalTermFreq, err = fp.in.ReadVLong()
							}
							if err == nil {
								if sumDocFreq, err := fp.in.ReadVLong(); err == nil {
									if docCount, err := fp.in.ReadVInt(); err == nil {
										if docCount < 0 || docCount > info.docCount { // #docs with field must be <= #docs
											return fp, errors.New(fmt.Sprintf(
												"invalid docCount: %v maxDoc: %v (resource=%v)",
												docCount, info.docCount, fp.in))
										}
										if sumDocFreq < int64(docCount) { // #postings must be >= #docs with field
											return fp, errors.New(fmt.Sprintf(
												"invalid sumDocFreq: %v docCount: %v (resource=%v)",
												sumDocFreq, docCount, fp.in))
										}
										if sumTotalTermFreq != -1 && sumTotalTermFreq < sumDocFreq { // #positions must be >= #postings
											return fp, errors.New(fmt.Sprintf(
												"invalid sumTotalTermFreq: %v sumDocFreq: %v (resource=%v)",
												sumTotalTermFreq, sumDocFreq, fp.in))
										}

										var indexStartFP int64
										if indexDivisor != -1 {
											indexStartFP, err = indexIn.ReadVLong()
											if err != nil {
												return fp, err
											}
										}
										if _, ok := fp.fields[fieldInfo.name]; ok {
											return fp, errors.New(fmt.Sprintf(
												"duplicate field: %v (resource=%v)", fieldInfo.name, fp.in))
										}
										fp.fields[fieldInfo.name], err = newFieldReader(
											fieldInfo, numTerms, rootCode, sumTotalTermFreq,
											sumDocFreq, docCount, indexStartFP, indexIn)
									}
								}
							}
						}
					}
				}
			}
			if err != nil {
				return fp, err
			}
		}

		if indexDivisor != -1 {
			err = indexIn.Close()
			if err != nil {
				return fp, err
			}
		}

		success = true
	}

	return fp, nil
}

func asInt(n int32, err error) (n2 int, err2 error) {
	return int(n), err
}

func (r *BlockTreeTermsReader) readHeader(input store.IndexInput) (version int, err error) {
	version, err = asInt(codec.CheckHeader(input, BTT_CODEC_NAME, BTT_VERSION_START, BTT_VERSION_CURRENT))
	if err != nil {
		return int(version), err
	}
	if version < BTT_VERSION_APPEND_ONLY {
		r.dirOffset, err = input.ReadLong()
		if err != nil {
			return int(version), err
		}
	}
	return int(version), nil
}

func (r *BlockTreeTermsReader) readIndexHeader(input store.IndexInput) (version int, err error) {
	version, err = asInt(codec.CheckHeader(input, BTT_INDEX_CODEC_NAME, BTT_INDEX_VERSION_START, BTT_INDEX_VERSION_CURRENT))
	if err != nil {
		return version, err
	}
	if version < BTT_INDEX_VERSION_APPEND_ONLY {
		r.indexDirOffset, err = input.ReadLong()
		if err != nil {
			return version, err
		}
	}
	return version, nil
}

func (r *BlockTreeTermsReader) seekDir(input store.IndexInput, dirOffset int64) error {
	if r.version >= BTT_INDEX_VERSION_APPEND_ONLY {
		input.Seek(input.Length() - 8)
		dirOffset, err := input.ReadLong()
		if err != nil {
			return err
		}
	}
	input.Seek(dirOffset)
	return nil
}

func (r *BlockTreeTermsReader) Terms(field string) Terms {
	ans := r.fields[field]
	return &ans
}

func (r *BlockTreeTermsReader) Close() error {
	defer func() {
		// Clear so refs to terms index is GCable even if
		// app hangs onto us:
		r.fields = make(map[string]FieldReader)
	}()
	return util.Close(r.in, r.postingsReader)
}

type FieldReader struct {
	numTerms         int64
	fieldInfo        FieldInfo
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int32
	indexStartFP     int64
	rootBlockFP      int64
	rootCode         []byte
	index            *util.FST
}

func newFieldReader(fieldInfo FieldInfo, numTerms int64, rootCode []byte,
	sumTotalTermFreq, sumDocFreq int64, docCount int32, indexStartFP int64,
	indexIn store.IndexInput) (r FieldReader, err error) {
	// assert numTerms > 0
	r = FieldReader{
		fieldInfo:        fieldInfo,
		numTerms:         numTerms,
		sumTotalTermFreq: sumTotalTermFreq,
		sumDocFreq:       sumDocFreq,
		docCount:         docCount,
		indexStartFP:     indexStartFP,
		rootCode:         rootCode}

	in := store.NewByteArrayDataInput(rootCode)
	n, err := in.ReadVLong()
	if err != nil {
		return r, err
	}
	r.rootBlockFP = int64(uint64(n) >> BTT_OUTPUT_FLAGS_NUM_BITS)

	if indexIn != nil {
		clone := indexIn.Clone()
		clone.Seek(indexStartFP)
		r.index, err = util.LoadFST(clone, util.ByteSequenceOutputsSingleton())
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

type SegmentTermsEnum struct {
	*TermsEnumImpl
	owner *FieldReader
	eof   bool
	term  []byte
}

func newSegmentTermsEnum(r *FieldReader) *SegmentTermsEnum {
	return &SegmentTermsEnum{owner: r}
}

func (e *SegmentTermsEnum) Comparator() sort.Interface {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) SeekExactUsingCache(target []byte, useCache bool) bool {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) SeekCeilUsingCache(text []byte, useCache bool) SeekStatus {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) Next() (buf []byte, err error) {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) Term() []byte {
	if e.eof {
		panic("assertion error")
	}
	return e.term
}

func (e *SegmentTermsEnum) DocFreq() int {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) TotalTermFreq() int64 {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) DocsByFlags(skipDocs util.Bits, reuse DocsEnum, flags int) DocsEnum {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) DocsAndPositionsByFlags(skipDocs util.Bits, reuse DocsAndPositionsEnum, flags int) DocsAndPositionsEnum {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) SeekExactFromLast(target []byte, otherState TermState) error {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) TermState() TermState {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) SeekExactByPosition(ord int64) error {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) Ord() int64 {
	panic("not supported!")
}

type PostingsReaderBase interface {
	io.Closer
	init(termsIn store.IndexInput) error
	// newTermState() BlockTermState
	// nextTerm(fieldInfo FieldInfo, state BlockTermState)
	// docs(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits, reuse DocsEnum, flags int)
	// docsAndPositions(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits)
}
