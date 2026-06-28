package files

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LocalOptions configures local filesystem storage.
type LocalOptions struct {
	BaseURL   string
	Overwrite bool
}

// LocalStorage stores files under one filesystem root.
type LocalStorage struct {
	root      string
	baseURL   string
	overwrite bool
}

func NewLocalStorage(root string, options LocalOptions) *LocalStorage {
	return &LocalStorage{
		root:      filepath.Clean(root),
		baseURL:   options.BaseURL,
		overwrite: options.Overwrite,
	}
}

func (s *LocalStorage) Open(_ context.Context, name string) (io.ReadCloser, error) {
	fullPath, err := s.fullPath(name)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, name)
	}
	return file, err
}

func (s *LocalStorage) Save(ctx context.Context, name string, content io.Reader) (string, error) {
	normalized, err := NormalizeName(name)
	if err != nil {
		return "", err
	}
	for index := 0; ; index++ {
		candidate := normalized
		if index > 0 {
			candidate = collisionName(normalized, index)
		}
		if !s.overwrite {
			exists, err := s.Exists(ctx, candidate)
			if err != nil {
				return "", err
			}
			if exists {
				continue
			}
		}
		if err := s.write(candidate, content); err != nil {
			if !s.overwrite && os.IsExist(err) {
				continue
			}
			return "", err
		}
		return candidate, nil
	}
}

func (s *LocalStorage) Delete(_ context.Context, name string) error {
	fullPath, err := s.fullPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) Exists(_ context.Context, name string) (bool, error) {
	fullPath, err := s.fullPath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) List(_ context.Context, prefix string) ([]string, error) {
	root := s.root
	if strings.TrimSpace(prefix) != "" {
		normalized, err := NormalizeName(prefix)
		if err != nil {
			return nil, err
		}
		root, err = s.fullPath(normalized)
		if err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	var names []string
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}
		names = append(names, filepath.ToSlash(relative))
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func (s *LocalStorage) Size(_ context.Context, name string) (int64, error) {
	info, err := s.stat(name)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *LocalStorage) URL(name string) (string, error) {
	normalized, err := NormalizeName(name)
	if err != nil {
		return "", err
	}
	if s.baseURL == "" {
		return normalized, nil
	}
	return strings.TrimRight(s.baseURL, "/") + "/" + normalized, nil
}

func (s *LocalStorage) ModifiedTime(_ context.Context, name string) (time.Time, error) {
	info, err := s.stat(name)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func (s *LocalStorage) Path(name string) (string, error) {
	return s.fullPath(name)
}

func (s *LocalStorage) write(name string, content io.Reader) error {
	fullPath, err := s.fullPath(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE
	if s.overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	file, err := os.OpenFile(fullPath, flags, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, content)
	return err
}

func (s *LocalStorage) stat(name string) (os.FileInfo, error) {
	fullPath, err := s.fullPath(name)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, name)
	}
	return info, err
}

func (s *LocalStorage) fullPath(name string) (string, error) {
	normalized, err := NormalizeName(name)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(s.root, filepath.FromSlash(normalized))
	relative, err := filepath.Rel(s.root, fullPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." || filepath.IsAbs(relative) {
		return "", fmt.Errorf("%w: %s", ErrUnsafeName, name)
	}
	return fullPath, nil
}
