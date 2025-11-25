package service

import (
	"context"
	"fmt"

	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/utils"
	"go.uber.org/zap"
)

// CacheService handles caching of generated content
type CacheService struct {
	redisCache *repository.RedisCache
	logger     *zap.Logger
	config     domain.CacheConfig
}

// NewCacheService creates a new cache service
func NewCacheService(
	redisCache *repository.RedisCache,
	logger *zap.Logger,
	config domain.CacheConfig,
) *CacheService {
	return &CacheService{
		redisCache: redisCache,
		logger:     logger,
		config:     config,
	}
}

// GenerateScriptCacheKey generates a deterministic key for script caching
func (s *CacheService) GenerateScriptCacheKey(job *domain.Job) string {
	keyData := struct {
		Prompt           string
		Duration         int
		AspectRatio      string
		Style            string
		Tone             string
		Platform         string
		IsPharmaceutical bool
	}{
		Prompt:           utils.NormalizePrompt(job.Prompt),
		Duration:         job.Duration,
		AspectRatio:      job.AspectRatio,
		Style:            job.Style,
		Tone:             job.Tone,
		Platform:         job.Platform,
		IsPharmaceutical: job.Voice != "" || job.SideEffectsText != "",
	}

	hash := utils.HashContent(keyData)
	return fmt.Sprintf("cache:script:%s", hash)
}

// GetScript retrieves a script from cache
func (s *CacheService) GetScript(ctx context.Context, key string) (*domain.Script, bool) {
	if !s.config.EnableCache {
		return nil, false
	}

	var script domain.Script
	found, err := s.redisCache.Get(ctx, key, &script)
	if err != nil {
		s.logger.Warn("Cache lookup failed", zap.Error(err))
		return nil, false
	}

	return &script, found
}

// StoreScript stores a script in cache
func (s *CacheService) StoreScript(ctx context.Context, key string, script *domain.Script) error {
	if !s.config.EnableCache {
		return nil
	}
	return s.redisCache.Set(ctx, key, script)
}

// GenerateAudioCacheKey generates a key for audio caching
func (s *CacheService) GenerateAudioCacheKey(audioSpec domain.AudioSpec, duration int) string {
	keyData := struct {
		Mood     string
		Style    string
		Duration int
	}{
		Mood:     audioSpec.MusicMood,
		Style:    audioSpec.MusicStyle,
		Duration: duration,
	}

	hash := utils.HashContent(keyData)
	return fmt.Sprintf("cache:audio:%s", hash)
}

// GetAudio retrieves cached audio URL
func (s *CacheService) GetAudio(ctx context.Context, key string) (string, bool) {
	if !s.config.EnableCache {
		return "", false
	}

	var audioURL string
	found, err := s.redisCache.Get(ctx, key, &audioURL)
	if err != nil {
		return "", false
	}

	return audioURL, found
}

// StoreAudio stores audio URL in cache
func (s *CacheService) StoreAudio(ctx context.Context, key string, audioURL string) error {
	if !s.config.EnableCache {
		return nil
	}
	return s.redisCache.Set(ctx, key, audioURL)
}
