package store

import (
	"github.com/balzaczyy/golucene/core/util"
	"io"
)

// store/IndexOutput.java

type IndexOutput interface {
	io.Closer
	util.DataOutput
}

type IndexOutputImpl struct {
	*util.DataOutputImpl
}

func newIndexOutput(part util.DataWriter) *IndexOutputImpl {
	return &IndexOutputImpl{util.NewDataOutput(part)}
}
