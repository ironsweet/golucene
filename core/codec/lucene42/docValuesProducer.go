package lucene42

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"reflect"
	"sync"
	"sync/atomic"
)

const (
	LUCENE42_DV_DATA_CODEC         = "Lucene42DocValuesData"
	LUCENE42_DV_DATA_EXTENSION     = "dvd"
	LUCENE42_DV_METADATA_CODEC     = "Lucene42DocValuesMetadata"
	LUCENE42_DV_METADATA_EXTENSION = "dvm"

	LUCENE42_DV_VERSION_START           = 0
	LUCENE42_DV_VERSION_GCD_COMPRESSION = 1
	LUCENE42_DV_VERSION_CHECKSUM        = 2
	LUCENE42_DV_VERSION_CURRENT         = LUCENE42_DV_VERSION_CHECKSUM

	LUCENE42_DV_NUMBER = 0
	LUCENE42_DV_BYTES  = 1
	LUCENE42_DV_FST    = 2

	LUCENE42_DV_DELTA_COMPRESSED = 0
	LUCENE42_DV_TABLE_COMPRESSED = 1
	LUCENE42_DV_UNCOMPRESSED     = 2
	LUCENE42_DV_GCD_COMPRESSED   = 3
)

type Lucene42DocValuesProducer struct {
	lock sync.Mutex

	numerics map[int]NumericEntry
	binaries map[int]BinaryEntry
	fsts     map[int]FSTEntry
	data     store.IndexInput

	numericInstances map[int]NumericDocValues

	maxDoc       int
	ramBytesUsed int64
}

func newLucene42DocValuesProducer(state SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (dvp *Lucene42DocValuesProducer, err error) {

	fmt.Println("Initializing Lucene42DocValuesProducer...")
	dvp = &Lucene42DocValuesProducer{
		numericInstances: make(map[int]NumericDocValues),
	}
	dvp.maxDoc = state.SegmentInfo.DocCount()

	metaName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, metaExtension)
	fmt.Println("Reading", metaName)
	// read in the entries from the metadata file.
	var in store.ChecksumIndexInput
	if in, err = state.Dir.OpenChecksumInput(metaName, state.Context); err != nil {
		return nil, err
	}
	dvp.ramBytesUsed = util.ShallowSizeOfInstance(reflect.TypeOf(dvp))

	var version int32
	func() {
		var success = false
		defer func() {
			if success {
				err = util.Close(in)
			} else {
				util.CloseWhileSuppressingError(in)
			}
		}()

		if version, err = codec.CheckHeader(in, metaCodec,
			LUCENE42_DV_VERSION_START, LUCENE42_DV_VERSION_CURRENT); err != nil {
			return
		}

		dvp.numerics = make(map[int]NumericEntry)
		dvp.binaries = make(map[int]BinaryEntry)
		dvp.fsts = make(map[int]FSTEntry)
		if err = dvp.readFields(in, state.FieldInfos); err != nil {
			return
		}

		if version >= LUCENE42_DV_VERSION_CHECKSUM {
			_, err = codec.CheckFooter(in)
		} else {
			err = codec.CheckEOF(in)
		}
	}()
	if err != nil {
		return nil, err
	}

	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(dvp.data)
		}
	}()

	dataName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, dataExtension)
	fmt.Println("Reading", dataName)
	if dvp.data, err = state.Dir.OpenInput(dataName, state.Context); err != nil {
		return nil, err
	}
	var version2 int32
	if version2, err = codec.CheckHeader(dvp.data, dataCodec,
		LUCENE42_DV_VERSION_START, LUCENE42_DV_VERSION_CURRENT); err != nil {
		return nil, err
	}

	if version != version2 {
		return nil, errors.New("Format versions mismatch")
	}

	if version >= LUCENE42_DV_VERSION_CHECKSUM {
		panic("niy")
	}

	success = true

	return dvp, nil
}

/*
Lucene42DocValuesProducer.java/4.5.1/L138
*/
func (dvp *Lucene42DocValuesProducer) readFields(meta store.IndexInput,
	infos FieldInfos) (err error) {

	var fieldNumber int
	var fieldType byte
	fieldNumber, err = asInt(meta.ReadVInt())
	for fieldNumber != -1 && err == nil {
		if infos.FieldInfoByNumber(fieldNumber) == nil {
			// tricker to validate more: because we re-use for norms,
			// becaue we use multiple entries for "composite" types like
			// sortedset, etc.
			return errors.New(fmt.Sprintf(
				"Invalid field number: %v (resource=%v)",
				fieldNumber, meta))
		}

		fieldType, err = meta.ReadByte()
		if err != nil {
			return
		}
		switch fieldType {
		case LUCENE42_DV_NUMBER:
			entry := NumericEntry{}
			entry.offset, err = meta.ReadLong()
			if err != nil {
				return
			}
			entry.format, err = meta.ReadByte()
			if err != nil {
				return
			}
			switch entry.format {
			case LUCENE42_DV_DELTA_COMPRESSED:
			case LUCENE42_DV_TABLE_COMPRESSED:
			case LUCENE42_DV_GCD_COMPRESSED:
			case LUCENE42_DV_UNCOMPRESSED:
			default:
				return errors.New(fmt.Sprintf("Unknown format: %v, input=%v", entry.format, meta))
			}
			if entry.format != LUCENE42_DV_UNCOMPRESSED {
				entry.packedIntsVersion, err = asInt(meta.ReadVInt())
				if err != nil {
					return
				}
			}
			fmt.Printf("Found entry [offset=%v, format=%v, packedIntsVersion=%v\n",
				entry.offset, entry.format, entry.packedIntsVersion)
			dvp.numerics[fieldNumber] = entry
		case LUCENE42_DV_BYTES:
			panic("not implemented yet")
		case LUCENE42_DV_FST:
			panic("not implemented yet")
		default:
			return errors.New(fmt.Sprintf("invalid entry type: %v, input=%v", fieldType, meta))
		}
		fieldNumber, err = asInt(meta.ReadVInt())
	}
	return
}

func asInt(n int32, err error) (n2 int, err2 error) {
	return int(n), err
}

func (dvp *Lucene42DocValuesProducer) Numeric(field *FieldInfo) (v NumericDocValues, err error) {
	dvp.lock.Lock()
	defer dvp.lock.Unlock()

	v, exists := dvp.numericInstances[int(field.Number)]
	if !exists {
		if v, err = dvp.loadNumeric(field); err == nil {
			dvp.numericInstances[int(field.Number)] = v
		}
	}
	return
}

func (dvp *Lucene42DocValuesProducer) loadNumeric(field *FieldInfo) (v NumericDocValues, err error) {
	entry := dvp.numerics[int(field.Number)]
	if err = dvp.data.Seek(entry.offset); err != nil {
		return
	}

	switch entry.format {
	case LUCENE42_DV_TABLE_COMPRESSED:
		panic("not implemented yet")
	case LUCENE42_DV_DELTA_COMPRESSED:
		panic("not implemented yet")
	case LUCENE42_DV_UNCOMPRESSED:
		bytes := make([]byte, dvp.maxDoc)
		if err = dvp.data.ReadBytes(bytes); err == nil {
			atomic.AddInt64(&dvp.ramBytesUsed, util.SizeOf(bytes))
			return func(docID int) int64 {
				return int64(bytes[docID])
			}, nil
		}
	case LUCENE42_DV_GCD_COMPRESSED:
		panic("not implemented yet")
	default:
		panic("assert fail")
	}
	return
}

func (dvp *Lucene42DocValuesProducer) Binary(field *FieldInfo) (v BinaryDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Sorted(field *FieldInfo) (v SortedDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) SortedSet(field *FieldInfo) (v SortedSetDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Close() error {
	if dvp == nil {
		return nil
	}
	return dvp.data.Close()
}

type NumericEntry struct {
	offset            int64
	format            byte
	packedIntsVersion int
}

type BinaryEntry struct {
	offset            int64
	numBytes          int64
	minLength         int
	maxLength         int
	packedIntsVersion int
	blockSize         int
}

type FSTEntry struct {
	offset  int64
	numOrds int64
}
