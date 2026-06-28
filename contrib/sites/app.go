package sites

import "github.com/cybersaksham/gogo/app"

func AppConfig() app.Config {
	return app.BaseConfig{
		AppName:        "gogo.contrib.sites",
		AppLabel:       "sites",
		AppPath:        "contrib/sites",
		AppVerboseName: "Sites",
	}
}

type Settings struct {
	SiteID int
}
