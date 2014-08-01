package compressing

import (
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"math"
)

/* number of chunks to serialize at once */
const BLOCK_SIZE = 1024

/*
Efficient index format for block-based Codecs.

This writer generates a file which be loaded into memory using
memory-efficient data structures to quickly locate the block that
contains any document.

In order to have a compact in-memory representation, for every block
of 1024 chunks, this index computes the average number of bytes per
chunk and for every chunk, only stores the difference between

- ${chunk number} * ${average length of a chunk}
- and the actual start offset of the chunk

Data is written as follows:

	- PackedIntsVersion, <Block>^BlockCount, BlocksEndMarker
	- PackedIntsVersion --> VERSION_CURRENT as a vint
	- BlocksEndMarker --> 0 as a vint, this marks the end of blocks since blocks are not allowed to start with 0
	- Block --> BlockChunks, <Docbases>, <StartPointers>
	- BlockChunks --> a vint which is the number of chunks encoded in the block
	- DocBases --> DocBase, AvgChunkDocs, BitsPerDocbaseDelta, DocBaseDeltas
	- DocBase --> first document ID of the block of chunks, as a vint
	- AvgChunkDocs --> average number of documents in a single chunk, as a vint
	- BitsPerDocBaseDelta --> number of bits required to represent a delta from the average using ZigZag encoding
	- DocBaseDeltas --> packed array of BlockChunks elements of BitsPerDocBaseDelta bits each, representing the deltas from the average doc base using ZigZag encoding.
	- StartPointers --> StartointerBase, AveChunkSize, BitsPerStartPointerDelta, StartPointerDeltas
	- StartPointerBase --> the first start ointer of the block, as a vint64
	- AvgChunkSize --> the average size of a chunk of compressed documents, as a vint64
	- BitsPerStartPointerDelta --> number of bits required to represent a delta from the average using ZigZag encoding
	- StartPointerDeltas --> packed array of BlockChunks elements of BitsPerStartPointerDelta bits each, representing the deltas from the average start pointer using ZigZag encoding

Notes

- For any block, the doc base of the n-th chunk can be restored with
DocBase + AvgChunkDocs * n + DOcBsaeDeltas[n].
- For any block, the start pointer of the n-th chunk can be restored
with StartPointerBase + AvgChunkSize * n + StartPointerDeltas[n].
- Once data is loaded into memory, you can lookup the start pointer
of any document by performing two binary searches: a first one based
on the values of DocBase in order to find the right block, and then
inside the block based on DocBaseDeltas (by reconstructing the doc
bases for every chunk).
*/
type StoredFieldsIndexWriter struct {
	fieldsIndexOut     store.IndexOutput
	totalDocs          int
	blockDocs          int
	blockChunks        int
	firstStartPointer  int64
	maxStartPointer    int64
	docBaseDeltas      []int
	startPointerDeltas []int64
}

func NewStoredFieldsIndexWriter(indexOutput store.IndexOutput) (*StoredFieldsIndexWriter, error) {
	err := indexOutput.WriteVInt(packed.VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	return &StoredFieldsIndexWriter{
		fieldsIndexOut:     indexOutput,
		blockChunks:        0,
		blockDocs:          0,
		firstStartPointer:  -1,
		totalDocs:          0,
		docBaseDeltas:      make([]int, BLOCK_SIZE),
		startPointerDeltas: make([]int64, BLOCK_SIZE),
	}, nil
}

func (w *StoredFieldsIndexWriter) reset() {
	w.blockChunks = 0
	w.blockDocs = 0
	w.firstStartPointer = -1 // means unset
}

func (w *StoredFieldsIndexWriter) writeBlock() error {
	assert(w.blockChunks > 0)
	err := w.fieldsIndexOut.WriteVInt(int32(w.blockChunks))
	if err != nil {
		return err
	}

	// The trick here is that we only store the difference from the
	// average start pointer or doc base, this helps save bits per
	// value. And in order to prevent a few chunks that would be far
	// from the average to raise the number of bits per value for all
	// of them, we only encode blocks of 1024 chunks at once.
	// See LUCENE-4512

	// doc bases
	var avgChunkDocs int
	if w.blockChunks == 1 {
		avgChunkDocs = 0
	} else {
		avgChunkDocs = int(math.Floor(float64(w.blockDocs-w.docBaseDeltas[w.blockChunks-1])/float64(w.blockChunks-1) + 0.5))
	}
	err = w.fieldsIndexOut.WriteVInt(int32(w.totalDocs - w.blockDocs)) // doc base
	if err == nil {
		err = w.fieldsIndexOut.WriteVInt(int32(avgChunkDocs))
	}
	if err != nil {
		return err
	}
	var docBase int = 0
	var maxDelta int64 = 0
	for i := 0; i < w.blockChunks; i++ {
		delta := docBase - avgChunkDocs*i
		maxDelta |= util.ZigZagEncodeLong(int64(delta))
		docBase += w.docBaseDeltas[i]
	}

	bitsPerDocbase := packed.BitsRequired(maxDelta)
	err = w.fieldsIndexOut.WriteVInt(int32(bitsPerDocbase))
	if err != nil {
		return err
	}
	writer := packed.WriterNoHeader(w.fieldsIndexOut,
		packed.PackedFormat(packed.PACKED), w.blockChunks, bitsPerDocbase, 1)
	docBase = 0
	for i := 0; i < w.blockChunks; i++ {
		delta := docBase - avgChunkDocs*i
		assert(packed.BitsRequired(util.ZigZagEncodeLong(int64(delta))) <= writer.BitsPerValue())
		err = writer.Add(util.ZigZagEncodeLong(int64(delta)))
		if err != nil {
			return err
		}
		docBase += w.docBaseDeltas[i]
	}
	err = writer.Finish()
	if err != nil {
		return err
	}

	// start pointers
	w.fieldsIndexOut.WriteVLong(w.firstStartPointer)
	var avgChunkSize int64
	if w.blockChunks == 1 {
		avgChunkSize = 0
	} else {
		avgChunkSize = (w.maxStartPointer - w.firstStartPointer) / int64(w.blockChunks-1)
	}
	err = w.fieldsIndexOut.WriteVLong(avgChunkSize)
	if err != nil {
		return err
	}
	var startPointer int64 = 0
	maxDelta = 0
	for i := 0; i < w.blockChunks; i++ {
		startPointer += w.startPointerDeltas[i]
		delta := startPointer - avgChunkSize*int64(i)
		maxDelta |= util.ZigZagEncodeLong(delta)
	}

	bitsPerStartPointer := packed.BitsRequired(maxDelta)
	err = w.fieldsIndexOut.WriteVInt(int32(bitsPerStartPointer))
	if err != nil {
		return err
	}
	writer = packed.WriterNoHeader(w.fieldsIndexOut,
		packed.PackedFormat(packed.PACKED), w.blockChunks, bitsPerStartPointer, 1)
	startPointer = 0
	for i := 0; i < w.blockChunks; i++ {
		startPointer += w.startPointerDeltas[i]
		delta := startPointer - avgChunkSize*int64(i)
		assert(packed.BitsRequired(util.ZigZagEncodeLong(delta)) <= writer.BitsPerValue())
		err = writer.Add(util.ZigZagEncodeLong(delta))
		if err != nil {
			return err
		}
	}
	return writer.Finish()
}

func (w *StoredFieldsIndexWriter) writeIndex(numDocs int, startPointer int64) error {
	if w.blockChunks == BLOCK_SIZE {
		err := w.writeBlock()
		if err != nil {
			return err
		}
		w.reset()
	}

	if w.firstStartPointer == -1 {
		w.firstStartPointer, w.maxStartPointer = startPointer, startPointer
	}
	assert(w.firstStartPointer > 0 && startPointer >= w.firstStartPointer)

	w.docBaseDeltas[w.blockChunks] = numDocs
	w.startPointerDeltas[w.blockChunks] = startPointer - w.maxStartPointer

	w.blockChunks++
	w.blockDocs += numDocs
	w.totalDocs += numDocs
	w.maxStartPointer = startPointer
	return nil
}

func (w *StoredFieldsIndexWriter) finish(numDocs int, maxPointer int64) (err error) {
	assert(w != nil)
	assert2(numDocs == w.totalDocs, "Expected %v docs, but got %v", numDocs, w.totalDocs)
	if w.blockChunks > 0 {
		if err = w.writeBlock(); err != nil {
			return
		}
	}
	if err = w.fieldsIndexOut.WriteVInt(0); err != nil { // end marker
		return
	}
	if err = w.fieldsIndexOut.WriteVLong(maxPointer); err != nil {
		return
	}
	return codec.WriteFooter(w.fieldsIndexOut)
}

func (w *StoredFieldsIndexWriter) Close() error {
	if w == nil {
		return nil
	}
	assert(w.fieldsIndexOut != nil)
	return w.fieldsIndexOut.Close()
}
