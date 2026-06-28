package redirects

import "github.com/cybersaksham/gogo/app"

func AppConfig() app.Config {
	return app.BaseConfig{AppName: "gogo.contrib.redirects", AppLabel: "redirects", AppPath: "contrib/redirects", AppVerboseName: "Redirects"}
}
