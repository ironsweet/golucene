package model

type Terms interface {
	Iterator(reuse TermsEnum) TermsEnum
	DocCount() int
	SumTotalTermFreq() int64
	SumDocFreq() int64
}
