package test_framework

import (
	"fmt"
	"github.com/balzaczyy/golucene/core/analysis"
	. "github.com/balzaczyy/golucene/core/codec/spi"
	docu "github.com/balzaczyy/golucene/core/document"
	"github.com/balzaczyy/golucene/core/index"
	"github.com/balzaczyy/golucene/core/search"
	"github.com/balzaczyy/golucene/core/store"
	"github.com/balzaczyy/golucene/core/util"
	ti "github.com/balzaczyy/golucene/test_framework/index"
	ts "github.com/balzaczyy/golucene/test_framework/search"
	. "github.com/balzaczyy/golucene/test_framework/util"
	. "github.com/balzaczyy/gounit"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
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
var TEST_VERSION_CURRENT = util.VERSION_45

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

// Class environment setup rule.
var ClassEnvRule = &TestRuleSetupAndRestoreClassEnv{}

// -----------------------------------------------------------------
// Test facilities and facades for subclasses.
// -----------------------------------------------------------------

// Create a new index writer config with random defaults
func NewIndexWriterConfig(v util.Version, a analysis.Analyzer) *index.IndexWriterConfig {
	return newRandomIndexWriteConfig(Random(), v, a)
}

// Create a new index write config with random defaults using the specified random
func newRandomIndexWriteConfig(r *rand.Rand, v util.Version, a analysis.Analyzer) *index.IndexWriterConfig {
	c := index.NewIndexWriterConfig(v, a)
	c.SetSimilarity(ClassEnvRule.similarity)
	if VERBOSE {
		// Even through TestRuleSetupAndRestoreClassEnv calls
		// infoStream.SetDefault, we do it again here so that the
		// PrintStreamInfoStream.messageID increments so that when there
		// are separate instance of IndexWriter created we see "IW 0",
		// "IW 1", "IW 2", ... instead of just always "IW 0":
		c.SetInfoStream(newThreadNameFixingPrintStreamInfoStream(os.Stdout))
	}

	if r.Intn(2) == 0 {
		c.SetMergeScheduler(index.NewSerialMergeScheduler())
	} else if Rarely(r) {
		log.Println("Use ConcurrentMergeScheduler")
		maxRoutineCount := NextInt(Random(), 1, 4)
		maxMergeCount := NextInt(Random(), maxRoutineCount, maxRoutineCount+4)
		cms := index.NewConcurrentMergeScheduler()
		cms.SetMaxMergesAndRoutines(maxMergeCount, maxRoutineCount)
		c.SetMergeScheduler(cms)
	}
	if r.Intn(2) == 0 {
		if Rarely(r) {
			log.Println("Use crazy value for buffered docs")
			// crazy value
			c.SetMaxBufferedDocs(NextInt(r, 2, 15))
		} else {
			// reasonable value
			c.SetMaxBufferedDocs(NextInt(r, 16, 1000))
		}
	}
	// Go doesn't need thread-affinity state.
	// if r.Intn(2) == 0 {
	// 	maxNumRoutineState := either(Rarely(r),
	// 		NextInt(r, 5, 20), // crazy value
	// 		NextInt(r, 1, 4))  // reasonable value

	// 	if Rarely(r) {
	// 		// reandom thread pool
	// 		c.SetIndexerThreadPool(newRandomDocumentsWriterPerThreadPool(maxNumRoutineState, r))
	// 	} else {
	// 		// random thread pool
	// 		c.SetMaxThreadStates(maxNumRoutineState)
	// 	}
	// }

	c.SetMergePolicy(newMergePolicy(r))

	if Rarely(r) {
		log.Println("Use SimpleMergedSegmentWarmer")
		c.SetMergedSegmentWarmer(index.NewSimpleMergedSegmentWarmer(c.InfoStream()))
	}
	c.SetUseCompoundFile(r.Intn(2) == 0)
	// c.SetUseCompoundFile(false)
	c.SetReaderPooling(r.Intn(2) == 0)
	c.SetReaderTermsIndexDivisor(NextInt(r, 1, 4))
	return c
}

func newMergePolicy(r *rand.Rand) index.MergePolicy {
	if Rarely(r) {
		log.Println("Use MockRandomMergePolicy")
		return ti.NewMockRandomMergePolicy(r)
	} else if r.Intn(2) == 0 {
		return newTieredMergePolicy(r)
	} else if r.Intn(5) == 0 {
		return newAlcoholicMergePolicy(r /*, ClassEnvRule.timeZone*/)
	} else {
		return newLogMergePolicy(r)
	}
}

// L883
func newTieredMergePolicy(r *rand.Rand) *index.TieredMergePolicy {
	tmp := index.NewTieredMergePolicy()
	if Rarely(r) {
		log.Println("Use crazy value for max merge at once")
		tmp.SetMaxMergeAtOnce(NextInt(r, 2, 9))
		tmp.SetMaxMergeAtOnceExplicit(NextInt(r, 2, 9))
	} else {
		tmp.SetMaxMergeAtOnce(NextInt(r, 10, 50))
		tmp.SetMaxMergeAtOnceExplicit(NextInt(r, 10, 50))
	}
	if Rarely(r) {
		log.Println("Use crazy value for max merge segment MB")
		tmp.SetMaxMergedSegmentMB(0.2 + r.Float64()*100)
	} else {
		tmp.SetMaxMergedSegmentMB(r.Float64() * 100)
	}
	tmp.SetFloorSegmentMB(0.2 + r.Float64()*2)
	tmp.SetForceMergeDeletesPctAllowed(0 + r.Float64()*30)
	if Rarely(r) {
		log.Println("Use crazy value for max merge per tire")
		tmp.SetSegmentsPerTier(float64(NextInt(r, 2, 20)))
	} else {
		tmp.SetSegmentsPerTier(float64(NextInt(r, 10, 50)))
	}
	configureRandom(r, tmp)
	tmp.SetReclaimDeletesWeight(r.Float64() * 4)
	return tmp
}

func newAlcoholicMergePolicy(r *rand.Rand /*, tz TimeZone*/) *ti.AlcoholicMergePolicy {
	return ti.NewAlcoholicMergePolicy(rand.New(rand.NewSource(r.Int63())))
}

func newLogMergePolicy(r *rand.Rand) *index.LogMergePolicy {
	var logmp *index.LogMergePolicy
	if r.Intn(2) == 0 {
		logmp = index.NewLogDocMergePolicy()
	} else {
		logmp = index.NewLogByteSizeMergePolicy()
	}
	if Rarely(r) {
		log.Println("Use crazy value for merge factor")
		logmp.SetMergeFactor(NextInt(r, 2, 9))
	} else {
		logmp.SetMergeFactor(NextInt(r, 10, 50))
	}
	configureRandom(r, logmp)
	return logmp
}

func configureRandom(r *rand.Rand, mergePolicy index.MergePolicy) {
	if r.Intn(2) == 0 {
		mergePolicy.SetNoCFSRatio(0.1 + r.Float64()*0.8)
	} else if r.Intn(2) == 0 {
		mergePolicy.SetNoCFSRatio(1.0)
	} else {
		mergePolicy.SetNoCFSRatio(0)
	}

	if Rarely(r) {
		log.Println("Use crazy value for max CFS segment size MB")
		mergePolicy.SetMaxCFSSegmentSizeMB(0.2 + r.Float64()*2)
	} else {
		mergePolicy.SetMaxCFSSegmentSizeMB(math.Inf(1))
	}
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
		log.Println("Use NRTCachingDirectory")
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
func NewTextField(name, value string, stored bool) *docu.Field {
	flag := docu.TEXT_FIELD_TYPE_STORED
	if !stored {
		flag = docu.TEXT_FIELD_TYPE_NOT_STORED
	}
	return NewField(Random(), name, value, flag)
}

func NewField(r *rand.Rand, name, value string, typ *docu.FieldType) *docu.Field {
	panic("not implemented yet")
	// if Usually(r) || !typ.Indexed() {
	// 	// most of the time, don't modify the params
	// 	return docu.NewStringField(name, value, typ)
	// }

	// newType := docu.NewFieldTypeFrom(typ)
	// if !newType.Stored() && r.Intn(2) == 0 {
	// 	newType.SetStored(true) // randonly store it
	// }

	// if !newType.StoreTermVectors() && r.Intn(2) == 0 {
	// 	newType.SetStoreTermVectors(true)
	// 	if !newType.StoreTermVectorOffsets() {
	// 		newType.SetStoreTermVectorOffsets(r.Intn(2) == 0)
	// 	}
	// 	if !newType.StoreTermVectorPositions() {
	// 		newType.SetStoreTermVectorPositions(r.Intn(2) == 0)

	// 		if newType.StoreTermVectorPositions() && !newType.StoreTermVectorPayloads() && !PREFLEX_IMPERSONATION_IS_ACTIVE {
	// 			newType.SetStoreTermVectorPayloads(r.Intn(2) == 2)
	// 		}
	// 	}
	// }

	// return docu.NewStringField(name, value, newType)
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

func NewDefaultIOContext(r *rand.Rand) store.IOContext {
	return NewIOContext(r, store.IO_CONTEXT_DEFAULT)
}

func NewIOContext(r *rand.Rand, oldContext store.IOContext) store.IOContext {
	randomNumDocs := r.Intn(4192)
	size := r.Int63n(512) * int64(randomNumDocs)
	if oldContext.FlushInfo != nil {
		// Always return at least the estimatedSegmentSize of the
		// incoming IOContext:
		if size < oldContext.FlushInfo.EstimatedSegmentSize {
			size = oldContext.FlushInfo.EstimatedSegmentSize
		}
		return store.NewIOContextForFlush(&store.FlushInfo{randomNumDocs, size})
	} else if oldContext.MergeInfo != nil {
		// Always return at least the estimatedMergeBytes of the
		// incoming IOContext:
		if size < oldContext.MergeInfo.EstimatedMergeBytes {
			size = oldContext.MergeInfo.EstimatedMergeBytes
		}
		return store.NewIOContextForMerge(
			&store.MergeInfo{randomNumDocs, size, r.Intn(2) == 0, NextInt(r, 1, 100)})
	} else {
		// Make a totally random IOContext:
		switch r.Intn(5) {
		case 1:
			return store.IO_CONTEXT_READ
		case 2:
			return store.IO_CONTEXT_READONCE
		case 3:
			return store.NewIOContextForMerge(&store.MergeInfo{randomNumDocs, size, true, -1})
		case 4:
			return store.NewIOContextForFlush(&store.FlushInfo{randomNumDocs, size})
		default:
			return store.IO_CONTEXT_DEFAULT
		}
	}
}

// L1193
/*
Sometimes wrap the IndexReader as slow, parallel or filter reader (or
combinations of that)
*/
func maybeWrapReader(r index.IndexReader) (index.IndexReader, error) {
	random := Random()
	if Rarely(random) {
		panic("not implemented yet")
	}
	return r, nil
}

// L1305
// Create a new searcher over the reader. This searcher might randomly use threads
func NewSearcher(r index.IndexReader) *ts.AssertingIndexSearcher {
	return newSearcher(r, true)
}

/*
Create a new searcher over the reader. This searcher might randomly
use threads. If maybeWrap is true, this searcher migt wrap the reader
with one that return nil for sequentialSubReaders.
*/
func newSearcher(r index.IndexReader, maybeWrap bool) *ts.AssertingIndexSearcher {
	random := Random()
	var err error
	// By default, GoLucene would make use of Goroutines to do
	// concurrent search and collect
	//
	// if util.Usually(random) {
	if maybeWrap {
		r, err = maybeWrapReader(r)
		assert(err == nil)
	}
	if random.Intn(2) == 0 {
		ss := ts.NewAssertingIndexSearcher(random, r)
		ss.SetSimilarity(ClassEnvRule.similarity)
		return ss
	}
	ss := ts.NewAssertingIndexSearcherFromContext(random, r.Context())
	ss.SetSimilarity(ClassEnvRule.similarity)
	return ss
	// }
}

// util/TestRuleSetupAndRestoreClassEnv.java

var suppressedCodecs string

func SuppressCodecs(name string) {
	suppressedCodecs = name
}

type ThreadNameFixingPrintStreamInfoStream struct {
	*util.PrintStreamInfoStream
}

func newThreadNameFixingPrintStreamInfoStream(w io.Writer) *ThreadNameFixingPrintStreamInfoStream {
	return &ThreadNameFixingPrintStreamInfoStream{util.NewPrintStreamInfoStream(w)}
}

func (is *ThreadNameFixingPrintStreamInfoStream) Message(component, message string, args ...interface{}) {
	if "TP" == component {
		return // ignore test points!
	}
	is.PrintStreamInfoStream.Message(component, message, args...)
}

/*
Setup and restore suite-level environment (fine grained junk that
doesn't fit anywhere else)
*/
type TestRuleSetupAndRestoreClassEnv struct {
	savedCodec      Codec
	savedInfoStream util.InfoStream

	similarity search.Similarity
	codec      Codec

	avoidCodecs map[string]bool
}

func (rule *TestRuleSetupAndRestoreClassEnv) Before() error {
	// if verbose: print some debugging stuff about which codecs are loaded.
	if VERBOSE {
		for _, codec := range AvailableCodecs() {
			log.Printf("Loaded codec: '%v': %v", codec,
				reflect.TypeOf(LoadCodec(codec)))
		}

		for _, postingFormat := range AvailablePostingsFormats() {
			log.Printf("Loaded postingsFormat: '%v': %v", postingFormat,
				reflect.TypeOf(LoadPostingsFormat(postingFormat)))
		}
	}

	rule.savedInfoStream = util.DefaultInfoStream()
	random := Random()
	if INFOSTREAM {
		util.SetDefaultInfoStream(newThreadNameFixingPrintStreamInfoStream(os.Stdout))
	} else if random.Intn(2) == 0 {
		util.SetDefaultInfoStream(NewNullInfoStream())
	}

	rule.avoidCodecs = make(map[string]bool)
	if suppressedCodecs != "" {
		rule.avoidCodecs[suppressedCodecs] = true
	}

	rule.savedCodec = DefaultCodec()
	randomVal := random.Intn(10)
	if "Lucene3x" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			"random" == TEST_DOCVALUESFORMAT &&
			randomVal == 3 &&
			!rule.shouldAvoidCodec("Lucene3x") { // preflex-only setup
		panic("not supported yet")
	} else if "Lucene40" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			randomVal == 0 &&
			!rule.shouldAvoidCodec("Lucene40") { // 4.0 setup
		panic("not supported yet")
	} else if "Lucene41" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			"random" == TEST_DOCVALUESFORMAT &&
			randomVal == 1 &&
			!rule.shouldAvoidCodec("Lucene41") {
		panic("not supported yet")
	} else if "Lucene42" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			"random" == TEST_DOCVALUESFORMAT &&
			randomVal == 2 &&
			!rule.shouldAvoidCodec("Lucene42") {
		panic("not supported yet")
	} else if "Lucene45" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			"random" == TEST_DOCVALUESFORMAT &&
			randomVal == 3 &&
			!rule.shouldAvoidCodec("Lucene45") {
		panic("not supported yet")
	} else if "Lucene46" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			"random" == TEST_DOCVALUESFORMAT &&
			randomVal == 4 &&
			!rule.shouldAvoidCodec("Lucene46") {
		panic("not supported yet")
	} else if "Lucene49" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			"random" == TEST_POSTINGSFORMAT &&
			"random" == TEST_DOCVALUESFORMAT &&
			randomVal == 5 &&
			!rule.shouldAvoidCodec("Lucene49") {

		rule.codec = LoadCodec("Lucene49")
		OLD_FORMAT_IMPERSONATION_IS_ACTIVE = true

	} else if "random" != TEST_POSTINGSFORMAT ||
		"random" != TEST_DOCVALUESFORMAT {
		// the user wired postings or DV: this is messy
		// refactor into RandomCodec...

		panic("not supported yet")
	} else if "SimpleText" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			randomVal == 9 &&
			Rarely(random) &&
			!rule.shouldAvoidCodec("SimpleText") {
		panic("not supported yet")
	} else if "Appending" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			randomVal == 8 &&
			!rule.shouldAvoidCodec("Appending") {
		panic("not supported yet")
	} else if "CheapBastard" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			randomVal == 8 &&
			!rule.shouldAvoidCodec("CheapBastard") &&
			!rule.shouldAvoidCodec("Lucene41") {
		panic("not supported yet")
	} else if "Asserting" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			randomVal == 6 &&
			!rule.shouldAvoidCodec("Asserting") {
		panic("not implemented yet")
	} else if "Compressing" == TEST_CODEC ||
		"random" == TEST_CODEC &&
			randomVal == 5 &&
			!rule.shouldAvoidCodec("Compressing") {
		panic("not supported yet")
	} else if "random" != TEST_CODEC {
		rule.codec = LoadCodec(TEST_CODEC)
	} else if "random" == TEST_POSTINGSFORMAT {
		panic("not supported yet")
	} else {
		panic("should not be here")
	}
	log.Printf("Use codec: %v", rule.codec)
	DefaultCodec = func() Codec { return rule.codec }

	// Initialize locale/ timezone
	// testLocale := or(os.Getenv("tests.locale"), "random")
	// testTimeZon := or(os.Getenv("tests.timezone"), "random")

	// Always pick a random one for consistency (whether tests.locale
	// was specified or not)
	// Ian: it's not supported yet
	// rule.savedLocale := DefaultLocale()
	// if "random" == testLocale {
	// 	rule.locale = randomLocale(random)
	// } else {
	// 	rule.locale = localeForName(testLocale)
	// }
	// SetDefaultLocale(rule.locale)

	// SetDefaultTimeZone() will set user.timezone to the default
	// timezone of the user's locale. So store the original property
	// value and restore it at end.
	// rule.restoreProperties["user.timezone"] = os.Getenv("user.timezone")
	// rule.savedTimeZone = DefaultTimeZone()
	// if "random" == testTimeZone {
	// 	rule.timeZone = randomTimeZone(random)
	// } else {
	// 	rule.timeZone = TimeZone(testTimeZone)
	// }
	// SetDefaultTImeZone(rule.timeZone)

	if random.Intn(2) == 0 {
		rule.similarity = search.NewDefaultSimilarity()
	} else {
		rule.similarity = ts.NewRandomSimilarityProvider(random)
	}

	// Check codec restrictions once at class level.
	err := rule.checkCodecRestrictions(rule.codec)
	if err != nil {
		log.Printf("NOTE: %v Suppressed codecs: %v", err, rule.avoidCodecs)
		return err
	}

	// We have "stickiness" so taht sometimes all we do is vary the RAM
	// buffer size, other times just the doc count to flush by, else
	// both. This way the assertMemory in DWFC sometimes runs (when we
	// always flush by RAM).
	// setLiveIWCFlushMode(LiveIWCFlushMode(random.Intn(3)))

	return nil
}

func or(a, b string) string {
	if len(a) > 0 {
		return a
	}
	return b
}

// Check codec restrictions.
func (rule *TestRuleSetupAndRestoreClassEnv) checkCodecRestrictions(codec Codec) error {
	assert(codec != nil)
	AssumeTrue(fmt.Sprintf("Class not allowed to use codec: %v.", codec.Name),
		rule.shouldAvoidCodec(codec.Name()))

	if _, ok := codec.(*index.RandomCodec); ok && len(rule.avoidCodecs) > 0 {
		panic("not implemented yet")
	}

	pf := codec.PostingsFormat()
	AssumeFalse(fmt.Sprintf("Class not allowed to use postings format: %v.", pf.Name()),
		rule.shouldAvoidCodec(pf.Name()))

	AssumeFalse(fmt.Sprintf("Class not allowed to use postings format: %v.", TEST_POSTINGSFORMAT),
		rule.shouldAvoidCodec(TEST_POSTINGSFORMAT))

	return nil
}

func (rule *TestRuleSetupAndRestoreClassEnv) After() error {
	panic("not implemented yet")
	// savedCodec := rule.savedCodec
	// index.DefaultCodec = func() Codec { return savedCodec }
	// util.SetDefaultInfoStream(rule.savedInfoStream)
	// // if rule.savedLocale != nil {
	// // 	SetDefaultLocale(rule.savedLocale)
	// // }
	// // if rule.savedTimeZone != nil {
	// // 	SetDefaultTimeZone(rule.savedTimeZone)
	// // }
	// return nil
}

// Should a given codec be avoided for the currently executing suite?
func (rule *TestRuleSetupAndRestoreClassEnv) shouldAvoidCodec(codec string) bool {
	if len(rule.avoidCodecs) == 0 {
		return false
	}
	_, ok := rule.avoidCodecs[codec]
	return ok
}
