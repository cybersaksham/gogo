package redirects

import redirectmigrations "github.com/cybersaksham/gogo/contrib/redirects/migrations"

func Migration() redirectmigrations.MigrationInfo {
	return redirectmigrations.Initial()
}
