package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestBaseConfigImplementsValidConfig(t *testing.T) {
	config := BaseConfig{
		AppName:        "example.blog",
		AppLabel:       "blog",
		AppPath:        t.TempDir(),
		AppVerboseName: "Blog",
		AppDependencies: []string{
			"example.accounts",
		},
	}

	if err := ValidateConfig(config); err != nil {
		t.Fatalf("ValidateConfig() error = %v", err)
	}

	if config.Name() != "example.blog" {
		t.Fatalf("Name() = %q, want example.blog", config.Name())
	}
	if config.Label() != "blog" {
		t.Fatalf("Label() = %q, want blog", config.Label())
	}
	if config.VerboseName() != "Blog" {
		t.Fatalf("VerboseName() = %q, want Blog", config.VerboseName())
	}
	if err := config.Ready(context.Background(), &Registry{}); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}
	if err := config.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestValidateConfigRejectsInvalidNames(t *testing.T) {
	config := validConfig(t)
	config.AppName = "not a package"

	err := ValidateConfig(config)
	if !errors.Is(err, ErrInvalidApp) {
		t.Fatalf("ValidateConfig() error = %v, want ErrInvalidApp", err)
	}
}

func TestValidateConfigRejectsInvalidLabels(t *testing.T) {
	config := validConfig(t)
	config.AppLabel = "bad-label"

	err := ValidateConfig(config)
	if !errors.Is(err, ErrInvalidApp) {
		t.Fatalf("ValidateConfig() error = %v, want ErrInvalidApp", err)
	}
}

func TestValidateConfigRejectsMissingPath(t *testing.T) {
	config := validConfig(t)
	config.AppPath = ""

	err := ValidateConfig(config)
	if !errors.Is(err, ErrInvalidApp) {
		t.Fatalf("ValidateConfig() error = %v, want ErrInvalidApp", err)
	}
}

func TestValidateConfigRejectsInvalidDependencyNames(t *testing.T) {
	config := validConfig(t)
	config.AppDependencies = []string{"bad dependency"}

	err := ValidateConfig(config)
	if !errors.Is(err, ErrInvalidApp) {
		t.Fatalf("ValidateConfig() error = %v, want ErrInvalidApp", err)
	}
}

func TestValidateConfigsRejectsDuplicateLabels(t *testing.T) {
	first := validConfig(t)
	second := validConfig(t)
	second.AppName = "example.other"
	second.AppPath = filepath.Join(t.TempDir(), "other")

	err := ValidateConfigs(first, second)
	if !errors.Is(err, ErrDuplicateApp) {
		t.Fatalf("ValidateConfigs() error = %v, want ErrDuplicateApp", err)
	}
}

func validConfig(t *testing.T) BaseConfig {
	t.Helper()

	return BaseConfig{
		AppName:        "example.blog",
		AppLabel:       "blog",
		AppPath:        t.TempDir(),
		AppVerboseName: "Blog",
	}
}
