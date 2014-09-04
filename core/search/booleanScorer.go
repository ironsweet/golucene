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

func (c *BooleanScorerCollector) Collect(doc int) error {
	panic("not implemented yet")
}

func (c *BooleanScorerCollector) SetNextReader(*index.AtomicReaderContext) {}
func (c *BooleanScorerCollector) SetScorer(Scorer)                         {}
func (c *BooleanScorerCollector) AcceptsDocsOutOfOrder() bool              { return true }

type Bucket struct {
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
		ans.buckets[i] = new(Bucket)
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

func (s *BooleanScorer) ScoreAndCollectUpto(collector Collector, max int) (bool, error) {
	panic("not implemented yet")
}

func (s *BooleanScorer) String() string {
	panic("not implemented yet")
}
