package util

import (
	"fmt"
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
var GOOD_FAST_HASH_SEED = func() uint32 {
	if prop := os.Getenv("tests_seed"); prop != "" {
		// so if there is a test failure that relied on hash order, we
		// remain reproducible based on the test seed:
		if len(prop) > 8 {
			prop = prop[len(prop)-8:]
		}
		n, err := strconv.ParseInt(prop, 16, 64)
		if err != nil {
			panic(err)
		}
		fmt.Printf("tests_seed=%v\n", uint32(n))
		return uint32(n)
	} else {
		return uint32(time.Now().Nanosecond())
	}
}()

/* Returns true iff the ref starts with the given prefix. Otherwise false. */
func StartsWith(ref, prefix []byte) bool {
	return sliceEquals(ref, prefix, 0)
}

func sliceEquals(sliceToTest, other []byte, pos int) bool {
	panic("niy")
}

/*
Returns the MurmurHash3_x86_32 hash.
Original source/tests at https://github.com/yonik/java_util/
*/
func MurmurHash3_x86_32(data []byte, seed uint32) uint32 {
	c1, c2 := uint32(0xcc9e2d51), uint32(0x1b873593)

	h1 := seed
	roundedEnd := (len(data) & 0xfffffffc) // round down to 4 byte block

	for i := 0; i < roundedEnd; i += 4 {
		// little endian load order
		k1 := uint32(data[i]) | (uint32(data[i+1]) << 8) |
			(uint32(data[i+2]) << 16) | (uint32(data[i+3]) << 24)
		k1 *= c1
		k1 = rotateLeft(k1, 15)
		k1 *= c2

		h1 ^= k1
		h1 = rotateLeft(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	// tail
	k1 := uint32(0)

	left := len(data) & 0x03
	if left == 3 {
		k1 = uint32(data[roundedEnd+2]) << 16
	}
	if left >= 2 {
		k1 |= uint32(data[roundedEnd+1]) << 8
	}
	if left >= 1 {
		k1 |= uint32(data[roundedEnd])
		k1 *= c1
		k1 = rotateLeft(k1, 15)
		k1 *= c2
		h1 ^= k1
	}

	// finalization
	h1 ^= uint32(len(data))

	// fmix(h1);
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}

func rotateLeft(n uint32, offset uint) uint32 {
	mask := uint32(1<<(32-offset)) - 1
	hi := (n & mask) << offset
	low := (n - (n & mask)) >> (32 - offset)
	return hi | low
}
