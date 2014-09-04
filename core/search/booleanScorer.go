package search

type BooleanScorer struct {
	*BulkScorerImpl
}

func newBooleanScorer(weight *BooleanWeight,
	disableCoord bool, minNrShouldMatch int,
	optionalScorers, prohibitedScorers []BulkScorer,
	maxCoord int) (*BooleanScorer, error) {

	panic("not implemented yet")
}

func (s *BooleanScorer) ScoreAndCollectUpto(collector Collector, max int) (bool, error) {
	panic("not implemented yet")
}

func (s *BooleanScorer) String() string {
	panic("not implemented yet")
}
