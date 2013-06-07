package search

type Weight interface {
	ValueForNormalization() float32
	Normalize(norm, topLevelBoost float32) float32
}
