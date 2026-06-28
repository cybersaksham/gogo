package cache

import "time"

type DatabaseOptions struct {
	BackendOptions
	Table string
	Now   func() time.Time
}

type DatabaseCache struct {
	*LocalCache
	Table string
}

func NewDatabaseCache(options DatabaseOptions) *DatabaseCache {
	if options.Table == "" {
		options.Table = "gogo_cache"
	}
	return &DatabaseCache{LocalCache: NewLocalCache(LocalOptions{BackendOptions: options.BackendOptions, Now: options.Now}), Table: options.Table}
}
