export tests_codec=Lucene49

go test github.com/balzaczyy/golucene/core/util
go test github.com/balzaczyy/golucene/core/util/automaton
go test github.com/balzaczyy/golucene/core/util/fst
go test github.com/balzaczyy/golucene/core/util/packed
go test github.com/balzaczyy/golucene/core/analysis/tokenattributes
go test github.com/balzaczyy/golucene/core/analysis
go test github.com/balzaczyy/golucene/core/document
go test github.com/balzaczyy/golucene/core/store
go test github.com/balzaczyy/golucene/core/index/model
go test github.com/balzaczyy/golucene/core/search/model
go test github.com/balzaczyy/golucene/core/codec
go test github.com/balzaczyy/golucene/core/codec/spi
go test github.com/balzaczyy/golucene/core/codec/blocktree
go test github.com/balzaczyy/golucene/core/codec/lucene42
go test github.com/balzaczyy/golucene/core/index
go test github.com/balzaczyy/golucene/core/search
go test github.com/balzaczyy/golucene/core_test
