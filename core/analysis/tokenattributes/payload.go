package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

/*
The payload of a Token.

The payload is stored in the index at each position, and can be used
to influence scoring when using Payload-based queries in the payloads
and spans packages.

NOTE: becaues the payload will be stored at each position, its
usually best to use the minimum number of bytes necessary. Some codec
implementations may optimize payload storage when all payloads have
the same length.
*/
type PayloadAttribute interface {
	util.Attribute
	// Returns this Token's payload
	Payload() []byte
	// Sets this Token's payload.
	SetPayload([]byte)
}

/* Default implementation of PayloadAttirbute */
type PayloadAttributeImpl struct {
	payload []byte
}

func newPayloadAttributeImpl() util.AttributeImpl {
	return new(PayloadAttributeImpl)
}

func (a *PayloadAttributeImpl) Interfaces() []string      { return []string{"PayloadAttribute"} }
func (a *PayloadAttributeImpl) Payload() []byte           { return a.payload }
func (a *PayloadAttributeImpl) SetPayload(payload []byte) { a.payload = payload }
func (a *PayloadAttributeImpl) Clear()                    { a.payload = nil }

func (a *PayloadAttributeImpl) Clone() util.AttributeImpl {
	return &PayloadAttributeImpl{
		payload: a.payload,
	}
}

func (a *PayloadAttributeImpl) CopyTo(target util.AttributeImpl) {
	target.(PayloadAttribute).SetPayload(a.payload)
}
