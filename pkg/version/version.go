package version

import (
	"fmt"

	"golang.org/x/mod/semver"
)

var (
	version    = "v0.0.0-master+$Format:%h$"
	revision   = "" // sha1 from git, output of $(git rev-parse HEAD)
	preRelease = "dev"

	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

// Version represents service version.
type Version struct {
	Version    string
	PreRelease string
	Revision   string
	BuildDate  string
}

// Get provides agent version.
func Get() *Version {
	return &Version{
		Version:    version,
		PreRelease: preRelease,
		Revision:   revision,
		BuildDate:  buildDate,
	}
}

// String implements fmt.Stringer interface.
func (v *Version) String() string {
	return fmt.Sprintf("ver: %s, rev: %s, pre: %s, date: %s",
		v.Version, v.Revision, v.PreRelease, v.BuildDate)
}

func Compare(w string) int {
	return semver.Compare(version, w)
}
