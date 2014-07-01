package util

import (
	"strings"
)

// util/Constants.java

/*
This is the internal Lucene version, recorded into each segment.
*/
const LUCENE_MAIN_VERSION = "4.9"

/*
This is the Lucene version for display purpose.
*/
var LUCENE_VERSION = func() string {
	return mainVersionWithoutAlphaBeta() + "-SNAPSHOT"
}()

/* Returns a LUCENE_MAIN_VERSION without any ALPHA/BETA qualifier used by test only! */
func mainVersionWithoutAlphaBeta() string {
	if parts := strings.Split(LUCENE_MAIN_VERSION, "."); len(parts) == 4 && "0" == parts[2] {
		return parts[0] + "." + parts[1]
	} else {
		return LUCENE_MAIN_VERSION
	}
}
