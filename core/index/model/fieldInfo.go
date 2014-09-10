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

	dvGen int64
}

func NewFieldInfo(name string, indexed bool, number int32,
	storeTermVector, omitNorms, storePayloads bool,
	indexOptions IndexOptions, docValues, normsType DocValuesType,
	dvGen int64, attributes map[string]string) *FieldInfo {

	assert(indexOptions > 0)
	assert(indexOptions <= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)

	fi := &FieldInfo{Name: name, indexed: indexed, Number: number, docValueType: docValues}
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
	fi.dvGen = dvGen
	fi.checkConsistency()
	return fi
}

func (info *FieldInfo) checkConsistency() {
	if !info.indexed {
		assert(!info.storeTermVector)
		assert(!info.storePayloads)
		assert(!info.omitNorms)
		assert(int(info.normType) == 0)
		assert(int(info.indexOptions) == 0)
	} else {
		assert(int(info.indexOptions) != 0)
		if info.omitNorms {
			assert(int(info.normType) == 0)
		}
		// Cannot store payloads unless positions are indexed:
		assert(int(info.indexOptions) >= int(INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS) || !info.storePayloads)
	}

	assert(info.dvGen == -1 || int(info.docValueType) != 0)
}

func (info *FieldInfo) Update(ft IndexableFieldType) {
	info.update(ft.Indexed(), false, ft.OmitNorms(), false, ft.IndexOptions())
}

func (info *FieldInfo) update(indexed, storeTermVector, omitNorms, storePayloads bool, indexOptions IndexOptions) {
	info.indexed = info.indexed || indexed // once indexed, always indexed
	if indexed {                           // if updated field data is not for indexing, leave the updates out
		info.storeTermVector = info.storeTermVector || storeTermVector // once vector, always vector
		info.storePayloads = info.storePayloads || storePayloads
		if info.omitNorms != omitNorms {
			info.omitNorms = true // if one require omitNorms at least once, it remains off for life
			info.normType = 0
		}
		if info.indexOptions != indexOptions {
			if info.indexOptions == 0 {
				info.indexOptions = indexOptions
			} else {
				// downgrade
				if info.indexOptions > indexOptions {
					info.indexOptions = indexOptions
				}
			}
			if info.indexOptions < INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS {
				// cannot store payloads if we don't store positions:
				info.storePayloads = false
			}
		}
	}
	info.checkConsistency()
}

func (info *FieldInfo) SetDocValueType(v DocValuesType) {
	assert2(int(info.docValueType) != 0 && info.docValueType != v,
		"cannot change DocValues type from %v to %v for field '%v'",
		info.docValueType, v, info.Name)
	info.docValueType = v
	info.checkConsistency()
}

/* Returns IndexOptions for the field, or 0 if the field is not indexed */
func (info *FieldInfo) IndexOptions() IndexOptions { return info.indexOptions }

/* Returns true if this field has any docValues. */
func (info *FieldInfo) HasDocValues() bool {
	return int(info.docValueType) != 0
}

/* Returns DocValueType of the docValues. This may be 0 if the fiel dhas no docValues. */
func (info *FieldInfo) DocValuesType() DocValuesType {
	return info.docValueType
}

func (info *FieldInfo) DocValuesGen() int64 {
	return info.dvGen
}

/* Returns DocValuesType of the norm. This may be 0 if the field has no norms. */
func (info *FieldInfo) NormType() DocValuesType {
	return info.normType
}

func (info *FieldInfo) SetNormValueType(typ DocValuesType) {
	assert2(info.normType == DocValuesType(0) || info.normType == typ,
		"cannot change Norm type from %v to %v for field '%v'",
		info.normType, typ, info.Name)
	info.normType = typ
	info.checkConsistency()
}

/* Returns true if norms are explicitly omitted for this field */
func (info *FieldInfo) OmitsNorms() bool { return info.omitNorms }

/* Returns true if this field actually has any norms. */
func (info *FieldInfo) HasNorms() bool { return int(info.normType) != 0 }

/* Returns true if this field is indexed. */
func (info *FieldInfo) IsIndexed() bool { return info.indexed }

/* Returns true if any payloads exist for this field. */
func (info *FieldInfo) HasPayloads() bool { return info.storePayloads }

/* Returns true if any term vectors exist for this field. */
func (info *FieldInfo) HasVectors() bool { return info.storeTermVector }

/*
Puts a codec attribute value.

This is a key-value mapping for the field that the codec can use to
store additional metadata, and will be available to the codec when
reading the segment via Attribute()

If a value already exists ofr the field, it will be replaced with the
new value.
*/
func (fi *FieldInfo) PutAttribute(key, value string) string {
	if fi.attributes == nil {
		fi.attributes = make(map[string]string)
	}
	var v string
	v, fi.attributes[key] = fi.attributes[key], value
	return v
}

func (fi *FieldInfo) String() string {
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
	DOC_VALUES_TYPE_NUMERIC        = DocValuesType(1)
	DOC_VALUES_TYPE_BINARY         = DocValuesType(2)
	DOC_VALUES_TYPE_SORTED         = DocValuesType(3)
	DOC_VALUES_TYPE_SORTED_SET     = DocValuesType(4)
	DOC_VALUES_TYPE_SORTED_NUMERIC = DocValuesType(5)
)

// index/IndexableFieldType.java

/**
 * Describes the properties of a field.
 */
type IndexableFieldType interface {
	/** True if this field should be indexed (inverted) */
	Indexed() bool
	/** True if the field's value should be stored */
	Stored() bool
	/**
	 * True if this field's value should be analyzed by the
	 * {@link Analyzer}.
	 * <p>
	 * This has no effect if {@link #indexed()} returns false.
	 */
	Tokenized() bool
	/**
	 * True if this field's indexed form should be also stored
	 * into term vectors.
	 * <p>
	 * This builds a miniature inverted-index for this field which
	 * can be accessed in a document-oriented way from
	 * {@link IndexReader#getTermVector(int,String)}.
	 * <p>
	 * This option is illegal if {@link #indexed()} returns false.
	 */
	StoreTermVectors() bool
	/**
	 * True if this field's token character offsets should also
	 * be stored into term vectors.
	 * <p>
	 * This option is illegal if term vectors are not enabled for the field
	 * ({@link #storeTermVectors()} is false)
	 */
	StoreTermVectorOffsets() bool

	/**
	 * True if this field's token positions should also be stored
	 * into the term vectors.
	 * <p>
	 * This option is illegal if term vectors are not enabled for the field
	 * ({@link #storeTermVectors()} is false).
	 */
	StoreTermVectorPositions() bool

	/**
	 * True if this field's token payloads should also be stored
	 * into the term vectors.
	 * <p>
	 * This option is illegal if term vector positions are not enabled
	 * for the field ({@link #storeTermVectors()} is false).
	 */
	StoreTermVectorPayloads() bool

	/**
	 * True if normalization values should be omitted for the field.
	 * <p>
	 * This saves memory, but at the expense of scoring quality (length normalization
	 * will be disabled), and if you omit norms, you cannot use index-time boosts.
	 */
	OmitNorms() bool

	/** {@link IndexOptions}, describing what should be
	 * recorded into the inverted index */
	IndexOptions() IndexOptions

	/**
	 * DocValues {@link DocValuesType}: if non-null then the field's value
	 * will be indexed into docValues.
	 */
	DocValueType() DocValuesType
}
