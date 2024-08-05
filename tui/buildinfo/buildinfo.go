// Package buildinfo provides functionality to store and retrieve information about a build,
// such as version, date, and commit hash. It includes a struct to hold this information and
// functions to create and format instances of this struct.
package buildinfo

import (
	"cmp"
	"fmt"
)

// BuildInfo holds information about a build.
type BuildInfo struct {
	version string
	date    string
	commit  string
}

// New creates a new BuildInfo instance with provided version, date, and commit.
// If any of the parameters are empty, it sets them to "N/A".
func New(version, date, commit string) BuildInfo {
	emptyData := "N/A"

	version = cmp.Or(version, emptyData)
	date = cmp.Or(date, emptyData)
	commit = cmp.Or(commit, emptyData)

	return BuildInfo{
		version: fmt.Sprintf("Build version: %s", version),
		date:    fmt.Sprintf("Build date: %s", date),
		commit:  fmt.Sprintf("Build commit: %s", commit),
	}
}

// String returns a formatted string representation of the BuildInfo.
func (bi BuildInfo) String() string {
	return fmt.Sprintf("%s; %s; %s;", bi.version, bi.date, bi.commit)
}
