package blocktree

import (
	"fmt"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"github.com/balzaczyy/golucene/core/util/fst"
	"sort"
	// "strconv"
)

// blocktree/SegmentTermsEnum.java

var fstOutputs = fst.ByteSequenceOutputsSingleton()
var noOutput = fstOutputs.NoOutput()

// Iterates through terms in this field
type SegmentTermsEnum struct {
	*TermsEnumImpl

	// lazy init:
	in store.IndexInput

	stack        []*segmentTermsEnumFrame
	staticFrame  *segmentTermsEnumFrame
	currentFrame *segmentTermsEnumFrame
	termExists   bool
	fr           *FieldReader

	targetBeforeCurrentLength int

	// What prefix of the current term was present in the index:
	scratchReader *store.ByteArrayDataInput

	// What prefix of the current term was present in the index:; when
	// we only next() through the index, this stays at 0. It's only set
	// when we seekCeil/Exact:
	validIndexPrefix int

	// assert only:
	eof bool

	term      *util.BytesRefBuilder
	fstReader fst.BytesReader

	arcs []*fst.Arc
}

func newSegmentTermsEnum(r *FieldReader) *SegmentTermsEnum {
	ans := &SegmentTermsEnum{
		fr:            r,
		stack:         make([]*segmentTermsEnumFrame, 0),
		scratchReader: store.NewEmptyByteArrayDataInput(),
		term:          util.NewBytesRefBuilder(),
		arcs:          make([]*fst.Arc, 1),
	}
	ans.TermsEnumImpl = NewTermsEnumImpl(ans)
	// fmt.Printf("BTTR.init seg=%v\n", r.parent.segment)

	// Used to hold seek by TermState, or cached seek
	ans.staticFrame = newFrame(ans, -1)

	if r.index != nil {
		ans.fstReader = r.index.BytesReader()
	}

	// Init w/ root block; don't use index since it may
	// not (and need not) have been loaded
	for i, _ := range ans.arcs {
		ans.arcs[i] = &fst.Arc{}
	}

	ans.currentFrame = ans.staticFrame
	var arc *fst.Arc
	if r.index != nil {
		arc = r.index.FirstArc(ans.arcs[0])
		// Empty string prefix must have an output in the index!
		if !arc.IsFinal() {
			panic("assert fail")
		}
	}
	ans.validIndexPrefix = 0
	// fmt.Printf("init frame state %v\n", ans.currentFrame.ord)
	ans.printSeekState()

	// ans.computeBlockStats()

	return ans
}

func (e *SegmentTermsEnum) initIndexInput() {
	if e.in == nil {
		e.in = e.fr.parent.in.Clone()
	}
}

func (e *SegmentTermsEnum) frame(ord int) *segmentTermsEnumFrame {
	if ord == len(e.stack) {
		e.stack = append(e.stack, newFrame(e, ord))
	} else if ord > len(e.stack) {
		// TODO over-allocate to ensure performance
		next := make([]*segmentTermsEnumFrame, 1+ord)
		copy(next, e.stack)
		for i := len(e.stack); i < len(next); i++ {
			next[i] = newFrame(e, i)
		}
		e.stack = next
	}
	assert(e.stack[ord].ord == ord)
	return e.stack[ord]
}

func (e *SegmentTermsEnum) getArc(ord int) *fst.Arc {
	if ord == len(e.arcs) {
		e.arcs = append(e.arcs, &fst.Arc{})
	} else if ord > len(e.arcs) {
		// TODO over-allocate
		next := make([]*fst.Arc, 1+ord)
		copy(next, e.arcs)
		for i := len(e.arcs); i < len(next); i++ {
			next[i] = &fst.Arc{}
		}
		e.arcs = next
	}
	return e.arcs[ord]
}

func (e *SegmentTermsEnum) Comparator() sort.Interface {
	panic("not implemented yet")
}

// Pushes a frame we seek'd to
func (e *SegmentTermsEnum) pushFrame(arc *fst.Arc, frameData []byte, length int) (f *segmentTermsEnumFrame, err error) {
	// fmt.Println("Pushing frame...")
	e.scratchReader.Reset(frameData)
	code, err := e.scratchReader.ReadVLong()
	if err != nil {
		return nil, err
	}
	fpSeek := int64(uint64(code) >> BTT_OUTPUT_FLAGS_NUM_BITS)
	f = e.frame(1 + e.currentFrame.ord)
	f.hasTerms = (code & BTT_OUTPUT_FLAG_HAS_TERMS) != 0
	f.hasTermsOrig = f.hasTerms
	f.isFloor = (code & BTT_OUTPUT_FLAG_IS_FLOOR) != 0
	if f.isFloor {
		f.setFloorData(e.scratchReader, frameData)
	}
	e.pushFrameAt(arc, fpSeek, length)
	return f, err
}

// Pushes next'd frame or seek'd frame; we later
// lazy-load the frame only when needed
func (e *SegmentTermsEnum) pushFrameAt(arc *fst.Arc, fp int64, length int) (f *segmentTermsEnumFrame, err error) {
	f = e.frame(1 + e.currentFrame.ord)
	f.arc = arc
	if f.fpOrig == fp && f.nextEnt != -1 {
		// fmt.Printf("      push reused frame ord=%v fp=%v isFloor?=%v hasTerms=%v pref=%v nextEnt=%v targetBeforeCurrentLength=%v term.length=%v vs prefix=%v\n",
		// 	f.ord, f.fp, f.isFloor, f.hasTerms, e.term, f.nextEnt, e.targetBeforeCurrentLength, e.term.length, f.prefix)
		if f.ord > e.targetBeforeCurrentLength {
			f.rewind()
		} else {
			// fmt.Println("        skip rewind!")
		}
		if length != f.prefix {
			panic("assert fail")
		}
	} else {
		f.nextEnt = -1
		f.prefix = length
		f.state.TermBlockOrd = 0
		f.fpOrig, f.fp = fp, fp
		f.lastSubFP = -1
		// fmt.Printf("      push new frame ord=%v fp=%v hasTerms=%v isFloor=%v pref=%v\n",
		// 	f.ord, f.fp, f.hasTerms, f.isFloor, e.term)
	}
	return f, nil
}

func (e *SegmentTermsEnum) SeekExact(target []byte) (ok bool, err error) {
	assert2(e.fr.index != nil, "terms index was not loaded")

	e.term.Grow(1 + len(target))

	e.eof = false
	// fmt.Printf("BTTR.seekExact seg=%v target=%v:%v current=%v (exists?=%v) validIndexPrefix=%v\n",
	// 	e.fr.parent.segment, e.fr.fieldInfo.Name, brToString(target),
	// 	brToString(e.term.bytes), e.termExists, e.validIndexPrefix)
	e.printSeekState()

	var arc *fst.Arc
	var targetUpto int
	var output interface{}

	e.targetBeforeCurrentLength = e.currentFrame.ord

	// if e.currentFrame != e.staticFrame {
	if e.currentFrame.ord != e.staticFrame.ord {
		// We are already seek'd; find the common
		// prefix of new seek term vs current term and
		// re-use the corresponding seek state.  For
		// example, if app first seeks to foobar, then
		// seeks to foobaz, we can re-use the seek state
		// for the first 5 bytes.

		// fmt.Printf("  re-use current seek state validIndexPrefix=%v\n", e.validIndexPrefix)

		arc = e.arcs[0]
		assert(arc.IsFinal())
		output = arc.Output
		targetUpto = 0

		lastFrame := e.stack[0]
		assert(e.validIndexPrefix <= e.term.Length())

		targetLimit := len(target)
		if e.validIndexPrefix < targetLimit {
			targetLimit = e.validIndexPrefix
		}

		cmp := 0

		// TODO: reverse vLong byte order for better FST
		// prefix output sharing

		// noOutputs := e.fstOutputs.NoOutput()

		// First compare up to valid seek frames:
		for targetUpto < targetLimit {
			cmp = int(e.term.At(targetUpto)) - int(target[targetUpto])
			// fmt.Printf("    cycle targetUpto=%v (vs limit=%v) cmp=%v (targetLabel=%c vs termLabel=%c) arc.output=%v output=%v\n",
			// 	targetUpto, targetLimit, cmp, target[targetUpto], e.term.bytes[targetUpto], arc.Output, output)
			if cmp != 0 {
				break
			}

			arc = e.arcs[1+targetUpto]
			assert2(arc.Label == int(target[targetUpto]),
				"arc.label=%c targetLabel=%c", arc.Label, target[targetUpto])
			panic("not implemented yet")
			// if arc.Output != noOutputs {
			// 	output = e.fstOutputs.Add(output, arc.Output).([]byte)
			// }
			if arc.IsFinal() {
				lastFrame = e.stack[1+lastFrame.ord]
			}
			targetUpto++
		}

		if cmp == 0 {
			targetUptoMid := targetUpto

			// Second compare the rest of the term, but
			// don't save arc/output/frame; we only do this
			// to find out if the target term is before,
			// equal or after the current term
			targetLimit2 := len(target)
			if e.term.Length() < targetLimit2 {
				targetLimit2 = e.term.Length()
			}
			for targetUpto < targetLimit2 {
				cmp = int(e.term.At(targetUpto)) - int(target[targetUpto])
				// fmt.Printf("    cycle2 targetUpto=%v (vs limit=%v) cmp=%v (targetLabel=%c vs termLabel=%c)\n",
				// 	targetUpto, targetLimit, cmp, target[targetUpto], e.term.bytes[targetUpto])
				if cmp != 0 {
					break
				}
				targetUpto++
			}

			if cmp == 0 {
				cmp = e.term.Length() - len(target)
			}
			targetUpto = targetUptoMid
		}

		if cmp < 0 {
			// Common case: target term is after current
			// term, ie, app is seeking multiple terms
			// in sorted order
			// fmt.Printf("  target is after current (shares prefixLen=%v); frame.ord=%v\n", targetUpto, lastFrame.ord)
			e.currentFrame = lastFrame
		} else if cmp > 0 {
			// Uncommon case: target term
			// is before current term; this means we can
			// keep the currentFrame but we must rewind it
			// (so we scan from the start)
			e.targetBeforeCurrentLength = lastFrame.ord
			// fmt.Printf("  target is before current (shares prefixLen=%v); rewind frame ord=%v\n", targetUpto, lastFrame.ord)
			e.currentFrame = lastFrame
			e.currentFrame.rewind()
		} else {
			// Target is exactly the same as current term
			assert(e.term.Length() == len(target))
			if e.termExists {
				// fmt.Println("  target is same as current; return true")
				return true, nil
			} else {
				// fmt.Println("  target is same as current but term doesn't exist")
			}
		}
	} else {
		e.targetBeforeCurrentLength = -1
		arc = e.fr.index.FirstArc(e.arcs[0])

		// Empty string prefix must have an output (block) in the index!
		assert(arc.IsFinal() && arc.Output != nil)

		// fmt.Println("    no seek state; push root frame")

		output = arc.Output

		e.currentFrame = e.staticFrame

		targetUpto = 0
		if e.currentFrame, err = e.pushFrame(arc, fstOutputs.Add(output, arc.NextFinalOutput).([]byte), 0); err != nil {
			return false, err
		}
	}

	// fmt.Printf("  start index loop targetUpto=%v output=%v currentFrame.ord=%v targetBeforeCurrentLength=%v\n",
	// 	targetUpto, output, e.currentFrame.ord, e.targetBeforeCurrentLength)

	for targetUpto < len(target) {
		targetLabel := int(target[targetUpto])
		nextArc, err := e.fr.index.FindTargetArc(targetLabel, arc, e.getArc(1+targetUpto), e.fstReader)
		if err != nil {
			return false, err
		}
		if nextArc == nil {
			// Index is exhausted
			// fmt.Printf("    index: index exhausted label=%c %x\n", targetLabel, targetLabel)

			e.validIndexPrefix = e.currentFrame.prefix

			e.currentFrame.scanToFloorFrame(target)

			if !e.currentFrame.hasTerms {
				e.termExists = false
				e.term.Set(targetUpto, byte(targetLabel))
				e.term.SetLength(1 + targetUpto)
				// fmt.Printf("  FAST NOT_FOUND term=%v\n", e.term)
				return false, nil
			}

			if err := e.currentFrame.loadBlock(); err != nil {
				return false, err
			}

			status, err := e.currentFrame.scanToTerm(target, true)
			if err != nil {
				return false, err
			}
			if status == SEEK_STATUS_FOUND {
				// fmt.Printf("  return FOUND term=%v\n", e.term)
				return true, nil
			} else {
				// fmt.Printf("  got %v; return NOT_FOUND term=%v\n", status, e.term)
				return false, nil
			}
		} else {
			// Follow this arc
			arc = nextArc
			e.term.Set(targetUpto, byte(targetLabel))
			// aggregate output as we go:
			assert(arc.Output != nil)
			if !fst.CompareFSTValue(arc.Output, noOutput) {
				output = fstOutputs.Add(output, arc.Output)
			}
			// fmt.Printf("    index: follow label=%x arc.output=%v arc.nfo=%v\n",
			// 	strconv.FormatInt(int64(target[targetUpto]), 16), arc.Output, arc.NextFinalOutput)
			targetUpto++

			if arc.IsFinal() {
				// fmt.Println("    arc is final!")
				if e.currentFrame, err = e.pushFrame(arc,
					fstOutputs.Add(output, arc.NextFinalOutput).([]byte),
					targetUpto); err != nil {
					return false, err
				}
				// fmt.Printf("    curFrame.ord=%v hasTerms=%v\n", e.currentFrame.ord, e.currentFrame.hasTerms)
			}
		}
	}

	e.validIndexPrefix = e.currentFrame.prefix

	e.currentFrame.scanToFloorFrame(target)

	// Target term is entirely contained in the index:
	if !e.currentFrame.hasTerms {
		e.termExists = false
		e.term.SetLength(targetUpto)
		// log.Printf("  FAST NOT_FOUND term=%v", e.term)
		return false, nil
	}

	if err := e.currentFrame.loadBlock(); err != nil {
		return false, err
	}

	status, err := e.currentFrame.scanToTerm(target, true)
	if err != nil {
		return false, err
	}
	if status == SEEK_STATUS_FOUND {
		// log.Printf("  return FOUND term=%v", e.term)
		return true, nil
	} else {
		// log.Printf("  got result %v; return NOT_FOUND term=%v", status, e.term)
		return false, nil
	}
}

func (e *SegmentTermsEnum) SeekCeil(text []byte) SeekStatus {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) printSeekState() {
	if e.currentFrame == e.staticFrame {
		// log.Println("  no prior seek")
	} else {
		// log.Println("  prior seek state:")
		ord := 0
		isSeekFrame := true
		for {
			f := e.frame(ord)
			assert(f != nil)
			// prefix := e.term.Bytes()[:f.prefix]
			if f.nextEnt == -1 {
				// action := "(next)"
				// if isSeekFrame {
				// 	action = "(seek)"
				// }
				// fpOrigValue := ""
				// if f.isFloor {
				// 	fpOrigValue = fmt.Sprintf(" (fpOrig=%v", f.fpOrig)
				// }
				code := (f.fp << BTT_OUTPUT_FLAGS_NUM_BITS)
				if f.hasTerms {
					code += BTT_OUTPUT_FLAG_HAS_TERMS
				}
				if f.isFloor {
					code += BTT_OUTPUT_FLAG_IS_FLOOR
				}
				// log.Printf("    frame %v ord=%v fp=%v%v prefixLen=%v prefix=%v hasTerms=%v isFloor=%v code=%v isLastInFloor=%v mdUpto=%v tbOrd=%v",
				// 	action, ord, f.fp, fpOrigValue, f.prefix, prefix, f.hasTerms, f.isFloor, code, f.isLastInFloor, f.metaDataUpto, f.getTermBlockOrd())
			} else {
				// action := "(next, loaded)"
				// if isSeekFrame {
				// 	action = "(seek, loaded)"
				// }
				// fpOrigValue := ""
				// if f.isFloor {
				// 	fpOrigValue = fmt.Sprintf(" (fpOrig=%v", f.fpOrig)
				// }
				code := (f.fp << BTT_OUTPUT_FLAGS_NUM_BITS)
				if f.hasTerms {
					code += BTT_OUTPUT_FLAG_HAS_TERMS
				}
				if f.isFloor {
					code += BTT_OUTPUT_FLAG_IS_FLOOR
				}
				// log.Printf("    frame %v ord=%v fp=%v prefixLen=%v prefix=%v nextEnt=%v (of %v) hasTerms=%v isFloor=%v code=%v lastSubFP=%v isLastInFloor=%v mdUpto=%v tbOrd=%v",
				// 	action, ord, f.fp, fpOrigValue, f.prefix, prefix, f.nextEnt, f.entCount, f.hasTerms, f.isFloor, code, f.lastSubFP, f.isLastInFloor, f.metaDataUpto, f.getTermBlockOrd())
			}
			if e.fr.index != nil {
				assert2(!isSeekFrame || f.arc != nil,
					"isSeekFrame=%v f.arc=%v", isSeekFrame, f.arc)
				panic("not implemented yet")
				// ret, err := fst.GetFSTOutput(e.fr.index, prefix)
				// if err != nil {
				// 	panic(err)
				// }
				// output := ret.([]byte)
				// if output == nil {
				// 	// log.Println("      broken seek state: prefix is not final in index")
				// 	panic("seek state is broken")
				// } else if isSeekFrame && !f.isFloor {
				// 	reader := store.NewByteArrayDataInput(output)
				// 	codeOrig, _ := reader.ReadVLong()
				// 	code := f.fp << BTT_OUTPUT_FLAGS_NUM_BITS
				// 	if f.hasTerms {
				// 		code += BTT_OUTPUT_FLAG_HAS_TERMS
				// 	}
				// 	if f.isFloor {
				// 		code += BTT_OUTPUT_FLAG_IS_FLOOR
				// 	}
				// 	if codeOrig != code {
				// 		// log.Printf("      broken seek state: output code=%v doesn't match frame code=%v", codeOrig, code)
				// 		panic("seek state is broken")
				// 	}
				// }
			}
			if f == e.currentFrame {
				break
			}
			if f.prefix == e.validIndexPrefix {
				isSeekFrame = false
			}
			ord++
		}
	}
}

func (e *SegmentTermsEnum) Next() (buf []byte, err error) {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) Term() []byte {
	assert(!e.eof)
	return e.term.Bytes()
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}

func (e *SegmentTermsEnum) DocFreq() (df int, err error) {
	assert(!e.eof)
	// log.Println("BTTR.docFreq")
	err = e.currentFrame.decodeMetaData()
	df = e.currentFrame.state.DocFreq
	// log.Printf("  return %v", df)
	return
}

func (e *SegmentTermsEnum) TotalTermFreq() (tf int64, err error) {
	assert(!e.eof)
	err = e.currentFrame.decodeMetaData()
	tf = e.currentFrame.state.TotalTermFreq
	return
}

func (e *SegmentTermsEnum) DocsByFlags(skipDocs util.Bits, reuse DocsEnum, flags int) (de DocsEnum, err error) {
	assert(!e.eof)
	// log.Printf("BTTR.docs seg=%v", e.fr.parent.segment)
	err = e.currentFrame.decodeMetaData()
	if err != nil {
		return nil, err
	}
	// log.Printf("  state=%v", e.currentFrame.state)
	return e.fr.parent.postingsReader.Docs(e.fr.fieldInfo, e.currentFrame.state, skipDocs, reuse, flags)
}

func (e *SegmentTermsEnum) DocsAndPositionsByFlags(skipDocs util.Bits, reuse DocsAndPositionsEnum, flags int) DocsAndPositionsEnum {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) SeekExactFromLast(target []byte, otherState TermState) error {
	// log.Printf("BTTR.seekExact termState seg=%v target=%v state=%v",
	// 	e.fr.parent.segment, brToString(target), otherState)
	e.eof = false
	if !fst.CompareFSTValue(target, e.term.Get()) || !e.termExists {
		assert(otherState != nil)
		// TODO can not assert type conversion here
		// _, ok := otherState.(*BlockTermState)
		// assert(ok)
		e.currentFrame = e.staticFrame
		e.currentFrame.state.CopyFrom(otherState)
		e.term.Copy(target)
		e.currentFrame.metaDataUpto = e.currentFrame.getTermBlockOrd()
		assert(e.currentFrame.metaDataUpto > 0)
		e.validIndexPrefix = 0
	} else {
		// log.Printf("  skip seek: already on target state=%v", e.currentFrame.state)
	}
	return nil
}

func copyBytes(a, b []byte) []byte {
	if len(a) < len(b) {
		a = make([]byte, len(b))
	}
	copy(a, b)
	return a[0:len(b)]
}

func (e *SegmentTermsEnum) TermState() (ts TermState, err error) {
	assert(!e.eof)
	if err = e.currentFrame.decodeMetaData(); err != nil {
		return nil, err
	}
	ts = e.currentFrame.state.Clone() // <-- clone doesn't work here
	// log.Printf("BTTR.termState seg=%v state=%v", e.fr.parent.segment, ts)
	return
}

func (e *SegmentTermsEnum) SeekExactByPosition(ord int64) error {
	panic("not implemented yet")
}

func (e *SegmentTermsEnum) Ord() int64 {
	panic("not supported!")
}

func (e *SegmentTermsEnum) String() string {
	return "SegmentTermsEnum"
}
