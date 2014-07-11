package perfield

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

// perfield/PerFieldDocValuesFormat.java

/*
Enables per field docvalues support,

Note, when extending this class, the name Name() is written into the
index. In order for the field to be read, the name must resolve to
your implementation via LoadXYZ(). This method use hard-coded map to
resolve codec names.

Files written by each docvalues format have an additional suffix
containing the format name. For example, in a per-field configuration
instead of _1.dat fielnames would look like _1_Lucene40_0.dat.
*/
type PerFieldDocValuesFormat struct {
}

func NewPerFieldDocValuesFormat(f func(field string) DocValuesFormat) *PerFieldDocValuesFormat {
	return &PerFieldDocValuesFormat{}
}

func (pf *PerFieldDocValuesFormat) Name() string {
	return "PerFieldDV40"
}

func (pf *PerFieldDocValuesFormat) FieldsConsumer(state *SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("not implemented yet")
}

func (pf *PerFieldDocValuesFormat) FieldsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newPerFieldDocValuesReader(state)
}

func dvSuffix(format, suffix string) string {
	return format + "_" + suffix
}

func dvFullSegmentSuffix(outerSuffix, suffix string) string {
	if len(outerSuffix) == 0 {
		return suffix
	}
	return outerSuffix + "_" + suffix
}

type PerFieldDocValuesReader struct {
	fields  map[string]DocValuesProducer
	formats map[string]DocValuesProducer
}

func newPerFieldDocValuesReader(state SegmentReadState) (dvp DocValuesProducer, err error) {
	ans := PerFieldDocValuesReader{
		make(map[string]DocValuesProducer), make(map[string]DocValuesProducer)}
	// Read _X.per and init each format:
	success := false
	defer func() {
		if !success {
			fps := make([]DocValuesProducer, 0)
			for _, v := range ans.formats {
				fps = append(fps, v)
			}
			items := make([]io.Closer, len(fps))
			for i, v := range fps {
				items[i] = v
			}
			util.CloseWhileSuppressingError(items...)
		}
	}()
	// Read field name -> format name
	for _, fi := range state.FieldInfos.Values {
		if fi.HasDocValues() {
			fieldName := fi.Name
			if formatName := fi.Attribute(PER_FIELD_FORMAT_KEY); formatName != "" {
				// null formatName means the field is in fieldInfos, but has no docvalues!
				suffix := fi.Attribute(PER_FIELD_SUFFIX_KEY)
				// assert suffix != nil
				segmentSuffix := dvFullSegmentSuffix(state.SegmentSuffix, dvSuffix(formatName, suffix))
				if _, ok := ans.formats[segmentSuffix]; !ok {
					newReadState := state // clone
					newReadState.SegmentSuffix = formatName + "_" + suffix
					if p, err := LoadDocValuesProducer(formatName, newReadState); err == nil {
						ans.formats[segmentSuffix] = p
					}
				}
				ans.fields[fieldName] = ans.formats[segmentSuffix]
			}
		}
	}
	success = true
	return &ans, nil
}

func (dvp *PerFieldDocValuesReader) Numeric(field *FieldInfo) (v NumericDocValues, err error) {
	if p, ok := dvp.fields[field.Name]; ok {
		return p.Numeric(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Binary(field *FieldInfo) (v BinaryDocValues, err error) {
	if p, ok := dvp.fields[field.Name]; ok {
		return p.Binary(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Sorted(field *FieldInfo) (v SortedDocValues, err error) {
	if p, ok := dvp.fields[field.Name]; ok {
		return p.Sorted(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) SortedSet(field *FieldInfo) (v SortedSetDocValues, err error) {
	if p, ok := dvp.fields[field.Name]; ok {
		return p.SortedSet(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Close() error {
	fps := make([]DocValuesProducer, 0)
	for _, v := range dvp.formats {
		fps = append(fps, v)
	}
	items := make([]io.Closer, len(fps))
	for i, v := range fps {
		items[i] = v
	}
	return util.Close(items...)
}
