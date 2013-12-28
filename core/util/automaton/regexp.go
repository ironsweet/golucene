package automaton

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// util/automaton/RegExp.java

type Kind int

const (
	REGEXP_UNION         = Kind(1)
	REGEXP_CONCATENATION = Kind(2)
	REGEXP_INTERSECTION  = Kind(3)
	REGEXP_OPTIONAL      = Kind(4)
	REGEXP_REPEAT        = Kind(5)
	REGEXP_REPEAT_MIN    = Kind(6)
	REGEXP_REPEAT_MINMAX = Kind(7)
	REGEXP_COMPLEMENT    = Kind(8)
	REGEXP_CHAR          = Kind(9)
	REGEXP_CHAR_RANGE    = Kind(10)
	REGEXP_ANYCHAR       = Kind(11)
	REGEXP_EMPTY         = Kind(12)
	REGEXP_STRING        = Kind(13)
	REGEXP_ANYSTRING     = Kind(14)
	REGEXP_AUTOMATON     = Kind(15)
	REGEXP_INTERVAL      = Kind(16)
)

// Syntax flags
const (
	INTERSECTION = 1 // &
	COMPLEMENT   = 2 // ~
)

// Syntax flag, enables all optional regex syntax.
const ALL = 0xffff

/*
Regular Expression extension to Automaton.

Regular expressions are built from the following abstract syntax:

	regexp	::= unionexp
 			|
	unionexp ::= interexp | unionexp 	(union)
	 		| interexp
	interexp ::= concatexp & interexp 	(intersection) 						[OPTIONAL]
 			| concatexp
	concatexp ::= repeatexp concatexp	 (concatenation)
 			| repeatexp
	repeatexp ::= repeatexp ? 			(zero or one occurrence)
 			| repeatexp * 				(zero or more occurrences)
 			| repeatexp + 				(one or more occurrences)
 			| repeatexp {n} 			(n occurrences)
 			| repeatexp {n,} 			(n or more occurrences)
 			| repeatexp {n,m} 			(n to m occurrences, including both)
 			| complexp
	complexp ::= ~ complexp 			(complement) 						[OPTIONAL]
 			| charclassexp
	charclassexp ::= [ charclasses ] 	(character class)
 			| [^ charclasses ] 			(negated character class)
 			| simpleexp
	charclasses ::= charclass charclasses
 			| charclass
	charclass ::= charexp - charexp 	(character range, including end-points)
 			| charexp
	simpleexp ::= charexp
 			| . 						(any single character)
 			| # 						(the empty language) 				[OPTIONAL]
 			| @ 						(any string) 						[OPTIONAL]
			| " <Unicode string without double-quotes>  " (a string)
 			| ( ) 						(the empty string)
 			| ( unionexp ) 				(precedence override)
 			| < <identifier> > 			(named automaton) 					[OPTIONAL]
			| <n-m> 					(numerical interval) 				[OPTIONAL]
	charexp ::= <Unicode character> 	(a single non-reserved character)
 			| \ <Unicode character>  	(a single character)

The productions marked [OPTIONAL] are only allowed if specified by
the syntax flags passed to the RegExp constructor. The reserved
characters used in the (enabled) syntax must be escaped with
backslash (\) or double-quotes ("..."). (In contrast to other regexp
syntaxes, this is required also in character classes.) Be aware that
dash (-) has a special meaning in charclass expressions. An
identifier is a string not containing right angle bracket (>) or dash
(-). Numerical intervals are specified by non-negative decimal
integers and include both end points, and if n and m have the same
number of digits, then the conforming strings must have that length
(i.e. prefixed by 0's).
*/
type RegExp struct {
	kind             Kind
	exp1, exp2       *RegExp
	s                string
	c                int
	min, max, digits int
	from, to         int
	b                []rune
	flags            int
	pos              int
}

// Constructs new RegExp from a string. Same as RegExp(s, ALL)
func NewRegExp(s string) *RegExp {
	return NewRegExpWithFlag(s, ALL)
}

// Constructs new RegExp from a string.
func NewRegExpWithFlag(s string, flags int) *RegExp {
	ans := &RegExp{
		b:     []rune(s),
		flags: flags,
	}
	var e *RegExp
	if len(s) == 0 {
		e = makeString("")
	} else {
		e = ans.parseUnionExp()
		if ans.pos < len(ans.b) {
			panic(fmt.Sprintf("end-of-string expected at position %v", ans.pos))
		}
	}
	ans.kind = e.kind
	ans.exp1, ans.exp2 = e.exp1, e.exp2
	ans.s = e.s
	ans.min, ans.max, ans.digits = e.min, e.max, e.digits
	ans.from, ans.to = e.from, e.to
	ans.b = nil
	return ans
}

// Constructs new Automaton from this RegExp. Same as
// ToAutomaton(nil) (empty automaton map).
func (re *RegExp) ToAutomaton() *Automaton {
	return re.toAutomatonAllowMutate(nil, nil)
}

func (re *RegExp) toAutomatonAllowMutate(automata map[string]*Automaton,
	provider AutomatonProvider) *Automaton {
	// b := false
	if ALLOW_MUTATION {
		panic("not implemented yet")
	}
	a := re.toAutomaton(automata, provider)
	if ALLOW_MUTATION {
		panic("not implemented yet")
	}
	return a
}

func (re *RegExp) toAutomaton(automata map[string]*Automaton,
	provider AutomatonProvider) *Automaton {
	var list []*Automaton
	var a *Automaton = nil
	switch re.kind {
	case REGEXP_UNION:
		list = make([]*Automaton, 0)
		list = re.findLeaves(re.exp1, REGEXP_UNION, list, automata, provider)
		list = re.findLeaves(re.exp2, REGEXP_UNION, list, automata, provider)
		a = union(list)
		minimize(a)
	case REGEXP_CONCATENATION:
		panic("not implemented yet")
	case REGEXP_INTERSECTION:
		a = re.exp1.toAutomaton(automata, provider).intersection(
			re.exp2.toAutomaton(automata, provider))
		minimize(a)
	case REGEXP_OPTIONAL:
		panic("not implemented yet")
	case REGEXP_REPEAT:
		a = re.exp1.toAutomaton(automata, provider).repeat()
		minimize(a)
	case REGEXP_REPEAT_MIN:
		a = re.exp1.toAutomaton(automata, provider).repeatMin(re.min)
		minimize(a)
	case REGEXP_REPEAT_MINMAX:
		panic("not implemented yet")
	case REGEXP_COMPLEMENT:
		a = re.exp1.toAutomaton(automata, provider).complement()
		minimize(a)
	case REGEXP_CHAR:
		a = makeChar(re.c)
	case REGEXP_CHAR_RANGE:
		a = makeCharRange(re.from, re.to)
	case REGEXP_ANYCHAR:
		a = makeAnyChar()
	case REGEXP_EMPTY:
		panic("not implemented yet")
	case REGEXP_STRING:
		panic("not implemented yet")
	case REGEXP_ANYSTRING:
		panic("not implemented yet")
	case REGEXP_AUTOMATON:
		panic("not implemented yet")
	case REGEXP_INTERVAL:
		panic("not implemented yet")
	}
	return a
}

func (re *RegExp) findLeaves(exp *RegExp, kind Kind, list []*Automaton,
	automata map[string]*Automaton, provider AutomatonProvider) []*Automaton {
	if exp.kind == kind {
		list = re.findLeaves(exp.exp1, kind, list, automata, provider)
		list = re.findLeaves(exp.exp2, kind, list, automata, provider)
		return list
	} else {
		return append(list, exp.toAutomaton(automata, provider))
	}
}

// Constructs string from parsed regular expression
func (re *RegExp) String() string {
	var b bytes.Buffer
	return re.toStringBuilder(&b).String()
}

func (re *RegExp) toStringBuilder(b *bytes.Buffer) *bytes.Buffer {
	switch re.kind {
	case REGEXP_UNION:
		b.WriteString("(")
		re.exp1.toStringBuilder(b)
		b.WriteString("|")
		re.exp2.toStringBuilder(b)
		b.WriteString(")")
	case REGEXP_CONCATENATION:
		panic("not implemented yet")
	case REGEXP_INTERSECTION:
		b.WriteString("(")
		re.exp1.toStringBuilder(b)
		b.WriteString("&")
		re.exp2.toStringBuilder(b)
		b.WriteString(")")
	case REGEXP_OPTIONAL:
		panic("not implemented yet")
	case REGEXP_REPEAT:
		panic("not implemented yet")
	case REGEXP_REPEAT_MIN:
		b.WriteString("(")
		re.exp1.toStringBuilder(b)
		fmt.Fprintf(b, "){%v,}", re.min)
	case REGEXP_REPEAT_MINMAX:
		panic("not implemented yet")
	case REGEXP_COMPLEMENT:
		b.WriteString("~(")
		re.exp1.toStringBuilder(b)
		b.WriteString(")")
	case REGEXP_CHAR:
		b.WriteString("\\")
		if rune(re.c) == '\r' { // edge case
			b.WriteRune('r')
		} else if rune(re.c) == '\t' { // edge case
			b.WriteRune('t')
		} else if rune(re.c) == '\n' { // edge case
			b.WriteRune('n')
		} else {
			b.WriteRune(rune(re.c))
		}
	case REGEXP_CHAR_RANGE:
		panic("not implemented yet")
	case REGEXP_ANYCHAR:
		b.WriteString(".")
	case REGEXP_EMPTY:
		panic("not implemented yet")
	case REGEXP_STRING:
		panic("not implemented yet")
	case REGEXP_ANYSTRING:
		panic("not implemented yet")
	case REGEXP_AUTOMATON:
		panic("not implemented yet")
	case REGEXP_INTERVAL:
		panic("not implemented yet")
	default:
		panic("not supported yet")
	}
	return b
}

func makeUnion(exp1, exp2 *RegExp) *RegExp {
	return &RegExp{
		kind: REGEXP_UNION,
		exp1: exp1,
		exp2: exp2,
	}
}

func makeConcatenation(exp1, exp2 *RegExp) *RegExp {
	panic("not implemented yet")
}

func makeIntersection(exp1, exp2 *RegExp) *RegExp {
	return &RegExp{
		kind: REGEXP_INTERSECTION,
		exp1: exp1,
		exp2: exp2,
	}
}

func makeOptional(exp *RegExp) *RegExp {
	panic("not implemented yet")
}

func makeRepeat(exp *RegExp) *RegExp {
	return &RegExp{
		kind: REGEXP_REPEAT,
		exp1: exp,
	}
}

func makeRepeatMin(exp *RegExp, min int) *RegExp {
	return &RegExp{
		kind: REGEXP_REPEAT_MIN,
		exp1: exp,
		min:  min,
	}
}

func makeRepeatRange(exp *RegExp, min, max int) *RegExp {
	panic("not implemented yet")
}

func makeComplement(exp *RegExp) *RegExp {
	return &RegExp{
		kind: REGEXP_COMPLEMENT,
		exp1: exp,
	}
}

func makeCharRE(c int) *RegExp {
	return &RegExp{
		kind: REGEXP_CHAR,
		c:    c,
	}
}

func makeCharRangeRE(from, to int) *RegExp {
	assert2(from <= to, fmt.Sprintf("invalid range: from (%v) cannot be > to (%v)", from, to))
	return &RegExp{
		kind: REGEXP_CHAR_RANGE,
		from: from,
		to:   to,
	}
}

func assert2(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func makeAnyCharRE() *RegExp {
	return &RegExp{kind: REGEXP_ANYCHAR}
}

func makeString(s string) *RegExp {
	panic("not implemented yet")
}

func (re *RegExp) peek(s string) bool {
	return re.more() && strings.ContainsRune(s, re.b[re.pos])
}

func (re *RegExp) match(c rune) bool {
	if re.pos >= len(re.b) {
		return false
	}
	if re.b[re.pos] == c {
		re.pos++
		return true
	}
	return false
}

func (re *RegExp) more() bool {
	return re.pos < len(re.b)
}

func (re *RegExp) next() int {
	assert2(re.more(), "unexpected end-of-string")
	ch := re.b[re.pos]
	re.pos++
	return int(ch) // int >= rune
}

func (re *RegExp) check(flag int) bool {
	return (re.flags & flag) != 0
}

func (re *RegExp) parseUnionExp() *RegExp {
	e := re.parseInterExp()
	if re.match('|') {
		e = makeUnion(e, re.parseUnionExp())
	}
	return e
}

func (re *RegExp) parseInterExp() *RegExp {
	e := re.parseConcatExp()
	if re.check(INTERSECTION) && re.match('&') {
		e = makeIntersection(e, re.parseInterExp())
	}
	return e
}

func (re *RegExp) parseConcatExp() *RegExp {
	e := re.parseRepeatExp()
	if re.more() && !re.peek(")|") && (!re.check(INTERSECTION) || !re.peek("&")) {
		e = makeConcatenation(e, re.parseConcatExp())
	}
	return e
}

func (re *RegExp) parseRepeatExp() *RegExp {
	e := re.parseComplExp()
	for re.peek("?*+{") {
		if re.match('?') {
			e = makeOptional(e)
		} else if re.match('*') {
			e = makeRepeat(e)
		} else if re.match('+') {
			e = makeRepeatMin(e, 1)
		} else if re.match('{') {
			start := re.pos
			for re.peek("0123456789") {
				re.next()
			}
			assert2(start != re.pos, fmt.Sprintf("integer expected at position %v", re.pos))
			n, err := strconv.Atoi(string(re.b[start:re.pos]))
			assertNoError(err)
			m := -1
			if re.match(',') {
				start = re.pos
				for re.peek("0123456789") {
					re.next()
				}
				if start != re.pos {
					m, err = strconv.Atoi(string(re.b[start:re.pos]))
					assertNoError(err)
				}
			} else {
				m = n
			}
			assert2(re.match('}'), fmt.Sprintf("expected '}' at position %v", re.pos))
			if m == -1 {
				e = makeRepeatMin(e, n)
			} else {
				e = makeRepeatRange(e, n, m)
			}
		}
	}
	return e
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func (re *RegExp) parseComplExp() *RegExp {
	if re.check(COMPLEMENT) && re.match('~') {
		return makeComplement(re.parseComplExp())
	}
	return re.parseCharClassExp()
}

func (re *RegExp) parseCharClassExp() *RegExp {
	if re.match('[') {
		negate := re.match('^')
		e := re.parseCharClasses()
		if negate {
			e = makeIntersection(makeAnyCharRE(), makeComplement(e))
		}
		assert2(re.match(']'), fmt.Sprintf("expected ']' at position %v", re.pos))
		return e
	}
	return re.parseSimpleExp()
}

func (re *RegExp) parseCharClasses() *RegExp {
	e := re.parseCharClass()
	for re.more() && !re.peek("]") {
		e = makeUnion(e, re.parseCharClass())
	}
	return e
}

func (re *RegExp) parseCharClass() *RegExp {
	c := re.parseCharExp()
	if re.match('-') {
		return makeCharRangeRE(c, re.parseCharExp())
	}
	return makeCharRE(c)
}

func (re *RegExp) parseSimpleExp() *RegExp {
	if re.match('.') {
		return makeAnyCharRE()
	}
	panic("not implemented yet")
}

func (re *RegExp) parseCharExp() int {
	re.match('\\')
	return re.next()
}

// util/automaton/AutomatonProvider.java

// Automaton provider for RegExp.
type AutomatonProvider func(name string) *Automaton
