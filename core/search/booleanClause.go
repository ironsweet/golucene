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
	query Query
	occur Occur
}

func NewBooleanClause(query Query, occur Occur) *BooleanClause {
	return &BooleanClause{
		query: query,
		occur: occur,
	}
}

func (c *BooleanClause) IsProhibited() bool {
	return c.occur == MUST_NOT
}

func (c *BooleanClause) IsRequired() bool {
	return c.occur == MUST
}
