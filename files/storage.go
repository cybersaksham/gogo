package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
)

var (
	ErrUnsafeName        = errors.New("unsafe file name")
	ErrFileNotFound      = errors.New("file not found")
	ErrUploadTooLarge    = errors.New("upload too large")
	ErrUploadInterrupted = errors.New("upload interrupted")
	ErrInvalidUpload     = errors.New("invalid upload")
)

// Storage defines the framework file-storage contract.
type Storage interface {
	Open(context.Context, string) (io.ReadCloser, error)
	Save(context.Context, string, io.Reader) (string, error)
	Delete(context.Context, string) error
	Exists(context.Context, string) (bool, error)
	List(context.Context, string) ([]string, error)
	Size(context.Context, string) (int64, error)
	URL(string) (string, error)
	ModifiedTime(context.Context, string) (time.Time, error)
	Path(string) (string, error)
}

// NormalizeName canonicalizes a storage name and rejects traversal attempts.
func NormalizeName(name string) (string, error) {
	name = strings.ReplaceAll(strings.TrimSpace(name), "\\", "/")
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("%w: %q", ErrUnsafeName, name)
	}
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if part == ".." {
			return "", fmt.Errorf("%w: %q", ErrUnsafeName, name)
		}
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("%w: %q", ErrUnsafeName, name)
	}
	return cleaned, nil
}

func collisionName(name string, index int) string {
	extension := path.Ext(name)
	base := strings.TrimSuffix(name, extension)
	return fmt.Sprintf("%s_%d%s", base, index, extension)
}
