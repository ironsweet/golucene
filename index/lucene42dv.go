package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/store"
	"github.com/balzaczyy/golucene/util"
	"log"
	"sync"
)

const (
	LUCENE42_DV_DATA_CODEC         = "Lucene42DocValuesData"
	LUCENE42_DV_DATA_EXTENSION     = "dvd"
	LUCENE42_DV_METADATA_CODEC     = "Lucene42DocValuesMetadata"
	LUCENE42_DV_METADATA_EXTENSION = "dvm"

	LUCENE42_DV_VERSION_START           = 0
	LUCENE42_DV_VERSION_GCD_COMPRESSION = 1
	LUCENE42_DV_VERSION_CURRENT         = LUCENE42_DV_VERSION_GCD_COMPRESSION

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

	maxDoc int
}

func newLucene42DocValuesProducer(state SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (dvp *Lucene42DocValuesProducer, err error) {
	dvp = &Lucene42DocValuesProducer{
		numericInstances: make(map[int]NumericDocValues),
	}
	dvp.maxDoc = int(state.segmentInfo.docCount)
	metaName := util.SegmentFileName(state.segmentInfo.name, state.segmentSuffix, metaExtension)
	// read in the entries from the metadata file.
	in, err := state.dir.OpenInput(metaName, state.context)
	if err != nil {
		return dvp, err
	}
	success := false
	defer func() {
		if success {
			err = util.Close(in)
		} else {
			util.CloseWhileSuppressingError(in)
		}
	}()

	version, err := codec.CheckHeader(in, metaCodec, LUCENE42_DV_VERSION_START, LUCENE42_DV_VERSION_CURRENT)
	if err != nil {
		return dvp, err
	}
	dvp.numerics = make(map[int]NumericEntry)
	dvp.binaries = make(map[int]BinaryEntry)
	dvp.fsts = make(map[int]FSTEntry)
	err = dvp.readFields(in)
	if err != nil {
		return dvp, err
	}
	success = true

	success = false
	dataName := util.SegmentFileName(state.segmentInfo.name, state.segmentSuffix, dataExtension)
	dvp.data, err = state.dir.OpenInput(dataName, state.context)
	if err != nil {
		return dvp, err
	}
	version2, err := codec.CheckHeader(dvp.data, dataCodec, LUCENE42_DV_VERSION_START, LUCENE42_DV_VERSION_CURRENT)
	if err != nil {
		return dvp, err
	}

	if version != version2 {
		return dvp, errors.New("Format versions mismatch")
	}
	return dvp, nil
}

/*
Lucene42DocValuesProducer.java/4.5.1/L138
*/
func (dvp *Lucene42DocValuesProducer) readFields(meta store.IndexInput) (err error) {
	var fieldNumber int
	var fieldType byte
	fieldNumber, err = asInt(meta.ReadVInt())
	for fieldNumber != -1 && err == nil {
		fieldType, err = meta.ReadByte()
		if err != nil {
			break
		}
		switch fieldType {
		case LUCENE42_DV_NUMBER:
			entry := NumericEntry{}
			entry.offset, err = meta.ReadLong()
			if err != nil {
				return err
			}
			entry.format, err = meta.ReadByte()
			if err != nil {
				return err
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
					return err
				}
			}
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

func (dvp *Lucene42DocValuesProducer) Numeric(field FieldInfo) (v NumericDocValues, err error) {
	dvp.lock.Lock()
	defer dvp.lock.Unlock()

	v, exists := dvp.numericInstances[int(field.number)]
	if !exists {
		if v, err = dvp.loadNumeric(field); err == nil {
			dvp.numericInstances[int(field.number)] = v
		}
	}
	return
}

func (dvp *Lucene42DocValuesProducer) loadNumeric(field FieldInfo) (v NumericDocValues, err error) {
	entry := dvp.numerics[int(field.number)]
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
			return func(docID int) int64 {
				return int64(bytes[docID])
			}, nil
		}
	case LUCENE42_DV_GCD_COMPRESSED:
		panic("not implemented yet")
	default:
		err = errors.New("assert fail")
	}
	return
}

func (dvp *Lucene42DocValuesProducer) Binary(field FieldInfo) (v BinaryDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Sorted(field FieldInfo) (v SortedDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) SortedSet(field FieldInfo) (v SortedSetDocValues, err error) {
	panic("not implemented yet")
	return nil, nil
}

func (dvp *Lucene42DocValuesProducer) Close() error {
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
