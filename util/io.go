package util

type Closeable interface {
	Close() error
}

func CloseWhileHandlingException(priorErr error, objects ...Closeable) error {
	var th error = nil
	for _, object := range objects {
		if object == nil {
			continue
		}
		if t := object.Close(); t != nil && th == nil {
			th = t
		}
	}

	if priorErr != nil {
		return priorErr
	} else if th != nil {
		return th
	} else {
		return nil
	}
}
