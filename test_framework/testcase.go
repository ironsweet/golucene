package test_framework

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

// --------------------------------------------------------------------
// Test groups, system properties and other annotations modifying tests
// --------------------------------------------------------------------
const (
	SYSPROP_NIGHTLY = "tests.nightly"
)

// -----------------------------------------------------------------
// Truly immutable fields and constants, initialized once and valid
// for all suites ever since.
// -----------------------------------------------------------------

// Use this constant then creating Analyzers and any other version-dependent
// stuff. NOTE: Change this when developmenet starts for new Lucene version:
const TEST_VERSION_CURRENT = util.VERSION_45

// True if and only if tests are run in verbose mode. If this flag is false
// tests are not expected toprint and messages.
var VERBOSE = ("false" == or(os.Getenv("tests.verbose"), "false"))

// A random multiplier which you should use when writing random tests:
// multiply it by the number of iterations to scale your tests (for nightly builds).
var RANDOM_MULTIPLIER = func() int {
	n, err := strconv.Atoi(or(os.Getenv("tests.multiplier"), "1"))
	if err != nil {
		panic(err)
	}
	return n
}()

// Gets the directory to run tests with
var TEST_DIRECTORY = or(os.Getenv("tests.directory"), "random")

// Whether or not Nightly tests should run
var TEST_NIGHTLY = ("true" == or(os.Getenv(SYSPROP_NIGHTLY), "false"))

// Throttling
var TEST_THROTTLING = either(TEST_NIGHTLY, THROTTLING_SOMETIMES, THROTTLING_NEVER).(Throttling)

func or(a, b string) string {
	if len(a) > 0 {
		return a
	}
	return b
}

// L300

// -----------------------------------------------------------------
// Class level (suite) rules.
// -----------------------------------------------------------------

var suiteFailureMarker = &TestRuleMarkFailure{}

// Ian: I have to extend Go's testing framework to simulate JUnit's
// TestRule
func wrapTesting(t *testing.T) *T {
	ans := wrapT(t)
	suiteFailureMarker.T = ans
	return ans
}

var suiteClosers []func() error

type T struct {
	delegate *testing.T
}

func wrapT(t *testing.T) *T {
	return &T{t}
}

func (c *T) Error(args ...interface{}) {
	c.delegate.Error(args)
}

func (c *T) afterSuite() {
	for _, closer := range suiteClosers {
		closer() // ignore error
	}
}

// -----------------------------------------------------------------
// Test facilities and facades for subclasses.
// -----------------------------------------------------------------

/*
Note it's different from Lucene's Randomized Test Runner.

There is an overhead connected with getting the Random for a particular
context and thread. It is better to cache this Random locally if tight loops
with multiple invocations are present or create a derivative local Random for
millions of calls like this:

		r := rand.New(rand.NewSource(99))
		// tight loop with many invocations.

*/
func Random() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
}

// Registers a Closeable resource that shold be closed after the suite completes.
func closeAfterSuite(closer func() error) {
	suiteClosers = append(suiteClosers, closer)
}

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
	return wrapDirectory(r, newDirectoryImpl(r, TEST_DIRECTORY), rarely(r))
}

func wrapDirectory(random *rand.Rand, directory store.Directory, bare bool) BaseDirectoryWrapper {
	if rarely(random) {
		directory = store.NewNRTCachingDirectory(directory, random.Float64(), random.Float64())
	}

	if rarely(random) {
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
		closeAfterSuite(NewCloseableDirectory(base, suiteFailureMarker))
		return base
	} else {
		mock := NewMockDirectoryWrapper(random, directory)

		mock.SetThrottling(TEST_THROTTLING)
		closeAfterSuite(NewCloseableDirectory(mock, suiteFailureMarker))
		return mock
	}
}

// L643

/*
Returns a number of at least i

The actual number returned will be influenced by whether TEST_NIGHTLY
is active and RANDOM_MULTIPLIER, but also with some random fudge.
*/
func atLeastBy(random *rand.Rand, i int) int {
	min := i * RANDOM_MULTIPLIER
	if TEST_NIGHTLY {
		min = 2 * min
	}
	max := min + min/2
	return NextInt(random, min, max)
}

func AtLeast(i int) int {
	return atLeastBy(Random(), i)
}

/*
Returns true if something should happen rarely,

The actual number returned will be influenced by whether TEST_NIGHTLY
is active and RANDOM_MULTIPLIER
*/
func rarely(random *rand.Rand) bool {
	p := either(TEST_NIGHTLY, 10, 1).(int)
	p += int(float64(p) * math.Log(float64(RANDOM_MULTIPLIER)))
	if p < 50 {
		p = 50
	}
	min := 100 - p // never more than 50
	return random.Intn(100) >= min
}

func either(flag bool, value, orValue interface{}) interface{} {
	if flag {
		return value
	}
	return orValue
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
		if rarely(random) {
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
