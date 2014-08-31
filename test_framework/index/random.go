package index

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index"
	tu "github.com/balzaczyy/golucene/test_framework/util"
	"math"
	"math/rand"
	"time"
)

// index/MockRandomMergePolicy.java

// MergePolicy that makes random decisions for testing.
type MockRandomMergePolicy struct {
	*MergePolicyImpl
	random          *rand.Rand
	doNonBulkMerges bool
}

func NewMockRandomMergePolicy(r *rand.Rand) *MockRandomMergePolicy {
	// fork a private random, since we are called unpredicatably from threads:
	res := &MockRandomMergePolicy{
		random:          rand.New(rand.NewSource(r.Int63())),
		doNonBulkMerges: true,
	}
	res.MergePolicyImpl = NewDefaultMergePolicyImpl(res)
	return res
}

func (p *MockRandomMergePolicy) FindMerges(mergeTrigger MergeTrigger,
	segmentInfos *SegmentInfos, writer *IndexWriter) (MergeSpecification, error) {

	var segments []*SegmentCommitInfo
	merging := writer.MergingSegments()

	for _, sipc := range segmentInfos.Segments {
		if _, ok := merging[sipc]; !ok {
			segments = append(segments, sipc)
		}
	}

	var merges []*OneMerge
	if n := len(segments); n > 1 && (n > 30 || p.random.Intn(5) == 3) {
		segments2 := make([]*SegmentCommitInfo, len(segments))
		for i, v := range p.random.Perm(len(segments)) {
			segments2[i] = segments[v]
		}
		segments = segments2

		// TODO: sometimes make more than 1 merge?
		segsToMerge := tu.NextInt(p.random, 1, n)
		if p.doNonBulkMerges {
			panic("not implemented yet")
		} else {
			merges = append(merges, NewOneMerge(segments[:segsToMerge]))
		}
	}
	return MergeSpecification(merges), nil
}

func (p *MockRandomMergePolicy) FindForcedMerges(segmentsInfos *SegmentInfos,
	maxSegmentCount int, segmentsToMerge map[*SegmentCommitInfo]bool,
	writer *IndexWriter) (MergeSpecification, error) {
	panic("not implemented yet")
}

func (p *MockRandomMergePolicy) Close() error { return nil }

func (p *MockRandomMergePolicy) UseCompoundFile(infos *SegmentInfos,
	mergedInfo *SegmentCommitInfo, writer *IndexWriter) (bool, error) {
	// 80% of the time we create CFS:
	return p.random.Intn(5) != 1, nil
}

// index/AlcoholicMergePolicy.java

/*
Merge policy for testing, it is like an alcoholic. It drinks (merges)
at night, and randomly decides what to drink. During the daytime it
sleeps.

If tests pass with this, then they are likely to pass with any
bizarro merge policy users might write.

It is a fine bottle of champagne (Ordered by Martijn)
*/
type AlcoholicMergePolicy struct {
	*LogMergePolicy
	random *rand.Rand
}

func NewAlcoholicMergePolicy( /*tz TimeZone, */ r *rand.Rand) *AlcoholicMergePolicy {
	mp := &AlcoholicMergePolicy{
		NewLogMergePolicy(0, int64(tu.NextInt(r, 1024*1024, int(math.MaxInt32)))), r}
	mp.SizeSPI = mp
	return mp
}

func (p *AlcoholicMergePolicy) Size(info *SegmentCommitInfo,
	writer *IndexWriter) (int64, error) {

	n, err := info.SizeInBytes()
	now := time.Now()
	if hour := now.Hour(); err == nil && (hour < 6 || hour > 20 ||
		p.random.Intn(23) == 5) { // it's 5 o'clock somewhere
		// pick a random drink during the day
		return drinks[p.random.Intn(5)] * n, nil
	}
	return n, err
}

var drinks = []int64{15, 17, 21, 22, 30}
