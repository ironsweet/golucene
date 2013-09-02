package store

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/codec"
	"github.com/balzaczyy/golucene/util"
	"log"
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

	COMPOUND_FILE_EXTENSION         = "cfs"
	COMPOUND_FILE_ENTRIES_EXTENSION = "cfe"
)

type CompoundFileDirectory struct {
	*DirectoryImpl
	lock sync.Mutex

	directory      Directory
	fileName       string
	readBufferSize int
	entries        map[string]FileEntry
	openForWriter  bool
	// writer CompoundFileWriter
	handle IndexInputSlicer
}

func NewCompoundFileDirectory(directory Directory, fileName string, context IOContext, openForWrite bool) (d *CompoundFileDirectory, err error) {
	self := &CompoundFileDirectory{
		lock:           sync.Mutex{},
		directory:      directory,
		fileName:       fileName,
		readBufferSize: bufferSize(context),
		openForWriter:  openForWrite}
	self.DirectoryImpl = newDirectoryImpl(self)

	if !openForWrite {
		log.Printf("Open for read.")
		success := false
		defer func() {
			if !success {
				util.CloseWhileSuppressingError(self.handle)
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
		self.DirectoryImpl.isOpen = true
		return self, err
	} else {
		panic("not supported yet")
	}
}

func (d *CompoundFileDirectory) Close() error {
	log.Printf("Closing %v...", d)
	if d == nil { // interface not nil
		return nil
	}
	d.lock.Lock()
	defer d.lock.Unlock()

	if d == nil || !d.isOpen {
		log.Print("CompoundFileDirectory is already closed.")
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

func (d *CompoundFileDirectory) OpenInput(name string, context IOContext) (in IndexInput, err error) {
	d.ensureOpen()
	// assert !d.openForWrite
	id := util.StripSegmentName(name)
	if entry, ok := d.entries[id]; ok {
		is := d.handle.openSlice(name, entry.offset, entry.length)
		return is, nil
	}
	keys := make([]string, 0)
	for k := range d.entries {
		keys = append(keys, k)
	}
	panic(fmt.Sprintf("No sub-file with id %v found (fileName=%v files: %v)", id, name, keys))
}

func (d *CompoundFileDirectory) ListAll() (paths []string, err error) {
	d.ensureOpen()
	// if self.writer != nil {
	// 	return self.writer.ListAll()
	// }
	// Add the segment name
	seg := util.ParseSegmentName(d.fileName)
	keys := make([]string, 0, len(d.entries))
	for k := range d.entries {
		keys = append(keys, seg+k)
	}
	return keys, nil
}

func (d *CompoundFileDirectory) FileExists(name string) bool {
	d.ensureOpen()
	// if d.writer != nil {
	// 	return d.writer.FileExists(name)
	// }
	_, ok := d.entries[util.StripSegmentName(name)]
	return ok
}

const (
	CODEC_MAGIC_BYTE1 = byte(uint32(codec.CODEC_MAGIC) >> 24 & 0xFF)
	CODEC_MAGIC_BYTE2 = byte(uint32(codec.CODEC_MAGIC) >> 16 & 0xFF)
	CODEC_MAGIC_BYTE3 = byte(uint32(codec.CODEC_MAGIC) >> 8 & 0xFF)
	CODEC_MAGIC_BYTE4 = byte(codec.CODEC_MAGIC & 0xFF)
)

func readEntries(handle IndexInputSlicer, dir Directory, name string) (mapping map[string]FileEntry, err error) {
	var stream, entriesStream IndexInput = nil, nil
	defer func() {
		err = util.CloseWhileHandlingError(err, stream, entriesStream)
	}()
	// read the first VInt. If it is negative, it's the version number
	// otherwise it's the count (pre-3.1 indexes)
	mapping = make(map[string]FileEntry)
	stream = handle.openFullSlice()
	log.Printf("Reading from stream: %v", stream)
	firstInt, err := stream.ReadVInt()
	if err != nil {
		return mapping, err
	}
	// impossible for 3.0 to have 63 files in a .cfs, CFS writer was not visible
	// and separate norms/etc are outside of cfs.
	if firstInt == int32(CODEC_MAGIC_BYTE1) {
		if secondByte, err := stream.ReadByte(); err == nil {
			if thirdByte, err := stream.ReadByte(); err == nil {
				if fourthByte, err := stream.ReadByte(); err == nil {
					if secondByte != CODEC_MAGIC_BYTE2 ||
						thirdByte != CODEC_MAGIC_BYTE3 ||
						fourthByte != CODEC_MAGIC_BYTE4 {
						return mapping, errors.New(fmt.Sprintf(
							"Illegal/impossible header for CFS file: %v,%v,%v",
							secondByte, thirdByte, fourthByte))
					}
				}
			}
		}
		if err != nil {
			return mapping, err
		}

		_, err = codec.CheckHeaderNoMagic(stream, CFD_DATA_CODEC, CFD_VERSION_START, CFD_VERSION_START)
		if err != nil {
			return mapping, err
		}
		entriesFileName := util.SegmentFileName(util.StripExtension(name), "", COMPOUND_FILE_ENTRIES_EXTENSION)
		entriesStream, err = dir.OpenInput(entriesFileName, IO_CONTEXT_READONCE)
		if err != nil {
			return mapping, err
		}
		// log.Printf("DEBUG: %v", entriesStream)
		_, err = codec.CheckHeader(entriesStream, CFD_ENTRY_CODEC, CFD_VERSION_START, CFD_VERSION_START)
		if err != nil {
			return mapping, err
		}
		numEntries, err := entriesStream.ReadVInt()
		if err != nil {
			return mapping, err
		}
		log.Printf("Entries number: %v", numEntries)
		for i := int32(0); i < numEntries; i++ {
			id, err := entriesStream.ReadString()
			if err != nil {
				return mapping, err
			}
			if _, ok := mapping[id]; ok {
				return mapping, errors.New(fmt.Sprintf(
					"Duplicate cfs entry id=%v in CFS: %v", id, entriesStream))
			}
			log.Printf("Found entry: %v", id)
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
