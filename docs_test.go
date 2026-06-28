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
	settingsReference, err := os.ReadFile("docs/reference/settings.md")
	if err != nil {
		t.Fatalf("read docs/reference/settings.md: %v", err)
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
			t.Fatalf("docs/reference/settings.md does not document %s", key)
		}
	}
}
