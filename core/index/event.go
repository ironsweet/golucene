package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
)

/*
Interface for internal atomic events. See DocumentsWriter fo details.
Events are executed concurrently and no order is guaranteed. Each
event should only rely on the serializeability within its process
method. All actions that must happen before or after a certain action
must be encoded inside the process() method.
*/
type Event func(writer *IndexWriter, triggerMerge, clearBuffers bool) error

var applyDeletesEvent = Event(func(writer *IndexWriter, triggerMerge, forcePurge bool) error {
	panic("not implemented yet")
})

var mergePendingEvent = Event(func(writer *IndexWriter, triggerMerge, forcePurge bool) error {
	return writer.doAfterSegmentFlushed(triggerMerge, forcePurge)
})

var forcedPurgeEvent = Event(func(writer *IndexWriter, triggerMerge, forcePurge bool) error {
	_, err := writer.purge(true)
	return err
})

func newFlushFailedEvent(info *model.SegmentInfo) Event {
	return Event(func(writer *IndexWriter, triggerMerge, forcePurge bool) error {
		return writer.flushFailed(info)
	})
}

func newDeleteNewFilesEvent(files map[string]bool) Event {
	return Event(func(writer *IndexWriter, triggerMerge, forcePurge bool) error {
		writer.Lock()
		defer writer.Unlock()
		var fileList []string
		for file, _ := range files {
			fileList = append(fileList, file)
		}
		writer.deleter.deleteNewFiles(fileList)
		return nil
	})
}
