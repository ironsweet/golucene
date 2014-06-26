package fst

import (
	"github.com/balzaczyy/golucene/core/util/packed"
)

/* Used to dedup states (lookup already-frozen states) */
type NodeHash struct {
	table *packed.PagedGrowableWriter
	mask  int64
	fst   *FST
	in    BytesReader
}

func newNodeHash(fst *FST, in BytesReader) *NodeHash {
	return &NodeHash{
		table: packed.NewPagedGrowableWriter(16, 1<<30, 8, packed.PackedInts.COMPACT),
		mask:  15,
		fst:   fst,
		in:    in,
	}
}

func (h *NodeHash) add(nodeIn *UnCompiledNode) (int64, error) {
	panic("not implemented yet")
}
