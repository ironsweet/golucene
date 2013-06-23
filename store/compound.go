package store

import (
	"fmt"
	"github.com/balzaczyy/golucene/util"
	"sync"
)

type FileEntry struct {
	offset, length int64
}

const (
	CFD_DATA_CODEC      = "CompoundFileWriterData"
	CFD_VERSION_START   = 0
	CFD_VERSION_CURRENT = CFD_VERSION_START

	CFD_ENTRY_CODEC = "CompoundFileWriterEntries"

	COMPOUND_FILE_ENTRIES_EXTENSION = "cfe"
)

type CompoundFileDirectory struct {
	*Directory
	lock sync.Mutex

	directory      *Directory
	fileName       string
	readBufferSize int
	entries        map[string]FileEntry
	openForWriter  bool
	// writer CompoundFileWriter
	handle IndexInputSlicer
}

func NewCompoundFileDirectory(directory *Directory, fileName string, context IOContext, openForWrite bool) (d CompoundFileDirectory, err error) {
	self := CompoundFileDirectory{
		directory:      directory,
		fileName:       fileName,
		readBufferSize: bufferSize(context),
		openForWriter:  openForWrite}
	super := &Directory{isOpen: false}
	super.OpenInput = func(name string, context IOContext) (in *IndexInput, err error) {
		super.ensureOpen()
		// assert !self.openForWrite
		id := util.StripSegmentName(name)
		if entry, ok := self.entries[id]; ok {
			is := self.handle.openSlice(name, entry.offset, entry.length)
			return &is, nil
		}
		keys := make([]string, 0)
		for k := range self.entries {
			keys = append(keys, k)
		}
		panic(fmt.Sprintf("No sub-file with id %v found (fileName=%v files: %v)", id, name, keys))
	}
	super.ListAll = func() (paths []string, err error) {
		super.ensureOpen()
		// if self.writer != nil {
		// 	return self.writer.ListAll()
		// }
		// Add the segment name
		seg := util.ParseSegmentName(self.fileName)
		keys := make([]string, 0)
		for k := range self.entries {
			keys = append(keys, seg+k)
		}
		return keys, nil
	}
	super.FileExists = func(name string) bool {
		super.ensureOpen()
		// if self.writer != nil {
		// 	return self.writer.FileExists(name)
		// }
		if _, ok := self.entries[util.StripSegmentName(name)]; ok {
			return true
		}
		return false
	}

	if !openForWrite {
		success := false
		defer func() {
			if !success {
				util.CloseWhileSupressingError(self.handle)
			}
		}()
		self.handle, err = directory.createSlicer(fileName, context)
		if err != nil {
			return self, err
		}
		self.entries, err = readEntries(self.handle, directory, fileName)
		if err != nil {
			return self, err
		}
		success = true
		super.isOpen = true
		return self, err
	} else {
		panic("not supported yet")
	}
}

const (
	CODEC_MAGIC_BYTE1 = byte(uint32(CODEC_MAGIC) >> 24 & 0xFF)
	CODEC_MAGIC_BYTE2 = byte(uint32(CODEC_MAGIC) >> 16 & 0xFF)
	CODEC_MAGIC_BYTE3 = byte(uint32(CODEC_MAGIC) >> 8 & 0xFF)
	CODEC_MAGIC_BYTE4 = byte(CODEC_MAGIC & 0xFF)
)

func readEntries(handle IndexInputSlicer, dir *Directory, name string) (mapping map[string]FileEntry, err error) {
	var stream, entriesStream *IndexInput = nil, nil
	defer func(err *error) {
		*err = util.CloseWhileHandlingError(*err, stream, entriesStream)
	}(&err)
	// read the first VInt. If it is negative, it's the version number
	// otherwise it's the count (pre-3.1 indexes)
	mapping = make(map[string]FileEntry)
	slice := handle.openFullSlice()
	stream = &(slice)
	firstInt, err := stream.ReadVInt()
	if err != nil {
		return mapping, err
	}
	// impossible for 3.0 to have 63 files in a .cfs, CFS writer was not visible
	// and separate norms/etc are outside of cfs.
	if firstInt == int(CODEC_MAGIC_BYTE1) {
		secondByte, err := stream.ReadByte()
		if err != nil {
			return mapping, err
		}
		thirdByte, err := stream.ReadByte()
		if err != nil {
			return mapping, err
		}
		fourthByte, err := stream.ReadByte()
		if err != nil {
			return mapping, err
		}
		if secondByte != CODEC_MAGIC_BYTE2 ||
			thirdByte != CODEC_MAGIC_BYTE3 ||
			fourthByte != CODEC_MAGIC_BYTE4 {
			return mapping, &CorruptIndexError{fmt.Sprintf("Illegal/impossible header for CFS file: %v,%v,%v", secondByte, thirdByte, fourthByte)}
		}
		_, err = CheckHeaderNoMagic(stream.DataInput, CFD_DATA_CODEC, CFD_VERSION_START, CFD_VERSION_START)
		if err != nil {
			return mapping, err
		}
		entriesFileName := util.SegmentFileName(util.StripExtension(name), "", COMPOUND_FILE_ENTRIES_EXTENSION)
		entriesStream, err = dir.OpenInput(entriesFileName, IO_CONTEXT_READONCE)
		if err != nil {
			return mapping, err
		}
		_, err = CheckHeader(entriesStream.DataInput, CFD_ENTRY_CODEC, CFD_VERSION_START, CFD_VERSION_START)
		if err != nil {
			return mapping, err
		}
		numEntries, err := entriesStream.ReadVInt()
		if err != nil {
			return mapping, err
		}
		for i := 0; i < numEntries; i++ {
			id, err := entriesStream.ReadString()
			if err != nil {
				return mapping, err
			}
			if _, ok := mapping[id]; ok {
				return mapping, &CorruptIndexError{fmt.Sprintf("Duplicate cfs entry id=%v in CFS: %v", id, entriesStream)}
			}
			offset, err := entriesStream.ReadLong()
			if err != nil {
				return mapping, err
			}
			length, err := entriesStream.ReadLong()
			if err != nil {
				return mapping, err
			}
			mapping[id] = FileEntry{offset, length}
		}
	} else {
		// TODO remove once 3.x is not supported anymore
		panic("not supported yet; will also be obsolete soon")
	}
	return mapping, nil
}

func (d *CompoundFileDirectory) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.isOpen {
		// allow double close - usually to be consistent with other closeables
		return nil // already closed
	}
	d.isOpen = false
	/*
		if d.writer != nil {
			// assert d.openForWrite
			return writer.Close()
		} else {*/
	return util.Close(d.handle)
	// }
}
