package lucene40

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
)

// codecs/lucene40/Lucene40LiveDocsFormat.java

/*
Lucene 4.0 Live Documents Format.

The .del file is optional, and only exists when a segment contains
deletions.

Although per-segment, this file is maintained exterior to compound
segment files.

Deletions (.del) --> Format,Heaer,ByteCount,BitCount, Bits | DGaps
  (depending on Format)
	Format,ByteSize,BitCount --> uint32
	Bits --> <byte>^ByteCount
	DGaps --> <DGap,NonOnesByte>^NonzeroBytesCount
	DGap --> vint
	NonOnesByte --> byte
	Header --> CodecHeader

Format is 1: indicates cleard DGaps.

ByteCount indicates the number of bytes in Bits. It is typically
(SegSize/8)+1.

BitCount indicates the number of bits that are currently set in Bits.

Bits contains one bit for each document indexed. When the bit
corresponding to a document number is cleared, that document is
marked as deleted. Bit ordering is from least to most significant.
Thus, if Bits contains two bytes, 0x00 and 0x02, then document 9 is
marked as alive (not deleted).

DGaps represents sparse bit-vectors more efficiently than Bits. It is
makde of DGaps on indexes of nonOnes bytes in Bits, and the nonOnes
bytes themselves. The number of nonOnes byte in Bits
(NonOnesBytesCount) is not stored.

For example, if there are 8000 bits and only bits 10,12,32 are
cleared, DGaps would be used:

(vint) 1, (byte) 20, (vint) 3, (byte) 1
*/
type Lucene40LiveDocsFormat struct {
}

/* Extension of deletes */
const DELETES_EXTENSION = "del"

func (format *Lucene40LiveDocsFormat) NewLiveDocs(size int) util.MutableBits {
	ans := NewBitVector(size)
	ans.InvertAll()
	return ans
}

func (format *Lucene40LiveDocsFormat) WriteLiveDocs(bits util.MutableBits,
	dir store.Directory, info *SegmentCommitInfo, newDelCount int,
	ctx store.IOContext) error {

	filename := util.FileNameFromGeneration(info.Info.Name, DELETES_EXTENSION, info.NextDelGen())
	liveDocs := bits.(*BitVector)
	assert(liveDocs.Count() == info.Info.DocCount()-info.DelCount()-newDelCount)
	assert(liveDocs.Length() == info.Info.DocCount())
	return liveDocs.Write(dir, filename, ctx)
}

func (format *Lucene40LiveDocsFormat) Files(info *SegmentCommitInfo) []string {
	if info.HasDeletions() {
		return []string{util.FileNameFromGeneration(info.Info.Name, DELETES_EXTENSION, info.DelGen())}
	}
	return []string{}
}
