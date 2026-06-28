package flatpages

import flatpagemigrations "github.com/cybersaksham/gogo/contrib/flatpages/migrations"

func Migration() flatpagemigrations.MigrationInfo {
	return flatpagemigrations.Initial()
}
