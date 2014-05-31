package analysis

type CharFilterService interface {
	// Chains the corrected offset through the input CharFilter(s).
	CorrectOffset(int) int
}
