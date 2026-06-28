package cache

import "time"

type FileOptions struct {
	BackendOptions
	Directory string
	Now       func() time.Time
}

type FileCache struct {
	*LocalCache
	Directory string
}

func NewFileCache(options FileOptions) *FileCache {
	return &FileCache{LocalCache: NewLocalCache(LocalOptions{BackendOptions: options.BackendOptions, Now: options.Now}), Directory: options.Directory}
}
