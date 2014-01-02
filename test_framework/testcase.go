package test_framework

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	. "github.com/balzaczyy/golucene/test_framework/util"
	"log"
	"math/rand"
	"os"
)

// --------------------------------------------------------------------
// Test groups, system properties and other annotations modifying tests
// --------------------------------------------------------------------

// -----------------------------------------------------------------
// Truly immutable fields and constants, initialized once and valid
// for all suites ever since.
// -----------------------------------------------------------------

// Use this constant then creating Analyzers and any other version-dependent
// stuff. NOTE: Change this when developmenet starts for new Lucene version:
const TEST_VERSION_CURRENT = util.VERSION_45

// Throttling
var TEST_THROTTLING = either(TEST_NIGHTLY, THROTTLING_SOMETIMES, THROTTLING_NEVER).(Throttling)

func either(flag bool, value, orValue interface{}) interface{} {
	if flag {
		return value
	}
	return orValue
}

// L300

// -----------------------------------------------------------------
// Class level (suite) rules.
// -----------------------------------------------------------------

// -----------------------------------------------------------------
// Test facilities and facades for subclasses.
// -----------------------------------------------------------------

// Create a new index writer config with random defaults
func NewIndexWriterConfig(v util.Version, a analysis.Analyzer) *index.IndexWriterConfig {
	panic("not implemented yet")
}

/*
Returns a new Direcotry instance. Use this when the test does not care about
the specific Directory implementation (most tests).

The Directory is wrapped with BaseDirectoryWrapper. This menas usually it
will be picky, such as ensuring that you properly close it and all open files
in your test. It will emulate some features of Windows, such as not allowing
open files ot be overwritten.
*/
func NewDirectory() BaseDirectoryWrapper {
	return newDirectoryWithSeed(Random())
}

// Returns a new Directory instance, using the specified random.
// See NewDirecotry() for more information
func newDirectoryWithSeed(r *rand.Rand) BaseDirectoryWrapper {
	return wrapDirectory(r, newDirectoryImpl(r, TEST_DIRECTORY), Rarely(r))
}

func wrapDirectory(random *rand.Rand, directory store.Directory, bare bool) BaseDirectoryWrapper {
	if Rarely(random) {
		directory = store.NewNRTCachingDirectory(directory, random.Float64(), random.Float64())
	}

	if Rarely(random) {
		maxMBPerSec := 10 + 5*(random.Float64()-0.5)
		if VERBOSE {
			log.Printf("LuceneTestCase: will rate limit output IndexOutput to %v MB/sec", maxMBPerSec)
		}
		rateLimitedDirectoryWrapper := store.NewRateLimitedDirectoryWrapper(directory)
		switch random.Intn(10) {
		case 3: // sometimes rate limit on flush
			rateLimitedDirectoryWrapper.SetMaxWriteMBPerSec(maxMBPerSec, store.IO_CONTEXT_TYPE_FLUSH)
		case 2: // sometimes rate limit flush & merge
			rateLimitedDirectoryWrapper.SetMaxWriteMBPerSec(maxMBPerSec, store.IO_CONTEXT_TYPE_FLUSH)
			rateLimitedDirectoryWrapper.SetMaxWriteMBPerSec(maxMBPerSec, store.IO_CONTEXT_TYPE_MERGE)
		default:
			rateLimitedDirectoryWrapper.SetMaxWriteMBPerSec(maxMBPerSec, store.IO_CONTEXT_TYPE_MERGE)
		}
		directory = rateLimitedDirectoryWrapper
	}

	if bare {
		base := NewBaseDirectoryWrapper(directory)
		CloseAfterSuite(NewCloseableDirectory(base, SuiteFailureMarker))
		return base
	} else {
		mock := NewMockDirectoryWrapper(random, directory)

		mock.SetThrottling(TEST_THROTTLING)
		CloseAfterSuite(NewCloseableDirectory(mock, SuiteFailureMarker))
		return mock
	}
}

// L1064
func NewTextField(name, value string, stored bool) *index.Field {
	flag := index.TEXT_FIELD_TYPE_STORED
	if !stored {
		flag = index.TEXT_FIELD_TYPE_NOT_STORED
	}
	return NewField(Random(), name, value, flag)
}

func NewField(r *rand.Rand, name, value string, typ *index.FieldType) *index.Field {
	panic("not implemented yet")
}

// Ian: Different from Lucene's default random class initializer, I have to
// explicitly initialize different directory randomly.
func newDirectoryImpl(random *rand.Rand, clazzName string) store.Directory {
	if clazzName == "random" {
		if Rarely(random) {
			switch random.Intn(1) {
			case 0:
				clazzName = "SimpleFSDirectory"
			}
		} else {
			clazzName = "RAMDirectory"
		}
	}
	if clazzName == "RAMDirectory" {
		return store.NewRAMDirectory()
	} else {
		path := TempDir("index")
		if err := os.MkdirAll(path, os.ModeTemporary); err != nil {
			panic(err)
		}
		switch clazzName {
		case "SimpleFSDirectory":
			d, err := store.NewSimpleFSDirectory(path)
			if err != nil {
				panic(err)
			}
			return d
		}
		panic(fmt.Sprintf("not supported yet: %v", clazzName))
	}
}

// L1305
// Create a new searcher over the reader. This searcher might randomly use threads
func NewSearcher(r index.IndexReader) *search.IndexSearcher {
	panic("not implemented yet")
}
