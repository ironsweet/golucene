package blocktree

import (
	"bytes"
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/fst"
)

type segmentTermsEnumFrame struct {
	// Our index in stack[]:
	ord int

	hasTerms     bool
	hasTermsOrig bool
	isFloor      bool

	arc *fst.Arc

	// File pointer where this block was loaded from
	fp     int64
	fpOrig int64
	fpEnd  int64

	suffixBytes    []byte
	suffixesReader store.ByteArrayDataInput

	statBytes   []byte
	statsReader store.ByteArrayDataInput

	floorData       []byte
	floorDataReader store.ByteArrayDataInput

	// Length of prefix shared by all terms in this block
	prefix int

	// Number of entries (term or sub-block) in this block
	entCount int

	// Which term we will next read, or -1 if the block
	// isn't loaded yet
	nextEnt int

	// True if this block is either not a floor block,
	// or, it's the last sub-block of a floor block
	isLastInFloor bool

	// True if all entries are terms
	isLeafBlock bool

	lastSubFP int64

	nextFloorLabel       int
	numFollowFloorBlocks int

	// Next term to decode metaData; we decode metaData
	// lazily so that scanning to find the matching term is
	// fast and only if you find a match and app wants the
	// stats or docs/positions enums, will we decode the
	// metaData
	metaDataUpto int

	state *BlockTermState

	// metadata buffer, holding monotonic values
	longs []int64
	// metadata buffer, holding general values
	bytes       []byte
	bytesReader *store.ByteArrayDataInput

	ste *SegmentTermsEnum

	startBytePos int
	suffix       int
	subCode      int64
}

func newFrame(ste *SegmentTermsEnum, ord int) *segmentTermsEnumFrame {
	f := &segmentTermsEnumFrame{
		suffixBytes: make([]byte, 128),
		statBytes:   make([]byte, 64),
		floorData:   make([]byte, 32),
		ste:         ste,
		ord:         ord,
		longs:       make([]int64, ste.fr.longsSize),
	}
	f.state = ste.fr.parent.postingsReader.NewTermState()
	f.state.TotalTermFreq = -1
	return f
}

func (f *segmentTermsEnumFrame) setFloorData(in *store.ByteArrayDataInput, source []byte) {
	numBytes := len(source) - (in.Pos - 0)
	if numBytes > len(f.floorData) {
		// TODO over allocate
		f.floorData = make([]byte, numBytes)
	}
	copy(f.floorData, source[in.Pos:])
	f.floorDataReader.Reset(f.floorData)
	f.numFollowFloorBlocks, _ = asInt(f.floorDataReader.ReadVInt())
	b, _ := f.floorDataReader.ReadByte()
	f.nextFloorLabel = int(b)
	// fmt.Printf("    setFloorData fpOrig=%v bytes=%v numFollowFloorBlocks=%v nextFloorLabel=%x\n",
	// 	f.fpOrig, source[in.Pos:], f.numFollowFloorBlocks, f.nextFloorLabel)
}

func (f *segmentTermsEnumFrame) getTermBlockOrd() int {
	if f.isLeafBlock {
		return f.nextEnt
	} else {
		return f.state.TermBlockOrd
	}
}

/* Does initial decode of next block of terms; this
   doesn't actually decode the docFreq, totalTermFreq,
   postings details (frq/prx offset, etc.) metadata;
   it just loads them as byte[] blobs which are then
   decoded on-demand if the metadata is ever requested
   for any term in this block.  This enables terms-only
   intensive consumes (eg certain MTQs, respelling) to
   not pay the price of decoding metadata they won't
   use. */
func (f *segmentTermsEnumFrame) loadBlock() (err error) {
	// Clone the IndexInput lazily, so that consumers
	// that just pull a TermsEnum to
	// seekExact(TermState) don't pay this cost:
	f.ste.initIndexInput()

	if f.nextEnt != -1 {
		// Already loaded
		return
	}

	f.ste.in.Seek(f.fp)
	code, err := asInt(f.ste.in.ReadVInt())
	if err != nil {
		return err
	}
	f.entCount = int(uint(code) >> 1)
	assert(f.entCount > 0)
	f.isLastInFloor = (code & 1) != 0

	assert2(f.arc == nil || !f.isLastInFloor || !f.isFloor,
		"fp=%v arc=%v isFloor=%v isLastInFloor=%v",
		f.fp, f.arc, f.isFloor, f.isLastInFloor)

	// TODO: if suffixes were stored in random-access
	// array structure, then we could do binary search
	// instead of linear scan to find target term; eg
	// we could have simple array of offsets

	// term suffixes:
	code, err = asInt(f.ste.in.ReadVInt())
	if err != nil {
		return err
	}
	f.isLeafBlock = (code & 1) != 0
	numBytes := int(uint(code) >> 1)
	if len(f.suffixBytes) < numBytes {
		f.suffixBytes = make([]byte, numBytes)
	}
	err = f.ste.in.ReadBytes(f.suffixBytes[:numBytes])
	if err != nil {
		return err
	}
	f.suffixesReader.Reset(f.suffixBytes)

	// if f.arc == nil {
	// 	fmt.Printf("    loadBlock (next) fp=%v entCount=%v prefixLen=%v isLastInFloor=%v leaf?=%v\n",
	// 		f.fp, f.entCount, f.prefix, f.isLastInFloor, f.isLeafBlock)
	// } else {
	// 	fmt.Printf("    loadBlock (seek) fp=%v entCount=%v prefixLen=%v hasTerms?=%v isFloor?=%v isLastInFloor=%v leaf?=%v\n",
	// 		f.fp, f.entCount, f.prefix, f.hasTerms, f.isFloor, f.isLastInFloor, f.isLeafBlock)
	// }

	// stats
	numBytes, err = asInt(f.ste.in.ReadVInt())
	if err != nil {
		return err
	}
	if len(f.statBytes) < numBytes {
		f.statBytes = make([]byte, numBytes)
	}
	err = f.ste.in.ReadBytes(f.statBytes[:numBytes])
	if err != nil {
		return err
	}
	f.statsReader.Reset(f.statBytes)
	f.metaDataUpto = 0

	f.state.TermBlockOrd = 0
	f.nextEnt = 0
	f.lastSubFP = -1

	// TODO: we could skip this if !hasTerms; but
	// that's rare so won't help much
	// metadata
	if numBytes, err = asInt(f.ste.in.ReadVInt()); err != nil {
		return err
	}
	if f.bytes == nil {
		f.bytes = make([]byte, util.Oversize(numBytes, 1))
		f.bytesReader = store.NewEmptyByteArrayDataInput()
	} else if len(f.bytes) < numBytes {
		f.bytes = make([]byte, util.Oversize(numBytes, 1))
	}
	if err = f.ste.in.ReadBytes(f.bytes[:numBytes]); err != nil {
		return err
	}
	f.bytesReader.Reset(f.bytes)

	// Sub-blocks of a single floor block are always
	// written one after another -- tail recurse:
	f.fpEnd = f.ste.in.FilePointer()
	// fmt.Printf("      fpEnd=%v\n", f.fpEnd)
	return nil
}

func (f *segmentTermsEnumFrame) rewind() {
	// Force reload:
	f.fp = f.fpOrig
	f.nextEnt = -1
	f.hasTerms = f.hasTermsOrig
	if f.isFloor {
		f.floorDataReader.Rewind()
		f.numFollowFloorBlocks, _ = asInt(f.floorDataReader.ReadVInt())
		assert(f.numFollowFloorBlocks > 0)
		b, _ := f.floorDataReader.ReadByte()
		f.nextFloorLabel = int(b)
	}
}

func (f *segmentTermsEnumFrame) next() bool {
	if f.isLeafBlock {
		return f.nextLeaf()
	}
	return f.nextNonLeaf()
}

// Decodes next entry; returns true if it's a sub-block
func (f *segmentTermsEnumFrame) nextLeaf() bool {
	panic("not implemented yet")
}

func (f *segmentTermsEnumFrame) nextNonLeaf() bool {
	panic("not implemented yet")
}

// TODO: make this array'd so we can do bin search?
// likely not worth it?  need to measure how many
// floor blocks we "typically" get
func (f *segmentTermsEnumFrame) scanToFloorFrame(target []byte) {
	if !f.isFloor || len(target) <= f.prefix {
		// fmt.Printf("    scanToFloorFrame skip: isFloor=%v target.length=%v vs prefix=%v\n",
		// 	f.isFloor, len(target), f.prefix)
		return
	}

	targetLabel := int(target[f.prefix])
	fmt.Printf("    scanToFloorFrame fpOrig=%v targetLabel=%x vs nextFloorLabel=%x numFollowFloorBlocks=%v\n",
		f.fpOrig, targetLabel, f.nextFloorLabel, f.numFollowFloorBlocks)
	if targetLabel < f.nextFloorLabel {
		fmt.Println("      already on correct block")
		return
	}

	assert(f.numFollowFloorBlocks != 0)

	var newFP int64
	for {
		code, _ := f.floorDataReader.ReadVLong() // ignore error
		newFP = f.fpOrig + int64(uint64(code)>>1)
		f.hasTerms = (code & 1) != 0
		// fmt.Printf("      label=%x fp=%v hasTerms?=%v numFollorFloor=%v\n",
		// f.nextFloorLabel, newFP, f.hasTerms, f.numFollowFloorBlocks)

		f.isLastInFloor = f.numFollowFloorBlocks == 1
		f.numFollowFloorBlocks--

		if f.isLastInFloor {
			f.nextFloorLabel = 256
			fmt.Printf("        stop!  last block nextFloorLabel=%x\n", f.nextFloorLabel)
			break
		} else {
			panic("niy")
		}
	}

	if newFP != f.fp {
		// Force re-load of the block:
		fmt.Printf("      force switch to fp=%v oldFP=%v\n", newFP, f.fp)
		f.nextEnt = -1
		f.fp = newFP
	} else {
		//
	}
}

func (f *segmentTermsEnumFrame) decodeMetaData() (err error) {
	// fmt.Printf("BTTR.decodeMetadata seg=%v mdUpto=%v vs termBlockOrd=%v\n",
	// 	f.ste.fr.parent.segment, f.metaDataUpto, f.state.TermBlockOrd)

	// lazily catch up on metadata decode:
	limit := f.getTermBlockOrd()
	absolute := f.metaDataUpto == 0
	assert(limit > 0)

	// TODO: better API would be "jump straight to term=N"???
	for f.metaDataUpto < limit {
		// TODO: we could make "tiers" of metadata, ie,
		// decode docFreq/totalTF but don't decode postings
		// metadata; this way caller could get
		// docFreq/totalTF w/o paying decode cost for
		// postings

		// TODO: if docFreq were bulk decoded we could
		// just skipN here:

		// stats
		if f.state.DocFreq, err = asInt(f.statsReader.ReadVInt()); err != nil {
			return err
		}
		// fmt.Printf("    dF=%v\n", f.state.DocFreq)
		if f.ste.fr.fieldInfo.IndexOptions() != INDEX_OPT_DOCS_ONLY {
			var n int64
			if n, err = f.statsReader.ReadVLong(); err != nil {
				return err
			}
			f.state.TotalTermFreq = int64(f.state.DocFreq) + n
			// fmt.Printf("    totTF=%v\n", f.state.TotalTermFreq)
		}

		// metadata
		for i := 0; i < f.ste.fr.longsSize; i++ {
			if f.longs[i], err = f.bytesReader.ReadVLong(); err != nil {
				return err
			}
		}

		if err = f.ste.fr.parent.postingsReader.DecodeTerm(f.longs,
			f.bytesReader, f.ste.fr.fieldInfo, f.state, absolute); err != nil {
			return err
		}
		f.metaDataUpto++
		absolute = false
	}

	f.state.TermBlockOrd = f.metaDataUpto
	return nil
}

// Used only by assert
func (f *segmentTermsEnumFrame) prefixMatches(target []byte) bool {
	for i := 0; i < f.prefix; i++ {
		if target[i] != f.ste.term.At(i) {
			return false
		}
	}
	return true
}

// NOTE: sets startBytePos/suffix as a side effect
func (f *segmentTermsEnumFrame) scanToTerm(target []byte, exactOnly bool) (status SeekStatus, err error) {
	if f.isLeafBlock {
		return f.scanToTermLeaf(target, exactOnly)
	}
	return f.scanToTermNonLeaf(target, exactOnly)
}

// Target's prefix matches this block's prefix; we
// scan the entries check if the suffix matches.
func (f *segmentTermsEnumFrame) scanToTermLeaf(target []byte, exactOnly bool) (status SeekStatus, err error) {
	// fmt.Printf("    scanToTermLeaf: block fp=%v prefix=%v nextEnt=%v (of %v) target=%v term=%v\n",
	// 	f.fp, f.prefix, f.nextEnt, f.entCount, brToString(target), f.ste.term)
	assert(f.nextEnt != -1)

	f.ste.termExists = true
	f.subCode = 0
	if f.nextEnt == f.entCount {
		if exactOnly {
			f.fillTerm()
		}
		return SEEK_STATUS_END, nil
	}

	if !f.prefixMatches(target) {
		panic("assert fail")
	}

	// Loop over each entry (term or sub-block) in this block:
	//nextTerm: while(nextEnt < entCount) {
	for {
		f.nextEnt++
		f.suffix, err = asInt(f.suffixesReader.ReadVInt())
		if err != nil {
			return 0, err
		}

		// suffixReaderPos := f.suffixesReader.Pos
		// fmt.Printf("      cycle: term %v (of %v) suffix=%v\n",
		// 	f.nextEnt-1, f.entCount, brToString(f.suffixBytes[suffixReaderPos:suffixReaderPos+f.suffix]))

		termLen := f.prefix + f.suffix
		f.startBytePos = f.suffixesReader.Pos
		f.suffixesReader.SkipBytes(int64(f.suffix))

		targetLimit := termLen
		if len(target) < termLen {
			targetLimit = len(target)
		}
		targetPos := f.prefix

		// Loop over bytes in the suffix, comparing to
		// the target
		bytePos := f.startBytePos
		isDone := false
		for {
			var cmp int
			var stop bool
			if targetPos < targetLimit {
				cmp = int(f.suffixBytes[bytePos]) - int(target[targetPos])
				bytePos++
				targetPos++
				stop = false
			} else {
				if targetPos != targetLimit {
					panic("assert fail")
				}
				cmp = termLen - len(target)
				stop = true
			}

			if cmp < 0 {
				// Current entry is still before the target;
				// keep scanning

				if f.nextEnt == f.entCount {
					if exactOnly {
						f.fillTerm()
					}
					// We are done scanning this block
					isDone = true
				}
				break
			} else if cmp > 0 {
				// // Done!  Current entry is after target --
				//     // return NOT_FOUND:
				f.fillTerm()

				// fmt.Println("        not found")
				return SEEK_STATUS_NOT_FOUND, nil
			} else if stop {
				// Exact match!

				// This cannot be a sub-block because we
				// would have followed the index to this
				// sub-block from the start:

				assert(f.ste.termExists)
				f.fillTerm()
				// fmt.Println("        found!")
				return SEEK_STATUS_FOUND, nil
			}
		}
		if isDone {
			// double jump
			break
		}
	}

	// It is possible (and OK) that terms index pointed us
	// at this block, but, we scanned the entire block and
	// did not find the term to position to.  This happens
	// when the target is after the last term in the block
	// (but, before the next term in the index).  EG
	// target could be foozzz, and terms index pointed us
	// to the foo* block, but the last term in this block
	// was fooz (and, eg, first term in the next block will
	// bee fop).
	fmt.Println("      block end")
	if exactOnly {
		f.fillTerm()
	}

	// TODO: not consistent that in the
	// not-exact case we don't next() into the next
	// frame here
	return SEEK_STATUS_END, nil
}

// Target's prefix matches this block's prefix; we
// scan the entries check if the suffix matches.
func (f *segmentTermsEnumFrame) scanToTermNonLeaf(target []byte,
	exactOnly bool) (status SeekStatus, err error) {

	fmt.Printf(
		"    scanToTermNonLeaf: block fp=%v prefix=%v nextEnt=%v (of %v) target=%v term=%v",
		f.fp, f.prefix, f.nextEnt, f.entCount, brToString(target), "" /*brToString(term)*/)

	assert(f.nextEnt != -1)

	if f.nextEnt == f.entCount {
		panic("not implemented yet")
	}

	assert(f.prefixMatches(target))

	// Loop over each entry (term or sub-block) in this block:
	for {
		f.nextEnt++

		code, _ := f.suffixesReader.ReadVInt() // no error
		f.suffix = int(uint32(code) >> 1)

		f.ste.termExists = (code & 1) == 0
		termLen := f.prefix + f.suffix
		f.startBytePos = f.suffixesReader.Position()
		f.suffixesReader.SkipBytes(int64(f.suffix))
		if f.ste.termExists {
			f.state.TermBlockOrd++
			f.subCode = 0
		} else {
			f.subCode, _ = f.suffixesReader.ReadVLong() // no error
			f.lastSubFP = f.fp - f.subCode
		}

		targetLimit := termLen
		if len(target) < termLen {
			targetLimit = len(target)
		}
		targetPos := f.prefix

		// Loop over bytes in the suffix, comparing to the target
		bytePos := f.startBytePos
		var toNextTerm, stopScan bool
		for {
			var cmp int
			var stop bool
			if targetPos < targetLimit {
				cmp = int(f.suffixBytes[bytePos]) - int(target[targetPos])
				bytePos++
				targetPos++
				stop = false
			} else {
				assert(targetPos == targetLimit)
				cmp = termLen - len(target)
				stop = true
			}

			if cmp < 0 {
				// Current entry is still before the target;
				// keep scanning

				if f.nextEnt == f.entCount {
					if exactOnly {
						f.fillTerm()
					}
					// We are done scanning this block
					stopScan = true
					break
				} else {
					toNextTerm = true
					break
				}
			} else if cmp > 0 {
				// Done! Current entry is after target -- return NOT_FOUND:
				f.fillTerm()

				if !exactOnly && !f.ste.termExists {
					panic("niy")
				}

				fmt.Println("        not found")
				return SEEK_STATUS_NOT_FOUND, nil
			} else if stop {
				// Exact match!

				// This cannot be a sub-block because we would have followed
				// the index to this sub-block from the start:

				assert(f.ste.termExists)
				f.fillTerm()
				fmt.Println("        found!")
				return SEEK_STATUS_FOUND, nil
			}
		}
		if toNextTerm {
			continue
		}
		if stopScan {
			break
		}
	}

	// It is possible (and OK) that terms index pointed us at this
	// block, but, we scanned the entire block and did not find the
	// term to position to. This happens when the taret is after the
	// last term in the block (but, before the next term in the index).
	// E.g., target could be foozzz, and terms index pointed us to the
	// foo* block, but the last term in this block was fooz (and, e.g.,
	// first term in the next block will be fop).
	fmt.Println("      block end")
	if exactOnly {
		f.fillTerm()
	}

	return SEEK_STATUS_END, nil
}

func (f *segmentTermsEnumFrame) fillTerm() {
	termLength := f.prefix + f.suffix
	f.ste.term.SetLength(termLength)
	f.ste.term.Grow(termLength)
	copy(f.ste.term.Bytes()[f.prefix:], f.suffixBytes[f.startBytePos:f.startBytePos+f.suffix])
}

// for debugging
func brToString(b []byte) string {
	if b == nil {
		return "nil"
	} else {
		var buf bytes.Buffer
		buf.WriteString("[")
		for i, v := range b {
			if i > 0 {
				buf.WriteString(" ")
			}
			fmt.Fprintf(&buf, "%x", v)
		}
		buf.WriteString("]")
		return fmt.Sprintf("%v %v", utf8ToString(b), buf.String())
	}
}

// Simpler version of Lucene's own method
func utf8ToString(iso8859_1_buf []byte) string {
	// buf := make([]rune, len(iso8859_1_buf))
	// for i, b := range iso8859_1_buf {
	// 	buf[i] = rune(b)
	// }
	// return string(buf)
	// TODO remove this method
	return string(iso8859_1_buf)
}

// // Lucene's BytesRef is basically Slice in Go, except here
// // that it's used as a local buffer when data is filled with
// // length unchanged temporarily.
// type bytesRef struct {
// 	/** The contents of the BytesRef. Should never be {@code null}. */
// 	bytes []byte
// 	/** Length of used bytes. */
// 	length int
// }

// func newBytesRef() *bytesRef {
// 	return &bytesRef{}
// }

// func (br *bytesRef) toBytes() []byte {
// 	return br.bytes[0:br.length]
// }

// func (br *bytesRef) ensureSize(minSize int) {
// 	assert(minSize >= 0)
// 	if cap(br.bytes) < minSize {
// 		next := make([]byte, util.Oversize(minSize, 1))
// 		copy(next, br.bytes)
// 		br.bytes = next
// 	}
// }

// func (br *bytesRef) String() string {
// 	return brToString(br.bytes[0:br.length])
// }

// /**
//  * Copies the bytes from the given {@link BytesRef}
//  * <p>
//  * NOTE: if this would exceed the array size, this method creates a
//  * new reference array.
//  */
// func (br *bytesRef) copyBytes(other []byte) {
// 	if cap(br.bytes) < len(other) {
// 		next := make([]byte, len(other))
// 		br.bytes = next
// 	} else if len(br.bytes) < len(other) {
// 		br.bytes = br.bytes[0:len(other)]
// 	}
// 	copy(br.bytes, other)
// 	br.length = len(other)
// }
