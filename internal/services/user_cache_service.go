package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/internal/domain"
)

type UserCacheService struct {
	client         *redis.Client
	userTTL        time.Duration
	tokenTTL       time.Duration
	userKeyPrefix  string
	tokenKeyPrefix string
}

type CachedUser struct {
	ID       int             `json:"id"`
	Role     domain.UserRole `json:"role"`
	IsActive bool            `json:"is_active"`
}

func NewUserCacheService(cfg config.RedisConfig) (*UserCacheService, error) {
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
		return nil, fmt.Errorf("ошибка подключения к Redis для user cache: %w", err)
	}

	return &UserCacheService{
		client:         client,
		userTTL:        15 * time.Minute, // Кеш пользователя на 15 минут
		tokenTTL:       5 * time.Minute,  // Кеш валидации токена на 5 минут
		userKeyPrefix:  "user:cache:",
		tokenKeyPrefix: "token:validation:",
	}, nil
}

func (s *UserCacheService) CacheUser(userID int, user *domain.User) error {
	ctx := context.Background()

	cachedUser := &CachedUser{
		ID:       user.ID,
		Role:     user.Role,
		IsActive: user.IsActive,
	}

	userJSON, err := json.Marshal(cachedUser)
	if err != nil {
		return fmt.Errorf("ошибка сериализации пользователя: %w", err)
	}

	key := s.getUserKey(userID)
	err = s.client.Set(ctx, key, userJSON, s.userTTL).Err()
	if err != nil {
		return fmt.Errorf("ошибка кеширования пользователя: %w", err)
	}

	return nil
}

func (s *UserCacheService) GetCachedUser(userID int) (*CachedUser, error) {
	ctx := context.Background()
	key := s.getUserKey(userID)

	userJSON, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения пользователя из кеша: %w", err)
	}

	var user CachedUser
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации пользователя: %w", err)
	}

	return &user, nil
}

func (s *UserCacheService) CacheTokenValidation(token string, user *domain.User) error {
	ctx := context.Background()

	cachedUser := &CachedUser{
		ID:       user.ID,
		Role:     user.Role,
		IsActive: user.IsActive,
	}

	userJSON, err := json.Marshal(cachedUser)
	if err != nil {
		return fmt.Errorf("ошибка сериализации валидации токена: %w", err)
	}

	key := s.getTokenKey(token)
	err = s.client.Set(ctx, key, userJSON, s.tokenTTL).Err()
	if err != nil {
		return fmt.Errorf("ошибка кеширования валидации токена: %w", err)
	}

	return nil
}

func (s *UserCacheService) GetCachedTokenValidation(token string) (*CachedUser, error) {
	ctx := context.Background()
	key := s.getTokenKey(token)

	userJSON, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения валидации токена из кеша: %w", err)
	}

	var user CachedUser
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации валидации токена: %w", err)
	}

	return &user, nil
}

func (s *UserCacheService) InvalidateUser(userID int) error {
	ctx := context.Background()
	key := s.getUserKey(userID)
	return s.client.Del(ctx, key).Err()
}

func (s *UserCacheService) InvalidateToken(token string) error {
	ctx := context.Background()
	key := s.getTokenKey(token)
	return s.client.Del(ctx, key).Err()
}

func (s *UserCacheService) InvalidateAllTokens() error {
	ctx := context.Background()
	pattern := s.tokenKeyPrefix + "*"
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("ошибка удаления ключа %s из кэша: %w", iter.Val(), err)
		}
	}
	return iter.Err()
}

func (s *UserCacheService) getUserKey(userID int) string {
	return fmt.Sprintf("%s%d", s.userKeyPrefix, userID)
}

func (s *UserCacheService) getTokenKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%s%x", s.tokenKeyPrefix, hash)
}
