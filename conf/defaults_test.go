package conf

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultSettingsProvidesSafeDevelopmentDefaults(t *testing.T) {
	settings := DefaultSettings()

	if settings.Env != "development" {
		t.Fatalf("Env = %q, want development", settings.Env)
	}
	if !settings.Debug {
		t.Fatalf("Debug = false, want true for development defaults")
	}
	if settings.HTTPAddr != ":8000" {
		t.Fatalf("HTTPAddr = %q, want :8000", settings.HTTPAddr)
	}
	if settings.StaticURL != "/static/" {
		t.Fatalf("StaticURL = %q, want /static/", settings.StaticURL)
	}
	if settings.MediaURL != "/media/" {
		t.Fatalf("MediaURL = %q, want /media/", settings.MediaURL)
	}
	if settings.DefaultAutoField != "BigAutoField" {
		t.Fatalf("DefaultAutoField = %q, want BigAutoField", settings.DefaultAutoField)
	}
	if settings.TimeZone != "UTC" {
		t.Fatalf("TimeZone = %q, want UTC", settings.TimeZone)
	}
	if settings.LanguageCode != "en-us" {
		t.Fatalf("LanguageCode = %q, want en-us", settings.LanguageCode)
	}
	if settings.SessionCookieName != "gogo_sessionid" {
		t.Fatalf("SessionCookieName = %q, want gogo_sessionid", settings.SessionCookieName)
	}
	if settings.CSRFCookieName != "gogo_csrftoken" {
		t.Fatalf("CSRFCookieName = %q, want gogo_csrftoken", settings.CSRFCookieName)
	}
	if settings.AdminPath != "/admin" {
		t.Fatalf("AdminPath = %q, want /admin", settings.AdminPath)
	}
}

func TestDefaultSettingsStillRequiresSecretsAndDatabase(t *testing.T) {
	err := DefaultSettings().Validate()
	if err == nil {
		t.Fatalf("Validate() error = nil, want required settings error")
	}

	for _, want := range []string{"GOGO_SECRET_KEY", "DATABASE_URL"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Validate() error = %q, want it to mention %s", err.Error(), want)
		}
	}
}

func TestLoadFromEnvAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=file-secret
DATABASE_URL=postgres://file
`)

	settings, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if settings.Env != "development" {
		t.Fatalf("Env = %q, want development default", settings.Env)
	}
	if settings.HTTPAddr != ":8000" {
		t.Fatalf("HTTPAddr = %q, want :8000 default", settings.HTTPAddr)
	}
	if !settings.Debug {
		t.Fatalf("Debug = false, want true for development default")
	}
}

func TestLoadFromEnvDisablesDebugByDefaultInProduction(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, filepath.Join(dir, ".env"), `
GOGO_ENV=production
GOGO_SECRET_KEY=file-secret
DATABASE_URL=postgres://file
GOGO_ALLOWED_HOSTS=example.com
`)

	settings, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if settings.Debug {
		t.Fatalf("Debug = true, want false for production default")
	}
}
