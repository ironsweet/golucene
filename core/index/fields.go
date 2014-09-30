package index

import (
	. "github.com/balzaczyy/golucene/core/index/model"
)

type MultiFields struct {
	subs      []Fields
	subSlices []ReaderSlice
	termsMap  map[string]Terms // synchronized
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
	// log.Print("Obtaining MultiFields from ", r)
	leaves := r.Leaves()
	switch len(leaves) {
	case 0:
		// log.Print("No fields are found.")
		// no fields
		return nil
	case 1:
		// already an atomic reader / reader with one leave
		return leaves[0].Reader().(AtomicReader).Fields()
	default:
		fields := make([]Fields, 0)
		slices := make([]ReaderSlice, 0)
		for _, ctx := range leaves {
			f := ctx.Reader().(AtomicReader).Fields()
			if f == nil {
				continue
			}
			fields = append(fields, f)
			slices = append(slices, ReaderSlice{ctx.DocBase, r.MaxDoc(), len(fields)})
		}
		// log.Printf("Found %v fields in %v slices.", len(fields), len(slices))
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
	// log.Printf("Loading field '%v' from %v", field, r)
	fields := GetMultiFields(r)
	if fields.Terms == nil {
		return nil
	}
	return fields.Terms(field)
}
