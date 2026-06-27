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
	builder.WriteString("// GeneratedMigration describes this generated migration.\n")
	builder.WriteString("var GeneratedMigration = Migration{\n")
	builder.WriteString(fmt.Sprintf("\tAppLabel: %q,\n", migration.AppLabel))
	builder.WriteString(fmt.Sprintf("\tName: %q,\n", migration.Name))
	builder.WriteString("\tDependencies: []Dependency{\n")
	for _, dependency := range migration.Dependencies {
		builder.WriteString(fmt.Sprintf("\t\t{AppLabel: %q, Name: %q},\n", dependency.AppLabel, dependency.Name))
	}
	builder.WriteString("\t},\n")
	builder.WriteString("\tOperations: []Operation{\n")
	for _, operation := range migration.Operations {
		builder.WriteString(fmt.Sprintf("\t\tManifestOperation{NameValue: %q},\n", operation.Name()))
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
