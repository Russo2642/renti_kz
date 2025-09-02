package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/internal/domain"
)

type redisQueueService struct {
	client          *redis.Client
	notificationKey string
	delayedKey      string
	processingKey   string
	enableQueue     bool
}

func NewRedisQueueService(redisConfig config.RedisConfig, notificationConfig config.NotificationConfig) (domain.MessageQueueService, error) {
	service := &redisQueueService{
		notificationKey: notificationConfig.RedisQueueName,
		delayedKey:      notificationConfig.RedisQueueName + ":delayed",
		processingKey:   notificationConfig.RedisQueueName + ":processing",
		enableQueue:     true,
	}

	if redisConfig.Host == "" {
		log.Println("⚠️ Redis host не указан, message queue отключена")
		service.enableQueue = false
		return service, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr(),
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
		MaxRetries:   redisConfig.MaxRetries,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
		PoolTimeout:  redisConfig.PoolTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	service.client = client
	log.Println("📨 Redis message queue инициализована")

	return service, nil
}

func (s *redisQueueService) PublishNotification(message *domain.PushMessage) error {
	if !s.enableQueue {
		log.Printf("📭 Сообщение пропущено (очередь отключена): %s", message.Title)
		return nil
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("ошибка сериализации сообщения: %w", err)
	}

	ctx := context.Background()
	err = s.client.LPush(ctx, s.notificationKey, messageJSON).Err()
	if err != nil {
		if strings.Contains(err.Error(), "READONLY") {
			log.Printf("⚠️ Redis в режиме только для чтения, сообщение не добавлено: %s", message.Title)
			return fmt.Errorf("redis недоступен для записи: %w", err)
		}
		return fmt.Errorf("ошибка добавления в очередь: %w", err)
	}

	log.Printf("📝 Сообщение добавлено в очередь: %s", message.Title)
	return nil
}

func (s *redisQueueService) PublishDelayedNotification(message *domain.PushMessage, delay time.Duration) error {
	if !s.enableQueue {
		log.Printf("📭 Отложенное сообщение пропущено (очередь отключена): %s", message.Title)
		return nil
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("ошибка сериализации отложенного сообщения: %w", err)
	}

	executeAt := time.Now().Add(delay).Unix()

	ctx := context.Background()
	err = s.client.ZAdd(ctx, s.delayedKey, redis.Z{
		Score:  float64(executeAt),
		Member: messageJSON,
	}).Err()

	if err != nil {
		if strings.Contains(err.Error(), "READONLY") {
			log.Printf("⚠️ Redis в режиме только для чтения, отложенное сообщение не добавлено: %s", message.Title)
			return fmt.Errorf("redis недоступен для записи: %w", err)
		}
		return fmt.Errorf("ошибка добавления отложенного сообщения: %w", err)
	}

	log.Printf("⏰ Отложенное сообщение добавлено (через %v): %s", delay, message.Title)
	return nil
}

func (s *redisQueueService) ConsumeNotifications(handler func(*domain.PushMessage) error) error {
	if !s.enableQueue {
		log.Println("📭 Consumer не запущен (очередь отключена)")
		return nil
	}

	log.Println("🔄 Запуск consumer для обработки уведомлений...")

	go s.processDelayedMessages(handler)

	ctx := context.Background()
	retryCount := 0
	maxRetries := 3

	for {
		result, err := s.client.BRPop(ctx, 1*time.Second, s.notificationKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}

			if strings.Contains(err.Error(), "READONLY") {
				retryCount++
				log.Printf("❌ Redis в режиме только для чтения (попытка %d/%d): %v", retryCount, maxRetries, err)

				if retryCount >= maxRetries {
					log.Printf("❌ Превышено максимальное количество попыток подключения к Redis")
					return fmt.Errorf("redis недоступен для записи после %d попыток: %w", maxRetries, err)
				}

				waitTime := time.Duration(retryCount*retryCount) * 10 * time.Second
				log.Printf("⏰ Ожидание %v перед повторной попыткой...", waitTime)
				time.Sleep(waitTime)
				continue
			}

			retryCount = 0
			log.Printf("❌ Ошибка чтения из очереди: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		retryCount = 0

		if len(result) < 2 {
			continue
		}

		messageJSON := result[1]
		err = s.processMessage(messageJSON, handler)
		if err != nil {
			log.Printf("❌ Ошибка обработки сообщения: %v", err)
		}
	}
}

func (s *redisQueueService) processMessage(messageJSON string, handler func(*domain.PushMessage) error) error {
	var message domain.PushMessage
	err := json.Unmarshal([]byte(messageJSON), &message)
	if err != nil {
		return fmt.Errorf("ошибка десериализации сообщения: %w", err)
	}

	log.Printf("⚡ Обрабатываем сообщение: %s для пользователя %d", message.Title, message.UserID)

	err = handler(&message)
	if err != nil {
		return fmt.Errorf("ошибка обработки handler: %w", err)
	}

	log.Printf("✅ Сообщение обработано успешно: %s", message.Title)
	return nil
}

func (s *redisQueueService) processDelayedMessages(handler func(*domain.PushMessage) error) {
	log.Println("⏰ Запуск процессора отложенных сообщений...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()

	for range ticker.C {
		now := time.Now().Unix()

		messages, err := s.client.ZRangeByScore(ctx, s.delayedKey, &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%d", now),
		}).Result()

		if err != nil {
			log.Printf("❌ Ошибка получения отложенных сообщений: %v", err)
			continue
		}

		if len(messages) == 0 {
			continue
		}

		log.Printf("⏰ Найдено %d отложенных сообщений для обработки", len(messages))

		for _, messageJSON := range messages {
			s.client.ZRem(ctx, s.delayedKey, messageJSON)

			err := s.processMessage(messageJSON, handler)
			if err != nil {
				log.Printf("❌ Ошибка обработки отложенного сообщения: %v", err)
			}
		}
	}
}

func (s *redisQueueService) GetQueueSize() (int64, error) {
	if !s.enableQueue {
		return 0, nil
	}

	ctx := context.Background()
	return s.client.LLen(ctx, s.notificationKey).Result()
}

func (s *redisQueueService) GetDelayedCount() (int64, error) {
	if !s.enableQueue {
		return 0, nil
	}

	ctx := context.Background()
	return s.client.ZCard(ctx, s.delayedKey).Result()
}

func (s *redisQueueService) ClearQueues() error {
	if !s.enableQueue {
		return nil
	}

	ctx := context.Background()

	err := s.client.Del(ctx, s.notificationKey, s.delayedKey, s.processingKey).Err()
	if err != nil {
		return fmt.Errorf("ошибка очистки очередей: %w", err)
	}

	log.Println("🧹 Очереди очищены")
	return nil
}

func (s *redisQueueService) GetStats() map[string]interface{} {
	if !s.enableQueue {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	queueSize, _ := s.GetQueueSize()
	delayedCount, _ := s.GetDelayedCount()

	return map[string]interface{}{
		"enabled":       true,
		"queue_size":    queueSize,
		"delayed_count": delayedCount,
		"redis_info":    s.getRedisInfo(),
	}
}

func (s *redisQueueService) getRedisInfo() map[string]string {
	ctx := context.Background()

	info, err := s.client.Info(ctx, "memory").Result()
	if err != nil {
		return map[string]string{"error": err.Error()}
	}

	return map[string]string{
		"memory_info": info,
	}
}
