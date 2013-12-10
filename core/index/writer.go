package index

import (
	"github.com/balzaczyy/golucene/core/store"
)

// index/IndexWriterConfig.java

/*
Holds all the configuration that is used to create an IndexWriter. Once
IndexWriter has been created with this object, changes to this object will not
affect the IndexWriter instance. For that, use LiveIndexWriterConfig that is
returned from IndexWriter.Config().

All setter methods return IndexWriterConfig to allow chaining settings
conveniently, for example:

		conf := NewIndexWriterConfig(analyzer)
						.setter1()
						.setter2()
*/
type IndexWriterConfig struct {
}

// index/IndexWriter.java

type IndexWriter struct {
}

/*
Constructs a new IndexWriter per the settings given in conf. If you want to
make "live" changes to this writer instance, use Config().

NOTE: after this writer is created, the given configuration instance cannot be
passed to another writer. If you intend to do so, you should clone it
beforehand.
*/
func NewIndexWriter(d store.Directory, conf *IndexWriterConfig) (w *IndexWriter, err error) {
	panic("not implemented yet")
}

// L906
func (w *IndexWriter) Close() error {
	panic("not implemented yet")
}

// L1201
func (w *IndexWriter) AddDocument(doc []IndexableField) error {
	panic("not implemented yet")
}
