package document

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/util"
)

// document/FieldType.java

// Data type of the numeric value
type NumericType int

const (
	FIELD_TYPE_NUMERIC_INT    = 1 // 32-bit integer numeric type
	FIELD_TYPE_NUMERIC_LONG   = 2 // 64-bit long numeric type
	FIELD_TYPE_NUMERIC_FLOAT  = 3 // 32-bit float numeric type
	FIELD_TYPE_NUMERIC_DOUBLE = 4 // 64-bit double numeric type
)

// Describes the properties of a field.
type FieldType struct {
	indexed                  bool
	stored                   bool
	_tokenized               bool
	storeTermVectors         bool
	storeTermVectorOffsets   bool
	storeTermVectorPositions bool
	storeTermVectorPayloads  bool
	_omitNorms               bool
	_indexOptions            model.IndexOptions
	numericType              NumericType
	frozen                   bool
	numericPrecisionStep     int
	_docValueType            model.DocValuesType
}

// Create a new mutable FieldType with all of the properties from <code>ref</code>
func NewFieldTypeFrom(ref *FieldType) *FieldType {
	ft := newFieldType()
	ft.indexed = ref.indexed
	ft.stored = ref.stored
	ft._tokenized = ref._tokenized
	ft.storeTermVectors = ref.storeTermVectors
	ft.storeTermVectorOffsets = ref.storeTermVectorOffsets
	ft.storeTermVectorPositions = ref.storeTermVectorPositions
	ft.storeTermVectorPayloads = ref.storeTermVectorPayloads
	ft._omitNorms = ref._omitNorms
	ft._indexOptions = ref._indexOptions
	ft._docValueType = ref._docValueType
	ft.numericType = ref.numericType
	// Do not copy frozen!
	return ft
}

// Create a new FieldType with default properties.
func newFieldType() *FieldType {
	return &FieldType{
		_tokenized:           true,
		_indexOptions:        model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS,
		numericPrecisionStep: util.NUMERIC_PRECISION_STEP_DEFAULT,
	}
}

func (ft *FieldType) checkIfFrozen() {
	assert2(!ft.frozen, "this FieldType is already frozen and cannot be changed")
}

func (ft *FieldType) Indexed() bool     { return ft.indexed }
func (ft *FieldType) SetIndexed(v bool) { ft.checkIfFrozen(); ft.indexed = v }
func (ft *FieldType) Stored() bool      { return ft.stored }
func (ft *FieldType) SetStored(v bool)  { ft.checkIfFrozen(); ft.stored = v }
func (ft *FieldType) Tokenized() bool   { return ft._tokenized }

func (ft *FieldType) StoreTermVectors() bool       { return ft.storeTermVectors }
func (ft *FieldType) SetStoreTermVectors(v bool)   { ft.checkIfFrozen(); ft.storeTermVectors = v }
func (ft *FieldType) StoreTermVectorOffsets() bool { return ft.storeTermVectorOffsets }
func (ft *FieldType) SetStoreTermVectorOffsets(v bool) {
	ft.checkIfFrozen()
	ft.storeTermVectorOffsets = v
}
func (ft *FieldType) StoreTermVectorPositions() bool { return ft.storeTermVectorPositions }
func (ft *FieldType) SetStoreTermVectorPositions(v bool) {
	ft.checkIfFrozen()
	ft.storeTermVectorPositions = v
}
func (ft *FieldType) StoreTermVectorPayloads() bool { return ft.storeTermVectorPayloads }
func (ft *FieldType) SetStoreTermVectorPayloads(v bool) {
	ft.checkIfFrozen()
	ft.storeTermVectorPayloads = v
}

func (ft *FieldType) OmitNorms() bool                   { return ft._omitNorms }
func (ft *FieldType) IndexOptions() model.IndexOptions  { return ft._indexOptions }
func (ft *FieldType) NumericType() NumericType          { return ft.numericType }
func (ft *FieldType) DocValueType() model.DocValuesType { return ft._docValueType }

// Prints a Field for human consumption.
func (ft *FieldType) String() string {
	var buf bytes.Buffer
	if ft.Stored() {
		buf.WriteString("stored")
	}
	if ft.Indexed() {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("indexed")
		if ft.Tokenized() {
			buf.WriteString(",tokenized")
		}
		if ft.StoreTermVectors() {
			buf.WriteString(",termVector")
		}
		if ft.StoreTermVectorOffsets() {
			buf.WriteString(",termVectorOffsets")
		}
		if ft.StoreTermVectorPositions() {
			buf.WriteString(",termVectorPosition")
		}
		if ft.StoreTermVectorPayloads() {
			buf.WriteString(",termVectorPayloads")
		}
		if ft.OmitNorms() {
			buf.WriteString(",omitNorms")
		}
		if ft.IndexOptions() != model.INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS {
			fmt.Fprintf(&buf, ",indexOptions=%v", ft.IndexOptions())
		}
		if ft.numericType != 0 {
			fmt.Fprintf(&buf, ",numericType=%v,numericPrecisionStep=%v", ft.numericType, ft.numericPrecisionStep)
		}
	}
	if ft.DocValueType() != 0 {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		fmt.Fprintf(&buf, "docValueType=%v", ft.DocValueType())
	}
	return buf.String()
}
