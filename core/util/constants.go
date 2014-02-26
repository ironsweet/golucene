package util

import (
	"strings"
)

/*
This is the internal Lucene version, recorded into each segment.
*/
const LUCENE_MAIN_VERSION = "4.5.1"

/*
This is the Lucene version for display purpose.
*/
var LUCENE_VERSION = func() string {
	if parts := strings.Split(LUCENE_MAIN_VERSION, "."); len(parts) == 4 {
		// alpha/beta
		assert(parts[2] == "0")
		return parts[0] + "." + parts[1] + "-SNAPSHOT"
	} else {
		return LUCENE_MAIN_VERSION + "-SNAPSHOT"
	}
}()
