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
		t := object.Close()
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
		object.Close()
	}
}

func Close(objects ...io.Closer) error {
	var th error = nil

	for _, object := range objects {
		if object == nil {
			continue
		}
		t := object.Close()
		addSuppressed(th, t)
		if th == nil {
			th = t
		}
	}

	return th
}

func addSuppressed(err error, suppressed error) error {
	if err == suppressed {
		panic("Self-suppression not permitted")
	}
	if suppressed == nil {
		return err
	}
	if ce, ok := err.(*CompoundError); ok {
		ce.errs = append(ce.errs, suppressed)
		return ce
	}
	return &CompoundError{[]error{suppressed}}
}
