package store

import (
	"github.com/balzaczyy/golucene/core/util"
	"sync"
)

// store/CompoundFileWriter.java

type AtomicBool struct {
	*sync.RWMutex
	v bool
}

func NewAtomicBool() *AtomicBool {
	return &AtomicBool{&sync.RWMutex{}, false}
}

func (b *AtomicBool) Get() bool {
	b.RLock()
	defer b.RUnlock()
	return b.v
}

func (b *AtomicBool) CompareAndSet(from, to bool) bool {
	b.Lock()
	defer b.Unlock()
	if b.v == from {
		b.v = to
	}
	return b.v
}

type FileEntry struct {
	file           string    // source file
	length, offset int64     // temporary holder for the start of this file's data section
	dir            Directory // which contains the file.
}

// Combines multiple files into a single compound file
type CompoundFileWriter struct {
	directory      Directory
	entries        map[string]FileEntry
	seenIDs        map[string]bool
	closed         bool
	outputTaken    *AtomicBool
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
		directory:   dir,
		entries:     make(map[string]FileEntry),
		seenIDs:     make(map[string]bool),
		outputTaken: NewAtomicBool(),
		entryTableName: util.SegmentFileName(
			util.StripExtension(name),
			"",
			COMPOUND_FILE_ENTRIES_EXTENSION,
		),
		dataFileName: name,
	}
}

func (w *CompoundFileWriter) output() (IndexOutput, error) {
	panic("not implemented yet")
}

/* Closes all resouces and writes the entry table */
func (w *CompoundFileWriter) Close() error {
	panic("not implemented yet")
}

func (w *CompoundFileWriter) ensureOpen() {
	assert2(!w.closed, "CFS Directory is already closed")
}

func (w *CompoundFileWriter) createOutput(name string, context IOContext) (IndexOutput, error) {
	w.ensureOpen()
	var success = false
	var outputLocked = false
	defer func() {
		if !success {
			delete(w.entries, name)
			if outputLocked { // release the output lock if not successful
				assert(w.outputTaken.Get())
				w.releaseOutputLock()
			}
		}
	}()

	assert2(name != "", "name must not be empty")
	_, ok := w.entries[name]
	assert2(!ok, "File %v already exists", name)
	entry := FileEntry{}
	entry.file = name
	w.entries[name] = entry
	id := util.StripSegmentName(name)
	_, ok = w.seenIDs[id]
	assert2(!ok, "file='%v' maps to id='%v', which was already written", name, id)
	w.seenIDs[id] = true

	var out *DirectCFSIndexOutput
	if outputLocked := w.outputTaken.CompareAndSet(false, true); outputLocked {
		o, err := w.output()
		if err != nil {
			return nil, err
		}
		out = newDirectCFSIndexOutput(w, o, entry, false)
	} else {
		entry.dir = w.directory
		assert2(!w.directory.FileExists(name), "File %v already exists", name)
		o, err := w.directory.CreateOutput(name, context)
		if err != nil {
			return nil, err
		}
		out = newDirectCFSIndexOutput(w, o, entry, true)
	}
	success = true
	return out, nil
}

func (w *CompoundFileWriter) releaseOutputLock() {
	w.outputTaken.CompareAndSet(true, false)
}

type DirectCFSIndexOutput struct {
	*IndexOutputImpl
}

func newDirectCFSIndexOutput(owner *CompoundFileWriter,
	delegate IndexOutput, entry FileEntry, isSeparate bool) *DirectCFSIndexOutput {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) Flush() error {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) Close() error {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) FilePointer() int64 {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) Length() (int64, error) {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) WriteByte(b byte) error {
	panic("not implemented yet")
}
func (out *DirectCFSIndexOutput) WriteBytes(b []byte) error {
	panic("not implemented yet")
}
