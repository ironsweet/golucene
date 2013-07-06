package store

import (
	"errors"
	"fmt"
)

func NewIndexFormatTooNewError(in *DataInput, version, minVersion, maxVersion int32) error {
	return errors.New(fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v)",
		in, version, minVersion, maxVersion))
}

func NewIndexFormatTooOldError(in *DataInput, version, minVersion, maxVersion int32) error {
	return errors.New(fmt.Sprintf(
		"Format version is not supported (resource: %v): %v (needs to be between %v and %v). This version of Lucene only supports indexes created with release 3.0 and later.",
		in, version, minVersion, maxVersion))
}
