package index

import (
	"github.com/balzaczyy/golucene/core/store"
)

/* Prefix codes term instances (prefixes are shared) */
type PrefixCodedTerms struct {
	buffer *store.RAMFile
}

func newPrefixCodedTerms(buffer *store.RAMFile) *PrefixCodedTerms {
	return &PrefixCodedTerms{buffer}
}

func (terms *PrefixCodedTerms) RamBytesUsed() int64 {
	return terms.buffer.RamBytesUsed()
}

/* Builds a PrefixCodedTerms: call add repeatedly, then finish. */
type PrefixCodedTermsBuilder struct {
	buffer *store.RAMFile
	output *store.RAMOutputStream
}

func newPrefixCodedTermsBuilder() *PrefixCodedTermsBuilder {
	f := store.NewRAMFileBuffer()
	return &PrefixCodedTermsBuilder{
		buffer: f,
		output: store.NewRAMOutputStream(f, false),
	}
}

func (b *PrefixCodedTermsBuilder) add(term *Term) {
	panic("not implemented yet")
}

func (b *PrefixCodedTermsBuilder) finish() *PrefixCodedTerms {
	err := b.output.Close()
	if err != nil {
		panic(err)
	}
	return newPrefixCodedTerms(b.buffer)
}
