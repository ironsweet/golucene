package fst

import (
	// "fmt"
	"github.com/balzaczyy/golucene/core/util/packed"
)

/* Used to dedup states (lookup already-frozen states) */
type NodeHash struct {
	table      *packed.PagedGrowableWriter
	count      int64
	mask       int64
	fst        *FST
	scratchArc *Arc
	in         BytesReader
}

func newNodeHash(fst *FST, in BytesReader) *NodeHash {
	return &NodeHash{
		table:      packed.NewPagedGrowableWriter(16, 1<<27, 8, packed.PackedInts.COMPACT),
		mask:       15,
		fst:        fst,
		scratchArc: new(Arc),
		in:         in,
	}
}

func (nh *NodeHash) nodesEqual(node *UnCompiledNode, address int64) (bool, error) {
	_, err := nh.fst.readFirstRealTargetArc(address, nh.scratchArc, nh.in)
	if err != nil {
		return false, err
	}
	if nh.scratchArc.bytesPerArc != 0 && node.NumArcs != nh.scratchArc.numArcs {
		return false, nil
	}
	for arcUpto := 0; arcUpto < node.NumArcs; arcUpto++ {
		if arc := node.Arcs[arcUpto]; arc.label != nh.scratchArc.Label ||
			arc.output != nh.scratchArc.Output ||
			arc.Target.(*CompiledNode).node != nh.scratchArc.target ||
			arc.nextFinalOutput != nh.scratchArc.NextFinalOutput ||
			arc.isFinal != nh.scratchArc.IsFinal() {
			return false, nil
		}

		if nh.scratchArc.isLast() {
			return arcUpto == node.NumArcs-1, nil
		}
		if _, err = nh.fst.readNextRealArc(nh.scratchArc, nh.in); err != nil {
			return false, err
		}
	}
	return false, err
}

const PRIME = 31

/* hash code for an unfrozen node. This must be identical to the frozen case (below) !! */
func (nh *NodeHash) hash(node *UnCompiledNode) int64 {
	// fmt.Println("hash unfrozen")
	h := int64(0)
	for arcIdx := 0; arcIdx < node.NumArcs; arcIdx++ {
		arc := node.Arcs[arcIdx]
		// fmt.Printf("  label=%v target=%v h=%v output=%v isFinal?=%v\n",
		// 	arc.label, arc.Target.(*CompiledNode).node, h,
		// 	nh.fst.outputs.outputToString(arc.output), arc.isFinal)
		h = PRIME*h + int64(arc.label)
		n := arc.Target.(*CompiledNode).node
		h = PRIME*h + int64(n^(n>>32))
		h = PRIME*h + hashPtr(arc.output)
		h = PRIME*h + hashPtr(arc.nextFinalOutput)
		if arc.isFinal {
			h += 17
		}
	}
	// fmt.Printf("  ret %v\n", int32(h))
	return h
}

/* hash code for a frozen node */
func (nh *NodeHash) hashFrozen(node int64) (int64, error) {
	// fmt.Printf("hash frozen node=%v\n", node)
	h := int64(0)
	_, err := nh.fst.readFirstRealTargetArc(node, nh.scratchArc, nh.in)
	if err != nil {
		return 0, err
	}
	for {
		// fmt.Printf("  label=%v target=%v h=%v output=%v next?=%v final?=%v pos=%v\n",
		// 	nh.scratchArc.Label, nh.scratchArc.target, h,
		// 	nh.fst.outputs.outputToString(nh.scratchArc.Output),
		// 	nh.scratchArc.flag(4), nh.scratchArc.IsFinal(), nh.in.getPosition())
		h = PRIME*h + int64(nh.scratchArc.Label)
		h = PRIME*h + int64(nh.scratchArc.target^(nh.scratchArc.target>>32))
		h = PRIME*h + hashPtr(nh.scratchArc.Output)
		h = PRIME*h + hashPtr(nh.scratchArc.NextFinalOutput)
		if nh.scratchArc.IsFinal() {
			h += 17
		}
		if nh.scratchArc.isLast() {
			break
		}
		if _, err = nh.fst.readNextRealArc(nh.scratchArc, nh.in); err != nil {
			return 0, err
		}
	}
	// fmt.Printf("  ret %v\n", int32(h))
	return h, nil
}

func hashPtr(obj interface{}) (h int64) {
	if obj != nil && obj != NO_OUTPUT {
		for _, b := range obj.([]byte) {
			h = PRIME*h + int64(b)
		}
	}
	return
}

func (nh *NodeHash) add(nodeIn *UnCompiledNode) (int64, error) {
	// fmt.Printf("hash: add count=%v vs %v mask=%v\n", nh.count, nh.table.Size(), nh.mask)
	h := nh.hash(nodeIn)
	pos := h & nh.mask
	c := int64(0)
	for {
		v := nh.table.Get(pos)
		if v == 0 {
			// freeze & add
			node, err := nh.fst.addNode(nodeIn)
			if err != nil {
				return 0, err
			}
			// fmt.Printf("  now freeze node=%v\n", node)
			h2, err := nh.hashFrozen(node)
			if err != nil {
				return 0, err
			}
			assert2(h2 == h, "frozenHash=%v vs h=%v", h2, h)
			nh.count++
			nh.table.Set(pos, node)
			// rehash at 2/3 occupancy:
			if nh.count > 2*nh.table.Size()/3 {
				panic("not implemented yet")
			}
			return node, nil
		} else {
			ok, err := nh.nodesEqual(nodeIn, v)
			if err != nil {
				return 0, err
			}
			if ok {
				// same node is already here
				return v, nil
			}
		}

		// quadratic probe
		c++
		pos = (pos + c) & nh.mask
	}
}
