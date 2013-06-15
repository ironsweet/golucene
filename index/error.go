package index

type CorruptIndexError struct {
	msg string
}

func (err *CorruptIndexError) Error() string {
	return err.msg
}
