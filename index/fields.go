package index

import (
	"fmt"
	"sort"
)

type Fields interface {
	// Iterator of string
	Terms(field string) Terms
}

type MultiFields struct {
	subs      []Fields
	subSlices []ReaderSlice
	termsMap  map[string]Terms
}

func NewMultiFields(subs []Fields, subSlices []ReaderSlice) MultiFields {
	return MultiFields{subs, subSlices, make(map[string]Terms)}
}

func (mf MultiFields) Terms(field string) Terms {
	if ans, ok := mf.termsMap[field]; ok {
		return ans
	}

	// Lazy init: first time this field is requested, we
	// create & add to terms:
	subs2 := make([]Terms, 0)
	slices2 := make([]ReaderSlice, 0)

	// Gather all sub-readers that share this field
	for i, v := range mf.subs {
		terms := v.Terms(field)
		if terms.Iterator != nil {
			subs2 = append(subs2, terms)
			slices2 = append(slices2, mf.subSlices[i])
		}
	}
	if len(subs2) == 0 {
		return nil
		// don't cache this case with an unbounded cache, since the number of fields that don't exist
		// is unbounded.
	}
	ans := NewMultiTerms(subs2, slices2)
	mf.termsMap[field] = ans
	return ans
}

func GetMultiFields(r IndexReader) Fields {
	leaves := r.Context().Leaves()
	switch len(leaves) {
	case 0:
		// no fields
		return nil
	case 1:
		// already an atomic reader / reader with one leave
		return leaves[0].Reader().(*AtomicReader).Fields()
	default:
		fields := make([]Fields, 0)
		slices := make([]ReaderSlice, 0)
		for _, ctx := range leaves {
			f := ctx.Reader().(*AtomicReader).Fields()
			if f == nil {
				continue
			}
			fields = append(fields, f)
			slices = append(slices, ReaderSlice{ctx.DocBase, r.MaxDoc(), len(fields)})
		}
		switch len(fields) {
		case 0:
			return nil
		case 1:
			return fields[0]
		default:
			return NewMultiFields(fields, slices)
		}
	}
}

func GetMultiTerms(r IndexReader, field string) Terms {
	fields := GetMultiFields(r)
	if fields.Terms == nil {
		return nil
	}
	return fields.Terms(field)
}

type FieldInfo struct {
	// Field's name
	name string
	// Internal field number
	number int32

	indexed      bool
	docValueType DocValuesType

	// True if any document indexed term vectors
	storeTermVector bool

	normType      DocValuesType
	omitNorms     bool
	indexOptions  IndexOptions
	storePayloads bool

	attributes map[string]string
}

func NewFieldInfo(name string, indexed bool, number int32, storeTermVector, omitNorms, storePayloads bool,
	indexOptions IndexOptions, docValues, normsType DocValuesType, attributes map[string]string) FieldInfo {
	fi := FieldInfo{name: name, indexed: indexed, number: number, docValueType: docValues, attributes: attributes}
	if indexed {
		fi.storeTermVector = storeTermVector
		fi.storePayloads = storePayloads
		fi.omitNorms = omitNorms
		fi.indexOptions = indexOptions
		if omitNorms {
			fi.normType = normsType
		}
	} // for non-indexed fields, leave defaults
	// assert checkConsistency()
	return fi
}

type Int32Slice []int32

func (p Int32Slice) Len() int           { return len(p) }
func (p Int32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type FieldInfos struct {
	hasFreq      bool
	hasProx      bool
	hasPayloads  bool
	hasOffsets   bool
	hasVectors   bool
	hasNorms     bool
	hasDocValues bool

	byNumber map[int32]FieldInfo
	byName   map[string]FieldInfo
	values   []FieldInfo // sorted by ID
}

func NewFieldInfos(infos []FieldInfo) FieldInfos {
	self := FieldInfos{byNumber: make(map[int32]FieldInfo), byName: make(map[string]FieldInfo)}

	numbers := make([]int32, 0)
	for _, info := range infos {
		if prev, ok := self.byNumber[info.number]; ok {
			panic(fmt.Sprintf("duplicate field numbers: %v and %v have: %v", prev.name, info.name, info.number))
		}
		self.byNumber[info.number] = info
		numbers = append(numbers, info.number)
		if prev, ok := self.byName[info.name]; ok {
			panic(fmt.Sprintf("duplicate field names: %v and %v have: %v", prev.number, info.number, info.name))
		}
		self.byName[info.name] = info

		self.hasVectors = self.hasVectors || info.storeTermVector
		self.hasProx = self.hasProx || info.indexed && info.indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
		self.hasFreq = self.hasFreq || info.indexed && info.indexOptions != INDEX_OPT_DOCS_ONLY
		self.hasOffsets = self.hasOffsets || info.indexed && info.indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		self.hasNorms = self.hasNorms || info.normType != 0
		self.hasDocValues = self.hasDocValues || info.docValueType != 0
		self.hasPayloads = self.hasPayloads || info.storePayloads
	}

	sort.Sort(Int32Slice(numbers))
	self.values = make([]FieldInfo, len(infos))
	for i, v := range numbers {
		self.values[int32(i)] = self.byNumber[v]
	}

	return self
}

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
