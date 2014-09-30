package index

import (
	"strconv"
)

import (
	"fmt"
	// docu "github.com/balzaczyy/golucene/core/document"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"sync/atomic"
)

// index/SegmentReader.java

/**
 * IndexReader implementation over a single segment.
 * <p>
 * Instances pointing to the same segment (but with different deletes, etc)
 * may share the same core data.
 * @lucene.experimental
 */
type SegmentReader struct {
	*AtomicReaderImpl
	si       *SegmentCommitInfo
	liveDocs util.Bits
	// Normally set to si.docCount - si.delDocCount, unless we
	// were created as an NRT reader from IW, in which case IW
	// tells us the docCount:
	numDocs int
	core    *SegmentCoreReaders

	fieldInfos FieldInfos
}

/**
 * Constructs a new SegmentReader with a new core.
 * @throws CorruptIndexException if the index is corrupt
 * @throws IOException if there is a low-level IO error
 */
// TODO: why is this public?
func NewSegmentReader(si *SegmentCommitInfo,
	termInfosIndexDivisor int, context store.IOContext) (r *SegmentReader, err error) {

	r = &SegmentReader{}
	r.AtomicReaderImpl = newAtomicReader(r)
	r.ARFieldsReader = r

	r.si = si
	if r.fieldInfos, err = ReadFieldInfos(si); err != nil {
		return nil, err
	}
	// log.Print("Obtaining SegmentCoreReaders...")
	if r.core, err = newSegmentCoreReaders(r, si.Info.Dir, si, context, termInfosIndexDivisor); err != nil {
		return nil, err
	}
	// r.segDocValues = newSegmentDocValues()

	var success = false
	defer func() {
		// With lock-less commits, it's entirely possible (and
		// fine) to hit a FileNotFound exception above.  In
		// this case, we want to explicitly close any subset
		// of things that were opened so that we don't have to
		// wait for a GC to do so.
		if !success {
			// log.Printf("Failed to initialize SegmentReader.")
			r.core.decRef()
		}
	}()

	codec := si.Info.Codec().(Codec)
	if si.HasDeletions() {
		panic("not supported yet")
	} else {
		assert(si.DelCount() == 0)
	}
	r.numDocs = si.Info.DocCount() - si.DelCount()

	if r.fieldInfos.HasDocValues {
		r.initDocValuesProducers(codec)
	}
	success = true
	return r, nil
}

/* initialize the per-field DocValuesProducer */
func (r *SegmentReader) initDocValuesProducers(codec Codec) error {
	// var dir store.Directory
	// if r.core.cfsReader != nil {
	// 	dir = r.core.cfsReader
	// } else {
	// 	dir = r.si.Info.Dir
	// }
	// dvFormat := codec.DocValuesFormat()

	// termsIndexDivisor := r.core.termsIndexDivisor
	if !r.si.HasFieldUpdates() {
		panic("not implemented yet")
	}

	panic("not implemented yet")
}

/* Reads the most recent FieldInfos of the given segment info. */
func ReadFieldInfos(info *SegmentCommitInfo) (fis FieldInfos, err error) {
	var dir store.Directory
	var closeDir bool
	if info.FieldInfosGen() == -1 && info.Info.IsCompoundFile() {
		// no fieldInfos gen and segment uses a compound file
		if dir, err = store.NewCompoundFileDirectory(info.Info.Dir,
			util.SegmentFileName(info.Info.Name, "", store.COMPOUND_FILE_EXTENSION),
			store.IO_CONTEXT_READONCE, false); err != nil {
			return
		}
		closeDir = true
	} else {
		// gen'd FIS are read outside CFS, or the segment doesn't use a compound file
		dir = info.Info.Dir
		closeDir = false
	}

	defer func() {
		if closeDir {
			err = mergeError(err, dir.Close())
		}
	}()

	var segmentSuffix string
	if n := info.FieldInfosGen(); n != -1 {
		segmentSuffix = strconv.FormatInt(n, 36)
	}
	codec := info.Info.Codec().(Codec)
	fisFormat := codec.FieldInfosFormat()
	return fisFormat.FieldInfosReader()(dir, info.Info.Name, segmentSuffix, store.IO_CONTEXT_READONCE)
}

func (r *SegmentReader) LiveDocs() util.Bits {
	r.ensureOpen()
	return r.liveDocs
}

func (r *SegmentReader) doClose() error {
	panic("not implemented yet")
	r.core.decRef()
	return nil
}

func (r *SegmentReader) FieldInfos() FieldInfos {
	r.ensureOpen()
	return r.fieldInfos
}

// Expert: retrieve thread-private StoredFieldsReader
func (r *SegmentReader) FieldsReader() StoredFieldsReader {
	r.ensureOpen()
	return r.core.fieldsReaderLocal()
}

func (r *SegmentReader) VisitDocument(docID int, visitor StoredFieldVisitor) error {
	r.checkBounds(docID)
	return r.FieldsReader().VisitDocument(docID, visitor)
}

func (r *SegmentReader) Fields() Fields {
	r.ensureOpen()
	return r.core.fields
}

func (r *SegmentReader) NumDocs() int {
	// Don't call ensureOpen() here (it could affect performance)
	return r.numDocs
}

func (r *SegmentReader) MaxDoc() int {
	// Don't call ensureOpen() here (it could affect performance)
	return r.si.Info.DocCount()
}

func (r *SegmentReader) TermVectorsReader() TermVectorsReader {
	panic("not implemented yet")
}

func (r *SegmentReader) TermVectors(docID int) (fs Fields, err error) {
	panic("not implemented yet")
}

func (r *SegmentReader) checkBounds(docID int) {
	if docID < 0 || docID >= r.MaxDoc() {
		panic(fmt.Sprintf("docID must be >= 0 and < maxDoc=%v (got docID=%v)", r.MaxDoc(), docID))
	}
}

// SegmentReader.java L179
func (r *SegmentReader) String() string {
	// SegmentInfo.toString takes dir and number of
	// *pending* deletions; so we reverse compute that here:
	return r.si.StringOf(r.si.Info.Dir, r.si.Info.DocCount()-r.numDocs-r.si.DelCount())
}

func (r *SegmentReader) SegmentName() string {
	return r.si.Info.Name
}

func (r *SegmentReader) SegmentInfos() *SegmentCommitInfo {
	return r.si
}

func (r *SegmentReader) Directory() store.Directory {
	// Don't ensureOpen here -- in certain cases, when a
	// cloned/reopened reader needs to commit, it may call
	// this method on the closed original reader
	return r.si.Info.Dir
}

func (r *SegmentReader) CoreCacheKey() interface{} {
	return r.core
}

func (r *SegmentReader) CombinedCoreAndDeletesKey() interface{} {
	return r
}

func (r *SegmentReader) TermInfosIndexDivisor() int {
	return r.core.termsIndexDivisor
}

func (r *SegmentReader) NumericDocValues(field string) (v NumericDocValues, err error) {
	r.ensureOpen()
	panic("not implemented yet")
}

func (r *SegmentReader) BinaryDocValues(field string) (v BinaryDocValues, err error) {
	r.ensureOpen()
	panic("not implemented yet")
}

func (r *SegmentReader) SortedDocValues(field string) (v SortedDocValues, err error) {
	r.ensureOpen()
	panic("not implemented yet")
}

func (r *SegmentReader) SortedSetDocValues(field string) (v SortedSetDocValues, err error) {
	r.ensureOpen()
	panic("not implemented yet")
}

func (r *SegmentReader) NormValues(field string) (v NumericDocValues, err error) {
	r.ensureOpen()
	return r.core.normValues(r.fieldInfos, field)
}

type CoreClosedListener interface {
	onClose(r interface{})
}

// index/SegmentCoreReaders.java

type SegmentCoreReaders struct {
	refCount int32 // synchronized

	fields        FieldsProducer
	normsProducer DocValuesProducer

	termsIndexDivisor int

	owner *SegmentReader

	fieldsReaderOrig      StoredFieldsReader
	termVectorsReaderOrig TermVectorsReader
	cfsReader             *store.CompoundFileDirectory

	/*
	 Lucene Java use ThreadLocal to serve as thread-level cache, to avoid
	 expensive read actions while limit memory consumption. Since Go doesn't
	 have thread or routine Local, a new object is always returned.

	 TODO redesign when ported to goroutines
	*/
	fieldsReaderLocal func() StoredFieldsReader
	normsLocal        func() map[string]interface{}

	addListener    chan CoreClosedListener
	removeListener chan CoreClosedListener
	notifyListener chan bool
}

func newSegmentCoreReaders(owner *SegmentReader, dir store.Directory, si *SegmentCommitInfo,
	context store.IOContext, termsIndexDivisor int) (self *SegmentCoreReaders, err error) {

	assert2(termsIndexDivisor != 0,
		"indexDivisor must be < 0 (don't load terms index) or greater than 0 (got 0)")
	// fmt.Println("Initializing SegmentCoreReaders from directory:", dir)

	self = &SegmentCoreReaders{
		refCount: 1,
		normsLocal: func() map[string]interface{} {
			return make(map[string]interface{})
		},
	}
	self.fieldsReaderLocal = func() StoredFieldsReader {
		return self.fieldsReaderOrig.Clone()
	}

	// fmt.Println("Initializing listeners...")
	self.addListener = make(chan CoreClosedListener)
	self.removeListener = make(chan CoreClosedListener)
	self.notifyListener = make(chan bool)
	// TODO re-enable later
	go func() { // ensure listners are synchronized
		coreClosedListeners := make([]CoreClosedListener, 0)
		isRunning := true
		var listener CoreClosedListener
		for isRunning {
			// fmt.Println("Listening for events...")
			select {
			case listener = <-self.addListener:
				coreClosedListeners = append(coreClosedListeners, listener)
			case listener = <-self.removeListener:
				n := len(coreClosedListeners)
				for i, v := range coreClosedListeners {
					if v == listener {
						newListeners := make([]CoreClosedListener, 0, n-1)
						newListeners = append(newListeners, coreClosedListeners[0:i]...)
						newListeners = append(newListeners, coreClosedListeners[i+1:]...)
						coreClosedListeners = newListeners
						break
					}
				}
			case <-self.notifyListener:
				fmt.Println("Shutting down SegmentCoreReaders...")
				isRunning = false
				for _, v := range coreClosedListeners {
					v.onClose(self)
				}
			}
		}
		fmt.Println("Listeners are done.")
	}()

	var success = false
	ans := self
	defer func() {
		if !success {
			fmt.Println("Failed to initialize SegmentCoreReaders.")
			ans.decRef()
		}
	}()

	codec := si.Info.Codec().(Codec)
	// fmt.Println("Obtaining CFS Directory...")
	var cfsDir store.Directory // confusing name: if (cfs) its the cfsdir, otherwise its the segment's directory.
	if si.Info.IsCompoundFile() {
		// fmt.Println("Detected CompoundFile.")
		name := util.SegmentFileName(si.Info.Name, "", store.COMPOUND_FILE_EXTENSION)
		if self.cfsReader, err = store.NewCompoundFileDirectory(dir, name, context, false); err != nil {
			return nil, err
		}
		// fmt.Println("CompoundFileDirectory: ", self.cfsReader)
		cfsDir = self.cfsReader
	} else {
		cfsDir = dir
	}
	// fmt.Println("CFS Directory:", cfsDir)

	// fmt.Println("Reading FieldInfos...")
	fieldInfos := owner.fieldInfos

	self.termsIndexDivisor = termsIndexDivisor
	format := codec.PostingsFormat()

	// fmt.Println("Obtaining SegmentReadState...")
	segmentReadState := NewSegmentReadState(cfsDir, si.Info, fieldInfos, context, termsIndexDivisor)
	// Ask codec for its Fields
	// fmt.Println("Obtaining FieldsProducer...")
	if self.fields, err = format.FieldsProducer(segmentReadState); err != nil {
		return nil, err
	}
	assert(self.fields != nil)
	// ask codec for its Norms:
	// TODO: since we don't write any norms file if there are no norms,
	// kinda jaky to assume the codec handles the case of no norms file at all gracefully?!

	if fieldInfos.HasNorms {
		// fmt.Println("Obtaining NormsDocValuesProducer...")
		if self.normsProducer, err = codec.NormsFormat().NormsProducer(segmentReadState); err != nil {
			return nil, err
		}
		assert(self.normsProducer != nil)
	}

	// fmt.Println("Obtaining StoredFieldsReader...")
	if self.fieldsReaderOrig, err = si.Info.Codec().(Codec).StoredFieldsFormat().FieldsReader(cfsDir, si.Info, fieldInfos, context); err != nil {
		return nil, err
	}

	if fieldInfos.HasVectors { // open term vector files only as needed
		// fmt.Println("Obtaining TermVectorsReader...")
		if self.termVectorsReaderOrig, err = si.Info.Codec().(Codec).TermVectorsFormat().VectorsReader(cfsDir, si.Info, fieldInfos, context); err != nil {
			return nil, err
		}
	}

	// fmt.Println("Success")
	success = true

	return self, nil
}

func (r *SegmentCoreReaders) normValues(infos FieldInfos,
	field string) (ndv NumericDocValues, err error) {

	if norms, ok := r.normsLocal()[field]; ok {
		ndv = norms.(NumericDocValues)
	} else if fi := infos.FieldInfoByName(field); fi != nil && fi.HasNorms() {
		assert(r.normsProducer != nil)
		if ndv, err = r.normsProducer.Numeric(fi); err == nil {
			r.normsLocal()[field] = norms
		} // else Field does not exist
	}
	return
}

func (r *SegmentCoreReaders) decRef() {
	if atomic.AddInt32(&r.refCount, -1) == 0 {
		fmt.Println("--- closing core readers")
		util.Close( /*self.termVectorsLocal, self.fieldsReaderLocal,  r.normsLocal,*/
			r.fields, r.termVectorsReaderOrig, r.fieldsReaderOrig,
			r.cfsReader, r.normsProducer)
		r.notifyListener <- true
	}
}
