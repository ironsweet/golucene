package index

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
)

// index/DocumentsWriterPerThread.java

// Returns the DocConsumer that the DocumentsWriter calls to
// process the documents.
type IndexingChain func(documentsWriterPerThread *DocumentsWriterPerThread) DocConsumer

var defaultIndexingChain = func(documentsWriterPerThread *DocumentsWriterPerThread) DocConsumer {
	/*
	   This is the current indexing chain:

	   DocConsumer / DocConsumerPerThread
	     --> code: DocFieldProcessor
	       --> DocFieldConsumer / DocFieldConsumerPerField
	         --> code: DocFieldConsumers / DocFieldConsumersPerField
	           --> code: DocInverter / DocInverterPerField
	             --> InvertedDocConsumer / InvertedDocConsumerPerField
	               --> code: TermsHash / TermsHashPerField
	                 --> TermsHashConsumer / TermsHashConsumerPerField
	                   --> code: FreqProxTermsWriter / FreqProxTermsWriterPerField
	                   --> code: TermVectorsTermsWriter / TermVectorsTermsWriterPerField
	             --> InvertedDocEndConsumer / InvertedDocConsumerPerField
	               --> code: NormsConsumer / NormsConsumerPerField
	       --> StoredFieldsConsumer
	         --> TwoStoredFieldConsumers
	           -> code: StoredFieldsProcessor
	           -> code: DocValuesProcessor
	*/

	// Build up indexing chain:

	termVectorsWriter := newTermVectorsConsumer(documentsWriterPerThread)
	freqProxWriter := new(FreqProxTermsWriter)

	termsHash := newTermsHash(documentsWriterPerThread, freqProxWriter, true,
		newTermsHash(documentsWriterPerThread, termVectorsWriter, false, nil))
	normsWriter := new(NormsConsumer)
	docInverter := newDocInverter(documentsWriterPerThread.docState, termsHash, normsWriter)
	storedFields := newTwoStoredFieldsConsumers(
		newStoredFieldsProcessor(documentsWriterPerThread),
		newDocValuesProcessor(documentsWriterPerThread.bytesUsed))
	return newDocFieldProcessor(documentsWriterPerThread, docInverter, storedFields)
}

type docState struct {
	docWriter  *DocumentsWriterPerThread
	infoStream util.InfoStream
	similarity Similarity
}

func newDocState(docWriter *DocumentsWriterPerThread, infoStream util.InfoStream) *docState {
	return &docState{
		docWriter:  docWriter,
		infoStream: infoStream,
	}
}

type DocumentsWriterPerThread struct {
	codec         Codec
	directory     *TrackingDirectoryWrapper
	directoryOrig store.Directory
	docState      *docState
	consumer      DocConsumer
	bytesUsed     util.Counter

	// Deletes for our still-in-RAM (to be flushed next) segment
	pendingDeletes *BufferedDeletes
	segmentInfo    *SegmentInfo // Current segment we are working on
	aborting       bool         // True if an abort is pending
	hasAborted     bool         // True if the last exception throws by #updateDocument was aborting

	fieldInfos         *FieldInfosBuilder
	infoStream         util.InfoStream
	numDocsInRAM       int // the number of RAM resident documents
	deleteQueue        *DocumentsWriterDeleteQueue
	deleteSlice        *DeleteSlice
	byteBlockAllocator util.Allocator
	intBlockAllocator  util.Allocator
	indexWriterConfig  *LiveIndexWriterConfig

	filesToDelete map[string]bool
}

func newDocumentsWriterPerThread(segmentName string, directory store.Directory,
	indexWriterConfig *LiveIndexWriterConfig, infoStream util.InfoStream,
	deleteQueue *DocumentsWriterDeleteQueue, fieldInfos *FieldInfosBuilder) *DocumentsWriterPerThread {

	counter := util.NewCounter()
	ans := &DocumentsWriterPerThread{
		directoryOrig:      directory,
		directory:          newTrackingDirectoryWrapper(directory),
		fieldInfos:         fieldInfos,
		indexWriterConfig:  indexWriterConfig,
		infoStream:         infoStream,
		codec:              indexWriterConfig.codec,
		bytesUsed:          counter,
		byteBlockAllocator: util.NewDirectTrackingAllocator(counter),
		pendingDeletes:     newBufferedDeletes(),
		intBlockAllocator:  newIntBlockAllocator(counter),
		deleteQueue:        deleteQueue,
		deleteSlice:        deleteQueue.newSlice(),
		segmentInfo: newSegmentInfo(directory, util.LUCENE_MAIN_VERSION,
			segmentName, -1, false, indexWriterConfig.codec, nil, nil),
	}
	ans.docState = newDocState(ans, infoStream)
	ans.docState.similarity = indexWriterConfig.similarity
	// assertn(ans.numDocsInRAM == 0, "num docs ", ans.numDocsInRAM)
	// ans.pendingDeletes.clear()
	if VERBOSE && infoStream.IsEnabled("DWPT") {
		infoStream.Message("DWPT", "init seg=%v delQueue=%v", segmentName, deleteQueue)
	}
	// this should be the last call in the ctor
	// it really sucks that we need to pull this within the ctor and pass this ref to the chain!
	ans.consumer = indexWriterConfig.indexingChain(ans)
	return ans
}

/*
Called if we hit an error at a bad time (when updating the index
files) and must discard all currently buffered docs. This resets our
state, discarding any docs added since last flush.
*/
func (dwpt *DocumentsWriterPerThread) abort(createdFiles map[string]bool) {
	log.Printf("now abort seg=%v", dwpt.segmentInfo.name)
	dwpt.hasAborted, dwpt.aborting = true, true
	defer func() {
		dwpt.aborting = false
		if dwpt.infoStream.IsEnabled("DWPT") {
			dwpt.infoStream.Message("DWPT", "done abort")
		}
	}()

	if dwpt.infoStream.IsEnabled("DWPT") {
		dwpt.infoStream.Message("DWPT", "now abort")
	}
	dwpt.consumer.abort()

	dwpt.pendingDeletes.clear()
	for file, _ := range dwpt.directory.createdFiles() {
		createdFiles[file] = true
	}
}

func (dwpt *DocumentsWriterPerThread) checkAndResetHasAborted() (res bool) {
	res, dwpt.hasAborted = dwpt.hasAborted, false
	return
}

func (dwpt *DocumentsWriterPerThread) updateDocument(doc []IndexableField, analyzer analysis.Analyzer, delTerm *Term) error {
	panic("not implemented yet")
}

// L600
// if you increase this, you must fix field cache impl for
// Terms/TermsIndex requires <= 32768
const MAX_TERM_LENGTH_UTF8 = util.BYTE_BLOCK_SIZE - 2

type IntBlockAllocator struct {
	bytesUsed util.Counter
}

func newIntBlockAllocator(bytesUsed util.Counter) *IntBlockAllocator {
	return &IntBlockAllocator{bytesUsed}
}
