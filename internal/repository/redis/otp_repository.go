package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/internal/domain"
)

type OTPRepository struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewOTPRepository(cfg config.RedisConfig) (*OTPRepository, error) {
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
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return &OTPRepository{
		client: client,
		prefix: "otp:session:",
		ttl:    10 * time.Minute,
	}, nil
}

func (r *OTPRepository) CreateSession(session *domain.OTPSession) error {
	ctx := context.Background()

	key := r.getKey(session.ID)

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("ошибка сериализации OTP сессии: %w", err)
	}

	err = r.client.Set(ctx, key, sessionJSON, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("ошибка сохранения OTP сессии в Redis: %w", err)
	}

	phoneKey := r.getPhoneKey(session.Phone)
	err = r.client.Set(ctx, phoneKey, session.ID, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("ошибка создания индекса по телефону: %w", err)
	}

	return nil
}

func (r *OTPRepository) GetSessionByID(id string) (*domain.OTPSession, error) {
	ctx := context.Background()

	key := r.getKey(id)

	sessionJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения OTP сессии из Redis: %w", err)
	}

	var session domain.OTPSession
	err = json.Unmarshal([]byte(sessionJSON), &session)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации OTP сессии: %w", err)
	}

	return &session, nil
}

func (r *OTPRepository) GetSessionByPhone(phone string) (*domain.OTPSession, error) {
	ctx := context.Background()

	phoneKey := r.getPhoneKey(phone)

	sessionID, err := r.client.Get(ctx, phoneKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения ID сессии по телефону: %w", err)
	}

	return r.GetSessionByID(sessionID)
}

func (r *OTPRepository) UpdateSession(session *domain.OTPSession) error {
	ctx := context.Background()

	key := r.getKey(session.ID)

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("ошибка сериализации OTP сессии при обновлении: %w", err)
	}

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		ttl = r.ttl
	}

	err = r.client.Set(ctx, key, sessionJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("ошибка обновления OTP сессии в Redis: %w", err)
	}

	return nil
}

func (r *OTPRepository) DeleteSession(id string) error {
	ctx := context.Background()

	session, err := r.GetSessionByID(id)
	if err == nil && session != nil {
		phoneKey := r.getPhoneKey(session.Phone)
		r.client.Del(ctx, phoneKey)
	}

	key := r.getKey(id)
	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("ошибка удаления OTP сессии из Redis: %w", err)
	}

	return nil
}

func (r *OTPRepository) DeleteExpiredSessions() error {
	ctx := context.Background()

	phonePattern := r.getPhoneKey("*")
	keys, err := r.client.Keys(ctx, phonePattern).Result()
	if err != nil {
		return fmt.Errorf("ошибка получения ключей индексов телефонов: %w", err)
	}

	var expiredKeys []string

	for _, phoneKey := range keys {
		sessionID, err := r.client.Get(ctx, phoneKey).Result()
		if err != nil {
			continue
		}

		sessionKey := r.getKey(sessionID)
		exists, err := r.client.Exists(ctx, sessionKey).Result()
		if err != nil || exists == 0 {
			expiredKeys = append(expiredKeys, phoneKey)
		}
	}

	if len(expiredKeys) > 0 {
		err = r.client.Del(ctx, expiredKeys...).Err()
		if err != nil {
			return fmt.Errorf("ошибка удаления протухших индексов: %w", err)
		}
	}

	return nil
}

func (r *OTPRepository) getKey(sessionID string) string {
	return fmt.Sprintf("%s%s", r.prefix, sessionID)
}

func (r *OTPRepository) getPhoneKey(phone string) string {
	return fmt.Sprintf("otp:phone:%s", phone)
}
