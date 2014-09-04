package search

import (
	"github.com/balzaczyy/golucene/core/index"
)

type BooleanScorerCollector struct {
	bucketTable *BucketTable
	mask        int
	scorer      Scorer
}

func newBooleanScorerCollector(mask int, bucketTable *BucketTable) *BooleanScorerCollector {
	return &BooleanScorerCollector{
		mask:        mask,
		bucketTable: bucketTable,
	}
}

func (c *BooleanScorerCollector) Collect(doc int) (err error) {
	table := c.bucketTable
	i := doc & BUCKET_TABLE_MASK
	bucket := table.buckets[i]

	var score float32
	if bucket.doc != doc {
		bucket.doc = doc
		if score, err = c.scorer.Score(); err != nil {
			return err
		}
		bucket.score = float64(score)
		bucket.bits = c.mask
		bucket.coord = 1

		bucket.next = table.first
		table.first = bucket
	} else {
		if score, err = c.scorer.Score(); err != nil {
			return err
		}
		bucket.score += float64(score)
		bucket.bits |= c.mask
		bucket.coord++
	}

	return nil
}

func (c *BooleanScorerCollector) SetNextReader(*index.AtomicReaderContext) {}
func (c *BooleanScorerCollector) SetScorer(scorer Scorer)                  { c.scorer = scorer }
func (c *BooleanScorerCollector) AcceptsDocsOutOfOrder() bool              { return true }

type Bucket struct {
	doc   int // tells if bucket is valid
	score float64
	bits  int
	coord int
	next  *Bucket
}

func newBucket() *Bucket {
	return &Bucket{
		doc: -1,
	}
}

const BUCKET_TABLE_SIZE = 1 << 11
const BUCKET_TABLE_MASK = BUCKET_TABLE_SIZE - 1

type BucketTable struct {
	buckets []*Bucket
	first   *Bucket // head of valid list
}

func newBucketTable() *BucketTable {
	ans := &BucketTable{
		buckets: make([]*Bucket, BUCKET_TABLE_SIZE),
	}
	// Pre-fill to save the lazy init when collecting each sub:
	for i, _ := range ans.buckets {
		ans.buckets[i] = newBucket()
	}
	return ans
}

func (t *BucketTable) newCollector(mask int) Collector {
	return newBooleanScorerCollector(mask, t)
}

type SubScorer struct {
	scorer     BulkScorer
	prohibited bool
	collector  Collector
	next       *SubScorer
	more       bool
}

func newSubScorer(scorer BulkScorer, required, prohibited bool,
	collector Collector, next *SubScorer) *SubScorer {

	assert2(!required, "this scorer cannot handle required=true")
	return &SubScorer{
		scorer:     scorer,
		more:       true,
		prohibited: prohibited,
		collector:  collector,
		next:       next,
	}
}

/* Any time a prohibited clause matches we set bit 0: */
const PROHIBITED_MASK = 1

type BooleanScorer struct {
	*BulkScorerImpl
	scorers          *SubScorer
	bucketTable      *BucketTable
	coordFactors     []float32
	minNrShouldMatch int
	end              int
	current          *Bucket
	weight           Weight
}

func newBooleanScorer(weight *BooleanWeight,
	disableCoord bool, minNrShouldMatch int,
	optionalScorers, prohibitedScorers []BulkScorer,
	maxCoord int) *BooleanScorer {

	ans := &BooleanScorer{
		bucketTable:      newBucketTable(),
		minNrShouldMatch: minNrShouldMatch,
		weight:           weight,
	}
	ans.BulkScorerImpl = newBulkScorer(ans)

	for _, scorer := range optionalScorers {
		ans.scorers = newSubScorer(scorer, false, false,
			ans.bucketTable.newCollector(0), ans.scorers)
	}

	for _, scorer := range prohibitedScorers {
		ans.scorers = newSubScorer(scorer, false, true,
			ans.bucketTable.newCollector(PROHIBITED_MASK), ans.scorers)
	}

	ans.coordFactors = make([]float32, len(optionalScorers)+1)
	for i, _ := range ans.coordFactors {
		if disableCoord {
			ans.coordFactors[i] = 1
		} else {
			ans.coordFactors[i] = weight.coord(i, maxCoord)
		}
	}

	return ans
}

func (s *BooleanScorer) ScoreAndCollectUpto(collector Collector, max int) (more bool, err error) {
	fs := newFakeScorer()

	// The internal loop will set the score and doc before calling collect.
	collector.SetScorer(fs)
	for {
		s.bucketTable.first = nil

		for s.current != nil { // more queued
			// check prohibited & required
			if (s.current.bits & PROHIBITED_MASK) == 0 {

				if s.current.doc >= max {
					panic("not implemented yet")
				}

				if s.current.coord >= s.minNrShouldMatch {
					fs.score = float32(s.current.score * float64(s.coordFactors[s.current.coord]))
					fs.doc = s.current.doc
					fs.freq = s.current.coord
					if err = collector.Collect(s.current.doc); err != nil {
						return false, err
					}
				}
			}

			s.current = s.current.next // pop the queue
		}

		if s.bucketTable.first != nil {
			panic("not implemented yet")
		}

		// refill the queue
		more = false
		s.end += BUCKET_TABLE_SIZE
		for sub := s.scorers; sub != nil; sub = sub.next {
			if sub.more {
				if sub.more, err = sub.scorer.ScoreAndCollectUpto(sub.collector, s.end); err != nil {
					return false, err
				}
				more = more || sub.more
			}
		}
		s.current = s.bucketTable.first

		if s.current == nil && !more {
			break
		}
	}
	return false, nil
}

func (s *BooleanScorer) String() string {
	panic("not implemented yet")
}

type FakeScorer struct {
	*abstractScorer
	score float32
	doc   int
	freq  int
}

func newFakeScorer() *FakeScorer {
	ans := &FakeScorer{
		doc:  -1,
		freq: -1,
	}
	ans.abstractScorer = newScorer(ans, nil)
	return ans
}

func (s *FakeScorer) Advance(int) (int, error) { panic("FakeScorer doesn't support advance(int)") }
func (s *FakeScorer) DocId() int               { return s.doc }
func (s *FakeScorer) Freq() (int, error)       { return s.freq, nil }
func (s *FakeScorer) NextDoc() (int, error)    { panic("FakeScorer doesn't support nextDoc()") }
func (s *FakeScorer) Score() (float32, error)  { return s.score, nil }

// func (s *FakeScorer) Cost() int64             { return 1 }
// func (s *FakeScorer) Weight() Weight          { panic("not supported") }
// func (s *FakeScorer) Children() []ChildScorer { panic("not supported") }
