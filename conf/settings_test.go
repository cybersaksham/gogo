package conf

import (
	"errors"
	"strings"
	"testing"
)

func TestSettingsValidateAcceptsDevelopmentSettings(t *testing.T) {
	settings := Settings{
		Env:         "development",
		SecretKey:   "dev-secret",
		HTTPAddr:    ":8000",
		DatabaseURL: "postgres://user:pass@localhost:5432/gogo?sslmode=disable",
	}

	if err := settings.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSettingsValidateAcceptsProductionSettings(t *testing.T) {
	settings := Settings{
		Env:          "production",
		SecretKey:    "prod-secret",
		AllowedHosts: []string{"example.com"},
		HTTPAddr:     "0.0.0.0:8000",
		DatabaseURL:  "postgres://user:pass@db:5432/gogo",
	}

	if err := settings.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSettingsValidateRejectsMissingRequiredValues(t *testing.T) {
	settings := Settings{
		Env:      "development",
		HTTPAddr: ":8000",
	}

	err := settings.Validate()
	if !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("Validate() error = %v, want ErrInvalidSettings", err)
	}

	message := err.Error()
	for _, want := range []string{"GOGO_SECRET_KEY", "DATABASE_URL"} {
		if !strings.Contains(message, want) {
			t.Fatalf("Validate() error = %q, want it to mention %q", message, want)
		}
	}
}

func TestSettingsValidateRejectsInvalidEnvironment(t *testing.T) {
	settings := validSettings()
	settings.Env = "staging"

	err := settings.Validate()
	if !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("Validate() error = %v, want ErrInvalidSettings", err)
	}
	if !strings.Contains(err.Error(), "GOGO_ENV") {
		t.Fatalf("Validate() error = %q, want it to mention GOGO_ENV", err.Error())
	}
}

func TestSettingsValidateRejectsInvalidHTTPAddress(t *testing.T) {
	settings := validSettings()
	settings.HTTPAddr = "not-an-address"

	err := settings.Validate()
	if !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("Validate() error = %v, want ErrInvalidSettings", err)
	}
	if !strings.Contains(err.Error(), "GOGO_HTTP_ADDR") {
		t.Fatalf("Validate() error = %q, want it to mention GOGO_HTTP_ADDR", err.Error())
	}
}

func TestSettingsValidateRequiresAllowedHostsInProduction(t *testing.T) {
	settings := validSettings()
	settings.Env = "production"
	settings.AllowedHosts = nil

	err := settings.Validate()
	if !errors.Is(err, ErrInvalidSettings) {
		t.Fatalf("Validate() error = %v, want ErrInvalidSettings", err)
	}
	if !strings.Contains(err.Error(), "GOGO_ALLOWED_HOSTS") {
		t.Fatalf("Validate() error = %q, want it to mention GOGO_ALLOWED_HOSTS", err.Error())
	}
}

func validSettings() Settings {
	return Settings{
		Env:          "development",
		SecretKey:    "secret",
		AllowedHosts: []string{"localhost"},
		HTTPAddr:     ":8000",
		DatabaseURL:  "postgres://user:pass@localhost:5432/gogo",
	}
}
