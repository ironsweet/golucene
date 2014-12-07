package blocktree

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// BlockTreeTermsReader.java

const (
	BTT_OUTPUT_FLAGS_NUM_BITS = 2
	BTT_OUTPUT_FLAG_IS_FLOOR  = 1
	BTT_OUTPUT_FLAG_HAS_TERMS = 2

	// BTT_INDEX_EXTENSION           = "tip"
	// BTT_INDEX_CODEC_NAME          = "BLOCK_TREE_TERMS_INDEX"
	// BTT_INDEX_VERSION_START       = 0
	// BTT_INDEX_VERSION_APPEND_ONLY = 1
	// BTT_INDEX_VERSION_CURRENT     = BTT_INDEX_VERSION_APPEND_ONLY
)

/* A block-based terms index and dictionary that assigns
terms to variable length blocks according to how they
share prefixes. The terms index is a prefix trie
whose leaves are term blocks. The advantage of this
approach is that seekExact is often able to
determine a term cannot exist without doing any IO, and
intersection with Automata is very fast. NOte that this
terms dictionary has its own fixed terms index (ie, it
does not support a pluggable terms index
implementation).

NOTE: this terms dictionary does not support
index divisor when opening an IndexReader. Instead, you
can change the min/maxItemsPerBlock during indexing.

The data strucure used by this implementation is very
similar to a [burst trie]
(http://citeseer.ist.psu.edu/viewdoc/summary?doi=10.1.1.18.3499),
but with added logic to break up too-large blocks of all
terms sharing a given prefix into smaller ones.

Use CheckIndex with the -verbose
option to see summary statistics on the blocks in the
dictionary. */
type BlockTreeTermsReader struct {
	// Open input to the main terms dict file (_X.tib)
	in store.IndexInput
	// Reads the terms dict entries, to gather state to
	// produce DocsEnum on demand
	postingsReader PostingsReaderBase
	fields         map[string]FieldReader
	// File offset where the directory starts in the terms file.
	dirOffset int64
	// File offset where the directory starts in the index file.
	indexDirOffset int64
	segment        string
	version        int
}

func NewBlockTreeTermsReader(dir store.Directory,
	fieldInfos FieldInfos, info *SegmentInfo,
	postingsReader PostingsReaderBase, ctx store.IOContext,
	segmentSuffix string, indexDivisor int) (p FieldsProducer, err error) {

	// log.Print("Initializing BlockTreeTermsReader...")
	fp := &BlockTreeTermsReader{
		postingsReader: postingsReader,
		fields:         make(map[string]FieldReader),
		segment:        info.Name,
	}
	fp.in, err = dir.OpenInput(util.SegmentFileName(info.Name, segmentSuffix, TERMS_EXTENSION), ctx)
	if err != nil {
		return nil, err
	}

	success := false
	var indexIn store.IndexInput
	defer func() {
		if !success {
			fmt.Println("Failed to initialize BlockTreeTermsReader.")
			if err != nil {
				fmt.Println("DEBUG ", err)
			}
			// this.close() will close in:
			util.CloseWhileSuppressingError(indexIn, fp)
		}
	}()

	fp.version, err = fp.readHeader(fp.in)
	if err != nil {
		return nil, err
	}
	// log.Printf("Version: %v", fp.version)

	if indexDivisor != -1 {
		filename := util.SegmentFileName(info.Name, segmentSuffix, TERMS_INDEX_EXTENSION)
		indexIn, err = dir.OpenInput(filename, ctx)
		if err != nil {
			return nil, err
		}

		indexVersion, err := fp.readIndexHeader(indexIn)
		if err != nil {
			return nil, err
		}
		// log.Printf("Index version: %v", indexVersion)
		if int(indexVersion) != fp.version {
			return nil, errors.New(fmt.Sprintf("mixmatched version files: %v=%v,%v=%v", fp.in, fp.version, indexIn, indexVersion))
		}
	}

	// verify
	if indexIn != nil && fp.version >= TERMS_VERSION_CURRENT {
		if _, err = store.ChecksumEntireFile(indexIn); err != nil {
			return nil, err
		}
	}

	// Have PostingsReader init itself
	postingsReader.Init(fp.in)

	if fp.version >= TERMS_VERSION_CHECKSUM {
		// NOTE: data file is too costly to verify checksum against all the
		// bytes on open, but for now we at least verify proper structure
		// of the checksum footer: which looks for FOOTER_MAGIC +
		// algorithmID. This is cheap and can detect some forms of
		// corruption such as file trucation.
		if _, err = codec.RetrieveChecksum(fp.in); err != nil {
			return nil, err
		}
	}

	// Read per-field details
	fp.seekDir(fp.in, fp.dirOffset)
	if indexDivisor != -1 {
		fp.seekDir(indexIn, fp.indexDirOffset)
	}

	numFields, err := fp.in.ReadVInt()
	if err != nil {
		return nil, err
	}
	// log.Printf("Fields number: %v", numFields)
	if numFields < 0 {
		return nil, errors.New(fmt.Sprintf("invalid numFields: %v (resource=%v)", numFields, fp.in))
	}

	for i := int32(0); i < numFields; i++ {
		// log.Printf("Next field...")
		field, err := fp.in.ReadVInt()
		if err != nil {
			return nil, err
		}
		// log.Printf("Field: %v", field)

		numTerms, err := fp.in.ReadVLong()
		if err != nil {
			return nil, err
		}
		assert2(numTerms > 0,
			"Illegal numTerms for field number: %v (resource=%v)", field, fp.in)
		// log.Printf("Terms number: %v", numTerms)

		numBytes, err := fp.in.ReadVInt()
		if err != nil {
			return nil, err
		}
		assert2(numBytes >= 0,
			"invalid rootCode for field number: %v, numBytes=%v (resource=%v)",
			field, numBytes, fp.in)
		// log.Printf("Bytes number: %v", numBytes)

		rootCode := make([]byte, numBytes)
		err = fp.in.ReadBytes(rootCode)
		if err != nil {
			return nil, err
		}
		fieldInfo := fieldInfos.FieldInfoByNumber(int(field))
		assert2(fieldInfo != nil, "invalid field numebr: %v (resource=%v)", field, fp.in)
		var sumTotalTermFreq int64
		if fieldInfo.IndexOptions() == INDEX_OPT_DOCS_ONLY {
			sumTotalTermFreq = -1
		} else {
			sumTotalTermFreq, err = fp.in.ReadVLong()
			if err != nil {
				return nil, err
			}
		}
		sumDocFreq, err := fp.in.ReadVLong()
		if err != nil {
			return nil, err
		}
		var docCount int
		if docCount, err = asInt(fp.in.ReadVInt()); err != nil {
			return nil, err
		}
		// fmt.Printf("DocCount: %v\n", docCount)
		var longsSize int
		if fp.version >= TERMS_VERSION_META_ARRAY {
			if longsSize, err = asInt(fp.in.ReadVInt()); err != nil {
				return nil, err
			}
		}
		assert2(longsSize >= 0,
			"invalid longsSize for field: %v, longsSize=%v (resource=%v)",
			fieldInfo.Name, longsSize, fp.in)
		var minTerm, maxTerm []byte
		if fp.version >= TERMS_VERSION_MIN_MAX_TERMS {
			if minTerm, err = readBytesRef(fp.in); err != nil {
				return nil, err
			}
			if maxTerm, err = readBytesRef(fp.in); err != nil {
				return nil, err
			}
		}
		if docCount < 0 || int(docCount) > info.DocCount() { // #docs with field must be <= #docs
			return nil, errors.New(fmt.Sprintf(
				"invalid docCount: %v maxDoc: %v (resource=%v)",
				docCount, info.DocCount(), fp.in))
		}
		if sumDocFreq < int64(docCount) { // #postings must be >= #docs with field
			return nil, errors.New(fmt.Sprintf(
				"invalid sumDocFreq: %v docCount: %v (resource=%v)",
				sumDocFreq, docCount, fp.in))
		}
		if sumTotalTermFreq != -1 && sumTotalTermFreq < sumDocFreq { // #positions must be >= #postings
			return nil, errors.New(fmt.Sprintf(
				"invalid sumTotalTermFreq: %v sumDocFreq: %v (resource=%v)",
				sumTotalTermFreq, sumDocFreq, fp.in))
		}

		var indexStartFP int64
		if indexDivisor != -1 {
			if indexStartFP, err = indexIn.ReadVLong(); err != nil {
				return nil, err
			}
		}
		// log.Printf("indexStartFP: %v", indexStartFP)
		if _, ok := fp.fields[fieldInfo.Name]; ok {
			return nil, errors.New(fmt.Sprintf(
				"duplicate field: %v (resource=%v)", fieldInfo.Name, fp.in))
		}
		if fp.fields[fieldInfo.Name], err = newFieldReader(fp,
			fieldInfo, numTerms, rootCode, sumTotalTermFreq,
			sumDocFreq, docCount, indexStartFP, longsSize,
			indexIn, minTerm, maxTerm); err != nil {
			return nil, err
		}
	}

	if indexDivisor != -1 {
		if err = indexIn.Close(); err != nil {
			return nil, err
		}
	}

	success = true

	return fp, nil
}

func asInt(n int32, err error) (n2 int, err2 error) {
	return int(n), err
}

func readBytesRef(in store.IndexInput) ([]byte, error) {
	length, err := asInt(in.ReadVInt())
	if err != nil {
		return nil, err
	}
	bytes := make([]byte, length)
	if err = in.ReadBytes(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func (r *BlockTreeTermsReader) readHeader(input store.IndexInput) (version int, err error) {
	version, err = asInt(codec.CheckHeader(input, TERMS_CODEC_NAME, TERMS_VERSION_START, TERMS_VERSION_CURRENT))
	if err != nil {
		return int(version), err
	}
	if version < TERMS_VERSION_APPEND_ONLY {
		r.dirOffset, err = input.ReadLong()
		if err != nil {
			return int(version), err
		}
	}
	return int(version), nil
}

func (r *BlockTreeTermsReader) readIndexHeader(input store.IndexInput) (version int, err error) {
	version, err = asInt(codec.CheckHeader(input, TERMS_INDEX_CODEC_NAME, TERMS_VERSION_START, TERMS_VERSION_CURRENT))
	if err != nil {
		return version, err
	}
	if version < TERMS_VERSION_APPEND_ONLY {
		r.indexDirOffset, err = input.ReadLong()
		if err != nil {
			return version, err
		}
	}
	return version, nil
}

func (r *BlockTreeTermsReader) seekDir(input store.IndexInput, dirOffset int64) (err error) {
	// log.Printf("Seeking to: %v", dirOffset)
	if r.version >= TERMS_VERSION_CHECKSUM {
		if err = input.Seek(input.Length() - codec.FOOTER_LENGTH - 8); err != nil {
			return
		}
		if dirOffset, err = input.ReadLong(); err != nil {
			return
		}
	} else if r.version >= TERMS_VERSION_APPEND_ONLY {
		if err = input.Seek(input.Length() - 8); err != nil {
			return
		}
		if dirOffset, err = input.ReadLong(); err != nil {
			return
		}
	}
	return input.Seek(dirOffset)
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
