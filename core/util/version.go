package util

import (
	"fmt"
)

type Version [4]int

var (
	// Match settings and bugs in Lucene's 3.1 release.
	VERSION_31 = Version([4]int{3, 1, 0, 0})
	// Match settings and bugs in Lucene's 4.5 release.
	VERSION_45 = Version([4]int{4, 5, 0, 0})
	// Match settings and bugs in Lucene's 4.9 release.
	// Use this to get the latest and greatest settings, bug fixes, etc,
	// for Lucnee.
	VERSION_49     = Version([4]int{4, 9, 0, 0})
	VERSION_4_10   = Version([4]int{4, 10, 0, 0})
	VERSION_4_10_1 = Version([4]int{4, 10, 1, 0})

	VERSION_LATEST = VERSION_4_10_1
)

func (v Version) OnOrAfter(other Version) bool {
	return encoded(v) >= encoded(other)
}

func encoded(v Version) int {
	return v[0]<<18 | v[1]<<10 | v[2]<<2 | v[3]
}

func (v Version) String() string {
	if v[3] == 0 {
		return fmt.Sprintf("%v.%v.%v", v[0], v[1], v[2])
	}
	return fmt.Sprintf("%v.%v.%v.%v", v[0], v[1], v[2], v[3])
}
