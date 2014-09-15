package fst

import (
	// "fmt"
	"github.com/balzaczyy/golucene/core/util"
)

type FSTEnumSPI interface {
	setCurrentLabel(int)
	grow()
}

type FSTEnum struct {
	spi FSTEnumSPI

	fst    *FST
	arcs   []*Arc
	output []interface{}

	NO_OUTPUT  interface{}
	fstReader  BytesReader
	scratchArc *Arc

	upto         int
	targetLength int
}

func newFSTEnum(spi FSTEnumSPI, fst *FST) *FSTEnum {
	arcs := make([]*Arc, 10)
	arcs[0] = new(Arc)
	fst.FirstArc(arcs[0])
	outputs := make([]interface{}, 10)
	outputs[0] = fst.outputs.NoOutput()

	return &FSTEnum{
		spi:       spi,
		fst:       fst,
		arcs:      arcs,
		output:    outputs,
		fstReader: fst.BytesReader(),
		NO_OUTPUT: outputs[0],
	}
}

func (e *FSTEnum) doNext() (err error) {
	// fmt.Printf("FE: next upto=%v\n", e.upto)
	if e.upto == 0 {
		// fmt.Println("  init")
		e.upto = 1
		if _, err = e.fst.readFirstTargetArc(e.Arc(0), e.Arc(1), e.fstReader); err != nil {
			return
		}
	} else {
		// pop
		// fmt.Printf("  check pop curArc target=%v label=%v isLast?=",
		// e.arcs[e.upto].target, e.arcs[e.upto].Label, e.arcs[e.upto].isLast())
		for e.arcs[e.upto].isLast() {
			if e.upto--; e.upto == 0 {
				// fmt.Println("  eof")
				return nil
			}
		}
		if _, err = e.fst.readNextArc(e.arcs[e.upto], e.fstReader); err != nil {
			return
		}
	}
	return e.pushFirst()
}

func (e *FSTEnum) incr() {
	e.upto++
	e.spi.grow()
	if len(e.arcs) <= e.upto {
		newArcs := make([]*Arc, util.Oversize(e.upto+1, util.NUM_BYTES_OBJECT_REF))
		copy(newArcs, e.arcs)
		e.arcs = newArcs
	}
	if len(e.output) < e.upto {
		newOutput := make([]interface{}, util.Oversize(e.upto+1, util.NUM_BYTES_OBJECT_REF))
		copy(newOutput, e.output)
		e.output = newOutput
	}
}

/*
Appends current arc, and then recuses from its target, appending
first arc all the way to the final node.
*/
func (e *FSTEnum) pushFirst() (err error) {
	arc := e.arcs[e.upto]
	assert(arc != nil)

	for {
		e.output[e.upto] = e.fst.outputs.Add(e.output[e.upto-1], arc.Output)
		if arc.Label == FST_END_LABEL {
			// final node
			break
		}
		// fmt.Printf("  pushFirst label=%c upto=%v output=%v\n",
		// 	rune(arc.Label), e.upto, e.fst.outputs.outputToString(e.output[e.upto]))
		e.spi.setCurrentLabel(arc.Label)
		e.incr()

		nextArc := e.Arc(e.upto)
		if _, err = e.fst.readFirstTargetArc(arc, nextArc, e.fstReader); err != nil {
			return
		}
		arc = nextArc
	}
	return nil
}

func (e *FSTEnum) Arc(idx int) *Arc {
	if e.arcs[idx] == nil {
		e.arcs[idx] = new(Arc)
	}
	return e.arcs[idx]
}
