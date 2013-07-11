package util

import (
	"fmt"
	"github.com/balzaczyy/golucene/codec"
)

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

	FST_FILE_FORMAT_NAME    = "FST"
	FST_VERSION_PACKED      = 3
	FST_VERSION_VINT_TARGET = 4

	FST_FINAL_END_NODE     = -1
	FST_NON_FINAL_END_NODE = 0

	FST_DEFAULT_MAX_BLOCK_BITS = 28 // 30 for 64 bit int
)

type Arc struct {
	label           int32
	output          interface{}
	node            int64 // from node
	target          int64 // to node
	flags           byte
	nextFinalOutput interface{}
	nextArc         int64
	posArcsStart    int64
	bytesPerArc     uint32
	arcIdx          int32
	numArcs         uint32
}

func newArcFrom(other Arc) Arc {
	arc := other
	if arc.bytesPerArc == 0 {
		arc.posArcsStart = 0
		arc.arcIdx = 0
		arc.numArcs = 0
	}
	return arc
}

func (arc Arc) isLast() bool {
	return arc.flag(FST_BIT_LAST_ARC)
}

func (arc Arc) flag(flag byte) bool {
	return hasFlag(arc.flags, flag)
}

func hasFlag(flags, bit byte) bool {
	return (flags & bit) != 0
}

type FST struct {
	inputType          InputType
	bytes              *BytesStore
	startNode          int64
	outputs            Outputs
	NO_OUTPUT          interface{}
	nodeCount          int64
	arcCount           int64
	arcWithOutputCount int64
	packed             bool
	nodeRefToAddress   PackedIntsReader
	allowArrayArcs     bool
	cachedRootArcs     []Arc
	version            int32
	emptyOutput        interface{}

	nodeAddress *GrowableWriter
}

func loadFST(in *DataInput, outputs Outputs) (fst FST, err error) {
	return loadFST3(in, outputs, FST_DEFAULT_MAX_BLOCK_BITS)
}

func loadFST3(in *DataInput, outputs Outputs, maxBlockBits uint32) (fst FST, err error) {
	fst = FST{outputs: outputs, startNode: -1}

	if maxBlockBits < 1 || maxBlockBits > 30 {
		panic(fmt.Sprintf("maxBlockBits should 1..30; got %v", maxBlockBits))
	}

	// NOTE: only reads most recent format; we don't have
	// back-compat promise for FSTs (they are experimental):
	fst.version, err = codec.CheckHeader(in, FST_FILE_FORMAT_NAME, FST_VERSION_PACKED, FST_VERSION_VINT_TARGET)
	if err != nil {
		return fst, err
	}
	if b, err := in.ReadByte(); err == nil {
		fst.packed = (b == 1)
	} else {
		return fst, err
	}
	if b, err := in.ReadByte(); err == nil {
		if b == 1 {
			// accepts empty string
			// 1 KB blocks:
			emptyBytes := newBytesStoreFromBits(10)
			if numBytes, err := in.ReadVInt(); err == nil {
				emptyBytes.CopyBytes(in, int64(numBytes))

				// De-serialize empty-string output:
				var reader *BytesReader
				if fst.packed {
					reader = emptyBytes.forwardReader()
				} else {
					reader = emptyBytes.reverseReader()
					// NoOutputs uses 0 bytes when writing its output,
					// so we have to check here else BytesStore gets
					// angry:
					if numBytes > 0 {
						reader.setPosition(int64(numBytes - 1))
					}
				}
				fst.emptyOutput, err = outputs.readFinalOutput(reader.DataInput)
			}
		} // else emptyOutput = nil
	}
	if err != nil {
		return fst, err
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
	}
	if err != nil {
		return fst, err
	}

	if fst.packed {
		fst.nodeRefToAddress, err = newPackedReader(in)
		if err != nil {
			return fst, err
		}
	} // else nodeRefToAddress = nil

	if fst.startNode, err = in.ReadVLong(); err == nil {
		if fst.nodeCount, err = in.ReadVLong(); err == nil {
			if fst.arcCount, err = in.ReadVLong(); err == nil {
				if fst.arcWithOutputCount, err = in.ReadVLong(); err == nil {
					if numBytes, err := in.ReadVLong(); err == nil {
						if fst.bytes, err = newBytesStoreFromInput(in, numBytes, 1<<maxBlockBits); err == nil {
							fst.NO_OUTPUT = outputs.noOutput()

							fst.cacheRootArcs()

							// NOTE: bogus because this is only used during
							// building; we need to break out mutable FST from
							// immutable
							fst.allowArrayArcs = false
						}
					}
				}
			}
		}
	}
	return fst, err
}

func (t *FST) getNodeAddress(node int64) int64 {
	if t.nodeAddress != nil { // Deref
		return t.nodeAddress.Get(int32(node))
	} else { // Straight
		return node
	}
}

func (t *FST) cacheRootArcs() {
	t.cachedRootArcs = make([]Arc, 0x80)
	arc := Arc{}
	t.getFirstArc(&arc)
	in := t.getBytesReader()
	if targetHasArcs(arc) {
		t.readFirstRealTargetArc(arc.target, &arc, in)
		for {
			// assert arc.label != END_LABEL
			if arc.label >= int32(len(t.cachedRootArcs)) {
				break
			}
			t.cachedRootArcs[arc.label] = newArcFrom(arc)
			if arc.isLast() {
				break
			}
			t.readNextRealArc(&arc, in)
		}
	}
}

func (t *FST) readLabel(in *DataInput) (v int32, err error) {
	switch t.inputType {
	case INPUT_TYPE_BYTE1: // Unsigned byte
		if b, err := in.ReadByte(); err == nil {
			v = int32(b)
		}
	case INPUT_TYPE_BYTE2: // Unsigned short
		if s, err := in.ReadShort(); err == nil {
			v = int32(s)
		}
	default:
		v, err = in.ReadVInt()
	}
	return v, err
}

func targetHasArcs(arc Arc) bool {
	return arc.target > 0
}

func (t *FST) getFirstArc(arc *Arc) *Arc {
	if t.emptyOutput != nil {
		arc.flags = FST_BIT_FINAL_ARC | FST_BIT_LAST_ARC
		arc.nextFinalOutput = t.emptyOutput
	} else {
		arc.flags = FST_BIT_LAST_ARC
		arc.nextFinalOutput = t.NO_OUTPUT
	}
	arc.output = t.NO_OUTPUT

	// If there are no nodes, ie, the FST only accepts the
	// empty string, then startNode is 0
	arc.target = t.startNode
	return arc
}

func (t *FST) readUnpackedNodeTarget(in *BytesReader) (target int64, err error) {
	if t.version < FST_VERSION_VINT_TARGET {
		return AsInt64(in.ReadInt())
	}
	return in.ReadVLong()
}

func AsUint32(n int32, err error) (n2 uint32, err2 error) {
	return uint32(n), err
}

func AsInt64(n int32, err error) (n2 int64, err2 error) {
	return int64(n), err
}

func (t *FST) readFirstRealTargetArc(node int64, arc *Arc, in *BytesReader) (ans *Arc, err error) {
	address := t.getNodeAddress(node)
	in.setPosition(address)
	arc.node = node

	flag, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	if flag == FST_ARCS_AS_FIXED_ARRAY {
		// this is first arc in a fixed-array
		arc.numArcs, err = AsUint32(in.ReadVInt())
		if err != nil {
			return nil, err
		}
		if t.packed || t.version >= FST_VERSION_VINT_TARGET {
			arc.bytesPerArc, err = AsUint32(in.ReadVInt())
		} else {
			arc.bytesPerArc, err = AsUint32(in.ReadInt())
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

func (t *FST) readNextRealArc(arc *Arc, in *BytesReader) (ans *Arc, err error) {
	// TODO: can't assert this because we call from readFirstArc
	// assert !flag(arc.flags, BIT_LAST_ARC);

	// this is a continuing arc in a fixed array
	if arc.bytesPerArc != 0 { // arcs are at fixed entries
		arc.arcIdx++
		// assert arc.arcIdx < arc.numArcs
		in.setPosition(arc.posArcsStart)
		in.skipBytes(arc.arcIdx * int32(arc.bytesPerArc))
	} else { // arcs are packed
		in.setPosition(arc.nextArc)
	}
	if arc.flags, err = in.ReadByte(); err == nil {
		arc.label, err = t.readLabel(in.DataInput)
	}
	if err != nil {
		return nil, err
	}

	if arc.flag(FST_BIT_ARC_HAS_OUTPUT) {
		arc.output, err = t.outputs.read(in.DataInput)
		if err != nil {
			return nil, err
		}
	} else {
		arc.output = t.outputs.noOutput()
	}

	if arc.flag(FST_BIT_ARC_HAS_FINAL_OUTPUT) {
		arc.nextFinalOutput, err = t.outputs.readFinalOutput(in.DataInput)
		if err != nil {
			return nil, err
		}
	} else {
		arc.nextFinalOutput = t.outputs.noOutput()
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
					in.skipBytes(int32(arc.bytesPerArc * arc.numArcs))
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
				arc.target = t.nodeRefToAddress.Get(int32(code))
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

func (t *FST) seekToNextNode(in *BytesReader) error {
	var err error
	var flags byte
	for {
		if flags, err = in.ReadByte(); err == nil {
			_, err = t.readLabel(in.DataInput)
		}
		if err != nil {
			return err
		}

		if hasFlag(flags, FST_BIT_ARC_HAS_OUTPUT) {
			_, err = t.outputs.read(in.DataInput)
			if err != nil {
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

func (t *FST) getBytesReader() *BytesReader {
	if t.packed {
		return t.bytes.forwardReader()
	}
	return t.bytes.reverseReader()
}

type BytesReader struct {
	*DataInput
	getPosition func() int64
	setPosition func(pos int64)
	reversed    func() bool
	skipBytes   func(count int32)
}

type Outputs interface {
	read(in *DataInput) (e interface{}, err error)
	readFinalOutput(in *DataInput) (e interface{}, err error)
	noOutput() interface{}
}

type abstractOutputs struct {
	Outputs
}

func (out *abstractOutputs) readFinalOutput(in *DataInput) (e interface{}, err error) {
	return out.Outputs.read(in)
}
