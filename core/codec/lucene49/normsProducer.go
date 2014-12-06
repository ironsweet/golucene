package lucene49

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"reflect"
	"sync"
	"sync/atomic"
)

// lucene49/Lucene49NormsProduer.java

type NormsEntry struct {
	format byte
	offset int64
}

type NormsProducer struct {
	sync.Locker

	norms   map[int]*NormsEntry
	data    store.IndexInput
	version int32

	instances map[int]NumericDocValues

	maxDoc       int
	ramBytesUsed int64 // atomic
}

func newLucene49NormsProducer(state SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (np *NormsProducer, err error) {

	np = &NormsProducer{
		Locker:       new(sync.Mutex),
		norms:        make(map[int]*NormsEntry),
		instances:    make(map[int]NumericDocValues),
		maxDoc:       state.SegmentInfo.DocCount(),
		ramBytesUsed: util.ShallowSizeOfInstance(reflect.TypeOf(np)),
	}
	metaName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, metaExtension)
	// read in the entries from the metadta file.
	var in store.ChecksumIndexInput
	if in, err = state.Dir.OpenChecksumInput(metaName, state.Context); err != nil {
		return nil, err
	}

	if err = func() error {
		var success = false
		defer func() {
			if success {
				err = util.Close(in)
			} else {
				util.CloseWhileSuppressingError(in)
			}
		}()

		if np.version, err = codec.CheckHeader(in, metaCodec, VERSION_START, VERSION_CURRENT); err != nil {
			return err
		}
		if err = np.readFields(in, state.FieldInfos); err != nil {
			return err
		}
		if _, err = codec.CheckFooter(in); err != nil {
			return err
		}
		success = true
		return nil
	}(); err != nil {
		return nil, err
	}

	dataName := util.SegmentFileName(state.SegmentInfo.Name, state.SegmentSuffix, dataExtension)
	if np.data, err = state.Dir.OpenInput(dataName, state.Context); err != nil {
		return nil, err
	}
	var success = false
	defer func() {
		if !success {
			util.CloseWhileSuppressingError(np.data)
		}
	}()

	var version2 int32
	if version2, err = codec.CheckHeader(np.data, dataCodec, VERSION_START, VERSION_CURRENT); err != nil {
		return nil, err
	}
	if version2 != np.version {
		return nil, errors.New("Format versions mismatch")
	}

	// NOTE: data file is too costly to verify checksum against all the
	// bytes on open, but fo rnow we at least verify proper structure
	// of the checksum footer: which looks for FOOTER_MATIC +
	// algorithmID. This is cheap and can detect some forms of
	// corruption such as file trucation.
	if _, err = codec.RetrieveChecksum(np.data); err != nil {
		return nil, err
	}

	success = true

	return np, nil
}

func (np *NormsProducer) readFields(meta store.IndexInput, infos FieldInfos) (err error) {
	var fieldNumber int32
	if fieldNumber, err = meta.ReadVInt(); err != nil {
		return err
	}
	for fieldNumber != -1 {
		info := infos.FieldInfoByNumber(int(fieldNumber))
		if info == nil {
			return errors.New(fmt.Sprintf("Invalid field number: %v (resource=%v)", fieldNumber, meta))
		} else if !info.HasNorms() {
			return errors.New(fmt.Sprintf("Invalid field: %v (resource=%v)", info.Name, meta))
		}
		var format byte
		if format, err = meta.ReadByte(); err != nil {
			return err
		}
		var offset int64
		if offset, err = meta.ReadLong(); err != nil {
			return err
		}
		entry := &NormsEntry{
			format: format,
			offset: offset,
		}
		if format > UNCOMPRESSED {
			return errors.New(fmt.Sprintf("Unknown format: %v, input=%v", format, meta))
		}
		np.norms[int(fieldNumber)] = entry
		if fieldNumber, err = meta.ReadVInt(); err != nil {
			return err
		}
	}
	return nil
}

func (np *NormsProducer) Numeric(field *FieldInfo) (NumericDocValues, error) {
	np.Lock()
	defer np.Unlock()

	instance, ok := np.instances[int(field.Number)]
	if !ok {
		var err error
		if instance, err = np.loadNorms(field); err != nil {
			return nil, err
		}
		np.instances[int(field.Number)] = instance
	}
	return instance, nil
}

func (np *NormsProducer) loadNorms(field *FieldInfo) (NumericDocValues, error) {
	entry, ok := np.norms[int(field.Number)]
	assert(ok)
	switch entry.format {
	case CONST_COMPRESSED:
		return func(int) int64 { return entry.offset }, nil
	case UNCOMPRESSED:
		panic("not implemented yet")
	case DELTA_COMPRESSED:
		panic("not implemented yet")
	case TABLE_COMPRESSED:
		var err error
		if err = np.data.Seek(entry.offset); err == nil {
			var packedVersion int32
			if packedVersion, err = np.data.ReadVInt(); err == nil {
				var size int
				if size, err = int32ToInt(np.data.ReadVInt()); err == nil {
					if size > 256 {
						return nil, errors.New(fmt.Sprintf(
							"TABLE_COMPRESSED cannot have more than 256 distinct values, input=%v",
							np.data))
					}
					decode := make([]int64, size)
					for i, _ := range decode {
						if decode[i], err = np.data.ReadLong(); err != nil {
							break
						}
					}
					if err == nil {
						var formatId int
						if formatId, err = int32ToInt(np.data.ReadVInt()); err == nil {
							var bitsPerValue int32
							if bitsPerValue, err = np.data.ReadVInt(); err == nil {
								var ordsReader packed.PackedIntsReader
								if ordsReader, err = packed.ReaderNoHeader(np.data,
									packed.PackedFormat(formatId), packedVersion,
									int32(np.maxDoc), uint32(bitsPerValue)); err == nil {

									atomic.AddInt64(&np.ramBytesUsed, util.SizeOf(decode)+ordsReader.RamBytesUsed())
									return func(docId int) int64 {
										return decode[int(ordsReader.Get(docId))]
									}, nil
								}
							}
						}
					}
				}
			}
		}
		if err != nil {
			return nil, err
		}
	default:
		panic("assert fail")
	}
	panic("should not be here")
}

func int32ToInt(n int32, err error) (int, error) {
	return int(n), err
}

func (np *NormsProducer) Binary(field *FieldInfo) (BinaryDocValues, error) {
	panic("not supported")
}

func (np *NormsProducer) Sorted(field *FieldInfo) (SortedDocValues, error) {
	panic("not supported")
}

func (np *NormsProducer) SortedSet(field *FieldInfo) (SortedSetDocValues, error) {
	panic("not supported")
}

func (np *NormsProducer) Close() error {
	return np.data.Close()
}
