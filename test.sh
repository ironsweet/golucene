export tests_codec=Lucene45

go test github.com/balzaczyy/golucene/core/analysis
go test github.com/balzaczyy/golucene/core/codec
go test github.com/balzaczyy/golucene/core/util
# go test github.com/balzaczyy/golucene/core/util/automaton
go test github.com/balzaczyy/golucene/core/util/fst
go test github.com/balzaczyy/golucene/core/util/packed
go test github.com/balzaczyy/golucene/core/store
go test github.com/balzaczyy/golucene/core/index
go test github.com/balzaczyy/golucene/core/search
go test github.com/balzaczyy/golucene/core_test
