package store

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/util"
	"sort"
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
		return true
	}
	return false
}

type FileEntry struct {
	file           string    // source file
	length, offset int64     // temporary holder for the start of this file's data section
	dir            Directory // which contains the file.
}

// Combines multiple files into a single compound file
type CompoundFileWriter struct {
	sync.Locker
	directory Directory
	entries   map[string]*FileEntry
	seenIDs   map[string]bool
	// all entries that are written to a sep. file but not yet moved into CFS
	pendingEntries *list.List
	closed         bool
	dataOut        IndexOutput
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
		Locker:         &sync.Mutex{},
		directory:      dir,
		entries:        make(map[string]*FileEntry),
		seenIDs:        make(map[string]bool),
		pendingEntries: list.New(),
		outputTaken:    NewAtomicBool(),
		entryTableName: util.SegmentFileName(
			util.StripExtension(name),
			"",
			COMPOUND_FILE_ENTRIES_EXTENSION,
		),
		dataFileName: name,
	}
}

func (w *CompoundFileWriter) output(ctx IOContext) (IndexOutput, error) {
	w.Lock()
	defer w.Unlock()
	if w.dataOut == nil {
		var success = false
		defer func() {
			if !success {
				util.CloseWhileSuppressingError(w.dataOut)
			}
		}()

		var err error
		w.dataOut, err = w.directory.CreateOutput(w.dataFileName, ctx)
		if err != nil {
			return nil, err
		}
		err = codec.WriteHeader(w.dataOut, CFD_DATA_CODEC, CFD_VERSION_CURRENT)
		if err != nil {
			return nil, err
		}
		success = true
	}
	return w.dataOut, nil
}

/* Closes all resouces and writes the entry table */
func (w *CompoundFileWriter) Close() (err error) {
	if w.closed {
		fmt.Println("CompoundFileWriter is already closed.")
		return nil
	}

	// TODO this code should clean up after itself (remove partial .cfs/.cfe)
	if err = func() (err error) {
		var success = false
		defer func() {
			if success {
				util.Close(w.dataOut)
			} else {
				util.CloseWhileSuppressingError(w.dataOut)
			}
		}()

		assert2(w.pendingEntries.Len() == 0 && !w.outputTaken.Get(),
			"CFS has pending open files")
		w.closed = true
		// open the compound stream; we can safely use IO_CONTEXT_DEFAULT
		// here because this will only open the output if no file was
		// added to the CFS
		_, err = w.output(IO_CONTEXT_DEFAULT)
		if err != nil {
			return
		}
		assert(w.dataOut != nil)
		err = codec.WriteFooter(w.dataOut)
		if err != nil {
			return
		}
		success = true
		return nil
	}(); err != nil {
		return
	}

	var entryTableOut IndexOutput
	var success = false
	defer func() {
		if success {
			util.Close(entryTableOut)
		} else {
			util.CloseWhileSuppressingError(entryTableOut)
		}
	}()
	entryTableOut, err = w.directory.CreateOutput(w.entryTableName, IO_CONTEXT_DEFAULT)
	if err != nil {
		return
	}
	err = w.writeEntryTable(w.entries, entryTableOut)
	if err != nil {
		return
	}
	success = true
	return
}

func (w *CompoundFileWriter) ensureOpen() {
	assert2(!w.closed, "CFS Directory is already closed")
}

/* Copy the contents of the file with specified extension into the provided output stream. */
func (w *CompoundFileWriter) copyFileEntry(dataOut IndexOutput, fileEntry *FileEntry) (n int64, err error) {
	var is IndexInput
	is, err = fileEntry.dir.OpenInput(fileEntry.file, IO_CONTEXT_READONCE)
	if err != nil {
		return 0, err
	}
	var success = false
	defer func() {
		if success {
			err = util.Close(is)
			// copy successful - delete file
			if err == nil {
				fileEntry.dir.DeleteFile(fileEntry.file) // ignore error
			}
		} else {
			util.CloseWhileSuppressingError(is)
		}
	}()

	startPtr := dataOut.FilePointer()
	length := fileEntry.length
	err = dataOut.CopyBytes(is, length)
	if err != nil {
		return 0, err
	}
	// verify that the output length diff is equal to original file
	endPtr := dataOut.FilePointer()
	diff := endPtr - startPtr
	if diff != length {
		return 0, errors.New(fmt.Sprintf(
			"Difference in the output file offsets %v does not match the original file length %v",
			diff, length))
	}
	fileEntry.offset = startPtr
	success = true
	return length, nil
}

func (w *CompoundFileWriter) writeEntryTable(entries map[string]*FileEntry,
	entryOut IndexOutput) (err error) {
	if err = codec.WriteHeader(entryOut, CFD_ENTRY_CODEC, CFD_VERSION_CURRENT); err == nil {
		if err = entryOut.WriteVInt(int32(len(entries))); err == nil {
			var names []string
			for name, _ := range entries {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				// for _, fe := range entries {
				fe := entries[name]
				if err = Stream(entryOut).
					WriteString(util.StripSegmentName(fe.file)).
					WriteLong(fe.offset).
					WriteLong(fe.length).
					Close(); err != nil {
					break
				}
			}
		}
	}
	if err == nil {
		err = codec.WriteFooter(entryOut)
	}
	return err
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
	entry := &FileEntry{}
	entry.file = name
	w.entries[name] = entry
	id := util.StripSegmentName(name)
	_, ok = w.seenIDs[id]
	assert2(!ok, "file='%v' maps to id='%v', which was already written", name, id)
	w.seenIDs[id] = true

	var out *DirectCFSIndexOutput
	if outputLocked := w.outputTaken.CompareAndSet(false, true); outputLocked {
		o, err := w.output(context)
		if err != nil {
			return nil, err
		}
		out = newDirectCFSIndexOutput(w, o, entry, false)
	} else {
		entry.dir = w.directory
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

func (w *CompoundFileWriter) prunePendingEntries() error {
	// claim the output and copy all pending files in
	if w.outputTaken.CompareAndSet(false, true) {
		defer func() {
			cas := w.outputTaken.CompareAndSet(true, false)
			assert(cas)
		}()
		for w.pendingEntries.Len() > 0 {
			head := w.pendingEntries.Front()
			w.pendingEntries.Remove(head)
			entry := head.Value.(*FileEntry)
			out, err := w.output(NewIOContextForFlush(&FlushInfo{0, entry.length}))
			if err == nil {
				_, err = w.copyFileEntry(out, entry)
			}
			if err != nil {
				return err
			}
			w.entries[entry.file] = entry
		}
	}
	return nil
}

type DirectCFSIndexOutput struct {
	*IndexOutputImpl
	owner        *CompoundFileWriter
	delegate     IndexOutput
	offset       int64
	closed       bool
	entry        *FileEntry
	writtenBytes int64
	isSeparate   bool
}

func newDirectCFSIndexOutput(owner *CompoundFileWriter,
	delegate IndexOutput, entry *FileEntry, isSeparate bool) *DirectCFSIndexOutput {
	ans := &DirectCFSIndexOutput{
		owner:      owner,
		delegate:   delegate,
		entry:      entry,
		offset:     delegate.FilePointer(),
		isSeparate: isSeparate,
	}
	ans.entry.offset = ans.offset
	ans.IndexOutputImpl = NewIndexOutput(ans)
	return ans
}

func (out *DirectCFSIndexOutput) Flush() error {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) Close() error {
	if out.closed {
		return nil
	}
	out.closed = true
	out.entry.length = out.writtenBytes
	if out.isSeparate {
		err := out.delegate.Close()
		if err != nil {
			return err
		}
		// we are a separate file - push into the pending entries
		out.owner.pendingEntries.PushBack(out.entry)
	} else {
		// we have been written into the CFS directly - release the lock
		out.owner.releaseOutputLock()
	}
	// now prune all pending entries and push them into the CFS
	return out.owner.prunePendingEntries()
}

func (out *DirectCFSIndexOutput) FilePointer() int64 {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) WriteByte(b byte) error {
	panic("not implemented yet")
}

func (out *DirectCFSIndexOutput) WriteBytes(b []byte) error {
	assert(!out.closed)
	out.writtenBytes += int64(len(b))
	return out.delegate.WriteBytes(b)
}

func (out *DirectCFSIndexOutput) Checksum() int64 {
	return out.delegate.Checksum()
}
