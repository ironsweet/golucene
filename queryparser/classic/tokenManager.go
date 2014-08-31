package classic

type TokenManager struct {
	curLexState     int
	defaultLexState int
	jjnewStateCnt   int
	jjround         int
	jjmatchedPos    int
	jjmatchedKind   int

	input_stream CharStream
	jjrounds     []int
	jjstateSet   []int
	curChar      rune
}

func newTokenManager(stream CharStream) *TokenManager {
	return &TokenManager{
		curLexState:     2,
		defaultLexState: 2,
		input_stream:    stream,
		jjrounds:        make([]int, 49),
		jjstateSet:      make([]int, 98),
	}
}

func (tm *TokenManager) ReInit(stream CharStream) {
	tm.jjmatchedPos = 0
	tm.jjnewStateCnt = 0
	tm.curLexState = tm.defaultLexState
	tm.reInitRounds()
}

func (tm *TokenManager) reInitRounds() {
	tm.jjround = 0x80000001
	for i := 48; i >= 0; i-- {
		tm.jjrounds[i] = 0x80000000
	}
}

// L1027

func (tm *TokenManager) jjFillToken() *Token {
	panic("not implemented yet")
}

func (tm *TokenManager) nextToken() (matchedToken *Token) {
	curPos := 0
	var err error
	var eof = false
	for !eof {
		if tm.curChar, err = tm.input_stream.beginToken(); err != nil {
			tm.jjmatchedKind = 0
			matchedToken = tm.jjFillToken()
			return
		}

		switch tm.curLexState {
		case 0:
			panic("not implemented yet")
		case 1:
			panic("not implemented yet")
		case 2:
			panic("not implemented yet")
		}

		if tm.jjmatchedKind != 0x7fffffff {
			panic("not implemented yet")
		}
		error_line := tm.input_stream.endLine()
		error_column := tm.input_stream.endColumn()
		var error_after string
		var eofSeen = false
		if _, err = tm.input_stream.readChar(); err == nil {
			tm.input_stream.backup(1)
			tm.input_stream.backup(1)
			if curPos > 1 {
				error_after = tm.input_stream.image()
			}
		} else {
			eofSeen = true
			if curPos > 1 {
				error_after = tm.input_stream.image()
			}
			if tm.curChar == '\n' || tm.curChar == '\r' {
				error_line++
				error_column = 0
			} else {
				error_column++
			}
		}
		panic(newTokenMgrError(eofSeen, tm.curLexState, error_line,
			error_column, error_after, tm.curChar, LEXICAL_ERROR))
	}
	panic("should not be here")
}
