package conf

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadEnvFileParsesCommentsAndQuotedValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	writeFile(t, path, `
# Framework
GOGO_ENV=development
GOGO_SECRET_KEY="secret with spaces"
DATABASE_URL='postgres://user:pass@localhost:5432/gogo'
GOGO_ALLOWED_HOSTS=localhost, 127.0.0.1
`)

	values, err := LoadEnvFile(path)
	if err != nil {
		t.Fatalf("LoadEnvFile() error = %v", err)
	}

	want := map[string]string{
		"GOGO_ENV":           "development",
		"GOGO_SECRET_KEY":    "secret with spaces",
		"DATABASE_URL":       "postgres://user:pass@localhost:5432/gogo",
		"GOGO_ALLOWED_HOSTS": "localhost, 127.0.0.1",
	}

	if !reflect.DeepEqual(values, want) {
		t.Fatalf("LoadEnvFile() = %#v, want %#v", values, want)
	}
}

func TestLoadEnvFileRejectsInvalidLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	writeFile(t, path, "GOGO_ENV\n")

	if _, err := LoadEnvFile(path); err == nil {
		t.Fatalf("LoadEnvFile() error = nil, want invalid line error")
	}
}

func TestLoadFromEnvUsesDotEnvAndProcessEnvironmentOverride(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, filepath.Join(dir, ".env"), `
GOGO_ENV=development
GOGO_SECRET_KEY=file-secret
DATABASE_URL=postgres://file
GOGO_HTTP_ADDR=:9000
GOGO_ALLOWED_HOSTS=localhost,127.0.0.1
`)
	t.Setenv("GOGO_SECRET_KEY", "process-secret")
	t.Setenv("DATABASE_URL", "postgres://process")

	settings, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if settings.SecretKey != "process-secret" {
		t.Fatalf("SecretKey = %q, want process override", settings.SecretKey)
	}
	if settings.DatabaseURL != "postgres://process" {
		t.Fatalf("DatabaseURL = %q, want process override", settings.DatabaseURL)
	}
	if settings.HTTPAddr != ":9000" {
		t.Fatalf("HTTPAddr = %q, want .env value", settings.HTTPAddr)
	}
	if !reflect.DeepEqual(settings.AllowedHosts, []string{"localhost", "127.0.0.1"}) {
		t.Fatalf("AllowedHosts = %#v, want parsed hosts", settings.AllowedHosts)
	}
}

func TestEnvExampleContainsEveryKnownEnvironmentKey(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("..", ".env.example"))
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}

	for _, key := range knownEnvKeys {
		if !strings.Contains(string(contents), key+"=") {
			t.Fatalf(".env.example does not contain %s", key)
		}
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
