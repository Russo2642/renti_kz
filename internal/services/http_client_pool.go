package services

import (
	"net/http"
	"time"
)

type HTTPClientPool struct {
	FastClient *http.Client

	SlowClient *http.Client

	WebHookClient *http.Client
}

var globalHTTPPool *HTTPClientPool

func InitHTTPClientPool() *HTTPClientPool {
	if globalHTTPPool != nil {
		return globalHTTPPool
	}

	globalHTTPPool = &HTTPClientPool{
		FastClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,              // Максимум неактивных соединений
				MaxIdleConnsPerHost: 20,               // На каждый хост
				IdleConnTimeout:     90 * time.Second, // Время жизни неактивного соединения
				DisableCompression:  false,            // Включаем сжатие для API
				ForceAttemptHTTP2:   true,             // HTTP/2 для лучшей производительности
			},
		},

		SlowClient: &http.Client{
			Timeout: 300 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     300 * time.Second,
				DisableCompression:  true,
				ForceAttemptHTTP2:   true,
			},
		},

		WebHookClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        30,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				ForceAttemptHTTP2:   true,
			},
		},
	}

	return globalHTTPPool
}

func GetHTTPClientPool() *HTTPClientPool {
	if globalHTTPPool == nil {
		return InitHTTPClientPool()
	}
	return globalHTTPPool
}

func GetFastClient() *http.Client {
	return GetHTTPClientPool().FastClient
}

func GetSlowClient() *http.Client {
	return GetHTTPClientPool().SlowClient
}

func GetWebHookClient() *http.Client {
	return GetHTTPClientPool().WebHookClient
}

func GetClientForTimeout(timeoutSeconds int) *http.Client {
	pool := GetHTTPClientPool()

	switch {
	case timeoutSeconds <= 30:
		return pool.FastClient
	case timeoutSeconds <= 60:
		return pool.WebHookClient
	default:
		return pool.SlowClient
	}
}
