package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Version [4]int

var (
	// Match settings and bugs in Lucene's 3.1 release.
	VERSION_31 = Version([4]int{3, 1, 0, 0})
	// Match settings and bugs in Lucene's 4.0 release.
	VERSION_4_0 = Version([4]int{4, 0, 0, 0})
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

/*
Parse a version number of the form major.minor.bugfix.prerlease

Part .bugfix and part .prerelease are optional.
Note that this is forwards compatible: the parsed version does not
have to exist as a constant.
*/
func ParseVersion(version string) (ver Version, err error) {
	if version == "" {
		return
	}

	parts := strings.SplitN(version, ".", 4)
	if len(parts) < 2 || len(parts) > 4 {
		err = errors.New(fmt.Sprintf(
			"Version is not in form major.minor(.bugfix.prerelease) (got: %v)",
			version))
		return
	}

	var major, minor, bugfix, prerelease int
	if major, err = parse(parts[0], version, "major"); err != nil {
		return
	}
	if minor, err = parse(parts[1], version, "minor"); err != nil {
		return
	}

	if len(parts) > 2 {
		if bugfix, err = parse(parts[2], version, "bugfix"); err != nil {
			return
		}

		if len(parts) > 3 {
			if prerelease, err = parse(parts[3], version, "prerelease"); err != nil {
				return
			}
			if prerelease == 0 {
				err = errors.New(fmt.Sprintf(
					"Invalid value %v for prerelease; should be 1 or 2 (got: %v)",
					prerelease, version))
				return
			}
		}
	}

	return newVersion(major, minor, bugfix, prerelease), nil
}

func parse(part, text, name string) (int, error) {
	if n, err := strconv.ParseInt(part, 10, 64); err == nil {
		return int(n), nil
	}
	return 0, errors.New(fmt.Sprintf(
		"Failed to parse %v version from '%v' (got: %v)",
		name, part, text))
}

func newVersion(major, minor, bugfix, prerelease int) Version {
	ans := Version([4]int{major, minor, bugfix, prerelease})

	// NOTE: do not enforce major version so we remain future proof,
	// except to make sure it fits in the 8 bits we encode it into:
	assert2(major >= 0 && major <= 255, "Illegal major version: %v", major)
	assert2(minor >= 0 && minor <= 255, "Illegal minor version: %v", minor)
	assert2(bugfix >= 0 && bugfix <= 255, "Illegal bugfix version: %v", bugfix)
	assert2(prerelease >= 0 || prerelease <= 2, "Illegal prerelease version: %v", prerelease)
	assert2(prerelease == 0 || minor == 0 && bugfix == 0,
		"Prerelease version only supported with major release (got prerelease: %v, minor: %v, bugfix: %v)",
		prerelease, minor, bugfix)

	return ans
}

func encoded(v Version) int {
	if len(v) < 4 {
		return 0
	}
	return v[0]<<18 | v[1]<<10 | v[2]<<2 | v[3]
}

func (v Version) OnOrAfter(other Version) bool {
	return encoded(v) >= encoded(other)
}

func (v Version) String() string {
	if v[3] == 0 {
		return fmt.Sprintf("%v.%v.%v", v[0], v[1], v[2])
	}
	return fmt.Sprintf("%v.%v.%v.%v", v[0], v[1], v[2], v[3])
}

func (v Version) Equals(other Version) bool {
	return encoded(v) == encoded(other)
}
