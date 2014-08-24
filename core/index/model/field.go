package model

import (
	"github.com/balzaczyy/golucene/core/analysis"
	"io"
)

// index/IndexableField.java

// TODO: how to handle versioning here...?

// TODO: we need to break out separate StoredField...

/** Represents a single field for indexing.  IndexWriter
 *  consumes Iterable&lt;IndexableField&gt; as a document.
 *
 *  @lucene.experimental */
type IndexableField interface {
	/** Field name */
	Name() string
	/** {@linkmodel.IndexableFieldType} describing the properties
	 * of this field. */
	FieldType() IndexableFieldType
	/**
	 * Returns the field's index-time boost.
	 * <p>
	 * Only fields can have an index-time boost, if you want to simulate
	 * a "document boost", then you must pre-multiply it across all the
	 * relevant fields yourself.
	 * <p>The boost is used to compute the norm factor for the field.  By
	 * default, in the {@link Similarity#computeNorm(FieldInvertState)} method,
	 * the boost value is multiplied by the length normalization factor and then
	 * rounded by {@link DefaultSimilarity#encodeNormValue(float)} before it is stored in the
	 * index.  One should attempt to ensure that this product does not overflow
	 * the range of that encoding.
	 * <p>
	 * It is illegal to return a boost other than 1.0f for a field that is not
	 * indexed ({@linkmodel.IndexableFieldType#indexed()} is false) or omits normalization values
	 * ({@linkmodel.IndexableFieldType#omitNorms()} returns true).
	 *
	 * @see Similarity#computeNorm(FieldInvertState)
	 * @see DefaultSimilarity#encodeNormValue(float)
	 */
	Boost() float32
	/** Non-null if this field has a binary value */
	BinaryValue() []byte

	/** Non-null if this field has a string value */
	StringValue() string

	/** Non-null if this field has a Reader value */
	ReaderValue() io.RuneReader

	/** Non-null if this field has a numeric value */
	NumericValue() interface{}

	// Creates the TokenStream used for indexing this field.  If appropriate,
	// implementations should use the given Analyzer to create the TokenStreams.
	TokenStream(analysis.Analyzer, analysis.TokenStream) (analysis.TokenStream, error)
}
