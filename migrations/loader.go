package migrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type manifest struct {
	AppLabel     string       `json:"app_label"`
	Name         string       `json:"name"`
	Dependencies []Dependency `json:"dependencies"`
	Operations   []string     `json:"operations"`
}

// Loader loads migration manifests from folders.
type Loader struct {
	Paths []string
}

// NewLoader creates a migration loader.
func NewLoader(paths []string) Loader {
	return Loader{Paths: append([]string(nil), paths...)}
}

// Load loads migrations from manifest files.
func (l Loader) Load() ([]Migration, error) {
	var migrations []Migration
	for _, root := range l.Paths {
		files, err := filepath.Glob(filepath.Join(root, "*.migration.json"))
		if err != nil {
			return nil, err
		}
		sort.Strings(files)
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}
			var item manifest
			if err := json.Unmarshal(content, &item); err != nil {
				return nil, err
			}
			migration := Migration{AppLabel: item.AppLabel, Name: item.Name, Dependencies: item.Dependencies}
			for _, name := range item.Operations {
				migration.Operations = append(migration.Operations, ManifestOperation{NameValue: name})
			}
			migrations = append(migrations, migration)
		}
	}
	return migrations, nil
}

// WriteManifest writes a deterministic migration manifest for tests and loaders.
func WriteManifest(dir string, migration Migration) error {
	item := manifest{AppLabel: migration.AppLabel, Name: migration.Name, Dependencies: append([]Dependency(nil), migration.Dependencies...)}
	for _, operation := range migration.Operations {
		item.Operations = append(item.Operations, operation.Name())
	}
	content, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(filepath.Join(dir, migration.Name+".migration.json"), content, 0o644)
}
