package static

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
)

// FoundFile is one discovered static file.
type FoundFile struct {
	Path     string
	Content  []byte
	Origin   string
	Priority int
}

// Duplicate describes a shadowed static path.
type Duplicate struct {
	Path   string
	Winner FoundFile
	Losers []FoundFile
}

func HashedName(name string, content []byte) string {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])[:12]
	extension := path.Ext(name)
	base := strings.TrimSuffix(name, extension)
	return fmt.Sprintf("%s.%s%s", base, hash, extension)
}

func BuildManifest(files []FoundFile) map[string]string {
	manifest := make(map[string]string, len(files))
	for _, file := range files {
		manifest[file.Path] = HashedName(file.Path, file.Content)
	}
	return manifest
}

func safeStaticPath(name string) (string, error) {
	name = strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("unsafe static path %q", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return "", fmt.Errorf("unsafe static path %q", name)
		}
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("unsafe static path %q", name)
	}
	return cleaned, nil
}
