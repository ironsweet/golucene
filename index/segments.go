package index

import (
	"fmt"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
	"log"
	"sync/atomic"
)

type SegmentInfoPerCommit struct {
	info            SegmentInfo
	delCount        int
	delGen          int64
	nextWriteDelGen int64
}

func NewSegmentInfoPerCommit(info SegmentInfo, delCount int, delGen int64) SegmentInfoPerCommit {
	nextWriteDelGen := int64(1)
	if delGen != -1 {
		nextWriteDelGen = delGen + 1
	}
	return SegmentInfoPerCommit{info, delCount, delGen, nextWriteDelGen}
}

func (si SegmentInfoPerCommit) HasDeletions() bool {
	return si.delGen != -1
}

func (si SegmentInfoPerCommit) StringOf(dir store.Directory, pendingDelCount int) string {
	return si.info.StringOf(dir, si.delCount+pendingDelCount)
}

func (si SegmentInfoPerCommit) String() string {
	s := si.info.StringOf(si.info.dir, si.delCount)
	if si.delGen != -1 {
		s = fmt.Sprintf("%v:delGen=%v", s, si.delGen)
	}
	return s
}

type SegmentReader struct {
	*AtomicReaderImpl
	si       SegmentInfoPerCommit
	liveDocs *util.Bits
	numDocs  int
	core     SegmentCoreReaders
}

func NewSegmentReader(si SegmentInfoPerCommit, termInfosIndexDivisor int, context store.IOContext) (r *SegmentReader, err error) {
	log.Print("Initializing SegmentReader...")
	r = &SegmentReader{}
	log.Print("Obtaining AtomicReader...")
	r.AtomicReaderImpl = newAtomicReader(r)
	r.FieldsReader = r
	r.si = si
	log.Print("Obtaining SegmentCoreReaders...")
	r.core, err = newSegmentCoreReaders(r, si.info.dir, si, context, termInfosIndexDivisor)
	if err != nil {
		return r, err
	}
	success := false
	defer func() {
		// With lock-less commits, it's entirely possible (and
		// fine) to hit a FileNotFound exception above.  In
		// this case, we want to explicitly close any subset
		// of things that were opened so that we don't have to
		// wait for a GC to do so.
		if !success {
			log.Printf("Failed to initialize SegmentReader.")
			if err != nil {
				log.Print(err)
			}
			r.core.decRef()
		}
	}()

	if si.HasDeletions() {
		panic("not supported yet")
	} else {
		// assert si.getDelCount() == 0
		// r.liveDocs = nil
	}
	r.numDocs = int(si.info.docCount) - si.delCount
	success = true
	return r, nil
}

// SegmentReader.java L179
func (r *SegmentReader) String() string {
	// SegmentInfo.toString takes dir and number of
	// *pending* deletions; so we reverse compute that here:
	return r.si.StringOf(r.si.info.dir, int(r.si.info.docCount)-r.numDocs-r.si.delCount)
}

type CoreClosedListener interface {
	onClose(r *SegmentReader)
}

type SegmentCoreReaders struct {
	refCount int32 // synchronized

	fieldInfos FieldInfos

	fields        FieldsProducer
	dvProducer    DocValuesProducer
	normsProducer DocValuesProducer

	termsIndexDivisor int

	owner *SegmentReader

	fieldsReaderOrig      StoredFieldsReader
	termVectorsReaderOrig TermVectorsReader
	cfsReader             *store.CompoundFileDirectory

	addListener    chan CoreClosedListener
	removeListener chan CoreClosedListener
	notifyListener chan *SegmentReader
}

func newSegmentCoreReaders(owner *SegmentReader, dir store.Directory, si SegmentInfoPerCommit,
	context store.IOContext, termsIndexDivisor int) (self SegmentCoreReaders, err error) {
	if termsIndexDivisor == 0 {
		panic("indexDivisor must be < 0 (don't load terms index) or greater than 0 (got 0)")
	}
	log.Printf("Initializing SegmentCoreReaders from directory: %v", dir)

	self = SegmentCoreReaders{refCount: 1}

	log.Print("Initializing listeners...")
	self.addListener = make(chan CoreClosedListener)
	self.removeListener = make(chan CoreClosedListener)
	self.notifyListener = make(chan *SegmentReader)
	// TODO re-enable later
	go func() { // ensure listners are synchronized
		coreClosedListeners := make([]CoreClosedListener, 0)
		isRunning := true
		var listener CoreClosedListener
		for isRunning {
			log.Print("Listening for events...")
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
			case owner := <-self.notifyListener:
				log.Print("Shutting down SegmentCoreReaders...")
				isRunning = false
				for _, v := range coreClosedListeners {
					v.onClose(owner)
				}
			}
		}
		log.Print("Listeners are done.")
	}()

	success := false
	defer func() {
		if !success {
			log.Print("Failed to initialize SegmentCoreReaders.")
			self.decRef()
		}
	}()

	codec := si.info.codec
	log.Print("Obtaining CFS Directory...")
	var cfsDir store.Directory // confusing name: if (cfs) its the cfsdir, otherwise its the segment's directory.
	if si.info.isCompoundFile {
		log.Print("Detected CompoundFile.")
		name := util.SegmentFileName(si.info.name, "", store.COMPOUND_FILE_EXTENSION)
		log.Print("DEBUG ", name)
		self.cfsReader, err = store.NewCompoundFileDirectory(dir, name, context, false)
		if err != nil {
			return self, err
		}
		log.Printf("CompoundFileDirectory: %v", self.cfsReader)
		cfsDir = self.cfsReader
	} else {
		cfsDir = dir
	}
	log.Printf("CFS Directory: %v", cfsDir)
	log.Print("Reading FieldInfos...")
	self.fieldInfos, err = codec.ReadFieldInfos(cfsDir, si.info.name, store.IO_CONTEXT_READONCE)
	if err != nil {
		return self, err
	}
	log.Printf("DEBUG FieldsInfos: %v", self.fieldInfos)
	self.termsIndexDivisor = termsIndexDivisor

	log.Print("Obtaining SegmentReadState...")
	segmentReadState := newSegmentReadState(cfsDir, si.info, self.fieldInfos, context, termsIndexDivisor)
	// Ask codec for its Fields
	log.Print("Obtaining FieldsProducer...")
	self.fields, err = codec.GetFieldsProducer(segmentReadState)
	if err != nil {
		return self, err
	}
	// assert fields != null;
	// ask codec for its Norms:
	// TODO: since we don't write any norms file if there are no norms,
	// kinda jaky to assume the codec handles the case of no norms file at all gracefully?!

	if self.fieldInfos.hasDocValues {
		log.Print("Obtaining DocValuesProducer...")
		self.dvProducer, err = codec.GetDocValuesProducer(segmentReadState)
		if err != nil {
			return self, err
		}
		// assert dvProducer != null;
	} else {
		// self.dvProducer = nil
	}

	if self.fieldInfos.hasNorms {
		log.Print("Obtaining NormsDocValuesProducer...")
		self.normsProducer, err = codec.GetNormsDocValuesProducer(segmentReadState)
		if err != nil {
			return self, err
		}
		// assert normsProducer != null;
	} else {
		// self.normsProducer = nil
	}

	log.Print("Obtaining StoredFieldsReader...")
	self.fieldsReaderOrig, err = si.info.codec.GetStoredFieldsReader(cfsDir, si.info, self.fieldInfos, context)
	if err != nil {
		return self, err
	}

	if self.fieldInfos.hasVectors { // open term vector files only as needed
		log.Print("Obtaining TermVectorsReader...")
		self.termVectorsReaderOrig, err = si.info.codec.GetTermVectorsReader(cfsDir, si.info, self.fieldInfos, context)
		if err != nil {
			return self, err
		}
	} else {
		// self.termVectorsReaderOrig = nil
	}

	log.Print("Success")
	success = true

	// Must assign this at the end -- if we hit an
	// exception above core, we don't want to attempt to
	// purge the FieldCache (will hit NPE because core is
	// not assigned yet).
	self.owner = owner
	return self, nil
}

func (r *SegmentCoreReaders) decRef() {
	if atomic.AddInt32(&r.refCount, -1) == 0 {
		util.Close( /*self.termVectorsLocal, self.fieldsReaderLocal, docValuesLocal, normsLocal,*/
			r.fields, r.dvProducer, r.termVectorsReaderOrig, r.fieldsReaderOrig,
			r.cfsReader, r.normsProducer)
		r.notifyListener <- r.owner
	}
}

type SegmentReadState struct {
	dir               store.Directory
	segmentInfo       SegmentInfo
	fieldInfos        FieldInfos
	context           store.IOContext
	termsIndexDivisor int
	segmentSuffix     string
}

func newSegmentReadState(dir store.Directory, info SegmentInfo, fieldInfos FieldInfos,
	context store.IOContext, termsIndexDivisor int) SegmentReadState {
	return SegmentReadState{dir, info, fieldInfos, context, termsIndexDivisor, ""}
}

type DocValuesProducer interface {
	io.Closer
	Numeric(field FieldInfo) (v NumericDocValues, err error)
	Binary(field FieldInfo) (v BinaryDocValues, err error)
	Sorted(field FieldInfo) (v SortedDocValues, err error)
	SortedSet(field FieldInfo) (v SortedSetDocValues, err error)
}

type NumericDocValues interface {
	value(docID int) int64
}
type BinaryDocValues interface {
	value(docID int, result []byte)
}
type SortedDocValues interface {
	BinaryDocValues
	ord(docID int) int
	lookupOrd(ord int, result []byte)
	valueCount() int
}
type SortedSetDocValues interface {
	nextOrd() int64
	setDocument(docID int)
	lookupOrd(ord int64, result []byte)
	valueCount() int64
}

type StoredFieldVisitor interface {
	binaryField(fi FieldInfo, value []byte) error
	stringField(fi FieldInfo, value string) error
	intField(fi FieldInfo, value int) error
	longField(fi FieldInfo, value int64) error
	floatField(fi FieldInfo, value float32) error
	doubleField(fi FieldInfo, value float64) error
	needsField(fi FieldInfo) StoredFieldVisitorStatus
}

type StoredFieldVisitorStatus int

const (
	SOTRED_FIELD_VISITOR_STATUS_YES  = StoredFieldVisitorStatus(1)
	SOTRED_FIELD_VISITOR_STATUS_NO   = StoredFieldVisitorStatus(2)
	SOTRED_FIELD_VISITOR_STATUS_STOP = StoredFieldVisitorStatus(3)
)
