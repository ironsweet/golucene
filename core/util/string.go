package util

import (
	"os"
	"strconv"
	"time"
)

const primeRK = 16777619

/* simple string hash used by Go strings package */
func Hashstr(sep string) int {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}
	return int(hash)
}

// util/StringHelper.java

/*
Poached from Guava: set a different salt/seed for each VM instance,
to frustrate hash key collision denial of service attacks, and to
catch any places that somehow rely on hash function/order across VM
instances:
*/
var GOOD_FAST_HASH_SEED int = func() int {
	if prop := os.Getenv("tests.seed"); prop != "" {
		// so if there is a test failure that relied on hash order, we
		// remain reproducible based on the test seed:
		if len(prop) > 8 {
			prop = prop[len(prop)-8:]
		}
		n, err := strconv.ParseInt(prop, 16, 64)
		if err != nil {
			panic(err)
		}
		return int(n)
	} else {
		return time.Now().Nanosecond()
	}
}()

/*
Returns the MurmurHash3_x86_32 hash.
Original source/tests at https://github.com/yonik/java_util/
*/
func MurmurHash3_x86_32(data []byte, seed int) int {
	panic("not implemented yet")
}
