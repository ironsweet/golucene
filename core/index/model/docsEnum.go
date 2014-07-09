package model

import (
	. "github.com/balzaczyy/golucene/core/search/model"
)

const (
	DOCS_ENUM_FLAG_FREQS = 1
)

type DocsEnum interface {
	DocIdSetIterator
	/**
	 * Returns term frequency in the current document, or 1 if the field was
	 * indexed with {@link IndexOptions#DOCS_ONLY}. Do not call this before
	 * {@link #nextDoc} is first called, nor after {@link #nextDoc} returns
	 * {@link DocIdSetIterator#NO_MORE_DOCS}.
	 *
	 * <p>
	 * <b>NOTE:</b> if the {@link DocsEnum} was obtain with {@link #FLAG_NONE},
	 * the result of this method is undefined.
	 */
	Freq() (n int, err error)
}
