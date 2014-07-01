package search

import (
	"bytes"
	"fmt"
)

// search/Explanation.java

type Explanation interface {
	// Indicate whether or not this Explanation models a good match.
	// By default, an Explanation represents a "match" if the value is positive.
	IsMatch() bool
	// The value assigned to this explanation node.
	Value() float32
}

type ExplanationSPI interface {
	Explanation
	// A short one line summary which should contian all high level information
	// about this Explanation, without the Details.
	Summary() string
	Details() []Explanation
}

/* Expert: Describes the score computation for document and query. */
type ExplanationImpl struct {
	spi         ExplanationSPI
	value       float32       // the value of this node
	description string        // what it represents
	details     []Explanation // sub-explanations
}

func newExplanation(value float32, description string) *ExplanationImpl {
	ans := &ExplanationImpl{value: value, description: description}
	ans.spi = ans
	return ans
}

func (exp *ExplanationImpl) IsMatch() bool       { return exp.value > 0.0 }
func (exp *ExplanationImpl) Value() float32      { return exp.value }
func (exp *ExplanationImpl) Description() string { return exp.description }

func (exp *ExplanationImpl) Summary() string {
	return fmt.Sprintf("%v = %v", exp.value, exp.description)
}

// The sub-nodes of this explanation node.
func (exp *ExplanationImpl) Details() []Explanation {
	return exp.details
}

// Adds a sub-node to this explanation node
func (exp *ExplanationImpl) addDetail(detail Explanation) {
	exp.details = append(exp.details, detail)
}

// Render an explanation as text.
func (exp *ExplanationImpl) String() string {
	return explanationToString(exp.spi, 0)
}

func explanationToString(exp ExplanationSPI, depth int) string {
	assert(depth <= 1000) // potential dead loop
	var buf bytes.Buffer
	for i := 0; i < depth; i++ {
		buf.WriteString("  ")
	}
	buf.WriteString(exp.Summary())
	buf.WriteString("\n")

	for _, v := range exp.Details() {
		buf.WriteString(explanationToString(v.(ExplanationSPI), depth+1))
	}

	return buf.String()
}

// search/ComplexExplanation.java

/*
Expert: Describes the score computation for the doucment and query,
and can distinguish a match independent of a postive value.
*/
type ComplexExplanation struct {
	*ExplanationImpl
	match interface{}
}

func newEmptyComplexExplanation() *ComplexExplanation {
	ans := new(ComplexExplanation)
	ans.ExplanationImpl = new(ExplanationImpl)
	ans.spi = ans
	return ans
}

func newComplexExplanation(match bool, value float32, desc string) *ComplexExplanation {
	ans := new(ComplexExplanation)
	ans.ExplanationImpl = newExplanation(value, desc)
	ans.spi = ans
	ans.match = match
	return ans
}

/*
Indicates whether or not this Explanation models a good match.

If the match status is explicitly set (i.e.: not nil) this method
uses it; otherwise it defers to the superclass.
*/
func (e *ComplexExplanation) IsMatch() bool {
	return e.match != nil && e.match.(bool) || e.match == nil && e.ExplanationImpl.IsMatch()
}

func (e *ComplexExplanation) Summary() string {
	if e.match == nil {
		return e.ExplanationImpl.Summary()
	}
	return fmt.Sprintf("%v = %v %v", e.Value(),
		map[bool]string{true: "(MATCH)", false: "(NON_MATCH)"}[e.IsMatch()],
		e.Description())
}
