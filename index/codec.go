package index

import (
	"fmt"
	"io"
	"lucene/store"
)

const (
	CODEC_MAGIC = 0x3fd76c17
)

func CheckHeader(in *store.DataInput, codec string, minVersion, maxVersion int) (v int, err error) {
	// Safety to guard against reading a bogus string:
	actualHeader, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	if actualHeader != CODEC_MAGIC {
		return 0, &CorruptIndexError{fmt.Sprintf(
			"codec header mismatch: actual header=%v vs expected header=%v (resource: %v)",
			actualHeader, CODEC_MAGIC, in)}
	}
	return CheckHeaderNoMagic(in, codec, minVersion, maxVersion)
}

func CheckHeaderNoMagic(in *store.DataInput, codec string, minVersion, maxVersion int) (v int, err error) {
	actualCodec, err := in.ReadString()
	if err != nil {
		return 0, err
	}
	if actualCodec != codec {
		return 0, &CorruptIndexError{fmt.Sprintf(
			"codec mismatch: actual codec=%v vs expected codec=%v (resource: %v)", actualCodec, codec, in)}
	}

	actualVersion, err := in.ReadInt()
	if err != nil {
		return 0, err
	}
	if actualVersion < minVersion {
		return 0, NewIndexFormatTooOldError(in, actualVersion, minVersion, maxVersion)
	}
	if actualVersion > maxVersion {
		return 0, NewIndexFormatTooNewError(in, actualVersion, minVersion, maxVersion)
	}

	return actualVersion, nil
}

type FieldsProducer interface {
	Fields
	io.Closer
}

type DocValuesProducer interface {
	io.Closer
	Numeric(field FieldInfo) (v NumericDocValues, err error)
	Binary(field FieldInfo) (v BinaryDocValues, err error)
	Sorted(field FieldInfo) (v SortedDocValues, err error)
	SortedSet(field FieldInfo) (v SortedSetDocValues, err error)
}

type NumericDocValues interface {
	value(docID int) int64
}
type BinaryDocValues interface {
	value(docID int, result []byte)
}
type SortedDocValues interface {
	BinaryDocValues
	ord(docID int) int
	lookupOrd(ord int, result []byte)
	valueCount() int
}
type SortedSetDocValues interface {
	nextOrd() int64
	setDocument(docID int)
	lookupOrd(ord int64, result []byte)
	valueCount() int64
}

type StoredFieldsReader interface {
	io.Closer
	visitDocument(n int, visitor StoredFieldVisitor) error
	clone() StoredFieldsReader
}

type StoredFieldVisitor interface {
	binaryField(fi FieldInfo, value []byte) error
	stringField(fi FieldInfo, value string) error
	intField(fi FieldInfo, value int) error
	longField(fi FieldInfo, value int64) error
	floatField(fi FieldInfo, value float32) error
	doubleField(fi FieldInfo, value float64) error
	needsField(fi FieldInfo) StoredFieldVisitorStatus
}

type StoredFieldVisitorStatus int

const (
	SOTRED_FIELD_VISITOR_STATUS_YES  = StoredFieldVisitorStatus(1)
	SOTRED_FIELD_VISITOR_STATUS_NO   = StoredFieldVisitorStatus(2)
	SOTRED_FIELD_VISITOR_STATUS_STOP = StoredFieldVisitorStatus(3)
)

type TermVectorsReader interface {
	io.Closer
	fields(doc int) Fields
	clone() TermVectorsReader
}
