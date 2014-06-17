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
	panic("not implemented yet")
}

/* Expert: holds a pending (seen but not yet serialized) Node. */
type UnCompiledNode struct {
}
