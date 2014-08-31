package classic

const (
	LEXICAL_ERROR = iota
	STATIC_LEXER_ERROR
	INVALID_LEXICAL_STATE
	LOOP_DETECTED
)

type TokenManagerError struct {
}

func newTokenMgrError(eofSeen bool, lexState, errorLine, errorColumn int,
	errorAfter string, curChar rune, reason int) *TokenManagerError {
	panic("not implemented yet")
}

func (err *TokenManagerError) Error() string {
	panic("not implemented yet")
}
