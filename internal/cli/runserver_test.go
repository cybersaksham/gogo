package cli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestRunserverUsesDefaultAddress(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=runserver-secret
DATABASE_URL=postgres://runserver
`)

	config := runRunserverWithCapture(t, nil)

	if config.Addr != ":8000" {
		t.Fatalf("Addr = %q, want default :8000", config.Addr)
	}
	if config.Reload {
		t.Fatalf("Reload = true, want default false")
	}
}

func TestRunserverUsesEnvironmentAddress(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=runserver-secret
DATABASE_URL=postgres://runserver
GOGO_HTTP_ADDR=:9000
`)

	config := runRunserverWithCapture(t, nil)

	if config.Addr != ":9000" {
		t.Fatalf("Addr = %q, want .env address", config.Addr)
	}
}

func TestRunserverFlagOverridesEnvironmentAddress(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=runserver-secret
DATABASE_URL=postgres://runserver
GOGO_HTTP_ADDR=:9000
`)

	config := runRunserverWithCapture(t, []string{"--addr", ":7000", "--reload=true"})

	if config.Addr != ":7000" {
		t.Fatalf("Addr = %q, want flag address", config.Addr)
	}
	if !config.Reload {
		t.Fatalf("Reload = false, want true from flag")
	}
}

func TestRunserverLoadsExplicitSettingsFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	settingsPath := filepath.Join(dir, "custom.env")
	writeTextFile(t, settingsPath, `
GOGO_SECRET_KEY=runserver-secret
DATABASE_URL=postgres://runserver
GOGO_HTTP_ADDR=:9100
`)

	config := runRunserverWithCapture(t, []string{"--settings", settingsPath})

	if config.Addr != ":9100" {
		t.Fatalf("Addr = %q, want explicit settings file address", config.Addr)
	}
	if config.SettingsPath != settingsPath {
		t.Fatalf("SettingsPath = %q, want %q", config.SettingsPath, settingsPath)
	}
}

func TestRunserverDefaultStarterIsUnavailable(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeTextFile(t, filepath.Join(dir, ".env"), `
GOGO_SECRET_KEY=runserver-secret
DATABASE_URL=postgres://runserver
`)

	command := NewRunserverCommand(nil)
	err := command.Run(context.Background(), nil)
	if !errors.Is(err, ErrCommandUnavailable) {
		t.Fatalf("Run() error = %v, want ErrCommandUnavailable", err)
	}
}

func runRunserverWithCapture(t *testing.T, args []string) RunserverConfig {
	t.Helper()

	var got RunserverConfig
	command := NewRunserverCommand(func(_ context.Context, config RunserverConfig) error {
		got = config
		return nil
	})

	if err := command.Run(context.Background(), args); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	return got
}
