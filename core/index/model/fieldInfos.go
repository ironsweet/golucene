package model

import (
	"fmt"
	"sort"
	"sync"
)

// Collection of FieldInfo(s) (accessible by number of by name)
type FieldInfos struct {
	HasFreq      bool
	HasProx      bool
	HasPayloads  bool
	HasOffsets   bool
	HasVectors   bool
	HasNorms     bool
	HasDocValues bool

	byNumber map[int32]*FieldInfo
	byName   map[string]*FieldInfo
	Values   []*FieldInfo // sorted by ID
}

func NewFieldInfos(infos []*FieldInfo) FieldInfos {
	self := FieldInfos{byNumber: make(map[int32]*FieldInfo), byName: make(map[string]*FieldInfo)}

	numbers := make([]int32, 0)
	for _, info := range infos {
		assert2(info.Number >= 0, "illegal field number: %v for field %v", info.Number, info.Name)
		if prev, ok := self.byNumber[info.Number]; ok {
			panic(fmt.Sprintf("duplicate field numbers: %v and %v have: %v", prev.Name, info.Name, info.Number))
		}
		self.byNumber[info.Number] = info
		numbers = append(numbers, info.Number)
		if prev, ok := self.byName[info.Name]; ok {
			panic(fmt.Sprintf("duplicate field names: %v and %v have: %v", prev.Number, info.Number, info.Name))
		}
		self.byName[info.Name] = info

		self.HasVectors = self.HasVectors || info.storeTermVector
		self.HasProx = self.HasProx || info.indexed && info.indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS
		self.HasFreq = self.HasFreq || info.indexed && info.indexOptions != INDEX_OPT_DOCS_ONLY
		self.HasOffsets = self.HasOffsets || info.indexed && info.indexOptions >= INDEX_OPT_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		self.HasNorms = self.HasNorms || info.normType != 0
		self.HasDocValues = self.HasDocValues || info.docValueType != 0
		self.HasPayloads = self.HasPayloads || info.storePayloads
	}

	sort.Sort(Int32Slice(numbers))
	self.Values = make([]*FieldInfo, len(infos))
	for i, v := range numbers {
		self.Values[int32(i)] = self.byNumber[v]
	}

	return self
}

/* Returns the number of fields */
func (infos FieldInfos) Size() int {
	assert(len(infos.byNumber) == len(infos.byName))
	return len(infos.byNumber)
}

/* Return the FieldInfo object referenced by the field name */
func (infos FieldInfos) FieldInfoByName(fieldName string) *FieldInfo {
	return infos.byName[fieldName]
}

/* Return the FieldInfo object referenced by the fieldNumber. */
func (infos FieldInfos) FieldInfoByNumber(fieldNumber int) *FieldInfo {
	assert2(fieldNumber >= 0, "Illegal field number: %v", fieldNumber)
	return infos.byNumber[int32(fieldNumber)]
}

func (fis FieldInfos) String() string {
	return fmt.Sprintf(`
hasFreq = %v
hasProx = %v
hasPayloads = %v
hasOffsets = %v
hasVectors = %v
hasNorms = %v
hasDocValues = %v
%v`, fis.HasFreq, fis.HasProx, fis.HasPayloads, fis.HasOffsets,
		fis.HasVectors, fis.HasNorms, fis.HasDocValues, fis.Values)
}

type FieldNumbers struct {
	sync.Locker
	numberToName map[int]string
	nameToNumber map[string]int
	// We use this to enforce that a given field never changes DV type,
	// even across segments / IndexWriter sessions:
	docValuesType map[string]DocValuesType
	// TODO: we should similarly catch an attempt to turn norms back on
	// after they were already ommitted; today we silently discard the
	// norm but this is badly trappy
	lowestUnassignedFieldNumber int
}

func NewFieldNumbers() *FieldNumbers {
	return &FieldNumbers{
		Locker:                      &sync.Mutex{},
		nameToNumber:                make(map[string]int),
		numberToName:                make(map[int]string),
		docValuesType:               make(map[string]DocValuesType),
		lowestUnassignedFieldNumber: -1,
	}
}

func (fn *FieldNumbers) AddOrGet(info *FieldInfo) int {
	return fn.addOrGet(info.Name, int(info.Number), info.docValueType)
}

/*
Returns the global field number for the given field name. If the name
does not exist yet it tries to add it with the given preferred field
number assigned if possible otherwise the first unassigned field
number is used as the field number.
*/
func (fn *FieldNumbers) addOrGet(name string, preferredNumber int, dv DocValuesType) int {
	fn.Lock()
	defer fn.Unlock()

	if dv != 0 {
		currentDv, ok := fn.docValuesType[name]
		if !ok || currentDv == 0 {
			fn.docValuesType[name] = dv
		} else {
			assert2(currentDv == dv,
				"cannot change DocValues type from %v to %v for field '%v'",
				currentDv, dv, name)
		}
	}
	number, ok := fn.nameToNumber[name]
	if !ok {
		_, ok = fn.numberToName[preferredNumber]
		if preferredNumber != -1 && !ok {
			// cool - we can use this number globally
			number = preferredNumber
		} else {
			// find a new FieldNumber
			fn.lowestUnassignedFieldNumber++
			for _, ok = fn.numberToName[fn.lowestUnassignedFieldNumber]; ok; {
				// might not be up to date - lets do the work once needed
				fn.lowestUnassignedFieldNumber++
				_, ok = fn.numberToName[fn.lowestUnassignedFieldNumber]
			}
			number = fn.lowestUnassignedFieldNumber
		}

		fn.numberToName[number] = name
		fn.nameToNumber[name] = number
	}
	return number
}

type FieldInfosBuilder struct {
	byName             map[string]*FieldInfo
	globalFieldNumbers *FieldNumbers
}

func NewFieldInfosBuilder(globalFieldNumbers *FieldNumbers) *FieldInfosBuilder {
	assert(globalFieldNumbers != nil)
	return &FieldInfosBuilder{
		byName:             make(map[string]*FieldInfo),
		globalFieldNumbers: globalFieldNumbers,
	}
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

/*
NOTE: this method does not carry over termVector booleans nor
docValuesType; the indexer chain  (TermVectorsConsumerPerField,
DocFieldProcessor) must set these fields when they succeed in
consuming the document
*/
func (b *FieldInfosBuilder) AddOrUpdate(name string, fieldType IndexableFieldType) *FieldInfo {
	// TODO: really, indexer shouldn't even call this method (it's only
	// called from DocFieldProcessor); rather, each component in the
	// chain should update what it "owns". E.g., fieldType.indexOptions()
	// should be updated by maybe FreqProxTermsWriterPerField:
	return b.addOrUpdateInternal(name, -1, fieldType.Indexed(), false,
		fieldType.OmitNorms(), false,
		fieldType.IndexOptions(), fieldType.DocValueType(), DocValuesType(0))
}

func (b *FieldInfosBuilder) addOrUpdateInternal(name string,
	preferredFieldNumber int, isIndexed bool, storeTermVector bool,
	omitNorms bool, storePayloads bool, indexOptions IndexOptions,
	docValues DocValuesType, normType DocValuesType) *FieldInfo {

	if fi, ok := b.byName[name]; ok {
		panic("not implemented yet")
		return fi
	} else {
		// This field wasn't yet added to this in-RAM segment's
		// FieldInfos, so now we get a global number for this field. If
		// the field was seen before then we'll get the same name and
		// number, else we'll allocate a new one:
		fieldNumber := int32(b.globalFieldNumbers.addOrGet(name, preferredFieldNumber, docValues))
		fi = NewFieldInfo(name, isIndexed, fieldNumber, storeTermVector,
			omitNorms, storePayloads, indexOptions, docValues, normType, -1, nil)
		b.byName[fi.Name] = fi
		return fi
	}
}

func (b *FieldInfosBuilder) Finish() FieldInfos {
	var infos []*FieldInfo
	for _, v := range b.byName {
		infos = append(infos, v)
	}
	return NewFieldInfos(infos)
}
