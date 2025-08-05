package model

import "time"

// CacheInfo contains statistics and metadata about the local cache
type CacheInfo struct {
	LastUpdated  time.Time `json:"last_updated"`
	TotalPrompts int       `json:"total_prompts"`
	CacheSize    int64     `json:"cache_size_bytes"`
}