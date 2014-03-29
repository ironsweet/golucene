package index

/* Prefix codes term instances (prefixes are shared) */
type PrefixCodedTerms struct{}

func (terms *PrefixCodedTerms) sizeInBytes() int64 {
	panic("not implemented yet")
}

/* Builds a PrefixCodedTerms: call add repeatedly, then finish. */
type PrefixCodedTermsBuilder struct {
}

func newPrefixCodedTermsBuilder() *PrefixCodedTermsBuilder {
	panic("not implemented yet")
}

func (b *PrefixCodedTermsBuilder) add(term *Term) {
	panic("not implemented yet")
}

func (b *PrefixCodedTermsBuilder) finish() *PrefixCodedTerms {
	panic("not implemented yet")
}
