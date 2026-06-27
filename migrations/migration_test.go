package migrations

import (
	"errors"
	"testing"
)

func TestMigrationIdentityNamingAndValidation(t *testing.T) {
	migration := Migration{
		AppLabel: "blog",
		Name:     InitialMigrationName(),
		Dependencies: []Dependency{
			{AppLabel: "auth", Name: "0001_initial"},
		},
		RunBefore: []Dependency{{AppLabel: "comments", Name: "0002_add_flags"}},
		Atomic:    true,
		Operations: []Operation{
			NoopOperation{NameValue: "CreateModel"},
		},
	}
	if migration.Identity() != "blog.0001_initial" {
		t.Fatalf("Identity() = %q", migration.Identity())
	}
	if NextMigrationName(2, "Add Post Model") != "0002_add_post_model" {
		t.Fatalf("NextMigrationName() mismatch")
	}
	if err := migration.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestMigrationValidationFailures(t *testing.T) {
	cases := []Migration{
		{Name: "0001_initial", Operations: []Operation{NoopOperation{NameValue: "op"}}},
		{AppLabel: "blog", Name: "bad name", Operations: []Operation{NoopOperation{NameValue: "op"}}},
		{AppLabel: "blog", Name: "0001_initial"},
		{AppLabel: "blog", Name: "0001_initial", Dependencies: []Dependency{{AppLabel: "auth", Name: "0001_initial"}, {AppLabel: "auth", Name: "0001_initial"}}, Operations: []Operation{NoopOperation{NameValue: "op"}}},
	}
	for _, migration := range cases {
		if err := migration.Validate(); !errors.Is(err, ErrInvalidMigration) {
			t.Fatalf("Validate(%#v) error = %v, want ErrInvalidMigration", migration, err)
		}
	}
}

type NoopOperation struct {
	NameValue string
}

func (o NoopOperation) Name() string {
	return o.NameValue
}
