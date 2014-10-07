package util

type Version int

const (
	// Match settings and bugs in Lucene's 3.1 release.
	VERSION_31 = Version(31)
	// Match settings and bugs in Lucene's 4.5 release.
	VERSION_45 = Version(45)
	// Match settings and bugs in Lucene's 4.9 release.
	// Use this to get the latest and greatest settings, bug fixes, etc,
	// for Lucnee.
	VERSION_49     = Version(49)
	VERSION_4_10   = Version(410)
	VERSION_4_10_1 = Version(4101)

	VERSION_LATEST = VERSION_4_10_1
)

func (v Version) OnOrAfter(other Version) bool {
	return int(v) >= int(other)
}
