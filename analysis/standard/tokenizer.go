package standard

import (
	. "github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

/*
A grammar-based tokenizer constructed with JFlex.

As of Lucene version 3.1, this class implements the Word Break rules
from the Unicode Text Segmentation algorithm, as specified in Unicode
standard Annex #29.

Many applications have specific tokenizer needs. If this tokenizer
does not suit your application, please consider copying this source
code directory to your project and maintaining your own grammar-based
tokenizer.

Version

You must specify the required Version compatibility when creating
StandardTokenizer:

	- As of 3.4, Hiragana and Han characters are no longer wrongly
	split from their combining characters. If you use a previous
	version number, you get the exact broken behavior for backwards
	compatibility.
	- As of 3.1, StandardTokenizer implements Unicode text segmentation.
	If you use a previous version number, you get the exact behavior of
	ClassicTokenizer for backwards compatibility.
*/
type StandardTokenizer struct {
	*Tokenizer
	maxTokenLength int
}

/*
Creates a new instance of the StandardTokenizer. Attaches the input
to the newly created JFlex scanner.
*/
func newStandardTokenizer(matchVersion util.Version, input io.ReadCloser) *StandardTokenizer {
	panic("not implemented yet")
}

func (t *StandardTokenizer) IncrementToken() (bool, error) {
	panic("not implemented yet")
}

func (t *StandardTokenizer) End() error {
	panic("not implemented yet")
}

func (t *StandardTokenizer) Reset() error {
	panic("not implemented yet")
}
