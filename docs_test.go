package gogo

import (
	"os"
	"strings"
	"testing"
)

func TestReadmeDocumentsEveryEnvExampleKey(t *testing.T) {
	envExample, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
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

		if !strings.Contains(string(readme), "`"+key+"`") {
			t.Fatalf("README.md does not document %s", key)
		}
	}
}
