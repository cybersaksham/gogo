package migrations

import (
	"context"
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

func (o NoopOperation) StateForwards(*ProjectState) error { return nil }
func (o NoopOperation) DatabaseForwards(context.Context, SchemaEditor) error {
	return nil
}
func (o NoopOperation) DatabaseBackwards(context.Context, SchemaEditor) error {
	return nil
}
func (o NoopOperation) Describe() string { return o.NameValue }
func (o NoopOperation) Reversible() bool { return true }
func (o NoopOperation) ReferencesModel(string, string) bool {
	return false
}
func (o NoopOperation) ReferencesField(string, string, string) bool {
	return false
}
