package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
)

type FieldsProducer interface {
	Fields
	io.Closer
}

const (
	/* BlockTreeTermsWriter */

	BTT_OUTPUT_FLAGS_NUM_BITS = 2
	BTT_EXTENSION             = "tim"

	/* BlockTreeTermsReader */

	BTT_INDEX_EXTENSION = "tip"
)

type BlockTreeTermsReader struct {
	in             *store.IndexInput
	postingsReader PostingsReaderBase
	fields         map[string]FieldReader
	dirOffset      int64
	indexDirOffset int64
	segment        string
	version        int
}

func newBlockTreeTermsReader(dir *store.Directory, fieldInfos FieldInfos, info SegmentInfo,
	postingsReader PostingsReaderBase, ctx store.IOContext,
	segmentSuffix string, indexDivisor int) (p FieldsProducer, err error) {
	fp := &BlockTreeTermsReader{postingsReader: postingsReader, segment: info.name}
	fp.in, err = dir.OpenInput(util.SegmentFileName(info.name, segmentSuffix, BTT_EXTENSION), ctx)
	if err != nil {
		return fp, err
	}

	success := false
	var indexIn *store.IndexInput
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
		if indexVersion != fp.version {
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
						}
					}
				}
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
					}
				}
			}
			if err != nil {
				return fp, err
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
			fp.fields[fieldInfo.name] = newFieldReader(
				fieldInfo, numTerms, rootCode, sumTotalTermFreq,
				sumDocFreq, docCount, indexStartFP, indexIn)
		}

		if indexDivisor != -1 {
			err = indexIn.Close()
			if err != nil {
				return fp, err
			}
		}

		success = true
	}
}

func (r *BlockTreeTermsReader) readHeader(input *store.IndexInput) (version int, err error) {
	version, err = store.CheckHeader(input, BTT_CODEC_NAME, BTT_VERSION_START, BTT_VERSION_CURRENT)
	if err != nil {
		return version, err
	}
	if version < BTT_VERSION_APPEND_ONLY {
		r.dirOffset, err = input.ReadLong()
		if err != nil {
			return version, err
		}
	}
	return version, nil
}

func (r *BlockTreeTermsReader) readIndexHeader(input *store.IndexInput) (version int, err error) {
	version, err = store.CheckHeader(input, BTT_INDEX_CODEC_NAME, BTT_INDEX_VERSION_START, BTT_INDEX_VERSION_CURRENT)
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

func (r *BlockTreeTermsReader) seekDir(input *store.IndexInput, dirOffset int64) error {
	if r.version >= BTT_INDEX_VERSION_APPEND_ONLY {
		input.Seek(input.Length() - 8)
		dirOffset, err := input.ReadLong()
		if err != nil {
			return err
		}
	}
	input.Seek(dirOffset)
}

func (r *BlockTreeTermsReader) Terms(field string) Terms {

}

func (r *BlockTreeTermsReader) Close() error {

}

type FieldReader struct {
	numTerms         int64
	fieldInfo        FieldInfo
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	indexStartFP     int64
	rootBlockFP      int64
	rootCode         []byte
	index            *FST
}

func newFieldReader(fieldInfo FieldInfo, numTerms int64, rootCode []byte,
	sumTotalTermFreq, sumDocFreq int64, docCount int, indexStartFP int64, indexIn *store.IndexInput) {
	// assert numTerms > 0
	self := FieldReader{
		fieldInfo:        fieldInfo,
		numTerms:         numTerms,
		sumTotalTermFreq: sumTotalTermFreq,
		sumDocFreq:       sumDocFreq,
		docCount:         docCount,
		indexStartFP:     indexStartFP,
		rootCode:         rootCode}

	self.rootBlockFP = uint64(newByteArrayDataInput(rootCode.bytes, rootCode.offset, rootCode.length).ReadVLong()) >> BTT_OUTPUT_FLAGS_NUM_BITS

	if indexIn != nil {
		clone = indexIn.Clone()
		clone.Seek(indexStartFP)
		self.index = loadFST(clone, ByteSequenceOutputs.getSingleton())
	} // else self.index = nil
}

type PostingsReaderBase interface {
	io.Closer
	init(termsIn *store.IndexInput) error
	// newTermState() BlockTermState
	// nextTerm(fieldInfo FieldInfo, state BlockTermState)
	// docs(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits, reuse DocsEnum, flags int)
	// docsAndPositions(fieldInfo FieldInfo, state BlockTermState, skipDocs util.Bits)
}
