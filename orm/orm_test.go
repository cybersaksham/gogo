package orm

import "testing"

func TestPackageExistsForModelInstanceContracts(t *testing.T) {
	if PackageName() != "orm" {
		t.Fatalf("PackageName() = %q, want orm", PackageName())
	}
}
