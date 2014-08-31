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

// L1059

func (tm *TokenManager) nextToken() *Token {
	panic("not implemented yet")
}
