package compressing

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
)

// codec/compressing/CompressingStoredFieldsIndexReader.java

// Random-access reader for CompressingStoredFieldsIndexWriter
type CompressingStoredFieldsIndexReader struct {
	maxDoc              int
	docBases            []int
	startPointers       []int64
	avgChunkDocs        []int
	avgChunkSizes       []int64
	docBasesDeltas      []packed.PackedIntsReader
	startPointersDeltas []packed.PackedIntsReader
}

func newCompressingStoredFieldsIndexReader(fieldsIndexIn store.IndexInput,
	si *model.SegmentInfo) (r *CompressingStoredFieldsIndexReader, err error) {

	r = &CompressingStoredFieldsIndexReader{}
	r.maxDoc = si.DocCount()
	r.docBases = make([]int, 0, 16)
	r.startPointers = make([]int64, 0, 16)
	r.avgChunkDocs = make([]int, 0, 16)
	r.avgChunkSizes = make([]int64, 0, 16)
	r.docBasesDeltas = make([]packed.PackedIntsReader, 0, 16)
	r.startPointersDeltas = make([]packed.PackedIntsReader, 0, 16)

	packedIntsVersion, err := fieldsIndexIn.ReadVInt()
	if err != nil {
		return nil, err
	}

	for blockCount := 0; ; blockCount++ {
		numChunks, err := fieldsIndexIn.ReadVInt()
		if err != nil {
			return nil, err
		}
		if numChunks == 0 {
			break
		}

		{ // doc bases
			n, err := fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			r.docBases = append(r.docBases, int(n))
			n, err = fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			r.avgChunkDocs = append(r.avgChunkDocs, int(n))
			bitsPerDocBase, err := fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			if bitsPerDocBase > 32 {
				return nil, errors.New(fmt.Sprintf("Corrupted bitsPerDocBase (resource=%v)", fieldsIndexIn))
			}
			pr, err := packed.ReaderNoHeader(fieldsIndexIn, packed.PACKED, packedIntsVersion, numChunks, uint32(bitsPerDocBase))
			if err != nil {
				return nil, err
			}
			r.docBasesDeltas = append(r.docBasesDeltas, pr)
		}

		{ // start pointers
			n, err := fieldsIndexIn.ReadVLong()
			if err != nil {
				return nil, err
			}
			r.startPointers = append(r.startPointers, n)
			n, err = fieldsIndexIn.ReadVLong()
			if err != nil {
				return nil, err
			}
			r.avgChunkSizes = append(r.avgChunkSizes, n)
			bitsPerStartPointer, err := fieldsIndexIn.ReadVInt()
			if err != nil {
				return nil, err
			}
			if bitsPerStartPointer > 64 {
				return nil, errors.New(fmt.Sprintf("Corrupted bitsPerStartPonter (resource=%v)", fieldsIndexIn))
			}
			pr, err := packed.ReaderNoHeader(fieldsIndexIn, packed.PACKED, packedIntsVersion, numChunks, uint32(bitsPerStartPointer))
			if err != nil {
				return nil, err
			}
			r.startPointersDeltas = append(r.startPointersDeltas, pr)
		}
	}

	return r, nil
}

func (r *CompressingStoredFieldsIndexReader) block(docID int) int {
	lo, hi := 0, len(r.docBases)-1
	for lo <= hi {
		mid := int(uint(lo+hi) >> 1)
		midValue := r.docBases[mid]
		if midValue == docID {
			return mid
		} else if midValue < docID {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return hi
}

func (r *CompressingStoredFieldsIndexReader) relativeDocBase(block, relativeChunk int) int {
	expected := r.avgChunkDocs[block] * relativeChunk
	delta := util.ZigZagDecodeLong(r.docBasesDeltas[block].Get(relativeChunk))
	return expected + int(delta)
}

func (r *CompressingStoredFieldsIndexReader) relativeStartPointer(block, relativeChunk int) int64 {
	expected := r.avgChunkSizes[block] * int64(relativeChunk)
	delta := util.ZigZagDecodeLong(r.startPointersDeltas[block].Get(relativeChunk))
	return expected + delta
}

func (r *CompressingStoredFieldsIndexReader) relativeChunk(block, relativeDoc int) int {
	lo, hi := 0, int(r.docBasesDeltas[block].Size())-1
	for lo <= hi {
		mid := int(uint(lo+hi) >> 1)
		midValue := r.relativeDocBase(block, mid)
		if midValue == relativeDoc {
			return mid
		} else if midValue < relativeDoc {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return hi
}

func (r *CompressingStoredFieldsIndexReader) startPointer(docID int) int64 {
	if docID < 0 || docID >= r.maxDoc {
		panic(fmt.Sprintf("docID out of range [0-%v]: %v", r.maxDoc, docID))
	}
	block := r.block(docID)
	relativeChunk := r.relativeChunk(block, docID-r.docBases[block])
	return r.startPointers[block] + r.relativeStartPointer(block, relativeChunk)
}

func (r *CompressingStoredFieldsIndexReader) Clone() *CompressingStoredFieldsIndexReader {
	return r
}
