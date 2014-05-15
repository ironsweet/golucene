package index

import (
	"errors"
	"fmt"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// index/BufferedDeletesStream.java

type ApplyDeletesResult struct {
	// True if any actual deletes took place:
	anyDeletes bool

	// Curreng gen, for the merged segment:
	gen int64

	// If non-nil, contains segments that are 100% deleted
	allDeleted []*SegmentInfoPerCommit
}

type SegInfoByDelGen []*SegmentInfoPerCommit

func (a SegInfoByDelGen) Len() int           { return len(a) }
func (a SegInfoByDelGen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SegInfoByDelGen) Less(i, j int) bool { return a[i].bufferedDeletesGen < a[j].bufferedDeletesGen }

type Query interface{}

type QueryAndLimit struct {
}

type CoalescedDeletes struct {
	_queries map[Query]int
}

func newCoalescedDeletes() *CoalescedDeletes {
	return &CoalescedDeletes{
		_queries: make(map[Query]int),
	}
}

func (cd *CoalescedDeletes) String() string {
	panic("not implemented yet")
}

func (cd *CoalescedDeletes) update(in *FrozenBufferedDeletes) {
	panic("not implemented yet")
}

func (cd *CoalescedDeletes) terms() []*Term {
	panic("not implemented yet")
}

func (cd *CoalescedDeletes) queries() []*QueryAndLimit {
	panic("not implemented yet")
}

/*
Tracks the stream of BufferedDeletes. When DocumentsWriterPerThread
flushes, its buffered deletes are appended to this stream. We later
apply these deletes (resolve them to the actual docIDs, per segment)
when a merge is started (only to the to-be-merged segments). We also
apply to all segments when NRT reader is pulled, commit/close is
called, or when too many deletes are buffered and must be flushed (by
RAM usage or by count).

Each packet is assigned a generation, and each flushed or merged
segment is also assigned a generation, so we can track when
BufferedDeletes packets to apply to any given segment.
*/
type BufferedDeletesStream struct {
	sync.Locker
	// TODO: maybe linked list?
	deletes []*FrozenBufferedDeletes

	// Starts at 1 so that SegmentInfos that have never had deletes
	// applied (whose bufferedDelGen defaults to 0) will be correct:
	nextGen int64

	// used only by assert
	lastDeleteTerm *Term

	infoStream util.InfoStream
	bytesUsed  int64 // atomic
	numTerms   int32 // atomic
}

func newBufferedDeletesStream(infoStream util.InfoStream) *BufferedDeletesStream {
	return &BufferedDeletesStream{
		Locker:     &sync.Mutex{},
		deletes:    make([]*FrozenBufferedDeletes, 0),
		nextGen:    1,
		infoStream: infoStream,
	}
}

/* Appends a new packet of buffered deletes to the stream, setting its generation: */
func (s *BufferedDeletesStream) push(packet *FrozenBufferedDeletes) int64 {
	panic("not implemented yet")
}

func (ds *BufferedDeletesStream) clear() {
	ds.Lock()
	defer ds.Unlock()

	ds.deletes = nil
	ds.nextGen = 1
	atomic.StoreInt32(&ds.numTerms, 0)
	atomic.StoreInt64(&ds.bytesUsed, 0)
}

func (ds *BufferedDeletesStream) any() bool {
	return atomic.LoadInt64(&ds.bytesUsed) != 0
}

/*
Resolves the buffered deleted Term/Query/docIDs, into actual deleted
docIDs in the liveDocs MutableBits for each SegmentReader.
*/
func (ds *BufferedDeletesStream) applyDeletes(readerPool *ReaderPool, infos []*SegmentInfoPerCommit) (*ApplyDeletesResult, error) {
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
		ds.infoStream.Message("BD", "applyDeletes: infos=%v packetCount=%v", infos, len(ds.deletes))
	}

	gen := ds.nextGen
	ds.nextGen++

	infos2 := make([]*SegmentInfoPerCommit, len(infos))
	copy(infos2, infos)
	sort.Sort(SegInfoByDelGen(infos2))

	var coalescedDeletes *CoalescedDeletes
	var anyNewDeletes bool

	infosIDX := len(infos2) - 1
	delIDX := len(ds.deletes) - 1

	var allDeleted []*SegmentInfoPerCommit

	for infosIDX >= 0 {
		log.Printf("BD: cycle delIDX=%v infoIDX=%v", delIDX, infosIDX)

		var packet *FrozenBufferedDeletes
		if delIDX >= 0 {
			packet = ds.deletes[delIDX]
		}
		info := infos2[infosIDX]
		segGen := info.bufferedDeletesGen

		if packet != nil && segGen < packet.gen {
			log.Println("  coalesce")
			if coalescedDeletes == nil {
				coalescedDeletes = newCoalescedDeletes()
			}
			if !packet.isSegmentPrivate {
				// Only coalesce if we are NOT on a segment private del
				// packet: the segment private del packet must only be
				// applied to segments with the same delGen. yet, if a
				// segment is already deleted from the SI since it had no
				// more documents remaining after some del packets younger
				// than its segPrivate packet (higher delGen) have been
				// applied, the segPrivate packet has not been removed.
				coalescedDeletes.update(packet)
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
				if coalescedDeletes != nil {
					log.Println("    del coalesced")
					var delta int64
					delta, err = ds._applyTermDeletes(coalescedDeletes.terms(), rld, reader)
					if err != nil {
						return
					}
					delCount += delta
					delta, err = applyQueryDeletes(coalescedDeletes.queries(), rld, reader)
					if err != nil {
						return
					}
					delCount += delta
				}
				log.Println("    del exact")
				// Don't delete by Term here; DWPT already did that on flush:
				delta, err := applyQueryDeletes(packet.queries(), rld, reader)
				delCount += delta
				fullDelCount := rld.info.delCount + rld.pendingDeleteCount()
				infoDocCount := rld.info.info.DocCount()
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
					info, segGen, packet, coalescedDeletes, delCount, suffix)
			}

			if coalescedDeletes == nil {
				coalescedDeletes = newCoalescedDeletes()
			}

			// Since we are on a segment private del packet we must not
			// update the coalescedDeletes here! We can simply advance to
			// the next packet and seginfo.
			delIDX--
			infosIDX--
			info.setBufferedDeletesGen(gen)

		} else {
			log.Println("  gt")

			if coalescedDeletes != nil {
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
					delta, err := ds._applyTermDeletes(coalescedDeletes.terms(), rld, reader)
					if err != nil {
						return
					}
					delCount += delta
					delta, err = applyQueryDeletes(coalescedDeletes.queries(), rld, reader)
					if err != nil {
						return
					}
					delCount += delta
					fullDelCount := rld.info.delCount + rld.pendingDeleteCount()
					infoDocCount := rld.info.info.DocCount()
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
						info, segGen, coalescedDeletes, delCount, suffix)
				}
			}
			info.setBufferedDeletesGen(gen)

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
Removes any BufferedDeletes that we no longer need to store because
all segments in the index have had the deletes applied.
*/
func (ds *BufferedDeletesStream) prune(infos *SegmentInfos) {
	ds.assertDeleteStats()
	var minGen int64 = math.MaxInt64
	for _, info := range infos.Segments {
		if info.bufferedDeletesGen < minGen {
			minGen = info.bufferedDeletesGen
		}
	}

	if ds.infoStream.IsEnabled("BD") {
		ds.infoStream.Message("BD", "prune sis=%v minGen=%v packetCount=%v",
			infos, minGen, len(ds.deletes))
	}
	for delIDX, limit := 0, len(ds.deletes); delIDX < limit; delIDX++ {
		if ds.deletes[delIDX].gen >= minGen {
			ds.pruneDeletes(delIDX)
			ds.assertDeleteStats()
			return
		}
	}

	// All deletes pruned
	ds.pruneDeletes(len(ds.deletes))
	assert(!ds.any())
	ds.assertDeleteStats()
}

func (ds *BufferedDeletesStream) pruneDeletes(count int) {
	if count > 0 {
		if ds.infoStream.IsEnabled("BD") {
			ds.infoStream.Message("BD", "pruneDeletes: prune %v packets; %v packets remain",
				count, len(ds.deletes)-count)
		}
		for delIDX := 0; delIDX < count; delIDX++ {
			packet := ds.deletes[delIDX]
			n := atomic.AddInt32(&ds.numTerms, -int32(packet.numTermDeletes))
			assert(n >= 0)
			n2 := atomic.AddInt64(&ds.bytesUsed, -int64(packet.bytesUsed))
			assert(n2 >= 0)
			ds.deletes[delIDX] = nil
		}
		ds.deletes = ds.deletes[count:]
	}
}

/* Delete by term */
func (ds *BufferedDeletesStream) _applyTermDeletes(terms []*Term,
	rld *ReadersAndLiveDocs, reader *SegmentReader) (int64, error) {
	panic("not implemented yet")
}

/* Delete by query */
func applyQueryDeletes(queries []*QueryAndLimit,
	rld *ReadersAndLiveDocs, reader *SegmentReader) (int64, error) {
	panic("not implemented yet")
}

func (ds *BufferedDeletesStream) assertDeleteStats() {
	var numTerms2 int
	var bytesUsed2 int64
	for _, packet := range ds.deletes {
		numTerms2 += packet.numTermDeletes
		bytesUsed2 += int64(packet.bytesUsed)
	}
	n1 := int(atomic.LoadInt32(&ds.numTerms))
	assertn(numTerms2 == n1, "numTerms2=%v vs %v", numTerms2, n1)
	n2 := int64(atomic.LoadInt64(&ds.bytesUsed))
	assertn(bytesUsed2 == n2, "bytesUsed2=%v vs %v", bytesUsed2, n2)
}
