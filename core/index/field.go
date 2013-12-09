package index

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
	name() string
	/** {@link IndexableFieldType} describing the properties
	 * of this field. */
	fieldType() IndexableFieldType
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
	 * indexed ({@link IndexableFieldType#indexed()} is false) or omits normalization values
	 * ({@link IndexableFieldType#omitNorms()} returns true).
	 *
	 * @see Similarity#computeNorm(FieldInvertState)
	 * @see DefaultSimilarity#encodeNormValue(float)
	 */
	boost() float32
	/** Non-null if this field has a binary value */
	binaryValue() []byte

	/** Non-null if this field has a string value */
	stringValue() string

	/** Non-null if this field has a Reader value */
	readerValue() io.Reader

	/** Non-null if this field has a numeric value */
	// numericValue() uint64

	/**
	 * Creates the TokenStream used for indexing this field.  If appropriate,
	 * implementations should use the given Analyzer to create the TokenStreams.
	 *
	 * @param analyzer Analyzer that should be used to create the TokenStreams from
	 * @return TokenStream value for indexing the document.  Should always return
	 *         a non-null value if the field is to be indexed
	 * @throws IOException Can be thrown while creating the TokenStream
	 */
	tokenStream(analyzer analysis.Analyzer) (ts analysis.TokenStream, err error)
}

// index/IndexableFieldType.java

/**
 * Describes the properties of a field.
 * @lucene.experimental
 */
type IndexableFieldType interface {
	/** True if this field should be indexed (inverted) */
	indexed() bool
	/** True if the field's value should be stored */
	stored() bool
	/**
	 * True if this field's value should be analyzed by the
	 * {@link Analyzer}.
	 * <p>
	 * This has no effect if {@link #indexed()} returns false.
	 */
	tokenized() bool
	/**
	 * True if this field's indexed form should be also stored
	 * into term vectors.
	 * <p>
	 * This builds a miniature inverted-index for this field which
	 * can be accessed in a document-oriented way from
	 * {@link IndexReader#getTermVector(int,String)}.
	 * <p>
	 * This option is illegal if {@link #indexed()} returns false.
	 */
	storeTermVectors() bool
	/**
	 * True if this field's token character offsets should also
	 * be stored into term vectors.
	 * <p>
	 * This option is illegal if term vectors are not enabled for the field
	 * ({@link #storeTermVectors()} is false)
	 */
	storeTermVectorOffsets() bool

	/**
	 * True if this field's token positions should also be stored
	 * into the term vectors.
	 * <p>
	 * This option is illegal if term vectors are not enabled for the field
	 * ({@link #storeTermVectors()} is false).
	 */
	storeTermVectorPositions() bool

	/**
	 * True if this field's token payloads should also be stored
	 * into the term vectors.
	 * <p>
	 * This option is illegal if term vector positions are not enabled
	 * for the field ({@link #storeTermVectors()} is false).
	 */
	storeTermVectorPayloads() bool

	/**
	 * True if normalization values should be omitted for the field.
	 * <p>
	 * This saves memory, but at the expense of scoring quality (length normalization
	 * will be disabled), and if you omit norms, you cannot use index-time boosts.
	 */
	omitNorms() bool

	/** {@link IndexOptions}, describing what should be
	 * recorded into the inverted index */
	indexOptions() IndexOptions

	/**
	 * DocValues {@link DocValuesType}: if non-null then the field's value
	 * will be indexed into docValues.
	 */
	docValueType() DocValuesType
}
