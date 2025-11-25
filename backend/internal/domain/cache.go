package domain

import (
	"time"
)

// CacheType represents the type of content being cached
type CacheType string

const (
	CacheTypeScript CacheType = "script"
	CacheTypeScene  CacheType = "scene"
	CacheTypeAudio  CacheType = "audio"
)

// CacheEntry represents a generic cache entry
type CacheEntry struct {
	Key           string      `json:"key"`
	Type          CacheType   `json:"type"`
	ContentHash   string      `json:"content_hash"`
	S3URL         string      `json:"s3_url,omitempty"`
	Metadata      interface{} `json:"metadata,omitempty"`
	CreatedAt     int64       `json:"created_at"`
	LastAccessed  int64       `json:"last_accessed"`
	AccessCount   int         `json:"access_count"`
	OriginalJobID string      `json:"original_job_id"`
	TTL           int64       `json:"ttl"` // Unix timestamp when entry expires
}

// CacheStats represents cache performance metrics
type CacheStats struct {
	ScriptHitRate   float64 `json:"script_hit_rate"`
	SceneHitRate    float64 `json:"scene_hit_rate"`
	AudioHitRate    float64 `json:"audio_hit_rate"`
	TotalSavings    float64 `json:"total_savings_usd"`
	CacheSize       int64   `json:"cache_size_bytes"`
	EntriesCount    int     `json:"entries_count"`
	AvgResponseTime int     `json:"avg_response_time_ms"`
}

// CacheConfig holds configuration for caching
type CacheConfig struct {
	RedisURL       string        `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
	DefaultTTL     time.Duration `envconfig:"CACHE_TTL" default:"1h"`
	MaxSize        int64         `envconfig:"CACHE_MAX_SIZE" default:"10737418240"` // 10GB
	EnableCache    bool          `envconfig:"ENABLE_CACHE" default:"true"`
	ScriptHitRatio float64       `envconfig:"SCRIPT_HIT_RATIO" default:"0.3"`
	SceneHitRatio  float64       `envconfig:"SCENE_HIT_RATIO" default:"0.15"`
}
