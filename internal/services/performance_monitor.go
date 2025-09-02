package services

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type PerformanceMetrics struct {
	RequestCount        uint64
	TotalResponseTime   time.Duration
	AverageResponseTime time.Duration
	MemoryUsage         uint64
	GoRoutines          int
	LastUpdated         time.Time
}

var globalMetrics = &PerformanceMetrics{}

func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		responseTime := time.Since(start)

		updateMetrics(responseTime)

		c.Header("X-Response-Time", responseTime.String())
	}
}

func updateMetrics(responseTime time.Duration) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	globalMetrics.RequestCount++
	globalMetrics.TotalResponseTime += responseTime
	globalMetrics.AverageResponseTime = globalMetrics.TotalResponseTime / time.Duration(globalMetrics.RequestCount)
	globalMetrics.MemoryUsage = m.Alloc
	globalMetrics.GoRoutines = runtime.NumGoroutine()
	globalMetrics.LastUpdated = time.Now()
}

func GetMetrics() *PerformanceMetrics {
	return globalMetrics
}

type MetricsResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    MetricsData `json:"data"`
}

type MetricsData struct {
	RequestsTotal       uint64 `json:"requests_total"`
	AverageResponseTime string `json:"average_response_time"`
	MemoryUsageBytes    uint64 `json:"memory_usage_bytes"`
	MemoryUsageMB       uint64 `json:"memory_usage_mb"`
	TotalMemoryMB       uint64 `json:"total_memory_mb"`
	Goroutines          int    `json:"goroutines"`
	GCCycles            uint32 `json:"gc_cycles"`
	Uptime              string `json:"uptime"`
	LastUpdated         string `json:"last_updated"`
}

// @Summary Получение метрик производительности
// @Description Возвращает текущие метрики производительности сервера: количество запросов, среднее время ответа, использование памяти, количество горутин и другие показатели
// @Tags monitoring
// @Accept json
// @Produce json
// @Success 200 {object} MetricsResponse
// @Router /metrics [get]
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		metricsData := MetricsData{
			RequestsTotal:       globalMetrics.RequestCount,
			AverageResponseTime: globalMetrics.AverageResponseTime.String(),
			MemoryUsageBytes:    m.Alloc,
			MemoryUsageMB:       m.Alloc / 1024 / 1024,
			TotalMemoryMB:       m.TotalAlloc / 1024 / 1024,
			Goroutines:          runtime.NumGoroutine(),
			GCCycles:            m.NumGC,
			Uptime:              time.Since(globalMetrics.LastUpdated).String(),
			LastUpdated:         globalMetrics.LastUpdated.Format(time.RFC3339),
		}

		response := MetricsResponse{
			Status:  "success",
			Message: "метрики производительности получены",
			Data:    metricsData,
		}

		c.JSON(http.StatusOK, response)
	}
}

func LogMetrics(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			log.Printf("📊 Performance Metrics: Requests=%d, Avg Response=%s, Memory=%dMB, Goroutines=%d",
				globalMetrics.RequestCount,
				globalMetrics.AverageResponseTime,
				m.Alloc/1024/1024,
				runtime.NumGoroutine(),
			)
		}
	}()
}

func StartPerformanceMonitoring() {
	globalMetrics.LastUpdated = time.Now()

	LogMetrics(5 * time.Minute)

	log.Println("🚀 Performance monitoring started")
}
