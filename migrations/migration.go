package migrations

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var migrationNamePattern = regexp.MustCompile(`^\d{4}_[a-z0-9_]+$`)

// Operation is the minimal migration operation contract.
type Operation interface {
	Name() string
}

// Dependency identifies another migration.
type Dependency struct {
	AppLabel string
	Name     string
}

// Identity returns app.name.
func (d Dependency) Identity() string {
	return d.AppLabel + "." + d.Name
}

// Migration describes a Go migration file.
type Migration struct {
	AppLabel     string
	Name         string
	Dependencies []Dependency
	Replaces     []Dependency
	Operations   []Operation
	Atomic       bool
	RunBefore    []Dependency
}

// Identity returns app.name.
func (m Migration) Identity() string {
	return m.AppLabel + "." + m.Name
}

// Validate checks migration contract invariants.
func (m Migration) Validate() error {
	if m.AppLabel == "" {
		return fmt.Errorf("%w: app label is required", ErrInvalidMigration)
	}
	if !migrationNamePattern.MatchString(m.Name) {
		return fmt.Errorf("%w: invalid migration name %q", ErrInvalidMigration, m.Name)
	}
	if len(m.Operations) == 0 {
		return fmt.Errorf("%w: operations are required", ErrInvalidMigration)
	}
	if err := validateUniqueDependencies("dependencies", m.Dependencies); err != nil {
		return err
	}
	if err := validateUniqueDependencies("run_before", m.RunBefore); err != nil {
		return err
	}
	return validateUniqueDependencies("replaces", m.Replaces)
}

// InitialMigrationName returns the first migration name.
func InitialMigrationName() string {
	return "0001_initial"
}

// NextMigrationName renders a deterministic numbered migration name.
func NextMigrationName(number int, slug string) string {
	return fmt.Sprintf("%04d_%s", number, slugify(slug))
}

func validateUniqueDependencies(label string, dependencies []Dependency) error {
	seen := map[string]struct{}{}
	for _, dependency := range dependencies {
		if dependency.AppLabel == "" || dependency.Name == "" {
			return fmt.Errorf("%w: %s contain empty dependency", ErrInvalidMigration, label)
		}
		if !migrationNamePattern.MatchString(dependency.Name) {
			return fmt.Errorf("%w: %s contain invalid dependency %s", ErrInvalidMigration, label, dependency.Identity())
		}
		key := dependency.Identity()
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: duplicate dependency %s", ErrInvalidMigration, key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func slugify(value string) string {
	var builder strings.Builder
	underscore := false
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			underscore = false
			continue
		}
		if !underscore {
			builder.WriteByte('_')
			underscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}
