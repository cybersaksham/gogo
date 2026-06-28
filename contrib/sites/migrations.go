package sites

import sitemigrations "github.com/cybersaksham/gogo/contrib/sites/migrations"

func Migration() sitemigrations.MigrationInfo {
	return sitemigrations.Initial()
}
