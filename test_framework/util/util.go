package util

import (
	. "github.com/balzaczyy/gounit"
	"io/ioutil"
	"math/rand"
	"os"
)

// util/_TestUtil.java

// Returns a temp directory, based on the given description. Creates the directory.
func TempDir(desc string) string {
	if len(desc) < 3 {
		panic("description must be at least 3 characters")
	}
	// Ian: I prefer Go's own way to obtain temp folder instead of Lucene's method
	f, err := ioutil.TempDir("", desc)
	if err != nil {
		panic(err)
	}
	f += ".tmp" // add suffix
	CloseAfterSuite(NewCloseableFile(f, SuiteFailureMarker))
	return f
}

// L264
func NextInt(r *rand.Rand, start, end int) int {
	return r.Intn(end-start) + start
}

// L314
// Returns random string, including full unicode range.
func RandomUnicodeString(r *rand.Rand) string {
	return randomUnicodeStringLength(r, 20)
}

// Returns a random string up to a certain length.
func randomUnicodeStringLength(r *rand.Rand, maxLength int) string {
	end := NextInt(r, 0, maxLength)
	if end == 0 {
		// allow 0 length
		return ""
	}
	buffer := make([]rune, end)
	randomFixedLengthUnicodeString(r, buffer)
	return string(buffer)
}

// Fills provided []rune with valid random unicode code unit sequence.
func randomFixedLengthUnicodeString(random *rand.Rand, chars []rune) {
	for i, length := 0, len(chars); i < length; i++ {
		t := random.Intn(5)
		if t == 0 && i < length-1 {
			// Make a surrogate pair
			// High surrogate
			chars[i] = rune(NextInt(random, 0xd800, 0xdbff))
			// Low surrogate
			i++
			chars[i] = rune(NextInt(random, 0xdc00, 0xdfff))
		} else if t <= 1 {
			chars[i] = rune(random.Intn(0x80))
		} else if t == 2 {
			chars[i] = rune(NextInt(random, 0x80, 0x7ff))
		} else if t == 3 {
			chars[i] = rune(NextInt(random, 0x800, 0xd7ff))
		} else if t == 4 {
			chars[i] = rune(NextInt(random, 0xe000, 0xffff))
		}
	}
}

// util/TestRuleMarkFailure.java

type TestRuleMarkFailure struct {
	failures bool
}

// Check if this object had any marked failures.
func (tr *TestRuleMarkFailure) hadFailures() bool {
	return tr.failures
}

// Check if this object was sucessful
func (tr *TestRuleMarkFailure) WasSuccessful() bool {
	return tr.hadFailures()
}

// util/CloseableFile.java

// A Closeable that attempts to remove a given file/folder
func NewCloseableFile(file string, failureMarker *TestRuleMarkFailure) func() error {
	return func() error {
		// only if there were no other test failures.
		if failureMarker.WasSuccessful() {
			os.RemoveAll(file) // ignore any error
			// no re-check
		}
		return nil
	}
}
