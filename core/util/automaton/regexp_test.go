package automaton

import (
	"fmt"
	"testing"
)

func TestRegExpSimple(t *testing.T) {
	r := NewRegExp("[^ \t\r\n]+")
	rr := r.String()
	assert2("((.&~((((\\ |\\t)|\\r)|\\n)))){1,}" == rr, rr)
	assert(REGEXP_REPEAT_MIN == r.kind)
	assert(1 == r.min)
	assert(8 == r.pos)
	r = r.exp1
	assert(REGEXP_INTERSECTION == r.kind)
	r = r.exp2
	assert(REGEXP_COMPLEMENT == r.kind)
	r = r.exp1
	assert(REGEXP_UNION == r.kind)
	r = r.exp1
	assert(REGEXP_UNION == r.kind)
	r = r.exp1
	assert(REGEXP_UNION == r.kind)
	r = r.exp1
	assert2(32 == r.c, fmt.Sprintf("r.c=%v", r.c))
	assert(REGEXP_CHAR == r.kind)
}

func TestRegExpSimple2(t *testing.T) {
	r := NewRegExpWithFlag("*.?-", NONE)
	assert(REGEXP_CONCATENATION == r.kind)
	r = r.exp1
	assert(REGEXP_CHAR == r.kind)
	assert(42 == r.c)
}
