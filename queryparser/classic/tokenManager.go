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

// L41

func (tm *TokenManager) jjMoveStringLiteralDfa0_2() int {
	switch tm.curChar {
	case 40:
		panic("not implemented yet")
	case 41:
		panic("not implemented yet")
	case 42:
		panic("not implemented yet")
	case 43:
		panic("not implemented yet")
	case 45:
		panic("not implemented yet")
	case 58:
		panic("not implemented yet")
	case 91:
		panic("not implemented yet")
	case 94:
		panic("not implemented yet")
	case 123:
		panic("not implemented yet")
	default:
		return tm.jjMoveNfa_2(0, 0)
	}
}

// L87

func (tm *TokenManager) jjMoveNfa_2(startState, curPos int) int {
	startsAt := 0
	tm.jjnewStateCnt = 49
	i := 1
	tm.jjstateSet[0] = startState
	kind := 0x7fffffff
	for {
		if tm.jjround++; tm.jjround == 0x7fffffff {
			tm.reInitRounds()
		}
		if tm.curChar < 64 {
			panic("not implemented yet")
		} else if tm.curChar < 128 {
			panic("not implemented yet")
		} else {
			panic("not implemented yet")
		}
		if kind != 0x7fffffff {
			tm.jjmatchedKind = kind
			tm.jjmatchedPos = curPos
			kind = 0x7fffffff
		}
		curPos++
		i = tm.jjnewStateCnt
		tm.jjnewStateCnt = startsAt
		startsAt = 49 - tm.jjnewStateCnt
		if i == startsAt {
			return curPos
		}
		var err error
		if tm.curChar, err = tm.input_stream.readChar(); err != nil {
			return curPos
		}
	}
	panic("should not be here")
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
			tm.jjmatchedKind = 0x7fffffff
			tm.jjmatchedPos = 0
			curPos = tm.jjMoveStringLiteralDfa0_2()
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
