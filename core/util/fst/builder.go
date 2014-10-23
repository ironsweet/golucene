package fst

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/util"
)

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

	lastInput *util.IntsRefBuilder

	// for packing
	doPackFST               bool
	acceptableOverheadRatio float32

	// current frontier
	frontier []*UnCompiledNode
}

/*
Instantiates an FST/FSA builder with all the possible tuning and
construction tweaks. Read parameter documentation carefully.

...
*/
func NewBuilder(inputType InputType, minSuffixCount1, minSuffixCount2 int,
	doShareSuffix, doShareNonSingletonNodes bool, shareMaxTailLength int,
	outputs Outputs, doPackFST bool,
	acceptableOverheadRatio float32, allowArrayArcs bool, bytesPageBits int) *Builder {

	fst := newFST(inputType, outputs, doPackFST, acceptableOverheadRatio, allowArrayArcs, bytesPageBits)
	f := make([]*UnCompiledNode, 10)
	ans := &Builder{
		minSuffixCount1:          minSuffixCount1,
		minSuffixCount2:          minSuffixCount2,
		doShareNonSingletonNodes: doShareNonSingletonNodes,
		shareMaxTailLength:       shareMaxTailLength,
		doPackFST:                doPackFST,
		acceptableOverheadRatio:  acceptableOverheadRatio,
		fst:       fst,
		NO_OUTPUT: outputs.NoOutput(),
		frontier:  f,
		lastInput: util.NewIntsRefBuilder(),
	}
	if doShareSuffix {
		ans.dedupHash = newNodeHash(fst, fst.bytes.reverseReaderAllowSingle(false))
	}
	for i, _ := range f {
		f[i] = NewUnCompiledNode(ans, i)
	}
	return ans
}

func (b *Builder) compileNode(nodeIn *UnCompiledNode, tailLength int) (*CompiledNode, error) {
	var node int64
	var err error
	if b.dedupHash != nil &&
		(b.doShareNonSingletonNodes || nodeIn.NumArcs <= 1) &&
		tailLength <= b.shareMaxTailLength {
		if nodeIn.NumArcs == 0 {
			node, err = b.fst.addNode(nodeIn)
		} else {
			node, err = b.dedupHash.add(nodeIn)
		}
	} else {
		node, err = b.fst.addNode(nodeIn)
	}
	if err != nil {
		return nil, err
	}
	assert(node != -2)

	nodeIn.Clear()

	return &CompiledNode{node}, nil
}

func (b *Builder) freezeTail(prefixLenPlus1 int) error {
	// fmt.Printf("  compileTail %v\n", prefixLenPlus1)
	downTo := prefixLenPlus1
	if downTo < 1 {
		downTo = 1
	}
	for idx := b.lastInput.Length(); idx >= downTo; idx-- {
		doPrune := false
		doCompile := false

		node := b.frontier[idx]
		parent := b.frontier[idx-1]

		if node.InputCount < int64(b.minSuffixCount1) {
			doPrune = true
			doCompile = true
		} else if idx > prefixLenPlus1 {
			// prune if parent's inputCount is less than suffixMinCount2
			if parent.InputCount < int64(b.minSuffixCount2) ||
				b.minSuffixCount2 == 1 && parent.InputCount == 1 && idx > 1 {
				// my parent, about to be compiled, doesn't make the cut, so
				// I'm definitely pruned

				// if minSuffixCount2 is 1, we keep only up
				// until the 'distinguished edge', ie we keep only the
				// 'divergent' part of the FST. if my parent, about to be
				// compiled, has inputCount 1 then we are already past the
				// distinguished edge.  NOTE: this only works if
				// the FST outputs are not "compressible" (simple
				// ords ARE compressible).
				doPrune = true
			} else {
				// my parent, about to be compiled, does make the cut, so
				// I'm definitely not pruned
				doPrune = false
			}
			doCompile = true
		} else {
			// if pruning is disabled (count is 0) we can always compile current node
			doCompile = b.minSuffixCount2 == 0
		}

		// fmt.Printf("    label=%c idx=%v inputCount=%v doCompile=%v doPrune=%v\n",
		// 	b.lastInput.At(idx-1), idx, b.frontier[idx].InputCount, doCompile, doPrune)
		if node.InputCount < int64(b.minSuffixCount2) ||
			(b.minSuffixCount2 == 1 && node.InputCount == 1 && idx > 1) {
			// drop all arcs
			panic("not implemented yet")
		}

		if doPrune {
			// tihs node doesn't make it -- deref it
			node.Clear()
			parent.deleteLast(b.lastInput.At(idx-1), node)
		} else {

			if b.minSuffixCount2 != 0 {
				b.compileAllTargets(node, b.lastInput.Length()-idx)
			}
			nextFinalOutput := node.output

			// we "fake" the node as being final if it has no outgoing arcs;
			// in theory we could leave it as non-final (the FST can
			// represent this), but FSTEnum, Util, etc., have trouble w/
			// non-final dead-end states:
			isFinal := node.IsFinal || node.NumArcs == 0

			if doCompile {
				// this node makes it and we now compile it. first, compile
				// any targets that were previously undecided:
				label := b.lastInput.At(idx - 1)
				node, err := b.compileNode(node, 1+b.lastInput.Length()-idx)
				if err != nil {
					return err
				}
				parent.replaceLast(label, node, nextFinalOutput, isFinal)
			} else {
				panic("not implemented yet")
			}
		}
	}
	return nil
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
	// { // debug
	// 	bytes := make([]byte, input.Length)
	// 	for i, _ := range bytes {
	// 		bytes[i] = byte(input.Ints[i])
	// 	}
	// 	if output == NO_OUTPUT {
	// 		fmt.Printf("\nFST ADD: input=%v %v\n", string(bytes), bytes)
	// 	} else {
	// 		panic("not implemented yet")
	// 		// fmt.Printf("\nFST ADD: input=%v %v output=%v", string(bytes), bytes, b.fst.outputs.outputToString(output)));
	// 	}
	// }

	// de-dup NO_OUTPUT since it must be a singleton:
	if output == NO_OUTPUT {
		output = NO_OUTPUT
	}

	assert2(b.lastInput.Length() == 0 || !input.Less(b.lastInput.Get()),
		"inputs are added out of order, lastInput=%v vs input=%v",
		b.lastInput.Get(), input)

	if input.Length == 0 {
		// empty input: only allowed as first input. We have to special
		// case this becaues the packed FST format cannot represent the
		// empty input since 'finalness' is stored on the incoming arc,
		// not on the node
		b.frontier[0].InputCount++
		b.frontier[0].IsFinal = true
		b.fst.setEmptyOutput(output)
		return nil
	}

	// compare shared prefix length
	pos1 := 0
	pos2 := input.Offset
	pos1Stop := b.lastInput.Length()
	if input.Length < pos1Stop {
		pos1Stop = input.Length
	}
	for {
		b.frontier[pos1].InputCount++
		if pos1 >= pos1Stop || b.lastInput.At(pos1) != input.Ints[pos2] {
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
			next[idx] = NewUnCompiledNode(b, idx)
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
		b.frontier[idx].InputCount++
	}

	lastNode := b.frontier[input.Length]
	if b.lastInput.Length() != input.Length || prefixLenPlus1 != input.Length+1 {
		lastNode.IsFinal = true
		lastNode.output = b.NO_OUTPUT
	}

	// push conflicting outputs forward, only as far as needed
	for idx := 1; idx < prefixLenPlus1; idx++ {
		node := b.frontier[idx]
		parentNode := b.frontier[idx-1]

		lastOutput := parentNode.lastOutput(input.Ints[input.Offset+idx-1])

		var commonOutputPrefix interface{}
		var wordSuffix interface{}

		if lastOutput != b.NO_OUTPUT {
			commonOutputPrefix = b.fst.outputs.Common(output, lastOutput)
			wordSuffix = b.fst.outputs.Subtract(lastOutput, commonOutputPrefix)
			parentNode.setLastOutput(input.Ints[input.Offset+idx-1], commonOutputPrefix)
			node.prependOutput(wordSuffix)
		} else {
			commonOutputPrefix = NO_OUTPUT
		}

		output = b.fst.outputs.Subtract(output, commonOutputPrefix)
	}

	if b.lastInput.Length() == input.Length && prefixLenPlus1 == 1+input.Length {
		// same input more than 1 time in a row, mapping to multiple outputs
		panic("not implemented yet")
	} else {
		// this new arc is private to this new input; set its arc output
		// to the leftover output:
		b.frontier[prefixLenPlus1-1].setLastOutput(input.At(prefixLenPlus1-1), output)
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

/*
Returns final FST. NOTE: this will return nil if nothing is accepted
by the FST.
*/
func (b *Builder) Finish() (*FST, error) {
	root := b.frontier[0]

	// minimize nodes in the last word's suffix
	err := b.freezeTail(0)
	if err != nil {
		return nil, err
	}
	if root.InputCount < int64(b.minSuffixCount1) ||
		root.InputCount < int64(b.minSuffixCount2) || root.NumArcs == 0 {
		if b.fst.emptyOutput == nil {
			return nil, nil
		} else if b.minSuffixCount1 > 0 || b.minSuffixCount2 > 0 {
			// emtpy string got pruned
			return nil, nil
		}
	} else {
		if b.minSuffixCount2 != 0 {
			err = b.compileAllTargets(root, b.lastInput.Length())
			if err != nil {
				return nil, err
			}
		}
	}
	d, err := b.compileNode(root, b.lastInput.Length())
	if err != nil {
		return nil, err
	}
	err = b.fst.finish(d.node)
	if err != nil {
		return nil, err
	}

	if b.doPackFST {
		n := b.fst.NodeCount() / 4
		if n < 10 {
			n = 10
		}
		return b.fst.pack(3, int(n), b.acceptableOverheadRatio)
	}
	return b.fst, nil
}

func (b *Builder) compileAllTargets(node *UnCompiledNode, tailLength int) error {
	panic("not implemented yet")
}

/* Expert: holds a pending (seen but not yet serialized) arc */
type builderArc struct {
	label           int // really an "unsigned" byte
	Target          Node
	isFinal         bool
	output          interface{}
	nextFinalOutput interface{}
}

/*
NOTE: not many instances of Node or CompiledNode are in memory while
the FST is being built; it's only the current "frontier":
*/
type Node interface {
	isCompiled() bool
}

type CompiledNode struct {
	node int64
}

func (n *CompiledNode) isCompiled() bool { return true }

/* Expert: holds a pending (seen but not yet serialized) Node. */
type UnCompiledNode struct {
	owner      *Builder
	NumArcs    int
	Arcs       []*builderArc
	output     interface{}
	IsFinal    bool
	InputCount int64

	// This node's depth, starting from the automaton root.
	depth int
}

func NewUnCompiledNode(owner *Builder, depth int) *UnCompiledNode {
	return &UnCompiledNode{
		owner:  owner,
		Arcs:   []*builderArc{new(builderArc)},
		output: owner.NO_OUTPUT,
		depth:  depth,
	}
}

func (n *UnCompiledNode) isCompiled() bool { return false }

func (n *UnCompiledNode) Clear() {
	n.NumArcs = 0
	n.IsFinal = false
	n.output = n.owner.NO_OUTPUT
	n.InputCount = 0

	// we don't clear the depth here becaues it never changes
	// for nodes on the frontier (even when reused).
}

func (n *UnCompiledNode) lastOutput(labelToMatch int) interface{} {
	assert(n.NumArcs > 0)
	assert(n.Arcs[n.NumArcs-1].label == labelToMatch)
	return n.Arcs[n.NumArcs-1].output
}

func (n *UnCompiledNode) addArc(label int, target Node) {
	assert(label >= 0)
	if n.NumArcs != 0 {
		assert2(label > n.Arcs[n.NumArcs-1].label,
			"arc[-1].label=%v new label=%v numArcs=%v",
			n.Arcs[n.NumArcs-1].label, label, n.NumArcs)
	}
	if n.NumArcs == len(n.Arcs) {
		newArcs := make([]*builderArc, util.Oversize(n.NumArcs+1, util.NUM_BYTES_OBJECT_REF))
		copy(newArcs, n.Arcs)
		for arcIdx := n.NumArcs; arcIdx < len(newArcs); arcIdx++ {
			newArcs[arcIdx] = new(builderArc)
		}
		n.Arcs = newArcs
	}
	arc := n.Arcs[n.NumArcs]
	n.NumArcs++
	arc.label = label
	arc.Target = target
	arc.output = n.owner.NO_OUTPUT
	arc.nextFinalOutput = n.owner.NO_OUTPUT
	arc.isFinal = false
}

func (n *UnCompiledNode) replaceLast(labelToMatch int, target Node, nextFinalOutput interface{}, isFinal bool) {
	assert(n.NumArcs > 0)
	arc := n.Arcs[n.NumArcs-1]
	assert2(arc.label == labelToMatch, "arc.label=%v vs %v", arc.label, labelToMatch)
	arc.Target = target
	arc.nextFinalOutput = nextFinalOutput
	arc.isFinal = isFinal
}

func (n *UnCompiledNode) deleteLast(label int, target Node) {
	assert(n.NumArcs > 0)
	assert(label == n.Arcs[n.NumArcs-1].label)
	assert(target == n.Arcs[n.NumArcs-1].Target)
	n.NumArcs--
}

func (n *UnCompiledNode) setLastOutput(labelToMatch int, newOutput interface{}) {
	assert(n.NumArcs > 0)
	arc := n.Arcs[n.NumArcs-1]
	assert(arc.label == labelToMatch)
	arc.output = newOutput
}

/* pushes an output prefix forward onto all arcs */
func (n *UnCompiledNode) prependOutput(outputPrefix interface{}) {
	for arcIdx := 0; arcIdx < n.NumArcs; arcIdx++ {
		n.Arcs[arcIdx].output = n.owner.fst.outputs.Add(outputPrefix, n.Arcs[arcIdx].output)
	}

	if n.IsFinal {
		n.output = n.owner.fst.outputs.Add(outputPrefix, n.output)
	}
}
