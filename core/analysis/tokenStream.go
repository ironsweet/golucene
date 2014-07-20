package analysis

import (
	. "github.com/balzaczyy/golucene/core/analysis/tokenattributes"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

/**
 * A <code>TokenStream</code> enumerates the sequence of tokens, either from
 * {@link Field}s of a {@link Document} or from query text.
 * <p>
 * This is an abstract class; concrete subclasses are:
 * <ul>
 * <li>{@link Tokenizer}, a <code>TokenStream</code> whose input is a Reader; and
 * <li>{@link TokenFilter}, a <code>TokenStream</code> whose input is another
 * <code>TokenStream</code>.
 * </ul>
 * A new <code>TokenStream</code> API has been introduced with Lucene 2.9. This API
 * has moved from being {@link Token}-based to {@link Attribute}-based. While
 * {@link Token} still exists in 2.9 as a convenience class, the preferred way
 * to store the information of a {@link Token} is to use {@link AttributeImpl}s.
 * <p>
 * <code>TokenStream</code> now extends {@link AttributeSource}, which provides
 * access to all of the token {@link Attribute}s for the <code>TokenStream</code>.
 * Note that only one instance per {@link AttributeImpl} is created and reused
 * for every token. This approach reduces object creation and allows local
 * caching of references to the {@link AttributeImpl}s. See
 * {@link #incrementToken()} for further details.
 * <p>
 * <b>The workflow of the new <code>TokenStream</code> API is as follows:</b>
 * <ol>
 * <li>Instantiation of <code>TokenStream</code>/{@link TokenFilter}s which add/get
 * attributes to/from the {@link AttributeSource}.
 * <li>The consumer calls {@link TokenStream#reset()}.
 * <li>The consumer retrieves attributes from the stream and stores local
 * references to all attributes it wants to access.
 * <li>The consumer calls {@link #incrementToken()} until it returns false
 * consuming the attributes after each call.
 * <li>The consumer calls {@link #end()} so that any end-of-stream operations
 * can be performed.
 * <li>The consumer calls {@link #close()} to release any resource when finished
 * using the <code>TokenStream</code>.
 * </ol>
 * To make sure that filters and consumers know which attributes are available,
 * the attributes must be added during instantiation. Filters and consumers are
 * not required to check for availability of attributes in
 * {@link #incrementToken()}.
 * <p>
 * You can find some example code for the new API in the analysis package level
 * Javadoc.
 * <p>
 * Sometimes it is desirable to capture a current state of a <code>TokenStream</code>,
 * e.g., for buffering purposes (see {@link CachingTokenFilter},
 * TeeSinkTokenFilter). For this usecase
 * {@link AttributeSource#captureState} and {@link AttributeSource#restoreState}
 * can be used.
 * <p>The {@code TokenStream}-API in Lucene is based on the decorator pattern.
 * Therefore all non-abstract subclasses must be final or have at least a final
 * implementation of {@link #incrementToken}! This is checked when Java
 * assertions are enabled.
 */
type TokenStream interface {
	// Releases resouces associated with this stream.
	//
	// If you override this method, always call TokenStreamImpl.Close(),
	// otherwise some internal state will not be correctly reset (e.g.,
	// Tokenizer will panic on reuse).
	io.Closer
	Attributes() *util.AttributeSource
	// Consumers (i.e., IndexWriter) use this method to advance the
	// stream to the next token. Implementing classes must implement
	// this method and update the appropriate AttributeImpls with the
	// attributes of he next token.
	//
	// The producer must make no assumptions about the attributes after
	// the method has been returned: the caller may arbitrarily change
	// it. If the producer needs to preserve the state for subsequent
	// calls, it can use captureState to create a copy of the current
	// attribute state.
	//
	// This method is called for every token of a docuent, so an
	// efficient implementation is crucial for good performance.l To
	// avoid calls to AddAttribute(clas) and GetAttribute(Class),
	// references to all AttributeImpls that this stream uses should be
	// retrived during instantiation.
	//
	// To ensure that filters and consumers know which attributes are
	// available, the attributes must be added during instantiation.
	// Filters and consumers are not required to check for availability
	// of attribute in IncrementToken().
	IncrementToken() (bool, error)
	// This method is called by the consumer after the last token has
	// been consumed, after IncrementToken() returned false (using the
	// new TokenSream API). Streams implementing the old API should
	// upgrade to use this feature.
	//
	// This method can be used to perform any end-of-stream operations,
	// such as setting the final offset of a stream. The final offset
	// of a stream might difer from the offset of the last token eg in
	// case one or more whitespaces followed after the last token, but
	// a WhitespaceTokenizer was used.
	//
	// Additionally any skipped positions (such as those removed by a
	// stopFilter) can be applied to the position increment, or any
	// adjustment or other attributes where the end-of-stream value may
	// be important.
	//
	// If you override this method, alwasy call TokenStreamImpl.End().
	End() error
	// This method is called by a consumer before it begins consumption
	// using IncrementToken().
	//
	// Resets this stream to a clean state. Stateful implementation
	// must implement this method so that they can be reused, just as
	// if they had been created fresh.
	//
	// If you override this method, alwasy call TokenStreamImpl.Reset(),
	// otherwise some internal state will not be correctly reset (e.g.,
	// Tokenizer will panic on further usage).
	Reset() error
}

type TokenStreamImpl struct {
	atts *util.AttributeSource
}

var DEFAULT_TOKEN_ATTRIBUTE_FACTORY = assembleAttributeFactory(
	DEFAULT_ATTRIBUTE_FACTORY,
	map[string]bool{
		"CharTermAttribute":          true,
		"TermToBytesRefAttribute":    true,
		"TypeAttribute":              true,
		"PositionIncrementAttribute": true,
		"PositionLengthAttribute":    true,
		"OffsetAttribute":            true,
	},
	NewPackedTokenAttribute,
)

/* A TokenStream using the default attribute factory. */
func NewTokenStream() *TokenStreamImpl {
	return &TokenStreamImpl{
		atts: util.NewAttributeSourceWith(DEFAULT_TOKEN_ATTRIBUTE_FACTORY),
	}
}

/* A TokenStream that uses the same attributes as the supplied one. */
func NewTokenStreamWith(input *util.AttributeSource) *TokenStreamImpl {
	return &TokenStreamImpl{
		atts: util.NewAttributeSourceFrom(input),
	}
}

func (ts *TokenStreamImpl) Attributes() *util.AttributeSource { return ts.atts }
func (ts *TokenStreamImpl) End() error {
	ts.atts.Clear() // LUCENE-3849: don't consume dirty atts
	if posIncAtt := ts.atts.Get("PositionIncrementAttribute").(PositionIncrementAttribute); posIncAtt != nil {
		posIncAtt.SetPositionIncrement(0)
	}
	return nil
}
func (ts *TokenStreamImpl) Reset() error { return nil }
func (ts *TokenStreamImpl) Close() error { return nil }
