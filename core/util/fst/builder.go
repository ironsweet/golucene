package fst

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

	// for packing
	doPackFST               bool
	acceptableOverheadRatio float32

	// current frontier
	frontier []*UnCompiledNode

	freezeTail FreezeTail
}

/* Expert: holds a pending (seen but not yet serialized) Node. */
type UnCompiledNode struct {
	owner  *Builder
	arcs   []*Arc
	output interface{}

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
		freezeTail:               freezeTail,
		doShareNonSingletonNodes: doShareNonSingletonNodes,
		shareMaxTailLength:       shareMaxTailLength,
		doPackFST:                doPackFST,
		acceptableOverheadRatio:  acceptableOverheadRatio,
		fst:       fst,
		NO_OUTPUT: outputs.NoOutput(),
		frontier:  f,
	}
	if doShareSuffix {
		ans.dedupHash = newNodeHash(fst, fst.bytes.reverseReaderAllowSingle(false))
	}
	for i, _ := range f {
		f[i] = newUnCompiledNode(ans, i)
	}
	return ans
}

/*
It's OK to add the same input twice in a row with different outputs,
as long as outputs impls the merge method. Note that input is fully
consumed after this method is returned (so caller is free to reuse),
but output is not. So if your outputs are changeable (eg
ByteSequenceOutputs or IntSequenceOutputs) then you cannot reuse
across calls.
*/
func (b *Builder) Add(input []int, output interface{}) error {
	panic("not implemented yet")
}
