package fst

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

/* Expert: invoked by Builder whenever a suffix is serialized. */
type FreezeTail func([]*UnCompiledNode, int, []int) error

/*
Builds a minimal FST (maps an []int term to an arbitrary output) from
pre-sorted terms with outputs. The FST becomes an FSA if you use
NoOutputs. The FST is written on-the-fly into a compact serialized
format byte array, which can be saved to / loaded from a Directory or
used directly for traversal. The FST is always finite (no cycles).

NOTE: the algorithm is described at
http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698

FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs
containing more than 2.1B nodes are also now possible, however they
cannot be packed.
*/
type Builder struct {
	dedupHash *NodeHash
	fst       *FST
	NO_OUTPUT interface{}

	// simplistic pruning: we prune node (and all following nodes) if
	// less than this number of terms go through it:
	minSuffixCount1 int

	// better pruning: we prune node (and all following nodes) if the
	// prior node has less than this number of terms go through it:
	minSuffixCount2 int

	doShareNonSingletonNodes bool
	shareMaxTailLength       int

	lastInput *util.IntsRef

	// for packing
	doPackFST               bool
	acceptableOverheadRatio float32

	// current frontier
	frontier []*UnCompiledNode

	_freezeTail FreezeTail
}

/*
NOTE: not many instances of Node or CompiledNode are in memory while
the FST is being built; it's only the current "frontier":
*/
type Node interface {
	isCompiled() bool
}

/* Expert: holds a pending (seen but not yet serialized) Node. */
type UnCompiledNode struct {
	owner      *Builder
	arcs       []*Arc
	output     interface{}
	isFinal    bool
	inputCount int64

	// This node's depth, starting from the automaton root.
	depth int
}

func newUnCompiledNode(owner *Builder, depth int) *UnCompiledNode {
	return &UnCompiledNode{
		owner:  owner,
		arcs:   []*Arc{new(Arc)},
		output: owner.NO_OUTPUT,
		depth:  depth,
	}
}

func (n *UnCompiledNode) isCompiled() bool { return false }

func (n *UnCompiledNode) lastOutput(labelToMatch int) interface{} {
	panic("not implementd yet")
}

func (n *UnCompiledNode) addArc(label int, target Node) {
	panic("not implemented yet")
}

/*
Instantiates an FST/FSA builder with all the possible tuning and
construction tweaks. Read parameter documentation carefully.

...
*/
func NewBuilder(inputType InputType, minSuffixCount1, minSuffixCount2 int,
	doShareSuffix, doShareNonSingletonNodes bool, shareMaxTailLength int,
	outputs Outputs, freezeTail FreezeTail, doPackFST bool,
	acceptableOverheadRatio float32, allowArrayArcs bool, bytesPageBits int) *Builder {

	fst := newFST(inputType, outputs, doPackFST, acceptableOverheadRatio, allowArrayArcs, bytesPageBits)
	f := make([]*UnCompiledNode, 10)
	ans := &Builder{
		minSuffixCount1:          minSuffixCount1,
		minSuffixCount2:          minSuffixCount2,
		_freezeTail:              freezeTail,
		doShareNonSingletonNodes: doShareNonSingletonNodes,
		shareMaxTailLength:       shareMaxTailLength,
		doPackFST:                doPackFST,
		acceptableOverheadRatio:  acceptableOverheadRatio,
		fst:       fst,
		NO_OUTPUT: outputs.NoOutput(),
		frontier:  f,
		lastInput: util.NewEmptyIntsRef(),
	}
	if doShareSuffix {
		ans.dedupHash = newNodeHash(fst, fst.bytes.reverseReaderAllowSingle(false))
	}
	for i, _ := range f {
		f[i] = newUnCompiledNode(ans, i)
	}
	return ans
}

func (b *Builder) freezeTail(prefixLenPlus1 int) error {
	if b._freezeTail != nil {
		// Custom plugin:
		return b._freezeTail(b.frontier, prefixLenPlus1, b.lastInput.Value())
	}
	panic("not implemented yet")
}

/*
It's OK to add the same input twice in a row with different outputs,
as long as outputs impls the merge method. Note that input is fully
consumed after this method is returned (so caller is free to reuse),
but output is not. So if your outputs are changeable (eg
ByteSequenceOutputs or IntSequenceOutputs) then you cannot reuse
across calls.
*/
func (b *Builder) Add(input *util.IntsRef, output interface{}) error {
	assert2(b.lastInput.Length == 0 || !input.CompareTo(b.lastInput),
		"inputs are added out of order, lastInput=%v vs input=%v",
		b.lastInput, input)

	if input.Length == 0 {
		// empty input: only allowed as first input. We have to special
		// case this becaues the packed FST format cannot represent the
		// empty input since 'finalness' is stored on the incoming arc,
		// not on the node
		b.frontier[0].inputCount++
		b.frontier[0].isFinal = true
		b.fst.setEmptyOutput(output)
		return nil
	}

	// compare shared prefix length
	pos1 := 0
	pos2 := input.Offset
	pos1Stop := b.lastInput.Length
	if input.Length < pos1Stop {
		pos1Stop = input.Length
	}
	for {
		b.frontier[pos1].inputCount++
		if pos1 >= pos1Stop || b.lastInput.Ints[pos1] != input.Ints[pos2] {
			break
		}
		pos1++
		pos2++
	}
	prefixLenPlus1 := pos1 + 1

	if len(b.frontier) < input.Length+1 {
		next := make([]*UnCompiledNode, util.Oversize(input.Length+1, util.NUM_BYTES_OBJECT_REF))
		copy(next, b.frontier)
		for idx := len(b.frontier); idx < len(next); idx++ {
			next[idx] = newUnCompiledNode(b, idx)
		}
		b.frontier = next
	}

	// minimize/compile states from previous input's orphan'd suffix
	err := b.freezeTail(prefixLenPlus1)
	if err != nil {
		return err
	}

	// init tail states for current input
	for idx := prefixLenPlus1; idx <= input.Length; idx++ {
		b.frontier[idx-1].addArc(input.Ints[input.Offset+idx-1], b.frontier[idx])
		b.frontier[idx].inputCount++
	}

	lastNode := b.frontier[input.Length]
	if b.lastInput.Length != input.Length || prefixLenPlus1 != input.Length+1 {
		lastNode.isFinal = true
		lastNode.output = b.NO_OUTPUT
	}

	// push conflicting outputs forward, only as far as needed
	for idx := 1; idx < prefixLenPlus1; idx++ {
		// node := b.frontier[idx]
		parentNode := b.frontier[idx-1]

		lastOutput := parentNode.lastOutput(input.Ints[input.Offset+idx-1])

		var commonOutputPrefix interface{}
		// var wordSuffix interface{}

		if lastOutput != b.NO_OUTPUT {
			panic("not implemented yet")
		} else {
			panic("not implemented yet")
		}

		output = b.fst.outputs.Subtract(output, commonOutputPrefix)
	}

	if b.lastInput.Length == input.Length && prefixLenPlus1 == 1+input.Length {
		// same input more than 1 time in a row, mapping to multiple outputs
		panic("not implemented yet")
	} else {
		panic("not implemented yet")
	}

	// save last input
	b.lastInput.CopyInts(input)
	return nil
}

func assert(ok bool) {
	assert2(ok, "assert fail")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}
