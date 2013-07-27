package index

import (
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"io"
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

type SegmentReader struct {
	*AtomicReader
	si       SegmentInfoPerCommit
	liveDocs *util.Bits
	numDocs  int
	core     SegmentCoreReaders
}

func NewSegmentReader(si SegmentInfoPerCommit, termInfosIndexDivisor int, context store.IOContext) (r *SegmentReader, err error) {
	r = &SegmentReader{}
	r.AtomicReader = newAtomicReader(r)
	r.si = si
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
			r.core.decRef()
		}
	}()

	if si.HasDeletions() {
		panic("not supported yet")
	} else {
		// assert si.getDelCount() == 0
		r.liveDocs = nil
	}
	r.numDocs = int(si.info.docCount) - si.delCount
	success = true
	return r, nil
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

	self = SegmentCoreReaders{refCount: 1}

	self.addListener = make(chan CoreClosedListener)
	self.removeListener = make(chan CoreClosedListener)
	self.notifyListener = make(chan *SegmentReader)
	go func() {
		coreClosedListeners := make([]CoreClosedListener, 0)
		nRemoved := 0
		isRunning := true
		var listener CoreClosedListener
		for isRunning {
			select {
			case listener = <-self.addListener:
				coreClosedListeners = append(coreClosedListeners, listener)
			case listener = <-self.removeListener:
				for i, v := range coreClosedListeners {
					if v == listener {
						coreClosedListeners[i] = nil
						nRemoved++
					}
				}
				if n := len(coreClosedListeners); n > 16 && nRemoved > n/2 {
					newListeners := make([]CoreClosedListener, 0)
					i := 0
					for _, v := range coreClosedListeners {
						if v == nil {
							continue
						}
						newListeners = append(newListeners, v)
					}
				}
			case owner := <-self.notifyListener:
				isRunning = false
				for _, v := range coreClosedListeners {
					v.onClose(owner)
				}
			}
		}
	}()

	success := false
	defer func() {
		if !success {
			self.decRef()
		}
	}()

	codec := si.info.codec
	var cfsDir store.Directory // confusing name: if (cfs) its the cfsdir, otherwise its the segment's directory.
	if si.info.isCompoundFile {
		self.cfsReader, err = store.NewCompoundFileDirectory(dir,
			util.SegmentFileName(si.info.name, "", store.COMPOUND_FILE_EXTENSION), context, false)
		if err != nil {
			return self, err
		}
		cfsDir = self.cfsReader.Directory
	} else {
		cfsDir = dir
	}
	self.fieldInfos, err = codec.ReadFieldInfos(cfsDir, si.info.name, store.IO_CONTEXT_READONCE)
	if err != nil {
		return self, err
	}
	self.termsIndexDivisor = termsIndexDivisor

	segmentReadState := newSegmentReadState(cfsDir, si.info, self.fieldInfos, context, termsIndexDivisor)
	// Ask codec for its Fields
	self.fields, err = codec.GetFieldsProducer(segmentReadState)
	if err != nil {
		return self, err
	}
	// assert fields != null;
	// ask codec for its Norms:
	// TODO: since we don't write any norms file if there are no norms,
	// kinda jaky to assume the codec handles the case of no norms file at all gracefully?!

	if self.fieldInfos.hasDocValues {
		self.dvProducer, err = codec.GetDocValuesProducer(segmentReadState)
		if err != nil {
			return self, err
		}
		// assert dvProducer != null;
	} else {
		self.dvProducer = nil
	}

	if self.fieldInfos.hasNorms {
		self.normsProducer, err = codec.GetNormsDocValuesProducer(segmentReadState)
		if err != nil {
			return self, err
		}
		// assert normsProducer != null;
	} else {
		self.normsProducer = nil
	}

	self.fieldsReaderOrig, err = si.info.codec.GetStoredFieldsReader(cfsDir, si.info, self.fieldInfos, context)
	if err != nil {
		return self, err
	}

	if self.fieldInfos.hasVectors { // open term vector files only as needed
		self.termVectorsReaderOrig, err = si.info.codec.GetTermVectorsReader(cfsDir, si.info, self.fieldInfos, context)
		if err != nil {
			return self, err
		}
	} else {
		self.termVectorsReaderOrig = nil
	}

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
		notifyListener <- r
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

type TermVectorsReader interface {
	io.Closer
	fields(doc int) Fields
	clone() TermVectorsReader
}
