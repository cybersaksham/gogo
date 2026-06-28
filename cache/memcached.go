package cache

import "time"

type MemcachedOptions struct {
	BackendOptions
	Servers []string
	Now     func() time.Time
}

type MemcachedCache struct {
	*LocalCache
	Servers []string
}

func NewMemcachedCache(options MemcachedOptions) *MemcachedCache {
	return &MemcachedCache{LocalCache: NewLocalCache(LocalOptions{BackendOptions: options.BackendOptions, Now: options.Now}), Servers: append([]string(nil), options.Servers...)}
}
