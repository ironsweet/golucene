package codec

/* Holder for per-term statistics. */
type TermStats struct {
	// How many documents have at least one occurrence of this term.
	DocFreq int
	// Total number of times this term occurs across all documents in
	// the field.
	TotalTermFreq int64
}

func NewTermStats(docFreq int, totalTermFreq int64) *TermStats {
	return &TermStats{docFreq, totalTermFreq}
}
