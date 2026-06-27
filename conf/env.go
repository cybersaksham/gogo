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
	"GOGO_CSRF_COOKIE_NAME",
	"GOGO_BROKER_URL",
	"GOGO_RESULT_BACKEND",
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

	return settingsFromMap(values), nil
}

func settingsFromMap(values map[string]string) Settings {
	return Settings{
		Env:               values["GOGO_ENV"],
		SecretKey:         values["GOGO_SECRET_KEY"],
		Debug:             parseBool(values["GOGO_DEBUG"]),
		AllowedHosts:      splitList(values["GOGO_ALLOWED_HOSTS"]),
		HTTPAddr:          values["GOGO_HTTP_ADDR"],
		DatabaseURL:       values["DATABASE_URL"],
		InstalledApps:     splitList(values["GOGO_INSTALLED_APPS"]),
		Middleware:        splitList(values["GOGO_MIDDLEWARE"]),
		RootURLConf:       values["GOGO_ROOT_URLCONF"],
		StaticURL:         values["GOGO_STATIC_URL"],
		StaticRoot:        values["GOGO_STATIC_ROOT"],
		MediaURL:          values["GOGO_MEDIA_URL"],
		MediaRoot:         values["GOGO_MEDIA_ROOT"],
		TemplateDirs:      splitList(values["GOGO_TEMPLATE_DIRS"]),
		DefaultAutoField:  values["GOGO_DEFAULT_AUTO_FIELD"],
		TimeZone:          values["GOGO_TIME_ZONE"],
		LanguageCode:      values["GOGO_LANGUAGE_CODE"],
		SessionCookieName: values["GOGO_SESSION_COOKIE_NAME"],
		CSRFCookieName:    values["GOGO_CSRF_COOKIE_NAME"],
		BrokerURL:         values["GOGO_BROKER_URL"],
		ResultBackend:     values["GOGO_RESULT_BACKEND"],
		CacheURL:          values["GOGO_CACHE_URL"],
		EmailURL:          values["GOGO_EMAIL_URL"],
	}
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
