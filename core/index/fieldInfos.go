package index

import (
	"fmt"
	"log"
	"sort"
)

// Collection of FieldInfo(s) (accessible by number of by name)
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

func (fis FieldInfos) String() string {
	return fmt.Sprintf(`
hasFreq = %v
hasProx = %v
hasPayloads = %v
hasOffsets = %v
hasVectors = %v
hasNorms = %v
hasDocValues = %v
%v`, fis.hasFreq, fis.hasProx, fis.hasPayloads, fis.hasOffsets, fis.hasVectors, fis.hasNorms, fis.hasDocValues, fis.values)
}

type FieldNumbers struct {
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

func newFieldNumbers() *FieldNumbers {
	return &FieldNumbers{
		nameToNumber:  make(map[string]int),
		numberToName:  make(map[int]string),
		docValuesType: make(map[string]DocValuesType),
	}
}

/*
Returns the global field number for the given field name. If the name
does not exist yet it tries to add it with the given preferred field
number assigned if possible otherwise the first unassigned field
number is used as the field number.
*/
func (fn *FieldNumbers) addOrGet(name string, preferredNumber int, dv DocValuesType) int {
	if dv != 0 {
		currentDv, ok := fn.docValuesType[name]
		if !ok || currentDv == 0 {
			fn.docValuesType[name] = dv
		} else if currentDv != dv {
			log.Panicf("cannot change DocValues type from %v to %v for field '%v'", currentDv, dv, name)
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
			for _, ok = fn.numberToName[fn.lowestUnassignedFieldNumber]; ok; {
				// might not be up to date - lets do the work once needed
				fn.lowestUnassignedFieldNumber++
			}
			number = fn.lowestUnassignedFieldNumber
		}

		fn.numberToName[number] = name
		fn.nameToNumber[name] = number
	}
	return number
}

type FieldInfosBuilder struct {
	byName             map[string]FieldInfo
	globalFieldNumbers *FieldNumbers
}

func newFieldInfosBuilder(globalFieldNumbers *FieldNumbers) *FieldInfosBuilder {
	assert(globalFieldNumbers != nil)
	return &FieldInfosBuilder{
		byName:             make(map[string]FieldInfo),
		globalFieldNumbers: globalFieldNumbers,
	}
}

func (b *FieldInfosBuilder) finish() FieldInfos {
	var infos []FieldInfo
	for _, v := range b.byName {
		infos = append(infos, v)
	}
	return NewFieldInfos(infos)
}
