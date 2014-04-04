package model

import (
	"fmt"
)

type FieldInfo struct {
	// Field's name
	Name string
	// Internal field number
	Number int32

	indexed      bool
	docValueType DocValuesType

	// True if any document indexed term vectors
	storeTermVector bool

	normType      DocValuesType
	omitNorms     bool
	indexOptions  IndexOptions
	storePayloads bool

	*AttributesMixin
}

func NewFieldInfo(name string, indexed bool, number int32, storeTermVector, omitNorms, storePayloads bool,
	indexOptions IndexOptions, docValues, normsType DocValuesType, attributes map[string]string) FieldInfo {
	fi := FieldInfo{Name: name, indexed: indexed, Number: number, docValueType: docValues}
	fi.AttributesMixin = &AttributesMixin{attributes}
	if indexed {
		fi.storeTermVector = storeTermVector
		fi.storePayloads = storePayloads
		fi.omitNorms = omitNorms
		fi.indexOptions = indexOptions
		if !omitNorms {
			fi.normType = normsType
		}
	} // for non-indexed fields, leave defaults
	// assert checkConsistency()
	return fi
}

/* Returns IndexOptions for the field, or 0 if the field is not indexed */
func (info FieldInfo) IndexOptions() IndexOptions { return info.indexOptions }

/* Returns true if this field has any docValues. */
func (info FieldInfo) HasDocValues() bool {
	return int(info.docValueType) != 0
}

/* Returns true if norms are explicitly omitted for this field */
func (info FieldInfo) OmitsNorms() bool { return info.omitNorms }

/* Returns true if this field actually has any norms. */
func (info FieldInfo) HasNorms() bool { return int(info.normType) != 0 }

/* Returns true if this field is indexed. */
func (info FieldInfo) IsIndexed() bool { return info.indexed }

/* Returns true if any payloads exist for this field. */
func (info FieldInfo) HasPayloads() bool { return info.storePayloads }

/* Returns true if any term vectors exist for this field. */
func (info FieldInfo) HasVectors() bool { return info.storeTermVector }

func (fi FieldInfo) String() string {
	return fmt.Sprintf("%v-%v, isIndexed=%v, docValueType=%v, hasVectors=%v, normType=%v, omitNorms=%v, indexOptions=%v, hasPayloads=%v, attributes=%v",
		fi.Number, fi.Name, fi.indexed, fi.docValueType, fi.storeTermVector, fi.normType, fi.omitNorms, fi.indexOptions, fi.storePayloads, fi.attributes)
}

type Int32Slice []int32

func (p Int32Slice) Len() int           { return len(p) }
func (p Int32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// index/FieldInfo.java

type IndexOptions int

const (
	INDEX_OPT_DOCS_ONLY                                = IndexOptions(1)
	INDEX_OPT_DOCS_AND_FREQS                           = IndexOptions(2)
	INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS             = IndexOptions(3)
	INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS = IndexOptions(4)
)

type DocValuesType int

const (
	DOC_VALUES_TYPE_NUMERIC    = DocValuesType(1)
	DOC_VALUES_TYPE_BINARY     = DocValuesType(2)
	DOC_VALUES_TYPE_SORTED     = DocValuesType(3)
	DOC_VALUES_TYPE_SORTED_SET = DocValuesType(4)
)
