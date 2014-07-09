package model

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/util"
	"sort"
)

// TermsEnum.java
/*
Iterator to seek, or step through terms to obtain frequency information, or
for the current term.

Term enumerations are always ordered by specified Comparator. Each term in the
enumeration is greater than the one before it.

The TermsEnum is unpositioned when you first obtain it and you must first
succesfully call next() or one of the seek methods.
*/
type TermsEnum interface {
	util.BytesRefIterator

	Attributes() *util.AttributeSource
	/* Attempts to seek to the exact term, returning
	true if the term is found. If this returns false, the
	enum is unpositioned. For some codecs, seekExact may
	be substantially faster than seekCeil. */
	SeekExact(text []byte) (ok bool, err error)
	/* Seeks to the specified term, if it exists, or to the
	next (ceiling) term. Returns SeekStatus to
	indicate whether exact term was found, a different
	term was found, or EOF was hit. The target term may
	be before or after the current term. If this returns
	SeekStatus.END, then enum is unpositioned. */
	SeekCeil(text []byte) SeekStatus
	/* Seeks to the specified term by ordinal (position) as
	previously returned by ord. The target ord
	may be before or after the current ord, and must be
	within bounds. */
	SeekExactByPosition(ord int64) error
	/* Expert: Seeks a specific position by TermState previously obtained
	from termState(). Callers shoudl maintain the TermState to
	use this method. Low-level implementations may position the TermsEnum
	without re-seeking the term dictionary.

	Seeking by TermState should only be used iff the state was obtained
	from the same instance.

	NOTE: Using this method with an incompatible TermState might leave
	this TermsEnum in undefiend state. On a segment level
	TermState instances are compatible only iff the source and the
	target TermsEnum operate on the same field. If operating on segment
	level, TermState instances must not be used across segments.

	NOTE: A seek by TermState might not restore the
	AttributeSource's state. AttributeSource state must be
	maintained separately if the method is used. */
	SeekExactFromLast(text []byte, state TermState) error
	/* Returns current term. Do not call this when enum
	is unpositioned. */
	Term() []byte
	/* Returns ordinal position for current term. This is an
	optional method (the codec may panic). Do not call this
	when the enum is unpositioned. */
	Ord() int64
	/* Returns the number of documentsw containing the current
	term. Do not call this when enum is unpositioned. */
	DocFreq() (df int, err error)
	/* Returns the total numnber of occurrences of this term
	across all documents (the sum of the freq() for each
	doc that has this term). This will be -1 if the
	codec doesn't support this measure. Note that, like
	other term measures, this measure does not take
	deleted documents into account. */
	TotalTermFreq() (tf int64, err error)
	/* Get DocsEnum for the current term. Do not
	call this when the enum is unpositioned. This method
	will not return nil. */
	Docs(liveDocs util.Bits, reuse DocsEnum) (de DocsEnum, err error)
	/* Get DocsEnum for the current term, with
	control over whether freqs are required. Do not
	call this when the enum is unpositioned. This method
	will not return nil. */
	DocsByFlags(liveDocs util.Bits, reuse DocsEnum, flags int) (de DocsEnum, err error)
	/* Get DocsAndPositionEnum for the current term.
	Do not call this when the enum is unpositioned. This
	method will return nil if positions were not
	indexed. */
	DocsAndPositions(liveDocs util.Bits, reuse DocsAndPositionsEnum) DocsAndPositionsEnum
	/* Get DocsAndPositionEnum for the current term,
	with control over whether offsets and payloads are
	required. Some codecs may be able to optimize their
	implementation when offsets and/or payloads are not required.
	Do not call this when the enum is unpositioned. This
	will return nil if positions were not indexed. */
	DocsAndPositionsByFlags(liveDocs util.Bits, reuse DocsAndPositionsEnum, flags int) DocsAndPositionsEnum
	/* Expert: Returns the TermsEnum internal state to position the TermsEnum
	without re-seeking the term dictionary.

	NOTE: A sek by TermState might not capture the
	AttributeSource's state. Callers must maintain the
	AttributeSource states separately. */
	TermState() (ts TermState, err error)
}

type SeekStatus int

const (
	SEEK_STATUS_END       = 1
	SEEK_STATUS_FOUND     = 2
	SEEK_STATUS_NOT_FOUND = 3
)

type TermsEnumImpl struct {
	TermsEnum
	atts *util.AttributeSource
}

func NewTermsEnumImpl(self TermsEnum) *TermsEnumImpl {
	return &TermsEnumImpl{self, util.NewAttributeSourceWith(tokenattributes.DEFAULT_ATTRIBUTE_FACTORY)}
}

func (e *TermsEnumImpl) Attributes() *util.AttributeSource {
	return e.atts
}

func (e *TermsEnumImpl) SeekExact(text []byte) (ok bool, err error) {
	return e.SeekCeil(text) == SEEK_STATUS_FOUND, nil
}

func (e *TermsEnumImpl) SeekExactFromLast(text []byte, state TermState) error {
	ok, err := e.SeekExact(text)
	if err != nil {
		return err
	}
	if !ok {
		panic(fmt.Sprintf("term %v does not exist", text))
	}
	return nil
}

func (e *TermsEnumImpl) Docs(liveDocs util.Bits, reuse DocsEnum) (DocsEnum, error) {
	assert(e != nil)
	return e.DocsByFlags(liveDocs, reuse, DOCS_ENUM_FLAG_FREQS)
}

func (e *TermsEnumImpl) DocsAndPositions(liveDocs util.Bits, reuse DocsAndPositionsEnum) DocsAndPositionsEnum {
	return e.DocsAndPositionsByFlags(liveDocs, reuse, DOCS_POSITIONS_ENUM_FLAG_OFF_SETS|DOCS_POSITIONS_ENUM_FLAG_PAYLOADS)
}

func (e *TermsEnumImpl) TermState() (ts TermState, err error) {
	return EMPTY_TERM_STATE, nil
}

var EMPTY_TERM_STATE = &EmptyTermState{}

type EmptyTermState struct{}

func (ts *EmptyTermState) CopyFrom(other TermState) {
	panic("not supported!")
}

func (ts *EmptyTermState) Clone() TermState {
	return ts
}

var (
	EMPTY_TERMS_ENUM = &EmptyTermsEnum{}
)

/* An empty TermsEnum for quickly returning an empty instance e.g.
in MultiTermQuery
Please note: This enum should be unmodifiable,
but it is currently possible to add Attributes to it.
This should not be a problem, as the enum is always empty and
the existence of unused Attributes does not matter. */
type EmptyTermsEnum struct {
	*TermsEnumImpl
}

func (e *EmptyTermsEnum) SeekCeilUsingCache(term []byte, useCache bool) SeekStatus {
	return SEEK_STATUS_END
}

func (e *EmptyTermsEnum) SeekExactByPosition(ord int64) error {
	return nil
}

func (e *EmptyTermsEnum) Term() []byte {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) Comparator() sort.Interface {
	return nil
}

func (e *EmptyTermsEnum) DocFreq() (df int, err error) {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) TotalTermFreq() (tf int64, err error) {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) Ord() int64 {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) DocsByFlags(liveDocs util.Bits, reuse DocsEnum, flags int) (de DocsEnum, err error) {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) DocsAndPositionsByFlags(liveDocs util.Bits, reuse DocsAndPositionsEnum, flags int) DocsAndPositionsEnum {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) Next() (term []byte, err error) {
	return nil, nil
}

func (e *EmptyTermsEnum) TermState() (ts TermState, err error) {
	panic("this method should never be called")
}

func (e *EmptyTermsEnum) SeekExactFromLast(term []byte, state TermState) error {
	panic("this method should never be called")
}
