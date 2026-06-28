package migrations

type MigrationInfo struct {
	Name   string
	Tables []string
}

func Initial() MigrationInfo {
	return MigrationInfo{Name: "0001_initial", Tables: []string{"flatpages_flatpage", "flatpages_flatpage_sites"}}
}
