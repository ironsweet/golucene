package search

type Weight interface {
	ValueForNormalization() float
	Normalize(norm, topLevelBoost float) float
}
