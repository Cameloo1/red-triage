package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// SetVersion sets the version information
func SetVersion(version, commit, buildDate string) {
	Version = version
	Commit = commit
	BuildDate = buildDate
}

// GetVersion returns the full version string
func GetVersion() string {
	return fmt.Sprintf("RedTriage v%s (%s) built on %s", Version, Commit, BuildDate)
}

// GetShortVersion returns just the version
func GetShortVersion() string {
	return Version
}

// GetBuildInfo returns build information
func GetBuildInfo() string {
	return fmt.Sprintf("Go %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
