package index

import (
	"github.com/balzaczyy/golucene/core/util"
)

// L600
// if you increase this, you must fix field cache impl for
// Terms/TermsIndex requires <= 32768
const MAX_TERM_LENGTH_UTF8 = util.BYTE_BLOCK_SIZE - 2

// index/FrozenBufferedDeletes.java

/*
Holds buffered deletes by term or query, once pushed. Pushed delets
are write-once, so we shift to more memory efficient data structure
to hold them. We don't hold docIDs because these are applied on flush.
*/
type FrozenBufferedDeletes struct {
}
