package migrations

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// Writer writes deterministic Go migration files.
type Writer struct {
	Dir string
}

// NewWriter creates a migration writer.
func NewWriter(dir string) Writer {
	return Writer{Dir: dir}
}

// Write writes a Go migration file and returns its path.
func (w Writer) Write(migration Migration) (string, error) {
	var builder strings.Builder
	builder.WriteString("package migrations\n\n")
	builder.WriteString("import gogomigrations \"github.com/cybersaksham/gogo/migrations\"\n\n")
	variableName := generatedMigrationVariableName(migration.Name)
	builder.WriteString("// " + variableName + " describes this generated migration.\n")
	builder.WriteString("var " + variableName + " = gogomigrations.Migration{\n")
	builder.WriteString(fmt.Sprintf("\tAppLabel: %q,\n", migration.AppLabel))
	builder.WriteString(fmt.Sprintf("\tName: %q,\n", migration.Name))
	builder.WriteString("\tDependencies: []gogomigrations.Dependency{\n")
	for _, dependency := range migration.Dependencies {
		builder.WriteString(fmt.Sprintf("\t\t{AppLabel: %q, Name: %q},\n", dependency.AppLabel, dependency.Name))
	}
	builder.WriteString("\t},\n")
	builder.WriteString("\tReplaces: []gogomigrations.Dependency{\n")
	for _, dependency := range migration.Replaces {
		builder.WriteString(fmt.Sprintf("\t\t{AppLabel: %q, Name: %q},\n", dependency.AppLabel, dependency.Name))
	}
	builder.WriteString("\t},\n")
	builder.WriteString("\tOperations: []gogomigrations.Operation{\n")
	for _, operation := range migration.Operations {
		builder.WriteString(fmt.Sprintf("\t\tgogomigrations.ManifestOperation{NameValue: %q},\n", operation.Name()))
	}
	builder.WriteString("\t},\n")
	builder.WriteString("}\n")
	content := []byte(builder.String())
	formatted, err := format.Source(content)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(w.Dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(w.Dir, migration.Name+".go")
	return path, os.WriteFile(path, formatted, 0o644)
}

func generatedMigrationVariableName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r < '0' || (r > '9' && r < 'A') || (r > 'Z' && r < 'a') || r > 'z'
	})
	var builder strings.Builder
	builder.WriteString("GeneratedMigration")
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}
