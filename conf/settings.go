package conf

import (
	"fmt"
	"net"
	"strings"
)

// Settings contains framework configuration loaded from code and environment.
type Settings struct {
	Env               string
	SecretKey         string
	Debug             bool
	AllowedHosts      []string
	HTTPAddr          string
	DatabaseURL       string
	InstalledApps     []string
	Middleware        []string
	RootURLConf       string
	StaticURL         string
	StaticRoot        string
	MediaURL          string
	MediaRoot         string
	TemplateDirs      []string
	DefaultAutoField  string
	TimeZone          string
	LanguageCode      string
	SessionCookieName string
	CSRFCookieName    string
	BrokerURL         string
	ResultBackend     string
	CacheURL          string
	EmailURL          string
}

// Validate returns an error when required settings are missing or invalid.
func (s Settings) Validate() error {
	var problems []string

	switch s.Env {
	case "development", "test", "production":
	default:
		problems = append(problems, "GOGO_ENV must be one of development, test, production")
	}

	if strings.TrimSpace(s.SecretKey) == "" {
		problems = append(problems, "GOGO_SECRET_KEY is required")
	}

	if strings.TrimSpace(s.DatabaseURL) == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	if strings.TrimSpace(s.HTTPAddr) == "" {
		problems = append(problems, "GOGO_HTTP_ADDR is required")
	} else if err := validateHTTPAddr(s.HTTPAddr); err != nil {
		problems = append(problems, fmt.Sprintf("GOGO_HTTP_ADDR is invalid: %v", err))
	}

	if s.Env == "production" && len(s.AllowedHosts) == 0 {
		problems = append(problems, "GOGO_ALLOWED_HOSTS must not be empty in production")
	}

	if len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidSettings, strings.Join(problems, "; "))
	}

	return nil
}

func validateHTTPAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	if port == "" {
		return fmt.Errorf("missing port")
	}

	if host != "" && net.ParseIP(host) == nil {
		if strings.Contains(host, " ") {
			return fmt.Errorf("host contains whitespace")
		}
	}

	return nil
}
