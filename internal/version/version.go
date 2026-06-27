package version

import "fmt"

var (
	// Version is the framework version. Release builds override it with ldflags.
	Version = "0.0.0-dev"

	// Commit is the source revision. Release builds override it with ldflags.
	Commit = "unknown"

	// BuildDate is the build timestamp. Release builds override it with ldflags.
	BuildDate = "unknown"
)

// Info returns stable human-readable version metadata.
func Info() string {
	return fmt.Sprintf("gogo %s (commit %s, built %s)", Version, Commit, BuildDate)
}
