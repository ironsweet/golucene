package classic

import (
// "fmt"
)

var jjbitVec0 = []int64{1, 0, 0, 0}
var jjbitVec4 = []uint64{
	0xfffefffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
}

var jjnextStates = []int{
	37, 39, 40, 17, 18, 20, 42, 45, 31, 46, 43, 22, 23, 25, 26, 24,
	25, 26, 45, 31, 46, 44, 47, 35, 22, 28, 29, 27, 27, 30, 30, 0,
	1, 2, 4, 5,
}

var jjstrLiteralImages = map[int]string{
	0: "", 11: "\u0053", 12: "\055",
	14: "\050", 15: "\051", 16: "\072", 17: "\052", 18: "\136",
	25: "\133", 26: "\173", 28: "\124\117", 29: "\135", 30: "\175",
}

var jjnewLexState = []int{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 0, -1, -1, -1, -1, -1, -1,
	1, 1, 2, -1, 2, 2, -1, -1,
}

var jjtoToken = []int64{0x1ffffff01}

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
			l := int64(1 << uint(tm.curChar))
			for {
				i--
				switch tm.jjstateSet[i] {
				case 49, 33:
					if (0xfbff7cf8ffffd9ff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
				case 0:
					if (0xfbff54f8ffffd9ff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					} else if (0x100002600 & l) != 0 {
						if kind > 7 {
							kind = 7
						}
					} else if (0x280200000000 & l) != 0 {
						panic("not implemented yet")
					} else if tm.curChar == 47 {
						panic("not implemented yet")
					} else if tm.curChar == 34 {
						panic("not implemented yet")
					}
					if (0x7bff50f8ffffd9ff & l) != 0 {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddStates(6, 10)
					} else if tm.curChar == 42 {
						panic("not implemented yet")
					} else if tm.curChar == 33 {
						panic("not implemented yet")
					}
					if tm.curChar == 38 {
						panic("not implemented yet")
					}

				case 4:
					panic("not implemented yet")
				case 5:
					panic("not implemented yet")
				case 13:
					panic("not implemented yet")
				case 14:
					panic("not implemented yet")
				case 15:
					panic("not implemented yet")
				case 16:
					panic("not implemented yet")
				case 17:
					panic("not implemented yet")
				case 19:
					panic("not implemented yet")
				case 20:
					panic("not implemented yet")
				case 22:
					panic("not implemented yet")
				case 23:
					panic("not implemented yet")
				case 24:
					panic("not implemented yet")
				case 25:
					panic("not implemented yet")
				case 27:
					panic("not implemented yet")
				case 28:
					panic("not implemented yet")
				case 30:
					panic("not implemented yet")
				case 31:
					if tm.curChar == 42 && kind > 22 {
						kind = 22
					}
				case 32:
					panic("not implemented yet")
				case 35:
					panic("not implemented yet")
				case 36, 38:
					panic("not implemented yet")
				case 37:
					panic("not implemented yet")
				case 40:
					panic("not implemented yet")
				case 41:
					panic("not implemented yet")
				case 42:
					if (0x7bff78f8ffffd9ff & l) != 0 {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddTwoStates(42, 43)
					}
				case 44:
					panic("not implemented yet")
				case 45:
					if (0x7bff78f8ffffd9ff & l) != 0 {
						tm.jjCheckNAddStates(18, 20)
					}
				case 47:
					panic("not implemented yet")
				}
				if i == startsAt {
					break
				}
			}
		} else if tm.curChar < 128 {
			l := int64(1) << (uint(tm.curChar) & 077)
			for {
				i--
				switch tm.jjstateSet[i] {
				case 49:
					panic("niy")
				case 0:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddStates(6, 10)
					} else if tm.curChar == 92 {
						panic("niy")
					} else if tm.curChar == 126 {
						panic("niy")
					}
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
					switch tm.curChar {
					case 78:
						panic("niy")
					case 124:
						panic("niy")
					case 79:
						panic("niy")
					case 65:
						panic("niy")
					}
				case 1:
					panic("niy")
				case 2:
					panic("niy")
				case 3:
					panic("niy")
				case 6:
					panic("niy")
				case 7:
					panic("niy")
				case 8:
					panic("niy")
				case 9:
					panic("niy")
				case 10:
					panic("niy")
				case 11:
					panic("niy")
				case 12:
					panic("niy")
				case 17:
					panic("niy")
				case 18:
					panic("niy")
				case 19:
					panic("niy")
				case 21:
					panic("niy")
				case 25:
					panic("niy")
				case 26:
					panic("niy")
				case 27:
					panic("niy")
				case 28:
					panic("niy")
				case 29:
					panic("niy")
				case 30:
					panic("niy")
				case 32:
					panic("niy")
				case 33:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
				case 34:
					if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(35, 35)
					}
				case 35:
					panic("niy")
				case 37:
					panic("niy")
				case 39:
					panic("niy")
				case 41:
					panic("niy")
				case 42:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddTwoStates(42, 43)
					}
				case 43:
					if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(44, 44)
					}
				case 44:
					panic("niy")
				case 45:
					if (0x97ffffff87ffffff & uint64(l)) != 0 {
						tm.jjCheckNAddStates(18, 20)
					}
				case 46:
					if tm.curChar == 92 {
						tm.jjCheckNAddTwoStates(47, 47)
					}
				case 47:
					panic("niy")
				case 48:
					panic("niy")
				}
				if i <= startsAt { // ==?
					break
				}
			}
		} else {
			hiByte := int(tm.curChar >> 8)
			i1 := hiByte >> 6
			l1 := int64(1 << (uint64(hiByte) & 077))
			i2 := int((tm.curChar & 0xff) >> 6)
			l2 := int64(1 << uint64(tm.curChar&077))
			for {
				i--
				switch tm.jjstateSet[i] {
				case 49, 33:
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
				case 0:
					if jjCanMove_0(hiByte, i1, i2, l1, l2) {
						if kind > 7 {
							kind = 7
						}
					}
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						if kind > 23 {
							kind = 23
						}
						tm.jjCheckNAddTwoStates(33, 34)
					}
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddStates(6, 10)
					}
				case 15:
					panic("not implemented yet")
				case 17, 19:
					panic("not implemented yet")
				case 25:
					panic("not implemented yet")
				case 27:
					panic("not implemented yet")
				case 28:
					panic("not implemented yet")
				case 30:
					panic("not implemented yet")
				case 32:
					panic("not implemented yet")
				case 35:
					panic("not implemented yet")
				case 37:
					panic("not implemented yet")
				case 41:
					panic("not implemented yet")
				case 42:
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						if kind > 20 {
							kind = 20
						}
						tm.jjCheckNAddTwoStates(42, 43)
					}
				case 44:
					panic("not implemented yet")
				case 45:
					if jjCanMove_2(hiByte, i1, i2, l1, l2) {
						tm.jjCheckNAddStates(18, 20)
					}
				case 47:
					panic("not implemented yet")
				}
				if i == startsAt {
					break
				}
			}
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

func jjCanMove_0(hiByte, i1, i2 int, l1, l2 int64) bool {
	switch hiByte {
	case 48:
		return (jjbitVec0[i2] & 12) != 0
	}
	return false
}

func jjCanMove_2(hiByte, i1, i2 int, l1, l2 int64) bool {
	switch hiByte {
	case 0:
		panic("not implemented yet")
	case 48:
		panic("not implemented yet")
	}
	return (jjbitVec4[i1] & uint64(l1)) != 0
}

func (tm *TokenManager) ReInit(stream CharStream) {
	tm.jjmatchedPos = 0
	tm.jjnewStateCnt = 0
	tm.curLexState = tm.defaultLexState
	tm.input_stream = stream
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
	var curTokenImage string
	if im, ok := jjstrLiteralImages[tm.jjmatchedKind]; ok {
		curTokenImage = im
	} else {
		curTokenImage = tm.input_stream.image()
	}
	beginLine := tm.input_stream.beginLine()
	beginColumn := tm.input_stream.beginColumn()
	endLine := tm.input_stream.endLine()
	endColumn := tm.input_stream.endColumn()
	t := newToken(tm.jjmatchedKind, curTokenImage)

	t.beginLine = beginLine
	t.endLine = endLine
	t.beginColumn = beginColumn
	t.endColumn = endColumn
	return t
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
			if tm.jjmatchedPos+1 < curPos {
				tm.input_stream.backup(curPos - tm.jjmatchedPos - 1)
			}
			if (jjtoToken[tm.jjmatchedKind>>6] & (int64(1) << uint(tm.jjmatchedKind&077))) != 0 {
				matchedToken = tm.jjFillToken()
				if jjnewLexState[tm.jjmatchedKind] != -1 {
					panic("not implemented yet")
				}
				return matchedToken
			} else {
				if n := jjnewLexState[tm.jjmatchedKind]; n != -1 {
					panic("niy")
					// tm.curLexState = n
				}
				continue
			}
		}
		error_line := tm.input_stream.endLine()
		error_column := tm.input_stream.endColumn()
		var error_after string
		var eofSeen = false
		if _, err = tm.input_stream.readChar(); err == nil {
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

// L1137
func (tm *TokenManager) jjCheckNAdd(state int) {
	if tm.jjrounds[state] != tm.jjround {
		tm.jjstateSet[tm.jjnewStateCnt] = state
		tm.jjnewStateCnt++
		tm.jjrounds[state] = tm.jjround
	}
}

// L1151

func (tm *TokenManager) jjCheckNAddTwoStates(state1, state2 int) {
	tm.jjCheckNAdd(state1)
	tm.jjCheckNAdd(state2)
}

func (tm *TokenManager) jjCheckNAddStates(start, end int) {
	assert(start < end)
	assert(start >= 0)
	assert(end <= len(jjnextStates))
	for {
		tm.jjCheckNAdd(jjnextStates[start])
		start++
		if start >= end {
			break
		}
	}
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}
