package index

import (
	"errors"
	"fmt"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// index/BufferedUpdatesStream.java

type ApplyDeletesResult struct {
	// True if any actual deletes took place:
	anyDeletes bool

	// Curreng gen, for the merged segment:
	gen int64

	// If non-nil, contains segments that are 100% deleted
	allDeleted []*SegmentCommitInfo
}

type SegInfoByDelGen []*SegmentCommitInfo

func (a SegInfoByDelGen) Len() int           { return len(a) }
func (a SegInfoByDelGen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SegInfoByDelGen) Less(i, j int) bool { return a[i].BufferedUpdatesGen < a[j].BufferedUpdatesGen }

type Query interface{}

type QueryAndLimit struct {
}

// index/CoalescedUpdates.java

type CoalescedUpdates struct {
	_queries         map[Query]int
	numericDVUpdates []*DocValuesUpdate
	binaryDVUpdates  []*DocValuesUpdate
}

func newCoalescedUpdates() *CoalescedUpdates {
	return &CoalescedUpdates{
		_queries: make(map[Query]int),
	}
}

func (cd *CoalescedUpdates) String() string {
	panic("not implemented yet")
}

func (cd *CoalescedUpdates) update(in *FrozenBufferedUpdates) {
	panic("not implemented yet")
}

func (cd *CoalescedUpdates) terms() []*Term {
	panic("not implemented yet")
}

func (cd *CoalescedUpdates) queries() []*QueryAndLimit {
	panic("not implemented yet")
}

/*
Tracks the stream of BufferedUpdates. When DocumentsWriterPerThread
flushes, its buffered deletes and updates are appended to this stream.
We later apply them (resolve them to the actual docIDs, per segment)
when a merge is started (only to the to-be-merged segments). We also
apply to all segments when NRT reader is pulled, commit/close is
called, or when too many deletes or updates are buffered and must be
flushed (by RAM usage or by count).

Each packet is assigned a generation, and each flushed or merged
segment is also assigned a generation, so we can track when
BufferedUpdates packets to apply to any given segment.
*/
type BufferedUpdatesStream struct {
	sync.Locker
	// TODO: maybe linked list?
	updates []*FrozenBufferedUpdates

	// Starts at 1 so that SegmentInfos that have never had deletes
	// applied (whose bufferedDelGen defaults to 0) will be correct:
	nextGen int64

	// used only by assert
	lastDeleteTerm *Term

	infoStream util.InfoStream
	bytesUsed  int64 // atomic
	numTerms   int32 // atomic
}

func newBufferedUpdatesStream(infoStream util.InfoStream) *BufferedUpdatesStream {
	return &BufferedUpdatesStream{
		Locker:     &sync.Mutex{},
		updates:    make([]*FrozenBufferedUpdates, 0),
		nextGen:    1,
		infoStream: infoStream,
	}
}

/* Appends a new packet of buffered deletes to the stream, setting its generation: */
func (s *BufferedUpdatesStream) push(packet *FrozenBufferedUpdates) int64 {
	panic("not implemented yet")
}

func (ds *BufferedUpdatesStream) clear() {
	ds.Lock()
	defer ds.Unlock()

	ds.updates = nil
	ds.nextGen = 1
	atomic.StoreInt32(&ds.numTerms, 0)
	atomic.StoreInt64(&ds.bytesUsed, 0)
}

func (ds *BufferedUpdatesStream) any() bool {
	return atomic.LoadInt64(&ds.bytesUsed) != 0
}

func (ds *BufferedUpdatesStream) RamBytesUsed() int64 {
	return atomic.LoadInt64(&ds.bytesUsed)
}

/*
Resolves the buffered deleted Term/Query/docIDs, into actual deleted
docIDs in the liveDocs MutableBits for each SegmentReader.
*/
func (ds *BufferedUpdatesStream) applyDeletesAndUpdates(readerPool *ReaderPool, infos []*SegmentCommitInfo) (*ApplyDeletesResult, error) {
	ds.Lock()
	defer ds.Unlock()

	if len(infos) == 0 {
		ds.nextGen++
		return &ApplyDeletesResult{false, ds.nextGen - 1, nil}, nil
	}

	t0 := time.Now()
	ds.assertDeleteStats()
	if !ds.any() {
		if ds.infoStream.IsEnabled("BD") {
			ds.infoStream.Message("BD", "applyDeletes: no deletes; skipping")
		}
		ds.nextGen++
		return &ApplyDeletesResult{false, ds.nextGen - 1, nil}, nil
	}

	if ds.infoStream.IsEnabled("BD") {
		ds.infoStream.Message("BD", "applyDeletes: infos=%v packetCount=%v", infos, len(ds.updates))
	}

	gen := ds.nextGen
	ds.nextGen++

	infos2 := make([]*SegmentCommitInfo, len(infos))
	copy(infos2, infos)
	sort.Sort(SegInfoByDelGen(infos2))

	var coalescedUpdates *CoalescedUpdates
	var anyNewDeletes bool

	infosIDX := len(infos2) - 1
	delIDX := len(ds.updates) - 1

	var allDeleted []*SegmentCommitInfo

	for infosIDX >= 0 {
		log.Printf("BD: cycle delIDX=%v infoIDX=%v", delIDX, infosIDX)

		var packet *FrozenBufferedUpdates
		if delIDX >= 0 {
			packet = ds.updates[delIDX]
		}
		info := infos2[infosIDX]
		segGen := info.BufferedUpdatesGen

		if packet != nil && segGen < packet.gen {
			log.Println("  coalesce")
			if coalescedUpdates == nil {
				coalescedUpdates = newCoalescedUpdates()
			}
			if !packet.isSegmentPrivate {
				// Only coalesce if we are NOT on a segment private del
				// packet: the segment private del packet must only be
				// applied to segments with the same delGen. yet, if a
				// segment is already deleted from the SI since it had no
				// more documents remaining after some del packets younger
				// than its segPrivate packet (higher delGen) have been
				// applied, the segPrivate packet has not been removed.
				coalescedUpdates.update(packet)
			}
			delIDX--

		} else if packet != nil && segGen == packet.gen {
			assertn(packet.isSegmentPrivate,
				"Packet and Segments deletegen can only match on a segment private del packet gen=%v",
				segGen)
			log.Println("  eq")

			// Lockorder: IW -> BD -> RP
			assert(readerPool.infoIsLive(info))
			rld := readerPool.get(info, true)
			reader, err := rld.reader(store.IO_CONTEXT_READ)
			if err != nil {
				return nil, err
			}
			delCount, segAllDeletes, err := func() (delCount int64, segAllDeletes bool, err error) {
				defer func() {
					err = mergeError(err, rld.release(reader))
					err = mergeError(err, readerPool.release(rld))
				}()
				dvUpdates := newDocValuesFieldUpdatesContainer()
				if coalescedUpdates != nil {
					fmt.Println("    del coalesced")
					var delta int64
					delta, err = ds._applyTermDeletes(coalescedUpdates.terms(), rld, reader)
					if err == nil {
						delCount += delta
						delta, err = applyQueryDeletes(coalescedUpdates.queries(), rld, reader)
						if err == nil {
							delCount += delta
							err = ds.applyDocValuesUpdates(coalescedUpdates.numericDVUpdates, rld, reader, dvUpdates)
							if err == nil {
								err = ds.applyDocValuesUpdates(coalescedUpdates.binaryDVUpdates, rld, reader, dvUpdates)
							}
						}
					}
					if err != nil {
						return
					}
				}
				fmt.Println("    del exact")
				// Don't delete by Term here; DWPT already did that on flush:
				var delta int64
				delta, err = applyQueryDeletes(packet.queries(), rld, reader)
				if err == nil {
					delCount += delta
					err = ds.applyDocValuesUpdates(packet.numericDVUpdates, rld, reader, dvUpdates)
					if err == nil {
						err = ds.applyDocValuesUpdates(packet.binaryDVUpdates, rld, reader, dvUpdates)
						if err == nil && dvUpdates.any() {
							err = rld.writeFieldUpdates(info.Info.Dir, dvUpdates)
						}
					}
				}
				if err != nil {
					return
				}
				fullDelCount := rld.info.DelCount() + rld.pendingDeleteCount()
				infoDocCount := rld.info.Info.DocCount()
				assert(fullDelCount <= infoDocCount)
				return delCount, fullDelCount == infoDocCount, nil
			}()
			if err != nil {
				return nil, err
			}
			anyNewDeletes = anyNewDeletes || (delCount > 0)

			if segAllDeletes {
				allDeleted = append(allDeleted, info)
			}

			if ds.infoStream.IsEnabled("BD") {
				var suffix string
				if segAllDeletes {
					suffix = " 100%% deleted"
				}
				ds.infoStream.Message("BD", "Seg=%v segGen=%v segDeletes=[%v]; coalesced deletes=[%v] newDelCount=%v%v",
					info, segGen, packet, coalescedUpdates, delCount, suffix)
			}

			if coalescedUpdates == nil {
				coalescedUpdates = newCoalescedUpdates()
			}

			// Since we are on a segment private del packet we must not
			// update the CoalescedUpdates here! We can simply advance to
			// the next packet and seginfo.
			delIDX--
			infosIDX--
			info.SetBufferedUpdatesGen(gen)

		} else {
			log.Println("  gt")

			if coalescedUpdates != nil {
				// Lock order: IW -> BD -> RP
				assert(readerPool.infoIsLive(info))
				rld := readerPool.get(info, true)
				reader, err := rld.reader(store.IO_CONTEXT_READ)
				if err != nil {
					return nil, err
				}
				delCount, segAllDeletes, err := func() (delCount int64, segAllDeletes bool, err error) {
					defer func() {
						err = mergeError(err, rld.release(reader))
						err = mergeError(err, readerPool.release(rld))
					}()
					var delta int64
					delta, err = ds._applyTermDeletes(coalescedUpdates.terms(), rld, reader)
					if err == nil {
						delCount += delta
						delta, err = applyQueryDeletes(coalescedUpdates.queries(), rld, reader)
						if err == nil {
							delCount += delta
							dvUpdates := newDocValuesFieldUpdatesContainer()
							err = ds.applyDocValuesUpdates(coalescedUpdates.numericDVUpdates, rld, reader, dvUpdates)
							if err == nil {
								err = ds.applyDocValuesUpdates(coalescedUpdates.binaryDVUpdates, rld, reader, dvUpdates)
								if err == nil && dvUpdates.any() {
									err = rld.writeFieldUpdates(info.Info.Dir, dvUpdates)
								}
							}
						}
					}
					if err != nil {
						return
					}

					fullDelCount := rld.info.DelCount() + rld.pendingDeleteCount()
					infoDocCount := rld.info.Info.DocCount()
					assert(fullDelCount <= infoDocCount)
					return delCount, fullDelCount == infoDocCount, nil
				}()
				if err != nil {
					return nil, err
				}
				anyNewDeletes = anyNewDeletes || (delCount > 0)

				if segAllDeletes {
					allDeleted = append(allDeleted, info)
				}

				if ds.infoStream.IsEnabled("BD") {
					var suffix string
					if segAllDeletes {
						suffix = " 100%% deleted"
					}
					ds.infoStream.Message("BD", "Seg=%v segGen=%v coalesced deletes=[%v] newDelCount=%v%v",
						info, segGen, coalescedUpdates, delCount, suffix)
				}
			}
			info.SetBufferedUpdatesGen(gen)

			infosIDX--
		}
	}

	ds.assertDeleteStats()
	if ds.infoStream.IsEnabled("BD") {
		ds.infoStream.Message("BD", "applyDeletes took %v", time.Now().Sub(t0))
	}

	return &ApplyDeletesResult{anyNewDeletes, gen, allDeleted}, nil
}

func mergeError(err, err2 error) error {
	if err == nil {
		return err2
	} else {
		return errors.New(fmt.Sprintf("%v\n  %v", err, err2))
	}
}

// Lock order IW -> BD
/*
Removes any BufferedUpdates that we no longer need to store because
all segments in the index have had the deletes applied.
*/
func (ds *BufferedUpdatesStream) prune(infos *SegmentInfos) {
	ds.assertDeleteStats()
	var minGen int64 = math.MaxInt64
	for _, info := range infos.Segments {
		if info.BufferedUpdatesGen < minGen {
			minGen = info.BufferedUpdatesGen
		}
	}

	if ds.infoStream.IsEnabled("BD") {
		var dir store.Directory
		if len(infos.Segments) > 0 {
			dir = infos.Segments[0].Info.Dir
		}
		ds.infoStream.Message("BD", "prune sis=%v minGen=%v packetCount=%v",
			infos.toString(dir), minGen, len(ds.updates))
	}
	for delIDX, update := range ds.updates {
		if update.gen >= minGen {
			ds.pruneUpdates(delIDX)
			ds.assertDeleteStats()
			return
		}
	}

	// All deletes pruned
	ds.pruneUpdates(len(ds.updates))
	assert(!ds.any())
	ds.assertDeleteStats()
}

func (ds *BufferedUpdatesStream) pruneUpdates(count int) {
	if count > 0 {
		if ds.infoStream.IsEnabled("BD") {
			ds.infoStream.Message("BD", "pruneDeletes: prune %v packets; %v packets remain",
				count, len(ds.updates)-count)
		}
		for delIDX := 0; delIDX < count; delIDX++ {
			packet := ds.updates[delIDX]
			n := atomic.AddInt32(&ds.numTerms, -int32(packet.numTermDeletes))
			assert(n >= 0)
			n2 := atomic.AddInt64(&ds.bytesUsed, -int64(packet.bytesUsed))
			assert(n2 >= 0)
			ds.updates[delIDX] = nil
		}
		ds.updates = ds.updates[count:]
	}
}

/* Delete by term */
func (ds *BufferedUpdatesStream) _applyTermDeletes(terms []*Term,
	rld *ReadersAndUpdates, reader *SegmentReader) (int64, error) {
	panic("not implemented yet")
}

/* DocValues updates */
func (ds *BufferedUpdatesStream) applyDocValuesUpdates(updates []*DocValuesUpdate,
	rld *ReadersAndUpdates, reader *SegmentReader,
	dvUpdatesCntainer *DocValuesFieldUpdatesContainer) error {
	panic("not implemented yet")
}

/* Delete by query */
func applyQueryDeletes(queries []*QueryAndLimit,
	rld *ReadersAndUpdates, reader *SegmentReader) (int64, error) {
	panic("not implemented yet")
}

func (ds *BufferedUpdatesStream) assertDeleteStats() {
	var numTerms2 int
	var bytesUsed2 int64
	for _, packet := range ds.updates {
		numTerms2 += packet.numTermDeletes
		bytesUsed2 += int64(packet.bytesUsed)
	}
	n1 := int(atomic.LoadInt32(&ds.numTerms))
	assertn(numTerms2 == n1, "numTerms2=%v vs %v", numTerms2, n1)
	n2 := int64(atomic.LoadInt64(&ds.bytesUsed))
	assertn(bytesUsed2 == n2, "bytesUsed2=%v vs %v", bytesUsed2, n2)
}
