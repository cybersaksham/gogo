package version

import "testing"

func TestInfoUsesDefaultBuildMetadata(t *testing.T) {
	restore := setBuildMetadata("0.0.0-dev", "unknown", "unknown")
	defer restore()

	got := Info()
	want := "gogo 0.0.0-dev (commit unknown, built unknown)"

	if got != want {
		t.Fatalf("Info() = %q, want %q", got, want)
	}
}

func TestInfoUsesOverriddenBuildMetadata(t *testing.T) {
	restore := setBuildMetadata("1.2.3", "abc123", "2026-06-27T00:00:00Z")
	defer restore()

	got := Info()
	want := "gogo 1.2.3 (commit abc123, built 2026-06-27T00:00:00Z)"

	if got != want {
		t.Fatalf("Info() = %q, want %q", got, want)
	}
}

func setBuildMetadata(version, commit, buildDate string) func() {
	oldVersion := Version
	oldCommit := Commit
	oldBuildDate := BuildDate

	Version = version
	Commit = commit
	BuildDate = buildDate

	return func() {
		Version = oldVersion
		Commit = oldCommit
		BuildDate = oldBuildDate
	}
}
