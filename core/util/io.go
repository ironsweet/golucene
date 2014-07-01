package util

import (
	"io"
)

type CompoundError struct {
	errs []error
}

func (e *CompoundError) Error() string {
	return e.errs[0].Error()
}

func CloseWhileHandlingError(priorErr error, objects ...io.Closer) error {
	var th error = nil

	for _, object := range objects {
		if object == nil {
			continue
		}
		t := safeClose(object)
		if t == nil {
			continue
		}
		if priorErr == nil {
			addSuppressed(th, t)
		} else {
			addSuppressed(priorErr, t)
		}
		if th == nil {
			th = t
		}
	}

	if priorErr != nil {
		return priorErr
	}
	return th
}

func CloseWhileSuppressingError(objects ...io.Closer) {
	for _, object := range objects {
		if object == nil {
			continue
		}
		safeClose(object)
	}
}

func Close(objects ...io.Closer) error {
	var th error = nil

	for _, object := range objects {
		if object == nil {
			continue
		}
		t := safeClose(object)
		if t != nil {
			addSuppressed(th, t)
			if th == nil {
				th = t
			}
		}
	}

	return th
}

func safeClose(obj io.Closer) (err error) {
	// defer func() {
	// 	if p := recover(); p != nil {
	// 		// err = errors.New(fmt.Sprintf("%v", p))
	// 		fmt.Println("Uncaught nil pointer of closer.", obj, p)
	// 	}
	// }()
	if obj != nil {
		err = obj.Close()
	}
	return
}

func addSuppressed(err error, suppressed error) error {
	assert2(err != suppressed, "Self-suppression not permitted")
	if suppressed == nil {
		return err
	}
	if ce, ok := err.(*CompoundError); ok {
		ce.errs = append(ce.errs, suppressed)
		return ce
	}
	return &CompoundError{[]error{suppressed}}
}

type FileDeleter interface {
	DeleteFile(name string) error
}

/*
Deletes all given files, suppressing all throw errors.

Note that the files should not be nil.
*/
func DeleteFilesIgnoringErrors(dir FileDeleter, files ...string) {
	for _, name := range files {
		dir.DeleteFile(name) // ignore error
	}
}
