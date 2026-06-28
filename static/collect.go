package static

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CollectOptions configures collectstatic output.
type CollectOptions struct {
	Finder      FinderConfig
	Destination string
	Manifest    bool
	Clear       bool
}

// CollectedFile describes one copied output file.
type CollectedFile struct {
	SourcePath string
	OutputPath string
	Origin     string
	Size       int64
}

// CollectResult summarizes one collectstatic run.
type CollectResult struct {
	Copied     []CollectedFile
	Duplicates []Duplicate
	Manifest   map[string]string
}

func Collect(ctx context.Context, options CollectOptions) (CollectResult, error) {
	if options.Destination == "" {
		return CollectResult{}, fmt.Errorf("static destination is required")
	}
	files, duplicates, err := Find(options.Finder)
	if err != nil {
		return CollectResult{}, err
	}
	if options.Clear {
		if err := os.RemoveAll(options.Destination); err != nil {
			return CollectResult{}, err
		}
	}
	if err := os.MkdirAll(options.Destination, 0o755); err != nil {
		return CollectResult{}, err
	}

	result := CollectResult{Duplicates: duplicates}
	if options.Manifest {
		result.Manifest = BuildManifest(files)
	} else {
		result.Manifest = map[string]string{}
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return CollectResult{}, err
		}
		outputPath := file.Path
		if options.Manifest {
			outputPath = result.Manifest[file.Path]
		}
		if err := writeCollectedFile(options.Destination, outputPath, file.Content); err != nil {
			return CollectResult{}, err
		}
		result.Copied = append(result.Copied, CollectedFile{
			SourcePath: file.Path,
			OutputPath: outputPath,
			Origin:     file.Origin,
			Size:       int64(len(file.Content)),
		})
	}
	if options.Manifest {
		encoded, err := json.MarshalIndent(result.Manifest, "", "  ")
		if err != nil {
			return CollectResult{}, err
		}
		if err := writeCollectedFile(options.Destination, "staticfiles.json", encoded); err != nil {
			return CollectResult{}, err
		}
	}
	return result, nil
}

func writeCollectedFile(root, name string, content []byte) error {
	normalized, err := safeStaticPath(name)
	if err != nil {
		return err
	}
	target := filepath.Join(root, filepath.FromSlash(normalized))
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if relative == ".." || filepath.IsAbs(relative) {
		return fmt.Errorf("unsafe static output path %q", name)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, content, 0o644)
}
