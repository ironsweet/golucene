package store

import (
	"fmt"
	"sync"
)

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
	conf := NewIndexWriterConfig(VERSION_45, analyzer)
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
	maxMergeSizeBytes float64
	maxCachedBytes    float64
	doCacheWrite      func(name string, context IOContext) bool
	uncacheLock       sync.Locker
}

/*
We will cache a newly created output if 1) it's a flush or a merge
and the estimated size of the merged segment is <= maxMergedSizeMB,
and 2) the total cached bytes is <= maxCachedMB.
*/
func NewNRTCachingDirectory(delegate Directory, maxMergeSizeMB, maxCachedMB float64) *NRTCachingDirectory {
	return &NRTCachingDirectory{
		Directory:         delegate,
		Locker:            &sync.Mutex{},
		cache:             NewRAMDirectory(),
		maxMergeSizeBytes: maxMergeSizeMB * 1024 * 1024,
		maxCachedBytes:    maxCachedMB * 1024 * 1024,
		doCacheWrite:      doCacheWrite,
		uncacheLock:       &sync.Mutex{},
	}
}

func (nrt *NRTCachingDirectory) String() string {
	return fmt.Sprintf("NRTCachingDirectory(%v; maxCacheMB=%v maxMergeSizeMB=%v)",
		nrt.Directory, nrt.maxCachedBytes/1024/1024, nrt.maxMergeSizeBytes/1024/1024)
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
	return nrt.cache.FileExists(name) || nrt.Directory.FileExists(name)
}

func (nrt *NRTCachingDirectory) DeleteFile(name string) error {
	panic("not implemented yet")
	// nrt.Lock() // synchronized
	// defer nrt.Unlock()
	// log.Printf("nrtdir.deleteFile name=%v", name)
	// if nrt.FileExists(name) {
	// 	assert2(!nrt.Directory.FileExists(name), fmt.Sprintf("name=%v", name))
	// 	return nrt.cache.DeleteFile(name)
	// } else {
	// 	return nrt.Directory.DeleteFile(name)
	// }
}

func assert2(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func (nrt *NRTCachingDirectory) FileLength(name string) (length int64, err error) {
	panic("not implemented yet")
	// nrt.Lock() // synchronized
	// defer nrt.Unlock()
	// if nrt.cache.FileExists(name) {
	// 	return nrt.cache.FileLength(name)
	// } else {
	// 	return nrt.Directory.FileLength(name)
	// }
}

func (nrt *NRTCachingDirectory) CreateOutput(name string, context IOContext) (out IndexOutput, err error) {
	panic("not implemented yet")
}

func (nrt *NRTCachingDirectory) Sync(fileNames []string) error {
	panic("not implemented yet")
}

func (nrt *NRTCachingDirectory) OpenInput(name string, context IOContext) (in IndexInput, err error) {
	panic("not implemented yet")
}

func (nrt *NRTCachingDirectory) CreateSlicer(name string, context IOContext) (slicer IndexInputSlicer, err error) {
	panic("not implemented yet")
}

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

// Subclass can override this to customize logic; return true if this
// file should be written to the RAMDirectory.
func doCacheWrite(name string, context IOContext) bool {
	panic("not implemented yet")
}

func (nrt *NRTCachingDirectory) unCache(fileName string) error {
	panic("not implemented yet")
}
