package model

// TermState.java
// Encapsulates all requried internal state to postiion the associated
// termsEnum without re-seeking
type TermState interface {
	CopyFrom(other TermState)
	Clone() TermState
}
