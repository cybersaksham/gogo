package conf

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var knownEnvKeys = []string{
	"GOGO_ENV",
	"GOGO_SECRET_KEY",
	"GOGO_DEBUG",
	"GOGO_ALLOWED_HOSTS",
	"GOGO_HTTP_ADDR",
	"DATABASE_URL",
	"GOGO_INSTALLED_APPS",
	"GOGO_MIDDLEWARE",
	"GOGO_ROOT_URLCONF",
	"GOGO_STATIC_URL",
	"GOGO_STATIC_ROOT",
	"GOGO_MEDIA_URL",
	"GOGO_MEDIA_ROOT",
	"GOGO_TEMPLATE_DIRS",
	"GOGO_DEFAULT_AUTO_FIELD",
	"GOGO_TIME_ZONE",
	"GOGO_LANGUAGE_CODE",
	"GOGO_SESSION_COOKIE_NAME",
	"GOGO_SESSION_COOKIE_SECURE",
	"GOGO_CSRF_COOKIE_NAME",
	"GOGO_CSRF_COOKIE_SECURE",
	"GOGO_HTTPS_ENABLED",
	"GOGO_CSRF_TRUSTED_ORIGINS",
	"GOGO_ADMIN_PATH",
	"GOGO_ADMIN_PATH_REVIEWED",
	"GOGO_DEPLOY_MIGRATIONS_APPLIED",
	"GOGO_DEPLOY_STATIC_COLLECTED",
	"GOGO_PASSWORD_RESET_ENABLED",
	"GOGO_BROKER_URL",
	"GOGO_RESULT_BACKEND",
	"GOGO_SCHEDULE_STORE",
	"GOGO_CACHE_URL",
	"GOGO_EMAIL_URL",
}

// LoadEnvFile parses a simple KEY=VALUE environment file.
func LoadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid env line %d: missing '='", lineNumber)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("invalid env line %d: key is required", lineNumber)
		}

		values[key] = unquoteEnvValue(strings.TrimSpace(value))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return values, nil
}

// LoadFromEnv loads settings from .env when present and process environment.
func LoadFromEnv() (Settings, error) {
	values := make(map[string]string)
	if fileValues, err := LoadEnvFile(".env"); err == nil {
		for key, value := range fileValues {
			values[key] = value
		}
	} else if !os.IsNotExist(err) {
		return Settings{}, err
	}

	for _, key := range knownEnvKeys {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}

	return SettingsFromMap(values), nil
}

// SettingsFromMap builds settings from parsed environment values and defaults.
func SettingsFromMap(values map[string]string) Settings {
	settings := DefaultSettings()

	if value := values["GOGO_ENV"]; value != "" {
		settings.Env = value
	}
	if value := values["GOGO_SECRET_KEY"]; value != "" {
		settings.SecretKey = value
	}
	if value, ok := values["GOGO_DEBUG"]; ok && strings.TrimSpace(value) != "" {
		settings.Debug = parseBool(value)
	} else {
		settings.Debug = settings.Env == "development"
	}
	if value := values["GOGO_ALLOWED_HOSTS"]; value != "" {
		settings.AllowedHosts = splitList(value)
	}
	if value := values["GOGO_HTTP_ADDR"]; value != "" {
		settings.HTTPAddr = value
	}
	if value := values["DATABASE_URL"]; value != "" {
		settings.DatabaseURL = value
	}
	if value := values["GOGO_INSTALLED_APPS"]; value != "" {
		settings.InstalledApps = splitList(value)
	}
	if value := values["GOGO_MIDDLEWARE"]; value != "" {
		settings.Middleware = splitList(value)
	}
	if value := values["GOGO_ROOT_URLCONF"]; value != "" {
		settings.RootURLConf = value
	}
	if value := values["GOGO_STATIC_URL"]; value != "" {
		settings.StaticURL = value
	}
	if value := values["GOGO_STATIC_ROOT"]; value != "" {
		settings.StaticRoot = value
	}
	if value := values["GOGO_MEDIA_URL"]; value != "" {
		settings.MediaURL = value
	}
	if value := values["GOGO_MEDIA_ROOT"]; value != "" {
		settings.MediaRoot = value
	}
	if value := values["GOGO_TEMPLATE_DIRS"]; value != "" {
		settings.TemplateDirs = splitList(value)
	}
	if value := values["GOGO_DEFAULT_AUTO_FIELD"]; value != "" {
		settings.DefaultAutoField = value
	}
	if value := values["GOGO_TIME_ZONE"]; value != "" {
		settings.TimeZone = value
	}
	if value := values["GOGO_LANGUAGE_CODE"]; value != "" {
		settings.LanguageCode = value
	}
	if value := values["GOGO_SESSION_COOKIE_NAME"]; value != "" {
		settings.SessionCookieName = value
	}
	if value, ok := values["GOGO_SESSION_COOKIE_SECURE"]; ok && strings.TrimSpace(value) != "" {
		settings.SessionCookieSecure = parseBool(value)
	}
	if value := values["GOGO_CSRF_COOKIE_NAME"]; value != "" {
		settings.CSRFCookieName = value
	}
	if value, ok := values["GOGO_CSRF_COOKIE_SECURE"]; ok && strings.TrimSpace(value) != "" {
		settings.CSRFCookieSecure = parseBool(value)
	}
	if value, ok := values["GOGO_HTTPS_ENABLED"]; ok && strings.TrimSpace(value) != "" {
		settings.HTTPSEnabled = parseBool(value)
	}
	if value := values["GOGO_CSRF_TRUSTED_ORIGINS"]; value != "" {
		settings.CSRFTrustedOrigins = splitList(value)
	}
	if value := values["GOGO_ADMIN_PATH"]; value != "" {
		settings.AdminPath = value
	}
	if value, ok := values["GOGO_ADMIN_PATH_REVIEWED"]; ok && strings.TrimSpace(value) != "" {
		settings.AdminPathReviewed = parseBool(value)
	}
	if value, ok := values["GOGO_DEPLOY_MIGRATIONS_APPLIED"]; ok && strings.TrimSpace(value) != "" {
		settings.MigrationsApplied = parseBool(value)
	}
	if value, ok := values["GOGO_DEPLOY_STATIC_COLLECTED"]; ok && strings.TrimSpace(value) != "" {
		settings.StaticFilesCollected = parseBool(value)
	}
	if value, ok := values["GOGO_PASSWORD_RESET_ENABLED"]; ok && strings.TrimSpace(value) != "" {
		settings.PasswordResetEnabled = parseBool(value)
	}
	if value := values["GOGO_BROKER_URL"]; value != "" {
		settings.BrokerURL = value
	}
	if value := values["GOGO_RESULT_BACKEND"]; value != "" {
		settings.ResultBackend = value
	}
	if value := values["GOGO_SCHEDULE_STORE"]; value != "" {
		settings.ScheduleStore = value
	}
	if value := values["GOGO_CACHE_URL"]; value != "" {
		settings.CacheURL = value
	}
	if value := values["GOGO_EMAIL_URL"]; value != "" {
		settings.EmailURL = value
	}

	return settings
}

func unquoteEnvValue(value string) string {
	if len(value) < 2 {
		return value
	}

	first := value[0]
	last := value[len(value)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return value[1 : len(value)-1]
	}

	return value
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
