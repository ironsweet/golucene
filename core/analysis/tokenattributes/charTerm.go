package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

/* The term text of a Token. */
type CharTermAttribute interface {
	// Copies the contents of buffer into the termBuffer array
	CopyBuffer(buffer []rune)
	// Returns the internal termBuffer rune slice which you can then
	// directly alter. If the slice is too small for your token, use
	// ResizeBuffer(int) to increase it. After altering the buffer, be
	// sure to call SetLength() to record the number of valid runes
	// that were placed into the termBuffer.
	//
	// NOTE: the returned buffer may be larger than the valid Length().
	Buffer() []rune
	Length() int
	// Appends teh specified string to this character sequence.
	//
	// The character of the string argument are appended, in order,
	// increasing the length of this sequence by the length of the
	// argument. If argument is "", then the three characters "nil" are
	// appended.
	AppendString(string) CharTermAttribute
}

const MIN_BUFFER_SIZE = 10

/* Default implementation of CharTermAttribute. */
type CharTermAttributeImpl struct {
	termBuffer []rune
	termLength int
	bytes      *util.BytesRefBuilder
}

func newCharTermAttributeImpl() *CharTermAttributeImpl {
	return &CharTermAttributeImpl{
		termBuffer: make([]rune, util.Oversize(MIN_BUFFER_SIZE, util.NUM_BYTES_CHAR)),
		bytes:      util.NewBytesRefBuilder(),
	}
}

func (a *CharTermAttributeImpl) Interfaces() []string {
	return []string{"CharTermAttribute", "TermToBytesRefAttribute"}
}

func (a *CharTermAttributeImpl) CopyBuffer(buffer []rune) {
	a.growTermBuffer(len(buffer))
	copy(a.termBuffer, buffer)
	a.termLength = len(buffer)
}

func (a *CharTermAttributeImpl) Buffer() []rune {
	return a.termBuffer
}

func (a *CharTermAttributeImpl) growTermBuffer(newSize int) {
	if len(a.termBuffer) < newSize {
		// not big enough: create a new slice with slight over allocation:
		a.termBuffer = make([]rune, util.Oversize(newSize, util.NUM_BYTES_CHAR))
	}
}

func (a *CharTermAttributeImpl) FillBytesRef() {
	s := string(a.termBuffer[:a.termLength])
	a.bytes.Copy([]byte(s))
}

func (a *CharTermAttributeImpl) BytesRef() *util.BytesRef {
	return a.bytes.Get()
}

func (a *CharTermAttributeImpl) Length() int {
	return a.termLength
}

func (a *CharTermAttributeImpl) AppendString(s string) CharTermAttribute {
	if s == "" { // needed for Appendable compliance
		return a.appendNil()
	}
	for _, ch := range s {
		if a.termLength < len(a.termBuffer) {
			a.termBuffer[a.termLength] = ch
		} else {
			a.termBuffer = append(a.termBuffer, ch)
		}
		a.termLength++
	}
	return a
}

func (a *CharTermAttributeImpl) appendNil() CharTermAttribute {
	a.termBuffer = append(a.termBuffer, 'n')
	a.termBuffer = append(a.termBuffer, 'i')
	a.termBuffer = append(a.termBuffer, 'l')
	a.termLength += 3
	return a
}

func (a *CharTermAttributeImpl) Clear() {
	a.termLength = 0
}

func (a *CharTermAttributeImpl) Clone() util.AttributeImpl {
	clone := new(CharTermAttributeImpl)
	// do a deep clone
	clone.termBuffer = make([]rune, a.termLength)
	copy(clone.termBuffer, a.termBuffer[:a.termLength])
	clone.termLength = a.termLength
	clone.bytes = util.NewBytesRefBuilder()
	clone.bytes.Copy(a.bytes.Bytes())
	return clone
}

func (a *CharTermAttributeImpl) String() string {
	return string(a.termBuffer[:a.termLength])
}

func (a *CharTermAttributeImpl) CopyTo(target util.AttributeImpl) {
	target.(CharTermAttribute).CopyBuffer(a.termBuffer[:a.termLength])
}
