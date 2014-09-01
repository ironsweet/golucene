package search

type Occur int

var (
	MUST     = Occur(1)
	SHOULD   = Occur(2)
	MUST_NOT = Occur(3)
)

func (occur Occur) String() string {
	switch occur {
	case MUST:
		return "+"
	case SHOULD:
		return ""
	case MUST_NOT:
		return "-"
	}
	panic("should not be here")
}

type BooleanClause struct {
}
