package index

// store/TrackingDirectoryWrapper.java

/*
A delegating Directory that records which files were written to and deleted.
*/
type TrackingDirectoryWrapper struct{}

func (w *TrackingDirectoryWrapper) createdFiles() map[string]bool {
	panic("not implemented yet")
}
