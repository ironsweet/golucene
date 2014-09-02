package tokenattributes

import (
	"github.com/balzaczyy/golucene/core/util"
)

type PositionLengthAttribute interface {
	util.Attribute
	SetPositionLength(int)
}
