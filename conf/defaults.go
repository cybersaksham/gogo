package conf

// DefaultSettings returns the framework defaults before environment overrides.
func DefaultSettings() Settings {
	return Settings{
		Env:               "development",
		Debug:             true,
		HTTPAddr:          ":8000",
		StaticURL:         "/static/",
		MediaURL:          "/media/",
		DefaultAutoField:  "BigAutoField",
		TimeZone:          "UTC",
		LanguageCode:      "en-us",
		SessionCookieName: "gogo_sessionid",
		CSRFCookieName:    "gogo_csrftoken",
	}
}
