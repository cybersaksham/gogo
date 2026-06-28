package gogo

import (
	"os"
	"strings"
	"testing"
)

func TestSettingsReferenceDocumentsEveryEnvExampleKey(t *testing.T) {
	envExample, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}
	const settingsReferencePath = "docs/code/reference/settings.md"
	settingsReference, err := os.ReadFile(settingsReferencePath)
	if err != nil {
		t.Fatalf("read %s: %v", settingsReferencePath, err)
	}

	for _, line := range strings.Split(string(envExample), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if !strings.Contains(string(settingsReference), "`"+key+"`") {
			t.Fatalf("%s does not document %s", settingsReferencePath, key)
		}
	}
}
