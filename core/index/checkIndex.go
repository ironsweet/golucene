package index

import (
	"github.com/balzaczyy/golucene/core/store"
	"io"
)

// index/CheckIndex.java

// Returned from checkIndex() detailing the health and status of the index
type CheckIndexStatus struct {
	// True if no problems found with the index.
	Clean bool
}

/*
Basic tool and API to check the health of an index and write a new
segments file that removes reference to problematic segments.

As this tool checks every byte in the index, on a large index it can
take a long time to run.
*/
type CheckIndex struct {
	infoStream            io.Writer
	dir                   store.Directory
	crossCheckTermVectors bool
}

func NewCheckIndex(dir store.Directory, crossCheckTermVectors bool, infoStream io.Writer) *CheckIndex {
	return &CheckIndex{
		dir: dir,
		crossCheckTermVectors: crossCheckTermVectors,
		infoStream:            infoStream,
	}
}

/*
Returns a Status instance detailing the state of the index.

As this method checks every byte in the specified segments, on a
large index it can take quite a long time to run.

WARNING: make sure you only call this when the index is not opened
by any writer.
*/
func (ch *CheckIndex) CheckIndex(onlySegments []string) (*CheckIndexStatus, error) {
	panic("not implemented yet")
}
