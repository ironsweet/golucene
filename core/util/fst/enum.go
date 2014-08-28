package fst

type FSTEnum struct {
	fst    *FST
	arcs   []*Arc
	output []interface{}

	NO_OUTPUT  interface{}
	fstReader  BytesReader
	scratchArc *Arc

	upto         int
	targetLength int
}

func newFSTEnum(fst *FST) *FSTEnum {
	arcs := make([]*Arc, 10)
	fst.FirstArc(arcs[0])
	outputs := make([]interface{}, 10)
	outputs[0] = fst.outputs.NoOutput()

	return &FSTEnum{
		fst:       fst,
		arcs:      arcs,
		output:    outputs,
		fstReader: fst.BytesReader(),
		NO_OUTPUT: outputs[0],
	}
}

func (e *FSTEnum) doNext() error {
	panic("not implemented yet")
}
