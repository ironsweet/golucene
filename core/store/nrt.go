package store

import (
	"fmt"
	"log"
	"sync"
)

const NRT_VERBOSE = false

// TODO
// - let subclass dictate policy...?
// - rename to MergeCachingDir? NRTCachingDIR

/*
Wraps a RAMDirectory around any provided delegate directory, to be
used during NRT search.

This class is likely only useful in a near-real-time context, where
indexing rate is lowish but reopen rate is highish, resulting in many
tiny files being written. This directory keeps such segments (as well
as the segments produced by merging them, as long as they are small
enough), in RAM.

This is safe to use: when you app calls IndexWriter.Commit(), all
cached files will be flushed from the cached and sync'd.

Here's a simple example usage:

	fsDir, _ := OpenFSDirectory("/path/to/index")
	cachedFSDir := NewNRTCachingDirectory(fsDir, 5.0, 60.0)
	conf := NewIndexWriterConfig(VERSION_49, analyzer)
	writer := NewIndexWriter(cachedFSDir, conf)

This will cache all newly flushed segments, all merged whose expected
segment size is <= 5 MB, unless the net cached bytes exceeds 60 MB at
which point all writes will not be cached (until the net bytes falls
below 60 MB).
*/
type NRTCachingDirectory struct {
	Directory
	sync.Locker

	cache             *RAMDirectory
	maxMergeSizeBytes int64
	maxCachedBytes    int64

	doCacheWrite func(name string, context IOContext) bool
	uncacheLock  sync.Locker
}

/*
We will cache a newly created output if 1) it's a flush or a merge
and the estimated size of the merged segment is <= maxMergedSizeMB,
and 2) the total cached bytes is <= maxCachedMB.
*/
func NewNRTCachingDirectory(delegate Directory, maxMergeSizeMB, maxCachedMB float64) (nrt *NRTCachingDirectory) {
	nrt = &NRTCachingDirectory{
		Directory:         delegate,
		Locker:            &sync.Mutex{},
		cache:             NewRAMDirectory(),
		maxMergeSizeBytes: int64(maxMergeSizeMB * 1024 * 1024),
		maxCachedBytes:    int64(maxCachedMB * 1024 * 1024),
		uncacheLock:       &sync.Mutex{},
	}
	// Subclass can override this to customize logic; return true if this
	// file should be written to the RAMDirectory.
	nrt.doCacheWrite = func(name string, context IOContext) bool {
		var bytes int64
		if context.MergeInfo != nil {
			bytes = context.MergeInfo.EstimatedMergeBytes
		} else if context.FlushInfo != nil {
			bytes = context.FlushInfo.EstimatedSegmentSize
		}
		if NRT_VERBOSE {
			log.Printf("CACHE check merge=%v flush=%v size=%v",
				context.MergeInfo, context.FlushInfo, bytes)
		}
		return name != "segments.gen" &&
			bytes <= nrt.maxMergeSizeBytes &&
			bytes+nrt.cache.RamBytesUsed() <= nrt.maxCachedBytes
	}
	return
}

func (nrt *NRTCachingDirectory) String() string {
	return fmt.Sprintf("NRTCachingDirectory(%v; maxCacheMB=%.2f maxMergeSizeMB=%.2f)",
		nrt.Directory, float32(nrt.maxCachedBytes/1024/1024),
		float32(nrt.maxMergeSizeBytes/1024/1024))
}

func (nrt *NRTCachingDirectory) ListAll() (all []string, err error) {
	nrt.Lock() // synchronized
	defer nrt.Unlock()
	files := make(map[string]bool)
	all, err = nrt.cache.ListAll()
	if err != nil {
		return
	}
	for _, f := range all {
		files[f] = true
	}
	// LUCENE-1468: our NRTCachingDirectory will actually exist (RAMDir!),
	// but if the underlying delegate is an FSDir and mkdirs() has not
	// yet been called, because so far everything is a cached write,
	// in this case, we don't want to throw a NoSuchDirectoryException
	all, err = nrt.Directory.ListAll()
	if err != nil {
		if _, ok := err.(*NoSuchDirectoryError); ok {
			// however if there are no cached files, then the directory truly
			// does not "exist"
			if len(files) == 0 {
				return
			}
		} else {
			return
		}
	}
	for _, f := range all {
		// Cannot do this -- if Lucene calls createOutput but files
		// already exists then this falsely trips:
		// assert2(!files[f], fmt.Sprintf("file '%v' is in both dirs", f)
		files[f] = true
	}
	all = make([]string, 0, len(files))
	for f, _ := range files {
		all = append(all, f)
	}
	return
}

// Returns how many bytes are being used by the RAMDirectory cache
// func (nrt *NRTCachingDirectory) sizeInBytes() int64 {
// 	return nrt.cache.sizeInBytes
// }

func (nrt *NRTCachingDirectory) FileExists(name string) bool {
	nrt.Lock() // synchronized
	defer nrt.Unlock()
	return nrt._fileExists(name)
}

func (nrt *NRTCachingDirectory) _fileExists(name string) bool {
	return nrt.cache.FileExists(name) || nrt.Directory.FileExists(name)
}

func (nrt *NRTCachingDirectory) DeleteFile(name string) error {
	assert(nrt.Directory != nil)
	nrt.Lock() // synchronized
	defer nrt.Unlock()

	if NRT_VERBOSE {
		log.Printf("nrtdir.deleteFile name=%v", name)
	}
	if nrt.cache.FileExists(name) {
		return nrt.cache.DeleteFile(name)
	} else {
		return nrt.Directory.DeleteFile(name)
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (nrt *NRTCachingDirectory) FileLength(name string) (length int64, err error) {
	nrt.Lock() // synchronized
	defer nrt.Unlock()
	if nrt.cache.FileExists(name) {
		return nrt.cache.FileLength(name)
	} else {
		return nrt.Directory.FileLength(name)
	}
}

func (nrt *NRTCachingDirectory) CreateOutput(name string, context IOContext) (out IndexOutput, err error) {
	if NRT_VERBOSE {
		log.Printf("nrtdir.createOutput name=%v", name)
	}
	if nrt.doCacheWrite(name, context) {
		if NRT_VERBOSE {
			log.Println("  to cache")
		}
		nrt.Directory.DeleteFile(name) // ignore IO error
		return nrt.cache.CreateOutput(name, context)
	}
	nrt.cache.DeleteFile(name) // ignore IO error
	return nrt.Directory.CreateOutput(name, context)
}

func (nrt *NRTCachingDirectory) Sync(fileNames []string) (err error) {
	if NRT_VERBOSE {
		log.Printf("nrtdir.sync files=%v", fileNames)
	}
	for _, fileName := range fileNames {
		err = nrt.unCache(fileName)
		if err != nil {
			return
		}
	}
	return nrt.Directory.Sync(fileNames)
}

func (nrt *NRTCachingDirectory) OpenInput(name string, context IOContext) (in IndexInput, err error) {
	nrt.Lock() // synchronized
	defer nrt.Unlock()
	if NRT_VERBOSE {
		log.Printf("nrtdir.openInput name=%v", name)
	}
	if nrt.cache.FileExists(name) {
		if NRT_VERBOSE {
			log.Println("  from cache")
		}
		return nrt.cache.OpenInput(name, context)
	}
	return nrt.Directory.OpenInput(name, context)
}

// func (nrt *NRTCachingDirectory) CreateSlicer(name string, context IOContext) (slicer IndexInputSlicer, err error) {
// 	nrt.EnsureOpen()
// 	if NRT_VERBOSE {
// 		log.Println("nrtdir.openInput name=%v", name)
// 	}
// 	if nrt.cache.FileExists(name) {
// 		if NRT_VERBOSE {
// 			log.Println("  from cache")
// 		}
// 		return nrt.cache.CreateSlicer(name, context)
// 	}
// 	return nrt.Directory.CreateSlicer(name, context)
// }

// Close this directory, which flushes any cached files to the
// delegate and then closes the delegate.
func (nrt *NRTCachingDirectory) Close() error {
	// NOTE: technically we shouldn't have to do this, ie,
	// IndexWriter should have sync'd all files, but we do
	// it for defensive reasons... or in case the app is
	// doing something custom (creating outputs directly w/o
	// using IndexWriter):
	all, err := nrt.cache.ListAll()
	if err != nil {
		return err
	}
	for _, fileName := range all {
		nrt.unCache(fileName)
	}
	err = nrt.cache.Close()
	if err != nil {
		return err
	}
	return nrt.Directory.Close()
}

func (nrt *NRTCachingDirectory) unCache(fileName string) (err error) {
	// Only let one goroutine uncache at a time; this only happens
	// during commit() or close():
	nrt.uncacheLock.Lock()
	defer nrt.uncacheLock.Unlock()

	log.Printf("nrtdir.unCache name=%v", fileName)
	if !nrt.cache.FileExists(fileName) {
		// Another goroutine beat us...
		return
	}
	context := IO_CONTEXT_DEFAULT
	var out IndexOutput
	out, err = nrt.Directory.CreateOutput(fileName, context)
	if err != nil {
		return
	}
	defer out.Close()
	var in IndexInput
	in, err = nrt.cache.OpenInput(fileName, context)
	if err != nil {
		return
	}
	defer in.Close()
	err = out.CopyBytes(in, in.Length())
	if err != nil {
		return
	}

	nrt.Lock() // Lock order: uncacheLock -> this
	defer nrt.Unlock()
	// Must sync here because other sync methods have
	// if nrt.cache.FileExists(name) { ... } else { ... }
	return nrt.cache.DeleteFile(fileName)
}
