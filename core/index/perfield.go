package index

import (
	"github.com/balzaczyy/golucene/core/util"
	"io"
	"log"
)

// perfield/PerFieldPostingsFormat.java

/*
Enables per field postings support.

Note, when extending this class, the name Name() is written into the
index. In order for the field to be read, the name must resolve to
your implementation via LoadXYZ(). This method use hard-coded map to
resolve codec names.

Files written by each posting format have an additional suffix containing
the format name. For example, in a per-field configuration instead of
_1.prx fielnames would look like _1_Lucene40_0.prx.
*/
type PerFieldPostingsFormat struct {
}

func newPerFieldPostingsFormat(f func(field string) PostingsFormat) *PerFieldPostingsFormat {
	return &PerFieldPostingsFormat{}
}

func (pf *PerFieldPostingsFormat) Name() string {
	return "PerField40"
}

func (pf *PerFieldPostingsFormat) FieldsConsumer(state SegmentWriteState) (w FieldsConsumer, err error) {
	panic("not implemented yet")
}

func (pf *PerFieldPostingsFormat) FieldsProducer(state SegmentReadState) (r FieldsProducer, err error) {
	return newPerFieldPostingsReader(state)
}

const (
	PER_FIELD_FORMAT_KEY = "PerFieldPostingsFormat.format"
	PER_FIELD_SUFFIX_KEY = "PerFieldPostingsFormat.suffix"
)

type PerFieldPostingsReader struct {
	fields  map[string]FieldsProducer
	formats map[string]FieldsProducer
}

func newPerFieldPostingsReader(state SegmentReadState) (fp FieldsProducer, err error) {
	ans := PerFieldPostingsReader{
		make(map[string]FieldsProducer),
		make(map[string]FieldsProducer),
	}
	// Read _X.per and init each format:
	success := false
	defer func() {
		if !success {
			log.Printf("Failed to initialize PerFieldPostingsReader.")
			if err != nil {
				log.Print("DEBUG ", err)
			}
			fps := make([]FieldsProducer, 0)
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
	for _, fi := range state.fieldInfos.values {
		log.Printf("Processing %v...", fi)
		if fi.indexed {
			fieldName := fi.name
			log.Printf("Name: %v", fieldName)
			if formatName, ok := fi.attributes[PER_FIELD_FORMAT_KEY]; ok {
				log.Printf("Format: %v", formatName)
				// null formatName means the field is in fieldInfos, but has no postings!
				suffix := fi.attributes[PER_FIELD_SUFFIX_KEY]
				log.Printf("Suffix: %v", suffix)
				// assert suffix != nil
				segmentSuffix := formatName + "_" + suffix
				log.Printf("Segment suffix: %v", segmentSuffix)
				if _, ok := ans.formats[segmentSuffix]; !ok {
					log.Printf("Loading fields producer: %v", segmentSuffix)
					newReadState := state // clone
					newReadState.segmentSuffix = formatName + "_" + suffix
					fp, err = LoadFieldsProducer(formatName, newReadState)
					if err != nil {
						return fp, err
					}
					ans.formats[segmentSuffix] = fp
				}
				ans.fields[fieldName] = ans.formats[segmentSuffix]
			}
		}
	}
	success = true
	return &ans, nil
}

func (r *PerFieldPostingsReader) Terms(field string) Terms {
	if p, ok := r.fields[field]; ok {
		return p.Terms(field)
	}
	return nil
}

func (r *PerFieldPostingsReader) Close() error {
	fps := make([]FieldsProducer, 0)
	for _, v := range r.formats {
		fps = append(fps, v)
	}
	items := make([]io.Closer, len(fps))
	for i, v := range fps {
		items[i] = v
	}
	return util.Close(items...)
}

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

func newPerFieldDocValuesFormat(f func(field string) DocValuesFormat) *PerFieldDocValuesFormat {
	return &PerFieldDocValuesFormat{}
}

func (pf *PerFieldDocValuesFormat) Name() string {
	return "PerFieldDV40"
}

func (pf *PerFieldDocValuesFormat) FieldsConsumer(state SegmentWriteState) (w DocValuesConsumer, err error) {
	panic("not implemented yet")
}

func (pf *PerFieldDocValuesFormat) FieldsProducer(state SegmentReadState) (r DocValuesProducer, err error) {
	return newPerFieldDocValuesReader(state)
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
	for _, fi := range state.fieldInfos.values {
		if fi.docValueType != 0 {
			fieldName := fi.name
			if formatName, ok := fi.attributes[PER_FIELD_FORMAT_KEY]; ok {
				// null formatName means the field is in fieldInfos, but has no docvalues!
				suffix := fi.attributes[PER_FIELD_SUFFIX_KEY]
				// assert suffix != nil
				segmentSuffix := formatName + "_" + suffix
				if _, ok := ans.formats[segmentSuffix]; !ok {
					newReadState := state // clone
					newReadState.segmentSuffix = formatName + "_" + suffix
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

func (dvp *PerFieldDocValuesReader) Numeric(field FieldInfo) (v NumericDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.Numeric(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Binary(field FieldInfo) (v BinaryDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.Binary(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) Sorted(field FieldInfo) (v SortedDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
		return p.Sorted(field)
	}
	return nil, nil
}

func (dvp *PerFieldDocValuesReader) SortedSet(field FieldInfo) (v SortedSetDocValues, err error) {
	if p, ok := dvp.fields[field.name]; ok {
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
