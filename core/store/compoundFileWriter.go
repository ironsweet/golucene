package store

import (
	"github.com/balzaczyy/golucene/core/util"
)

// store/CompoundFileWriter.java

// Combines multiple files into a single compound file
type CompoundFileWriter struct {
	directory      Directory
	entryTableName string
	dataFileName   string
}

/*
Create the compound stream in the specified file. The filename is the
entire name (no extensions are added).
*/
func newCompoundFileWriter(dir Directory, name string) *CompoundFileWriter {
	assert2(dir != nil, "directory cannot be nil")
	assert2(name != "", "name cannot be empty")
	return &CompoundFileWriter{
		directory: dir,
		entryTableName: util.SegmentFileName(
			util.StripExtension(name),
			"",
			COMPOUND_FILE_ENTRIES_EXTENSION,
		),
		dataFileName: name,
	}
}

/* Closes all resouces and writes the entry table */
func (w *CompoundFileWriter) Close() error {
	panic("not implemented yet")
}
