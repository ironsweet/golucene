package index

import (
	"github.com/balzaczyy/golucene/core/index/model"
)

type DocFieldConsumerPerField interface {
	abort()
	fieldInfo() *model.FieldInfo
}
