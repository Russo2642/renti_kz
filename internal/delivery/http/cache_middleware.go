package http

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/services"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// CacheMiddleware создает middleware для кеширования ответов
func CacheMiddleware(cacheService *services.ResponseCacheService, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Кешируем только GET запросы
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Создаем ключ кеша на основе URL и query параметров
		cacheKey := generateCacheKey(c)

		// Проверяем кеш
		var cachedResponse string
		found, err := cacheService.GetCachedResponse(cacheKey, &cachedResponse)
		if err == nil && found {
			// Возвращаем закешированный ответ
			c.Header("X-Cache", "HIT")
			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, cachedResponse)
			c.Abort()
			return
		}

		// Если не найдено в кеше, выполняем запрос
		c.Header("X-Cache", "MISS")

		// Создаем custom response writer для захвата ответа
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		c.Next()

		// Кешируем ответ только если статус 200
		if c.Writer.Status() == http.StatusOK && writer.body.Len() > 0 {
			// Используем переданный TTL или стандартный
			cacheTTL := ttl
			if cacheTTL == 0 {
				cacheTTL = cacheService.GetDefaultTTL()
			}

			// Кешируем сырой JSON ответ
			_ = cacheService.CacheResponse(cacheKey, writer.body.String(), cacheTTL)
		}
	}
}

// CacheMiddlewareWithTTL создает middleware с кастомным TTL
func CacheMiddlewareWithTTL(cacheService *services.ResponseCacheService, ttl time.Duration) gin.HandlerFunc {
	return CacheMiddleware(cacheService, ttl)
}

// LongCacheMiddleware для редко изменяемых данных (1 час)
func LongCacheMiddleware(cacheService *services.ResponseCacheService) gin.HandlerFunc {
	return CacheMiddleware(cacheService, cacheService.GetLongTTL())
}

// ShortCacheMiddleware для часто изменяемых данных (2 минуты)
func ShortCacheMiddleware(cacheService *services.ResponseCacheService) gin.HandlerFunc {
	return CacheMiddleware(cacheService, cacheService.GetShortTTL())
}

// generateCacheKey создает уникальный ключ кеша
func generateCacheKey(c *gin.Context) string {
	// Используем путь + query параметры + заголовки авторизации (если есть)
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	// Добавляем информацию о пользователе для персонализированного кеша
	userID := ""
	if val, exists := c.Get("user_id"); exists {
		if id, ok := val.(int); ok {
			userID = strconv.Itoa(id)
		}
	}

	// Создаем хеш для ключа
	keyString := fmt.Sprintf("%s?%s&user=%s", path, query, userID)
	hash := md5.Sum([]byte(keyString))

	return fmt.Sprintf("api:%s:%x", strings.ReplaceAll(path, "/", ":"), hash)
}
