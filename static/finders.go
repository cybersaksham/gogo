package static

import (
	"os"
	"path/filepath"
	"sort"
)

// FinderConfig configures static-file discovery.
type FinderConfig struct {
	ProjectDirs    []string
	AppDirs        []string
	FrameworkFiles map[string][]byte
}

// Find discovers static files with project > app > framework precedence.
func Find(config FinderConfig) ([]FoundFile, []Duplicate, error) {
	byPath := map[string][]FoundFile{}
	for name, content := range config.FrameworkFiles {
		normalized, err := safeStaticPath(name)
		if err != nil {
			return nil, nil, err
		}
		byPath[normalized] = append(byPath[normalized], FoundFile{Path: normalized, Content: append([]byte(nil), content...), Origin: "framework", Priority: 0})
	}
	for _, dir := range config.AppDirs {
		files, err := loadStaticDir(dir, "app", 1)
		if err != nil {
			return nil, nil, err
		}
		for _, file := range files {
			byPath[file.Path] = append(byPath[file.Path], file)
		}
	}
	for _, dir := range config.ProjectDirs {
		files, err := loadStaticDir(dir, "project", 2)
		if err != nil {
			return nil, nil, err
		}
		for _, file := range files {
			byPath[file.Path] = append(byPath[file.Path], file)
		}
	}

	paths := make([]string, 0, len(byPath))
	for name := range byPath {
		paths = append(paths, name)
	}
	sort.Strings(paths)

	winners := make([]FoundFile, 0, len(paths))
	duplicates := make([]Duplicate, 0)
	for _, name := range paths {
		candidates := byPath[name]
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Priority != candidates[j].Priority {
				return candidates[i].Priority > candidates[j].Priority
			}
			return candidates[i].Origin < candidates[j].Origin
		})
		winners = append(winners, candidates[0])
		if len(candidates) > 1 {
			duplicates = append(duplicates, Duplicate{Path: name, Winner: candidates[0], Losers: append([]FoundFile(nil), candidates[1:]...)})
		}
	}
	return winners, duplicates, nil
}

func loadStaticDir(root, origin string, priority int) ([]FoundFile, error) {
	var files []FoundFile
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return files, nil
	}
	err := filepath.WalkDir(root, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		normalized, err := safeStaticPath(filepath.ToSlash(relative))
		if err != nil {
			return err
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		files = append(files, FoundFile{Path: normalized, Content: content, Origin: origin, Priority: priority})
		return nil
	})
	return files, err
}
