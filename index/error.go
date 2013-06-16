package index

import (
	"fmt"
	"lucene/store"
)

type CorruptIndexError struct {
	msg string
}

func (err *CorruptIndexError) Error() string {
	return err.msg
}

func NewIndexFormatTooNewError(in *store.DataInput, version, minVersion, maxVersion int) *CorruptIndexError {
	return &CorruptIndexError{fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v)",
		in, version, minVersion, maxVersion)}
}

func NewIndexFormatTooOldError(in *store.DataInput, version, minVersion, maxVersion int) *CorruptIndexError {
	return &CorruptIndexError{fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v). This version of Lucene only supports indexes created with release 3.0 and later.",
		in, version, minVersion, maxVersion)}
}
