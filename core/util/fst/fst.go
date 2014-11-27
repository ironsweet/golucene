package fst

import (
	"bytes"
	"fmt"
	"github.com/balzaczyy/golucene/core/codec"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/packed"
	"math"
	"reflect"
)

// util/fst/FST.java

var ARC_SHALLOW_RAM_BYTES_USED = util.ShallowSizeOfInstance(reflect.TypeOf(Arc{}))

type InputType int

const (
	INPUT_TYPE_BYTE1 = 1
	INPUT_TYPE_BYTE2 = 2
	INPUT_TYPE_BYTE4 = 3
)

const (
	FST_BIT_FINAL_ARC            = byte(1 << 0)
	FST_BIT_LAST_ARC             = byte(1 << 1)
	FST_BIT_TARGET_NEXT          = byte(1 << 2)
	FST_BIT_STOP_NODE            = byte(1 << 3)
	FST_BIT_ARC_HAS_OUTPUT       = byte(1 << 4)
	FST_BIT_ARC_HAS_FINAL_OUTPUT = byte(1 << 5)
	FST_BIT_TARGET_DELTA         = byte(1 << 6)
	FST_ARCS_AS_FIXED_ARRAY      = FST_BIT_ARC_HAS_FINAL_OUTPUT

	FIXED_ARRAY_SHALLOW_DISTANCE = 3 // 0 => only root node
	FIXED_ARRAY_NUM_ARCS_SHALLOW = 5
	FIXED_ARRAY_NUM_ARCS_DEEP    = 10

	FST_FILE_FORMAT_NAME    = "FST"
	FST_VERSION_PACKED      = 3
	FST_VERSION_VINT_TARGET = 4

	VERSION_CURRENT = FST_VERSION_VINT_TARGET

	FST_FINAL_END_NODE     = -1
	FST_NON_FINAL_END_NODE = 0

	/** If arc has this label then that arc is final/accepted */
	FST_END_LABEL = -1

	FST_DEFAULT_MAX_BLOCK_BITS = 28 // 30 for 64 bit int
)

// Represents a single arc
type Arc struct {
	Label           int
	Output          interface{}
	node            int64 // from node
	target          int64 // to node
	flags           byte
	NextFinalOutput interface{}
	nextArc         int64
	posArcsStart    int64
	bytesPerArc     int
	arcIdx          int
	numArcs         int
}

func (arc *Arc) copyFrom(other *Arc) *Arc {
	arc.node = other.node
	arc.Label = other.Label
	arc.target = other.target
	arc.flags = other.flags
	arc.Output = other.Output
	arc.NextFinalOutput = other.NextFinalOutput
	arc.nextArc = other.nextArc
	arc.bytesPerArc = other.bytesPerArc
	if other.bytesPerArc != 0 {
		arc.posArcsStart = other.posArcsStart
		arc.arcIdx = other.arcIdx
		arc.numArcs = other.numArcs
	}
	return arc
}

func (arc *Arc) flag(flag byte) bool {
	return hasFlag(arc.flags, flag)
}

func (arc *Arc) isLast() bool {
	return arc.flag(FST_BIT_LAST_ARC)
}

func (arc *Arc) IsFinal() bool {
	return arc.flag(FST_BIT_FINAL_ARC)
}

func (arc *Arc) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "node=%v target=%v label=%v", arc.node, arc.target, util.ItoHex(int64(arc.Label)))
	if arc.flag(FST_BIT_FINAL_ARC) {
		fmt.Fprintf(&b, " final")
	}
	if arc.flag(FST_BIT_LAST_ARC) {
		fmt.Fprintf(&b, " last")
	}
	if arc.flag(FST_BIT_TARGET_NEXT) {
		fmt.Fprintf(&b, " targetNext")
	}
	if arc.flag(FST_BIT_STOP_NODE) {
		fmt.Fprintf(&b, " stop")
	}
	if arc.flag(FST_BIT_ARC_HAS_OUTPUT) {
		fmt.Fprintf(&b, " output=%v", arc.Output)
	}
	if arc.flag(FST_BIT_ARC_HAS_FINAL_OUTPUT) {
		fmt.Fprintf(&b, " nextFinalOutput=%v", arc.NextFinalOutput)
	}
	if arc.bytesPerArc != 0 {
		fmt.Fprintf(&b, " arcArray(idx=%v of %v)", arc.arcIdx, arc.numArcs)
	}
	return b.String()
}

func hasFlag(flags, bit byte) bool {
	return (flags & bit) != 0
}

type FST struct {
	inputType   InputType
	bytesPerArc []int
	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput interface{}

	bytes *BytesStore

	startNode int64

	outputs Outputs

	lastFrozenNode int64

	NO_OUTPUT interface{}

	nodeCount          int64
	arcCount           int64
	arcWithOutputCount int64

	packed           bool
	nodeRefToAddress packed.PackedIntsReader

	allowArrayArcs bool

	cachedRootArcs          []*Arc
	assertingCachedRootArcs []*Arc // only set wit assert

	version int32

	nodeAddress *packed.GrowableWriter

	// TODO: we could be smarter here, and prune periodically as we go;
	// high in-count nodes will "usually" become clear early on:
	inCounts *packed.GrowableWriter

	cachedArcsBytesUsed int
}

/* Make a new empty FST, for building; Builder invokes this ctor */
func newFST(inputType InputType, outputs Outputs, willPackFST bool,
	acceptableOverheadRatio float32, allowArrayArcs bool,
	bytesPageBits int) *FST {
	bytes := newBytesStoreFromBits(uint32(bytesPageBits))
	// pad: ensure no node gets address 0 which is reserved to mean
	// the stop state w/ no arcs
	bytes.WriteByte(0)
	ans := &FST{
		inputType:      inputType,
		outputs:        outputs,
		allowArrayArcs: allowArrayArcs,
		version:        VERSION_CURRENT,
		bytes:          bytes,
		NO_OUTPUT:      outputs.NoOutput(),
		startNode:      -1,
	}
	if willPackFST {
		ans.nodeAddress = packed.NewGrowableWriter(15, 8, acceptableOverheadRatio)
		ans.inCounts = packed.NewGrowableWriter(1, 8, acceptableOverheadRatio)
	}
	return ans
}

func LoadFST(in util.DataInput, outputs Outputs) (fst *FST, err error) {
	return loadFST3(in, outputs, FST_DEFAULT_MAX_BLOCK_BITS)
}

/** Load a previously saved FST; maxBlockBits allows you to
 *  control the size of the byte[] pages used to hold the FST bytes. */
func loadFST3(in util.DataInput, outputs Outputs, maxBlockBits uint32) (fst *FST, err error) {
	// log.Printf("Loading FST from %v and output to %v...", in, outputs)
	// defer func() {
	// 	if err != nil {
	// 		log.Print("Failed to load FST.")
	// 	}
	// }()
	fst = &FST{outputs: outputs, startNode: -1}

	if maxBlockBits < 1 || maxBlockBits > 30 {
		panic(fmt.Sprintf("maxBlockBits should 1..30; got %v", maxBlockBits))
	}

	// NOTE: only reads most recent format; we don't have
	// back-compat promise for FSTs (they are experimental):
	fst.version, err = codec.CheckHeader(in, FST_FILE_FORMAT_NAME, FST_VERSION_PACKED, FST_VERSION_VINT_TARGET)
	if err != nil {
		return nil, err
	}
	if b, err := in.ReadByte(); err == nil {
		fst.packed = (b == 1)
	} else {
		return nil, err
	}
	if b, err := in.ReadByte(); err == nil {
		if b == 1 {
			// accepts empty string
			// 1 KB blocks:
			emptyBytes := newBytesStoreFromBits(10)
			if numBytes, err := in.ReadVInt(); err == nil {
				// log.Printf("Number of bytes: %v", numBytes)
				emptyBytes.CopyBytes(in, int64(numBytes))

				// De-serialize empty-string output:
				var reader BytesReader
				if fst.packed {
					// log.Printf("Forward reader.")
					reader = emptyBytes.forwardReader()
				} else {
					// log.Printf("Reverse reader.")
					reader = emptyBytes.reverseReader()
					// NoOutputs uses 0 bytes when writing its output,
					// so we have to check here else BytesStore gets
					// angry:
					if numBytes > 0 {
						reader.setPosition(int64(numBytes - 1))
					}
				}
				// log.Printf("Reading final output from %v to %v...\n", reader, outputs)
				fst.emptyOutput, err = outputs.ReadFinalOutput(reader)
			}
		} // else emptyOutput = nil
	}
	if err != nil {
		return nil, err
	}

	if t, err := in.ReadByte(); err == nil {
		switch t {
		case 0:
			fst.inputType = INPUT_TYPE_BYTE1
		case 1:
			fst.inputType = INPUT_TYPE_BYTE2
		case 2:
			fst.inputType = INPUT_TYPE_BYTE4
		default:
			panic(fmt.Sprintf("invalid input type %v", t))
		}
	} else {
		return nil, err
	}

	if fst.packed {
		if fst.nodeRefToAddress, err = packed.NewPackedReader(in); err != nil {
			return nil, err
		}
	} // else nodeRefToAddress = nil

	if fst.startNode, err = in.ReadVLong(); err == nil {
		if fst.nodeCount, err = in.ReadVLong(); err == nil {
			if fst.arcCount, err = in.ReadVLong(); err == nil {
				if fst.arcWithOutputCount, err = in.ReadVLong(); err == nil {
					if numBytes, err := in.ReadVLong(); err == nil {
						if fst.bytes, err = newBytesStoreFromInput(in, numBytes, 1<<maxBlockBits); err == nil {
							fst.NO_OUTPUT = outputs.NoOutput()

							err = fst.cacheRootArcs()

							// NOTE: bogus because this is only used during
							// building; we need to break out mutable FST from
							// immutable
							// fst.allowArrayArcs = false
						}
					}
				}
			}
		}
	}
	return fst, err
}

func (t *FST) ramBytesUsed(arcs []*Arc) int64 {
	var size int64
	if arcs != nil {
		size += util.ShallowSizeOf(arcs)
		for _, arc := range arcs {
			if arc != nil {
				size += ARC_SHALLOW_RAM_BYTES_USED
				if arc.Output != nil && arc.Output != t.outputs.NoOutput() {
					size += t.outputs.ramBytesUsed(arc.Output)
				}
				if arc.NextFinalOutput != nil && arc.NextFinalOutput != t.outputs.NoOutput() {
					size += t.outputs.ramBytesUsed(arc.NextFinalOutput)
				}
			}
		}
	}
	return size
}

func (t *FST) finish(newStartNode int64) error {
	assert2(t.startNode == -1, "already finished")
	if newStartNode == FST_FINAL_END_NODE && t.emptyOutput != nil {
		newStartNode = 0
	}
	t.startNode = newStartNode
	t.bytes.finish()

	return t.cacheRootArcs()
}

func (t *FST) getNodeAddress(node int64) int64 {
	if t.nodeAddress != nil { // Deref
		return t.nodeAddress.Get(int(node))
	} else { // Straight
		return node
	}
}

func (t *FST) cacheRootArcs() error {
	t.cachedRootArcs = make([]*Arc, 0x80)
	t.readRootArcs(t.cachedRootArcs)
	t.cachedArcsBytesUsed += int(t.ramBytesUsed(t.cachedRootArcs))

	if err := t.setAssertingRootArcs(t.cachedRootArcs); err != nil {
		return err
	}
	t.assertRootArcs()
	return nil
}

func (t *FST) readRootArcs(arcs []*Arc) (err error) {
	arc := &Arc{}
	t.FirstArc(arc)
	in := t.BytesReader()
	if targetHasArcs(arc) {
		_, err = t.readFirstRealTargetArc(arc.target, arc, in)
		for err == nil {
			if arc.Label == FST_END_LABEL {
				panic("assert fail")
			}
			if arc.Label >= len(t.cachedRootArcs) {
				break
			}
			arcs[arc.Label] = (&Arc{}).copyFrom(arc)
			if arc.isLast() {
				break
			}
			_, err = t.readNextRealArc(arc, in)
		}
	}
	return err
}

func (t *FST) setAssertingRootArcs(arcs []*Arc) error {
	t.assertingCachedRootArcs = make([]*Arc, len(arcs))
	err := t.readRootArcs(t.assertingCachedRootArcs)
	if err == nil {
		t.cachedArcsBytesUsed *= 2
	}
	return err
}

func (t *FST) assertRootArcs() {
	if t.cachedRootArcs == nil || t.assertingCachedRootArcs == nil {
		panic("assert fail")
	}
	for i, v := range t.assertingCachedRootArcs {
		root := t.cachedRootArcs[i]
		asserting := v
		if root != nil {
			assert(root.arcIdx == asserting.arcIdx)
			assert(root.bytesPerArc == asserting.bytesPerArc)
			assert(root.flags == asserting.flags)
			assert(root.Label == asserting.Label)
			assert(root.nextArc == asserting.nextArc)
			assert2(equals(root.NextFinalOutput, asserting.NextFinalOutput),
				"%v != %v", root.NextFinalOutput, asserting.NextFinalOutput)
			assert(root.node == asserting.node)
			assert(root.numArcs == asserting.numArcs)
			assert(equals(root.Output, asserting.Output))
			assert(root.posArcsStart == asserting.posArcsStart)
			assert(root.target == asserting.target)
		} else {
			assert(asserting == nil)
		}
	}
}

// Since Go doesn't has Java's Object.equals() method,
// I have to implement my own.
func equals(a, b interface{}) bool {
	sameType := reflect.TypeOf(a) == reflect.TypeOf(b)
	if _, ok := a.([]byte); ok {
		if _, ok := b.([]byte); !ok {
			// panic(fmt.Sprintf("incomparable type: %v vs %v", a, b))
			return false
		}
		b1 := a.([]byte)
		b2 := b.([]byte)
		if len(b1) != len(b2) {
			return false
		}
		for i := 0; i < len(b1) && i < len(b2); i++ {
			if b1[i] != b2[i] {
				return false
			}
		}
		return true
	} else if _, ok := a.(int64); ok {
		if _, ok := b.(int64); !ok {
			// panic(fmt.Sprintf("incomparable type: %v vs %v", a, b))
			return false
		}
		return a.(int64) == b.(int64)
	} else if a == nil && b == nil {
		return true
	} else if sameType && a == b {
		return true
	}
	return false
}

func CompareFSTValue(a, b interface{}) bool {
	return equals(a, b)
}

func (t *FST) EmptyOutput() interface{} {
	return t.emptyOutput
}

// L493
func (t *FST) setEmptyOutput(v interface{}) {
	if t.emptyOutput != nil {
		t.emptyOutput = t.outputs.merge(t.emptyOutput, v)
	} else {
		t.emptyOutput = v
	}
}

func (t *FST) Save(out util.DataOutput) error {
	assert2(t.startNode != -1, "call finish first")
	assert2(t.nodeAddress == nil, "cannot save an FST pre-packaged FST; it must first be packed")
	_, ok := t.nodeRefToAddress.(packed.Mutable)
	assert2(!t.packed || ok, "cannot save a FST which has been loaded from disk ")
	err := codec.WriteHeader(out, FST_FILE_FORMAT_NAME, VERSION_CURRENT)
	if err == nil && t.packed {
		err = out.WriteByte(1)
	} else {
		err = out.WriteByte(0)
	}
	// TODO: really we should encode this as an arc, arriving
	// to the root node, instead of special casing here:
	if err == nil && t.emptyOutput != nil {
		// accepts empty string
		err = out.WriteByte(1)

		if err == nil {
			// serialize empty-string output:
			ros := store.NewRAMOutputStreamBuffer()
			err = t.outputs.writeFinalOutput(t.emptyOutput, ros)

			if err == nil {
				emptyOutputBytes := make([]byte, ros.FilePointer())
				err = ros.WriteToBytes(emptyOutputBytes)

				length := len(emptyOutputBytes)
				if err == nil && !t.packed {
					// reverse
					stopAt := length / 2
					for upto := 0; upto < stopAt; upto++ {
						emptyOutputBytes[upto], emptyOutputBytes[length-upto-1] =
							emptyOutputBytes[length-upto-1], emptyOutputBytes[upto]
					}
				}
				if err == nil {
					err = out.WriteVInt(int32(length))
					if err == nil {
						err = out.WriteBytes(emptyOutputBytes)
					}
				}
			}
		}
	} else if err == nil {
		err = out.WriteByte(0)
	}
	if err != nil {
		return err
	}

	var tb byte
	switch int(t.inputType) {
	case INPUT_TYPE_BYTE1:
		tb = 0
	case INPUT_TYPE_BYTE2:
		tb = 1
	default:
		tb = 2
	}
	err = out.WriteByte(tb)
	if err == nil && t.packed {
		err = t.nodeRefToAddress.(packed.Mutable).Save(out)
	}
	if err != nil {
		return err
	}

	err = out.WriteVLong(t.startNode)
	if err == nil {
		err = out.WriteVLong(t.nodeCount)
		if err == nil {
			err = out.WriteVLong(t.arcCount)
			if err == nil {
				err = out.WriteVLong(t.arcWithOutputCount)
				if err == nil {
					err = out.WriteVLong(t.bytes.position())
					if err == nil {
						err = t.bytes.writeTo(out)
					}
				}
			}
		}
	}
	return err
}

func (t *FST) writeLabel(out util.DataOutput, v int) error {
	assert2(v >= 0, "v=%v", v)
	if t.inputType == INPUT_TYPE_BYTE1 {
		assert2(v <= 255, "v=%v", v)
		return out.WriteByte(byte(v))
	} else if t.inputType == INPUT_TYPE_BYTE2 {
		panic("not implemented yet")
	} else {
		panic("not implemented yet")
	}
}

func (t *FST) readLabel(in util.DataInput) (v int, err error) {
	switch t.inputType {
	case INPUT_TYPE_BYTE1: // Unsigned byte
		if b, err := in.ReadByte(); err == nil {
			v = int(b)
		}
	case INPUT_TYPE_BYTE2: // Unsigned short
		if s, err := in.ReadShort(); err == nil {
			v = int(s)
		}
	default:
		v, err = AsInt(in.ReadVInt())
	}
	return v, err
}

func targetHasArcs(arc *Arc) bool {
	return arc.target > 0
}

/* Serializes new node by appending its bytes to the end of the current []byte */
func (t *FST) addNode(nodeIn *UnCompiledNode) (int64, error) {
	// fmt.Printf("FST.addNode pos=%v numArcs=%v\n", t.bytes.position(), nodeIn.NumArcs)
	if nodeIn.NumArcs == 0 {
		if nodeIn.IsFinal {
			return FST_FINAL_END_NODE, nil
		}
		return FST_NON_FINAL_END_NODE, nil
	}

	startAddress := t.bytes.position()
	// fmt.Printf("  startAddr=%v\n", startAddress)

	doFixedArray := t.shouldExpand(nodeIn)
	if doFixedArray {
		// fmt.Println("  fixedArray")
		if len(t.bytesPerArc) < nodeIn.NumArcs {
			t.bytesPerArc = make([]int, util.Oversize(nodeIn.NumArcs, 1))
		}
	}

	t.arcCount += int64(nodeIn.NumArcs)

	lastArc := nodeIn.NumArcs - 1

	lastArcStart := t.bytes.position()
	maxBytesPerArc := 0
	for arcIdx := 0; arcIdx < nodeIn.NumArcs; arcIdx++ {
		arc := nodeIn.Arcs[arcIdx]
		target := arc.Target.(*CompiledNode)
		flags := byte(0)
		// fmt.Printf("  arc %v label=%v -> target=%v\n", arcIdx, arc.label, target.node)

		if arcIdx == lastArc {
			flags += FST_BIT_LAST_ARC
		}

		if t.lastFrozenNode == target.node && !doFixedArray {
			flags += FST_BIT_TARGET_NEXT
		}

		if arc.isFinal {
			flags += FST_BIT_FINAL_ARC
			if arc.nextFinalOutput != NO_OUTPUT {
				flags += FST_BIT_ARC_HAS_FINAL_OUTPUT
			}
		} else {
			assert(arc.nextFinalOutput == NO_OUTPUT)
		}

		targetHasArcs := target.node > 0

		if !targetHasArcs {
			flags += FST_BIT_STOP_NODE
		} else if t.inCounts != nil {
			panic("not implemented yet")
		}

		if arc.output != NO_OUTPUT {
			flags += FST_BIT_ARC_HAS_OUTPUT
		}

		t.bytes.WriteByte(flags)
		var err error
		if err = t.writeLabel(t.bytes, arc.label); err != nil {
			return 0, err
		}

		// fmt.Printf("  write arc: label=%c flags=%v target=%v pos=%v output=%v\n",
		// 	rune(arc.label), flags, target.node, t.bytes.position(),
		// 	t.outputs.outputToString(arc.output))

		if arc.output != NO_OUTPUT {
			if err = t.outputs.Write(arc.output, t.bytes); err != nil {
				return 0, err
			}
			// fmt.Println("    write output")
			t.arcWithOutputCount++
		}

		if arc.nextFinalOutput != NO_OUTPUT {
			// fmt.Println("    write final output")
			if err = t.outputs.writeFinalOutput(arc.nextFinalOutput, t.bytes); err != nil {
				return 0, err
			}
		}

		if targetHasArcs && (flags&FST_BIT_TARGET_NEXT) == 0 {
			assert(target.node > 0)
			// fmt.Println("    write target")
			if err = t.bytes.WriteVLong(target.node); err != nil {
				return 0, err
			}
		}

		// just write the arcs "like normal" on first pass, but record
		// how many bytes each one took, and max byte size:
		if doFixedArray {
			t.bytesPerArc[arcIdx] = int(t.bytes.position() - lastArcStart)
			lastArcStart = t.bytes.position()
			if t.bytesPerArc[arcIdx] > maxBytesPerArc {
				maxBytesPerArc = t.bytesPerArc[arcIdx]
			}
		}
	}

	if doFixedArray {
		MAX_HEADER_SIZE := 11 // header(byte) + numArcs(vint) + numBytes(vint)
		assert(maxBytesPerArc > 0)
		// 2nd pass just "expands" all arcs to take up a fixed byte size
		// create the header
		header := make([]byte, MAX_HEADER_SIZE)
		bad := store.NewByteArrayDataOutput(header)
		// write a "false" first arc:
		bad.WriteByte(FST_ARCS_AS_FIXED_ARRAY)
		bad.WriteVInt(int32(nodeIn.NumArcs))
		bad.WriteVInt(int32(maxBytesPerArc))
		headerLen := bad.Position()

		fixedArrayStart := startAddress + int64(headerLen)

		// expand the arcs in place, backwards
		srcPos := t.bytes.position()
		destPos := fixedArrayStart + int64(nodeIn.NumArcs)*int64(maxBytesPerArc)
		assert(destPos >= srcPos)
		if destPos > srcPos {
			t.bytes.skipBytes(int(destPos - srcPos))
			for arcIdx := nodeIn.NumArcs - 1; arcIdx >= 0; arcIdx-- {
				destPos -= int64(maxBytesPerArc)
				srcPos -= int64(t.bytesPerArc[arcIdx])
				if srcPos != destPos {
					assert2(destPos > srcPos,
						"destPos=%v srcPos=%v arcIdx=%v maxBytesPerArc=%v bytesPerArc[arcIdx]=%v nodeIn.numArcs=%v",
						destPos, srcPos, arcIdx, maxBytesPerArc, t.bytesPerArc[arcIdx], nodeIn.NumArcs)
					t.bytes.copyBytesInside(srcPos, destPos, t.bytesPerArc[arcIdx])
				}
			}
		}

		// now write the header
		t.bytes.writeBytesAt(startAddress, header[:headerLen])
	}

	thisNodeAddress := t.bytes.position() - 1

	t.bytes.reverse(startAddress, thisNodeAddress)

	// PackedInts uses int as the index, so we cannot handle > 2.1B
	// nodes when packing:
	assert2(t.nodeAddress == nil || t.nodeCount < math.MaxInt32,
		"cannot create a packed FST with more than 2.1 billion nodes")

	t.nodeCount++
	var node int64
	if t.nodeAddress != nil {
		panic("not implemented yet")
	} else {
		node = thisNodeAddress
	}
	t.lastFrozenNode = node

	// fmt.Printf("  ret node=%v address=%v nodeAddress=%v",
	// 	node, thisNodeAddress, t.nodeAddress)
	return node, nil
}

func (t *FST) FirstArc(arc *Arc) *Arc {
	if t.emptyOutput != nil {
		arc.flags = FST_BIT_FINAL_ARC | FST_BIT_LAST_ARC
		arc.NextFinalOutput = t.emptyOutput
		if t.emptyOutput != NO_OUTPUT {
			arc.flags |= FST_BIT_ARC_HAS_FINAL_OUTPUT
		}
	} else {
		arc.flags = FST_BIT_LAST_ARC
		arc.NextFinalOutput = t.NO_OUTPUT
	}
	arc.Output = t.NO_OUTPUT

	// If there are no nodes, ie, the FST only accepts the
	// empty string, then startNode is 0
	arc.target = t.startNode
	return arc
}

func (t *FST) readUnpackedNodeTarget(in BytesReader) (target int64, err error) {
	if t.version < FST_VERSION_VINT_TARGET {
		return AsInt64(in.ReadInt())
	}
	return in.ReadVLong()
}

func AsInt(n int32, err error) (n2 int, err2 error) {
	return int(n), err
}

func AsInt64(n int32, err error) (n2 int64, err2 error) {
	return int64(n), err
}

func (t *FST) readFirstTargetArc(follow, arc *Arc, in BytesReader) (*Arc, error) {
	if follow.IsFinal() {
		// insert "fake" final first arc:
		arc.Label = FST_END_LABEL
		arc.Output = follow.NextFinalOutput
		arc.flags = FST_BIT_FINAL_ARC
		if follow.target <= 0 {
			arc.flags |= FST_BIT_LAST_ARC
		} else {
			arc.node = follow.target
			// NOTE: nextArc is a node (not an address!) in this case:
			arc.nextArc = follow.target
		}
		arc.target = FST_FINAL_END_NODE
		return arc, nil
	}
	return t.readFirstRealTargetArc(follow.target, arc, in)
}

func (t *FST) readFirstRealTargetArc(node int64, arc *Arc, in BytesReader) (ans *Arc, err error) {
	address := t.getNodeAddress(node)
	in.setPosition(address)
	arc.node = node

	flag, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	if flag == FST_ARCS_AS_FIXED_ARRAY {
		// this is first arc in a fixed-array
		arc.numArcs, err = AsInt(in.ReadVInt())
		if err != nil {
			return nil, err
		}
		if t.packed || t.version >= FST_VERSION_VINT_TARGET {
			arc.bytesPerArc, err = AsInt(in.ReadVInt())
		} else {
			arc.bytesPerArc, err = AsInt(in.ReadInt())
		}
		if err != nil {
			return nil, err
		}
		arc.arcIdx = -1
		pos := in.getPosition()
		arc.nextArc, arc.posArcsStart = pos, pos
	} else {
		// arc.flags = b
		arc.nextArc = address
		arc.bytesPerArc = 0
	}

	return t.readNextRealArc(arc, in)
}

func (t *FST) readNextArc(arc *Arc, in BytesReader) (*Arc, error) {
	if arc.Label == FST_END_LABEL {
		// this was a fake inserted "final" arc
		assert2(arc.nextArc > 0, "cannot readNextArc when arc.isLast()=true")
		return t.readFirstRealTargetArc(arc.nextArc, arc, in)
	} else {
		return t.readNextRealArc(arc, in)
	}
}

/** Never returns null, but you should never call this if
 *  arc.isLast() is true. */
func (t *FST) readNextRealArc(arc *Arc, in BytesReader) (ans *Arc, err error) {
	// TODO: can't assert this because we call from readFirstArc
	// assert !flag(arc.flags, BIT_LAST_ARC);

	// this is a continuing arc in a fixed array
	if arc.bytesPerArc != 0 { // arcs are at fixed entries
		arc.arcIdx++
		// assert arc.arcIdx < arc.numArcs
		in.setPosition(arc.posArcsStart)
		in.skipBytes(int64(arc.arcIdx * arc.bytesPerArc))
	} else { // arcs are packed
		in.setPosition(arc.nextArc)
	}
	if arc.flags, err = in.ReadByte(); err == nil {
		arc.Label, err = t.readLabel(in)
	}
	if err != nil {
		return nil, err
	}

	if arc.flag(FST_BIT_ARC_HAS_OUTPUT) {
		arc.Output, err = t.outputs.Read(in)
		if err != nil {
			return nil, err
		}
	} else {
		arc.Output = t.outputs.NoOutput()
	}

	if arc.flag(FST_BIT_ARC_HAS_FINAL_OUTPUT) {
		arc.NextFinalOutput, err = t.outputs.ReadFinalOutput(in)
		if err != nil {
			return nil, err
		}
	} else {
		arc.NextFinalOutput = t.outputs.NoOutput()
	}

	if arc.flag(FST_BIT_STOP_NODE) {
		if arc.flag(FST_BIT_FINAL_ARC) {
			arc.target = FST_FINAL_END_NODE
		} else {
			arc.target = FST_NON_FINAL_END_NODE
		}
		arc.nextArc = in.getPosition()
	} else if arc.flag(FST_BIT_TARGET_NEXT) {
		arc.nextArc = in.getPosition()
		// TODO: would be nice to make this lazy -- maybe
		// caller doesn't need the target and is scanning arcs...
		if t.nodeAddress == nil {
			if !arc.flag(FST_BIT_LAST_ARC) {
				if arc.bytesPerArc == 0 { // must scan
					t.seekToNextNode(in)
				} else {
					in.setPosition(arc.posArcsStart)
					in.skipBytes(int64(arc.bytesPerArc * arc.numArcs))
				}
			}
			arc.target = in.getPosition()
		} else {
			arc.target = arc.node - 1
			// assert arc.target > 0
		}
	} else {
		if t.packed {
			pos := in.getPosition()
			code, err := in.ReadVLong()
			if err != nil {
				return nil, err
			}
			if arc.flag(FST_BIT_TARGET_DELTA) { // Address is delta-coded from current address:
				arc.target = pos + code
			} else if code < int64(t.nodeRefToAddress.Size()) { // Deref
				arc.target = t.nodeRefToAddress.Get(int(code))
			} else { // Absolute
				arc.target = code
			}
		} else {
			arc.target, err = t.readUnpackedNodeTarget(in)
			if err != nil {
				return nil, err
			}
		}
		arc.nextArc = in.getPosition()
	}
	return arc, nil
}

// TODO: could we somehow [partially] tableize arc lookups
// look automaton?

/** Finds an arc leaving the incoming arc, replacing the arc in place.
 *  This returns null if the arc was not found, else the incoming arc. */
func (t *FST) FindTargetArc(labelToMatch int, follow *Arc, arc *Arc, in BytesReader) (target *Arc, err error) {
	if labelToMatch == FST_END_LABEL {
		if follow.IsFinal() {
			if follow.target <= 0 {
				arc.flags = FST_BIT_LAST_ARC
			} else {
				arc.flags = 0
				// NOTE: nextArc is a node (not an address!) in this case:
				arc.nextArc = follow.target
				arc.node = follow.target
			}
			arc.Output = follow.NextFinalOutput
			arc.Label = FST_END_LABEL
			return arc, nil
		} else {
			return nil, nil
		}
	}

	// Short-circuit if this arc is in the root arc cache:
	if follow.target == t.startNode && labelToMatch < len(t.cachedRootArcs) {
		// LUCENE-5152: detect tricky cases where caller
		// modified previously returned cached root-arcs:
		t.assertRootArcs()
		if result := t.cachedRootArcs[labelToMatch]; result != nil {
			arc.copyFrom(result)
			return arc, nil
		}
		return nil, nil
	}

	if !targetHasArcs(follow) {
		return nil, nil
	}

	in.setPosition(t.getNodeAddress(follow.target))

	arc.node = follow.target

	// log.Printf("fta label=%v", labelToMatch)

	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == FST_ARCS_AS_FIXED_ARRAY {
		// Arcs are full array; do binary search:
		arc.numArcs, err = AsInt(in.ReadVInt())
		if err != nil {
			return nil, err
		}
		if t.packed || t.version >= FST_VERSION_VINT_TARGET {
			arc.bytesPerArc, err = AsInt(in.ReadVInt())
			if err != nil {
				return nil, err
			}
		} else {
			arc.bytesPerArc, err = AsInt(in.ReadInt())
			if err != nil {
				return nil, err
			}
		}
		arc.posArcsStart = in.getPosition()
		for low, high := 0, arc.numArcs-1; low < high; {
			// log.Println("    cycle")
			mid := int(uint(low+high) / 2)
			in.setPosition(arc.posArcsStart)
			in.skipBytes(int64(arc.bytesPerArc*mid) + 1)
			midLabel, err := t.readLabel(in)
			if err != nil {
				return nil, err
			}
			cmp := midLabel - labelToMatch
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				arc.arcIdx = mid - 1
				// log.Println("    found!")
				return t.readNextRealArc(arc, in)
			}
		}

		return nil, nil
	}

	// Linear scan

	if _, err = t.readFirstRealTargetArc(follow.target, arc, in); err != nil {
		return nil, err
	}

	for {
		//System.out.println("  non-bs cycle");
		// TODO: we should fix this code to not have to create
		// object for the output of every arc we scan... only
		// for the matching arc, if found
		if arc.Label == labelToMatch {
			//System.out.println("    found!");
			return arc, nil
		} else if arc.Label > labelToMatch {
			return nil, nil
		} else if arc.isLast() {
			return nil, nil
		} else {
			if _, err = t.readNextRealArc(arc, in); err != nil {
				return nil, err
			}
		}
	}
}

func (t *FST) seekToNextNode(in BytesReader) error {
	var err error
	var flags byte
	for {
		if flags, err = in.ReadByte(); err == nil {
			_, err = t.readLabel(in)
		}
		if err != nil {
			return err
		}

		if hasFlag(flags, FST_BIT_ARC_HAS_OUTPUT) {
			if err = t.outputs.SkipOutput(in); err != nil {
				return err
			}
		}

		if hasFlag(flags, FST_BIT_ARC_HAS_FINAL_OUTPUT) {
			if err = t.outputs.SkipFinalOutput(in); err != nil {
				return err
			}
		}

		if !hasFlag(flags, FST_BIT_STOP_NODE) && !hasFlag(flags, FST_BIT_TARGET_NEXT) {
			if t.packed {
				_, err = in.ReadVLong()
			} else {
				_, err = t.readUnpackedNodeTarget(in)
			}
			if err != nil {
				return err
			}
		}

		if hasFlag(flags, FST_BIT_LAST_ARC) {
			return nil
		}
	}
}

func (t *FST) NodeCount() int64 {
	return t.nodeCount + 1
}

func (t *FST) shouldExpand(node *UnCompiledNode) bool {
	return t.allowArrayArcs &&
		(node.depth <= FIXED_ARRAY_SHALLOW_DISTANCE && node.NumArcs >= FIXED_ARRAY_NUM_ARCS_SHALLOW ||
			node.NumArcs >= FIXED_ARRAY_NUM_ARCS_DEEP)
}

func (t *FST) BytesReader() BytesReader {
	if t.packed {
		return t.bytes.forwardReader()
	}
	return t.bytes.reverseReader()
}

type RandomAccess interface {
	getPosition() int64
	setPosition(pos int64)
	reversed() bool
	skipBytes(count int64)
}

type BytesReader interface {
	// *util.DataInputImpl
	util.DataInput
	RandomAccess
}

// L1464
/*
Expert: creates an FST by packing this one. This process requires
substantial additional RAM (currently up to ~8 bytes per node
depending on acceptableOverheadRatio), but then should produce a
smaller FST.

The implementation of this method uses ideas from
<a target="_blank" href="http://www.cs.put.poznan.pl/dweiss/site/publications/download/fsacomp.pdf">Smaller Representation of Finite State Automata</a>
which describes techniques to reduce the size of a FST. However, this
is not a strict implementation of the algorithms described in this
paper.
*/
func (t *FST) pack(minInCountDeref, maxDerefNodes int,
	acceptableOverheadRatio float32) (*FST, error) {
	panic("not implemented yet")
}
