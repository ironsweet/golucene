package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"sync/atomic"
)

// index/DocumentsWriterPerThread.java

const DWPT_VERBOSE = false

// Returns the DocConsumer that the DocumentsWriter calls to
// process the documents.
type IndexingChain func(documentsWriterPerThread *DocumentsWriterPerThread) DocConsumer

var defaultIndexingChain = func(dwpt *DocumentsWriterPerThread) DocConsumer {
	return newDefaultIndexingChain(dwpt)
}

type docState struct {
	docWriter  *DocumentsWriterPerThread
	analyzer   analysis.Analyzer
	infoStream util.InfoStream
	similarity Similarity
	docID      int
	doc        []IndexableField
}

func newDocState(docWriter *DocumentsWriterPerThread, infoStream util.InfoStream) *docState {
	return &docState{
		docWriter:  docWriter,
		infoStream: infoStream,
	}
}

func (ds *docState) testPoint(name string) {
	ds.docWriter.testPoint(name)
}

func (ds *docState) clear() {
	// don't hold onto doc nor analyzer, in case it is largish:
	ds.doc = nil
	ds.analyzer = nil
}

type FlushedSegment struct {
	segmentInfo    *SegmentCommitInfo
	fieldInfos     FieldInfos
	segmentUpdates *FrozenBufferedUpdates
	liveDocs       util.MutableBits
	delCount       int
}

func newFlushedSegment(segmentInfo *SegmentCommitInfo,
	fieldInfos FieldInfos, segmentUpdates *BufferedUpdates,
	liveDocs util.MutableBits, delCount int) *FlushedSegment {

	var sd *FrozenBufferedUpdates
	if segmentUpdates != nil && segmentUpdates.any() {
		sd = freezeBufferedUpdates(segmentUpdates, true)
	}
	return &FlushedSegment{segmentInfo, fieldInfos, sd, liveDocs, delCount}
}

type DocumentsWriterPerThread struct {
	codec         Codec
	directory     *store.TrackingDirectoryWrapper
	directoryOrig store.Directory
	docState      *docState
	consumer      DocConsumer
	_bytesUsed    util.Counter

	// Deletes for our still-in-RAM (to be flushed next) segment
	pendingUpdates *BufferedUpdates
	segmentInfo    *SegmentInfo // Current segment we are working on
	aborting       bool         // True if an abort is pending
	hasAborted     bool         // True if the last exception throws by #updateDocument was aborting

	fieldInfos         *FieldInfosBuilder
	infoStream         util.InfoStream
	numDocsInRAM       int // the number of RAM resident documents
	deleteQueue        *DocumentsWriterDeleteQueue
	deleteSlice        *DeleteSlice
	byteBlockAllocator util.ByteAllocator
	intBlockAllocator  util.IntAllocator
	pendingNumDocs     *int64
	indexWriterConfig  LiveIndexWriterConfig

	filesToDelete map[string]bool
}

func newDocumentsWriterPerThread(segmentName string,
	directory store.Directory, indexWriterConfig LiveIndexWriterConfig,
	infoStream util.InfoStream, deleteQueue *DocumentsWriterDeleteQueue,
	fieldInfos *FieldInfosBuilder, pendingNumDocs *int64) *DocumentsWriterPerThread {

	counter := util.NewCounter()
	ans := &DocumentsWriterPerThread{
		directoryOrig:      directory,
		directory:          store.NewTrackingDirectoryWrapper(directory),
		fieldInfos:         fieldInfos,
		pendingNumDocs:     pendingNumDocs,
		indexWriterConfig:  indexWriterConfig,
		infoStream:         infoStream,
		codec:              indexWriterConfig.Codec(),
		_bytesUsed:         counter,
		byteBlockAllocator: util.NewDirectTrackingAllocator(counter),
		pendingUpdates:     newBufferedUpdates(),
		intBlockAllocator:  newIntBlockAllocator(counter),
		deleteQueue:        deleteQueue,
		deleteSlice:        deleteQueue.newSlice(),
		segmentInfo:        NewSegmentInfo(directory, util.VERSION_LATEST, segmentName, -1, false, indexWriterConfig.Codec(), nil),
		filesToDelete:      make(map[string]bool),
	}
	ans.docState = newDocState(ans, infoStream)
	ans.docState.similarity = indexWriterConfig.Similarity()
	assert2(ans.numDocsInRAM == 0, "num docs %v", ans.numDocsInRAM)
	if DWPT_VERBOSE && infoStream.IsEnabled("DWPT") {
		infoStream.Message("DWPT", "init seg=%v delQueue=%v", segmentName, deleteQueue)
	}
	// this should be the last call in the ctor
	// it really sucks that we need to pull this within the ctor and pass this ref to the chain!
	ans.consumer = indexWriterConfig.indexingChain()(ans)
	return ans
}

/*
Called if we hit an error at a bad time (when updating the index
files) and must discard all currently buffered docs. This resets our
state, discarding any docs added since last flush.
*/
func (dwpt *DocumentsWriterPerThread) abort(createdFiles map[string]bool) {
	assert(createdFiles != nil)
	log.Printf("now abort seg=%v", dwpt.segmentInfo.Name)
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

	dwpt.pendingUpdates.clear()
	dwpt.directory.EachCreatedFiles(func(file string) {
		createdFiles[file] = true
	})
}

func (dwpt *DocumentsWriterPerThread) setAborting() {
	dwpt.aborting = true
}

func (dwpt *DocumentsWriterPerThread) checkAndResetHasAborted() (res bool) {
	res, dwpt.hasAborted = dwpt.hasAborted, false
	return
}

func (dwpt *DocumentsWriterPerThread) testPoint(msg string) {
	if dwpt.infoStream.IsEnabled("TP") {
		dwpt.infoStream.Message("TP", msg)
	}
}

/*
Anything that will add N docs to the index should reserve first to
make sure it's allowed.
*/
func (dwpt *DocumentsWriterPerThread) reserveDoc() {
	if atomic.AddInt64(dwpt.pendingNumDocs, 1) > int64(actualMaxDocs) {
		// reserve failed
		atomic.AddInt64(dwpt.pendingNumDocs, -1)
		panic(fmt.Sprintf("number of documents in the index cannot exceed %v", actualMaxDocs))
	}
}

func (dwpt *DocumentsWriterPerThread) updateDocument(doc []IndexableField,
	analyzer analysis.Analyzer, delTerm *Term) error {

	dwpt.testPoint("DocumentsWriterPerThread addDocument start")
	assert(dwpt.deleteQueue != nil)
	dwpt.docState.doc = doc
	dwpt.docState.analyzer = analyzer
	dwpt.docState.docID = dwpt.numDocsInRAM
	if DWPT_VERBOSE && dwpt.infoStream.IsEnabled("DWPT") {
		dwpt.infoStream.Message("DWPT", "update delTerm=%v docID=%v seg=%v ",
			delTerm, dwpt.docState.docID, dwpt.segmentInfo.Name)
	}
	// Even on error, the document is still added (but marked deleted),
	// so we don't need to un-reserve at that point. Aborting errors
	// will actually "lose" more than one document, so the counter will
	// be "wrong" in that case, but it's very hard to fix (we can't
	// easily distinguish aborting vs non-aborting errors):
	dwpt.reserveDoc()
	if err := func() error {
		var success = false
		defer func() {
			if !success {
				if !dwpt.aborting {
					// mark document as deleted
					dwpt.deleteDocID(dwpt.docState.docID)
					dwpt.numDocsInRAM++
				} else {
					dwpt.abort(dwpt.filesToDelete)
				}
			}
		}()
		defer dwpt.docState.clear()
		if err := dwpt.consumer.processDocument(); err != nil {
			return err
		}
		success = true
		return nil
	}(); err != nil {
		return err
	}
	dwpt.finishDocument(delTerm)
	return nil
}

func (dwpt *DocumentsWriterPerThread) updateDocuments(doc []IndexableField,
	analyzer analysis.Analyzer, delTerm *Term) error {

	dwpt.testPoint("DocumentsWriterPerThread addDocument start")
	panic("not implemented yet")
}

func (w *DocumentsWriterPerThread) finishDocument(delTerm *Term) {
	// here we actually finish the document in two steps:
	// 1. push the delete into the queue and update out slice.
	// 2. increment the DWPT private document id.
	//
	// the updated slice we get from 1. holds all the deletes that have
	// occurred since we updated the slice the last time.
	applySlice := w.numDocsInRAM != 0
	if delTerm != nil {
		w.deleteQueue.add(delTerm, w.deleteSlice)
		assertn(w.deleteSlice.isTailItem(delTerm), "expected the delete term as the tail item")
	} else {
		if !w.deleteQueue.updateSlice(w.deleteSlice) {
			applySlice = false
		}
	}

	if applySlice {
		w.deleteSlice.apply(w.pendingUpdates, w.numDocsInRAM)
	} else { // if we don't need to apply we must reset!
		w.deleteSlice.reset()
	}
	w.numDocsInRAM++
}

/*
Buffer a specific docID for deletion. Currenlty only used when we hit
an error when adding a document
*/
func (dwpt *DocumentsWriterPerThread) deleteDocID(docIDUpto int) {
	dwpt.pendingUpdates.addDocID(docIDUpto)
	// NOTE: we do not trigger flush here.  This is
	// potentially a RAM leak, if you have an app that tries
	// to add docs but every single doc always hits a
	// non-aborting exception.  Allowing a flush here gets
	// very messy because we are only invoked when handling
	// exceptions so to do this properly, while handling an
	// exception we'd have to go off and flush new deletes
	// which is risky (likely would hit some other
	// confounding exception).
}

/*
Prepares this DWPT fo flushing. This method will freeze and return
the DWDQs global buffer and apply all pending deletes to this DWPT.
*/
func (dwpt *DocumentsWriterPerThread) prepareFlush() *FrozenBufferedUpdates {
	assert(dwpt.numDocsInRAM > 0)
	globalDeletes := dwpt.deleteQueue.freezeGlobalBuffer(dwpt.deleteSlice)
	// deleteSlice can possibly be nil if we have hit non-aborting
	// errors during adding a document.
	if dwpt.deleteSlice != nil {
		// apply all deletes before we flush and release the delete slice
		dwpt.deleteSlice.apply(dwpt.pendingUpdates, dwpt.numDocsInRAM)
		assert(dwpt.deleteSlice.isEmpty())
		dwpt.deleteSlice.reset()
	}
	return globalDeletes
}

/* Flush all pending docs to a new segment */
func (dwpt *DocumentsWriterPerThread) flush() (fs *FlushedSegment, err error) {
	assert(dwpt.numDocsInRAM > 0)
	assert2(dwpt.deleteSlice.isEmpty(), "all deletes must be applied in prepareFlush")
	dwpt.segmentInfo.SetDocCount(dwpt.numDocsInRAM)
	numBytesUsed := dwpt.bytesUsed()
	flushState := NewSegmentWriteState(dwpt.infoStream, dwpt.directory,
		dwpt.segmentInfo, dwpt.fieldInfos.Finish(),
		dwpt.indexWriterConfig.TermIndexInterval(), dwpt.pendingUpdates,
		store.NewIOContextForFlush(&store.FlushInfo{dwpt.numDocsInRAM, numBytesUsed}))
	startMBUsed := float64(numBytesUsed) / 1024 / 1024

	// Apply delete-by-docID now (delete-byDocID only happens when an
	// error is hit processing that doc, e.g., if analyzer has some
	// problem with the text):
	if delCount := len(dwpt.pendingUpdates.docIDs); delCount > 0 {
		flushState.LiveDocs = dwpt.codec.LiveDocsFormat().NewLiveDocs(dwpt.numDocsInRAM)
		for _, delDocID := range dwpt.pendingUpdates.docIDs {
			flushState.LiveDocs.Clear(delDocID)
		}
		flushState.DelCountOnFlush = delCount
		atomic.AddInt64(&dwpt.pendingUpdates.bytesUsed, -int64(delCount)*BYTES_PER_DEL_DOCID)
		dwpt.pendingUpdates.docIDs = nil
	}

	if dwpt.aborting {
		if dwpt.infoStream.IsEnabled("DWPT") {
			dwpt.infoStream.Message("DWPT", "flush: skip because aborting is set")
		}
		return nil, nil
	}

	if dwpt.infoStream.IsEnabled("DWPT") {
		dwpt.infoStream.Message("DWPT", "flush postings as segment %v numDocs=%v",
			flushState.SegmentInfo.Name, dwpt.numDocsInRAM)
	}

	var success = false
	defer func() {
		if !success {
			dwpt.abort(dwpt.filesToDelete)
		}
	}()

	err = dwpt.consumer.flush(flushState)
	if err != nil {
		return nil, err
	}
	dwpt.pendingUpdates.terms = make(map[*Term]int)
	files := make(map[string]bool)
	dwpt.directory.EachCreatedFiles(func(name string) {
		files[name] = true
	})
	dwpt.segmentInfo.SetFiles(files)

	info := NewSegmentCommitInfo(dwpt.segmentInfo, 0, -1, -1, -1)
	if dwpt.infoStream.IsEnabled("DWPT") {
		dwpt.infoStream.Message("DWPT", "new segment has %v deleted docs",
			check(flushState.LiveDocs == nil, 0,
				flushState.SegmentInfo.DocCount()-flushState.DelCountOnFlush))
		dwpt.infoStream.Message("DWPT", "new segment has %v; %v; %v; %v; %v",
			check(flushState.FieldInfos.HasVectors, "vectors", "no vectors"),
			check(flushState.FieldInfos.HasNorms, "norms", "no norms"),
			check(flushState.FieldInfos.HasDocValues, "docValues", "no docValues"),
			check(flushState.FieldInfos.HasProx, "prox", "no prox"),
			check(flushState.FieldInfos.HasFreq, "freqs", "no freqs"))
		dwpt.infoStream.Message("DWPT", "flushedFiles=%v", info.Files())
		dwpt.infoStream.Message("DWPT", "flushed codec=%v", dwpt.codec)
	}

	var segmentUpdates *BufferedUpdates
	if len(dwpt.pendingUpdates.queries) > 0 {
		segmentUpdates = dwpt.pendingUpdates
	}

	if dwpt.infoStream.IsEnabled("DWPT") {
		numBytes, err := info.SizeInBytes()
		if err != nil {
			return nil, err
		}
		newSegmentSize := float64(numBytes) / 1024 / 1024
		dwpt.infoStream.Message("DWPT",
			"flushed: segment=%v ramUsed=%v MB newFlushedSize(includes docstores)=%v MB docs/MB=%v",
			dwpt.segmentInfo.Name, startMBUsed, newSegmentSize,
			float64(flushState.SegmentInfo.DocCount())/newSegmentSize)
	}

	assert(dwpt.segmentInfo != nil)

	fs = newFlushedSegment(info, flushState.FieldInfos, segmentUpdates,
		flushState.LiveDocs, flushState.DelCountOnFlush)
	err = dwpt.sealFlushedSegment(fs)
	if err != nil {
		return nil, err
	}
	success = true

	return fs, nil
}

func check(ok bool, v1, v2 interface{}) interface{} {
	if ok {
		return v1
	}
	return v2
}

/*
Seals the SegmentInfo for the new flushed segment and persists the
deleted documents MutableBits
*/
func (dwpt *DocumentsWriterPerThread) sealFlushedSegment(flushedSegment *FlushedSegment) error {
	assert(flushedSegment != nil)

	newSegment := flushedSegment.segmentInfo

	setDiagnostics(newSegment.Info, SOURCE_FLUSH)

	segSize, err := newSegment.SizeInBytes()
	if err != nil {
		return err
	}
	context := store.NewIOContextForFlush(&store.FlushInfo{
		newSegment.Info.DocCount(),
		segSize,
	})

	var success = false
	defer func() {
		if !success {
			if dwpt.infoStream.IsEnabled("DWPT") {
				dwpt.infoStream.Message(
					"DWPT", "hit error relating compound file for newly flushed segment %v",
					newSegment.Info.Name)
			}
		}
	}()

	if dwpt.indexWriterConfig.UseCompoundFile() {
		files, err := createCompoundFile(
			dwpt.infoStream, dwpt.directory,
			CheckAbortNone(0), newSegment.Info, context)
		if err != nil {
			return err
		}
		for _, file := range files {
			dwpt.filesToDelete[file] = true
		}
		newSegment.Info.SetUseCompoundFile(true)
	}

	// Have codec write SegmentInfo. Must do this after creating CFS so
	// that 1) .si isn't slurped into CFS, and 2) .si reflects
	// useCompoundFile=true change above:
	err = dwpt.codec.SegmentInfoFormat().SegmentInfoWriter().Write(
		dwpt.directory,
		newSegment.Info,
		flushedSegment.fieldInfos,
		context)
	if err != nil {
		return err
	}

	// TODO: ideally we would freeze newSegment here!!
	// because any changes after writing the .si will be lost...

	// Must write deleted docs after the CFS so we don't slurp the del
	// file into CFS:
	if flushedSegment.liveDocs != nil {
		delCount := flushedSegment.delCount
		assert(delCount > 0)
		if dwpt.infoStream.IsEnabled("DWPT") {
			dwpt.infoStream.Message("DWPT", "flush: write %v deletes gen=%v",
				delCount, flushedSegment.segmentInfo.DelGen())
		}

		// TODO: we should prune the segment if it's 100% deleted... but
		// merge will also catch it.

		// TODO: in the NRT case id'd be better to hand this del vector
		// over to the shortly-to-be-opened SegmentReader and let it
		// carry the changes; there's no reason to use filesystem as
		// intermediary here.

		info := flushedSegment.segmentInfo
		codec := info.Info.Codec().(Codec)
		err = codec.LiveDocsFormat().WriteLiveDocs(flushedSegment.liveDocs,
			dwpt.directory, info, delCount, context)
		if err != nil {
			return err
		}
		newSegment.SetDelCount(delCount)
		newSegment.AdvanceDelGen()
	}

	success = true
	return nil
}

func (dwpt *DocumentsWriterPerThread) bytesUsed() int64 {
	return dwpt._bytesUsed.Get() + atomic.LoadInt64(&dwpt.pendingUpdates.bytesUsed)
}

// L600
// if you increase this, you must fix field cache impl for
// Terms/TermsIndex requires <= 32768
const MAX_TERM_LENGTH_UTF8 = util.BYTE_BLOCK_SIZE - 2

type IntBlockAllocator struct {
	*util.IntAllocatorImpl
	blockSize int
	bytesUsed util.Counter
}

func newIntBlockAllocator(bytesUsed util.Counter) *IntBlockAllocator {
	return &IntBlockAllocator{
		IntAllocatorImpl: util.NewIntAllocator(util.INT_BLOCK_SIZE),
		blockSize:        util.INT_BLOCK_SIZE,
		bytesUsed:        bytesUsed,
	}
}

func (alloc *IntBlockAllocator) Recycle(blocks [][]int) {
	alloc.bytesUsed.AddAndGet(int64(-len(blocks) * alloc.blockSize * util.NUM_BYTES_INT))
	for i, _ := range blocks {
		blocks[i] = nil
	}
}

func (w *DocumentsWriterPerThread) String() string {
	return fmt.Sprintf(
		"DocumentsWriterPerThread [pendingUpdates=%v, segment=%v, aborting=%v, numDocsInRAM=%v, deleteQueue=%v]",
		w.pendingUpdates, func() string {
			if w.segmentInfo != nil {
				return w.segmentInfo.Name
			}
			return "null"
		}(), w.aborting, w.numDocsInRAM, w.deleteQueue)
}
