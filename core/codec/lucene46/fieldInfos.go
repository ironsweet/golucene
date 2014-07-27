package lucene46

import (
	. "github.com/balzaczyy/golucene/core/codec/spi"
	. "github.com/balzaczyy/golucene/core/index/model"
	"github.com/balzaczyy/golucene/core/store"
)

type Lucene46FieldInfosFormat struct {
	r FieldInfosReader
	w FieldInfosWriter
}

func NewLucene46FieldInfosFormat() *Lucene46FieldInfosFormat {
	return &Lucene46FieldInfosFormat{
		r: Lucene46FieldInfosReader,
		w: Lucene46FieldInfosWriter,
	}
}

func (f *Lucene46FieldInfosFormat) FieldInfosReader() FieldInfosReader {
	return f.r
}

func (f *Lucene46FieldInfosFormat) FieldInfosWriter() FieldInfosWriter {
	return f.w
}

var Lucene46FieldInfosReader = func(dir store.Directory,
	segment, suffix string, context store.IOContext) (fi FieldInfos, err error) {
	panic("not implemented yet")
}

var Lucene46FieldInfosWriter = func(dir store.Directory,
	segName, suffix string, infos FieldInfos, ctx store.IOContext) (err error) {
	panic("not implemented yet")
}
