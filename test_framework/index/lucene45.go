package index

// index/RandomCodec.java

/*
Codec that assigns per-field random postings format.

The same field/format assignment will happen regardless of order, a
hash is computed up front that determines the mapping. This means
fields can be put into things like HashSets and added to documents
in different orders and the tests will still be deterministic and
reproducible.
*/

type RandomCodec struct{}
