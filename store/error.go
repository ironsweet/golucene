package store

import (
	"fmt"
)

type CorruptIndexError struct {
	Msg string
}

func (err *CorruptIndexError) Error() string {
	return err.Msg
}

func NewIndexFormatTooNewError(in *DataInput, version, minVersion, maxVersion int) *CorruptIndexError {
	return &CorruptIndexError{fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v)",
		in, version, minVersion, maxVersion)}
}

func NewIndexFormatTooOldError(in *DataInput, version, minVersion, maxVersion int) *CorruptIndexError {
	return &CorruptIndexError{fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v). This version of Lucene only supports indexes created with release 3.0 and later.",
		in, version, minVersion, maxVersion)}
}
