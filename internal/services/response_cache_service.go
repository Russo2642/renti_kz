package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/russo2642/renti_kz/internal/config"
)

type ResponseCacheService struct {
	client     *redis.Client
	defaultTTL time.Duration
	shortTTL   time.Duration
	longTTL    time.Duration
	keyPrefix  string
}

func NewResponseCacheService(cfg config.RedisConfig) (*ResponseCacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis для response cache: %w", err)
	}

	return &ResponseCacheService{
		client:     client,
		defaultTTL: 10 * time.Minute, // Стандартный TTL 10 минут
		shortTTL:   2 * time.Minute,  // Короткий TTL 2 минуты для часто изменяемых данных
		longTTL:    60 * time.Minute, // Длинный TTL 1 час для редко изменяемых данных
		keyPrefix:  "api:cache:",
	}, nil
}

func (s *ResponseCacheService) CacheResponse(key string, data interface{}, ttl time.Duration) error {
	ctx := context.Background()

	if ttl == 0 {
		ttl = s.defaultTTL
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных для кеша: %w", err)
	}

	cacheKey := s.getCacheKey(key)
	err = s.client.Set(ctx, cacheKey, dataJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("ошибка кеширования ответа: %w", err)
	}

	return nil
}

func (s *ResponseCacheService) GetCachedResponse(key string, dest interface{}) (bool, error) {
	ctx := context.Background()
	cacheKey := s.getCacheKey(key)

	dataJSON, err := s.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("ошибка получения данных из кеша: %w", err)
	}

	err = json.Unmarshal([]byte(dataJSON), dest)
	if err != nil {
		return false, fmt.Errorf("ошибка десериализации данных из кеша: %w", err)
	}

	return true, nil
}

func (s *ResponseCacheService) InvalidatePattern(pattern string) error {
	ctx := context.Background()

	iter := s.client.Scan(ctx, 0, s.getCacheKey(pattern), 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("ошибка поиска ключей для удаления: %w", err)
	}

	if len(keys) > 0 {
		err := s.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("ошибка удаления ключей из кеша: %w", err)
		}
	}

	return nil
}

func (s *ResponseCacheService) Invalidate(key string) error {
	ctx := context.Background()
	cacheKey := s.getCacheKey(key)
	return s.client.Del(ctx, cacheKey).Err()
}

func (s *ResponseCacheService) GetDefaultTTL() time.Duration {
	return s.defaultTTL
}

func (s *ResponseCacheService) GetShortTTL() time.Duration {
	return s.shortTTL
}

func (s *ResponseCacheService) GetLongTTL() time.Duration {
	return s.longTTL
}

func (s *ResponseCacheService) getCacheKey(key string) string {
	return fmt.Sprintf("%s%s", s.keyPrefix, key)
}
