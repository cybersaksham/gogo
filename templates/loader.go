package templates

import (
	"os"
	"path/filepath"
)

// LoaderConfig configures template discovery.
type LoaderConfig struct {
	ProjectDirs        []string
	AppDirs            []string
	FrameworkTemplates map[string]string
}

// LoadTemplates loads templates with project > app > framework precedence.
func LoadTemplates(config LoaderConfig) (map[string]string, error) {
	loaded := make(map[string]string, len(config.FrameworkTemplates))
	for name, source := range config.FrameworkTemplates {
		loaded[name] = source
	}
	for _, dir := range config.AppDirs {
		templates, err := loadTemplatePath(dir)
		if err != nil {
			return nil, err
		}
		for name, source := range templates {
			loaded[name] = source
		}
	}
	for _, dir := range config.ProjectDirs {
		templates, err := loadTemplatePath(dir)
		if err != nil {
			return nil, err
		}
		for name, source := range templates {
			loaded[name] = source
		}
	}
	return loaded, nil
}

func loadTemplatePath(root string) (map[string]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		content, err := os.ReadFile(root)
		if err != nil {
			return nil, err
		}
		return map[string]string{filepath.ToSlash(filepath.Base(root)): string(content)}, nil
	}

	templates := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		templates[filepath.ToSlash(relative)] = string(content)
		return nil
	}); err != nil {
		return nil, err
	}
	return templates, nil
}
