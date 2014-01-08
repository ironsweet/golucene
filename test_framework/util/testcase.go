package util

import (
	"math"
	"math/rand"
	"os"
	"strconv"
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

// Suite failure marker (any error in the test or suite scope)
var SuiteFailureMarker = &TestRuleMarkFailure{}

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
func Rarely(random *rand.Rand) bool {
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
