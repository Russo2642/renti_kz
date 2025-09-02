package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
)

type SchedulerHandler struct {
	schedulerService *services.SchedulerService
}

func NewSchedulerHandler(schedulerService *services.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerService: schedulerService,
	}
}

// @Summary Получить статистику Scheduler'а
// @Description Возвращает информацию о состоянии Redis Scheduler
// @Tags admin
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/scheduler/stats [get]
func (h *SchedulerHandler) GetStats(c *gin.Context) {
	ctx := context.Background()
	stats := h.schedulerService.GetSchedulerStats(ctx)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статистика scheduler получена", stats))
}

// @Summary Получение метрик производительности планировщика
// @Description Возвращает детальную информацию о производительности планировщика задач
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/scheduler/metrics [get]
func (h *SchedulerHandler) GetSchedulerMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	stats := h.schedulerService.GetSchedulerStats(ctx)
	metrics := h.schedulerService.GetMetrics()

	response := gin.H{
		"scheduler_stats": stats,
		"performance_metrics": gin.H{
			"tasks_processed_total":   metrics.TasksProcessed,
			"tasks_skipped_total":     metrics.TasksSkipped,
			"last_processing_time_ms": metrics.ProcessingTimeMs,
			"last_processing_at":      metrics.LastProcessingTime,
			"active_workers":          metrics.ActiveWorkers,
			"total_errors":            metrics.TotalErrors,
			"database_queries_count":  metrics.DatabaseQueriesCount,
		},
		"efficiency_metrics": gin.H{
			"tasks_processed_per_second": h.calculateTasksPerSecond(metrics),
			"average_processing_time_ms": metrics.ProcessingTimeMs,
			"error_rate_percent":         h.calculateErrorRate(metrics),
			"worker_utilization_percent": h.calculateWorkerUtilization(metrics, stats),
		},
		"recommendations": h.generateRecommendations(metrics, stats),
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("метрики планировщика получены", response))
}

func (h *SchedulerHandler) calculateTasksPerSecond(metrics *services.SchedulerMetrics) float64 {
	if metrics.ProcessingTimeMs == 0 {
		return 0
	}
	return float64(metrics.TasksProcessed) / (float64(metrics.ProcessingTimeMs) / 1000.0)
}

func (h *SchedulerHandler) calculateErrorRate(metrics *services.SchedulerMetrics) float64 {
	total := metrics.TasksProcessed + metrics.TotalErrors
	if total == 0 {
		return 0
	}
	return (float64(metrics.TotalErrors) / float64(total)) * 100
}

func (h *SchedulerHandler) calculateWorkerUtilization(metrics *services.SchedulerMetrics, stats map[string]interface{}) float64 {
	workerPoolCapacity, ok := stats["worker_pool_capacity"].(int)
	if !ok || workerPoolCapacity == 0 {
		return 0
	}
	return (float64(metrics.ActiveWorkers) / float64(workerPoolCapacity)) * 100
}

func (h *SchedulerHandler) generateRecommendations(metrics *services.SchedulerMetrics, stats map[string]interface{}) []string {
	var recommendations []string

	if metrics.ProcessingTimeMs > 5000 {
		recommendations = append(recommendations, "Время обработки задач превышает 5 секунд - рассмотрите увеличение количества воркеров")
	}

	errorRate := h.calculateErrorRate(metrics)
	if errorRate > 5 { 
		recommendations = append(recommendations, fmt.Sprintf("Высокий уровень ошибок (%.1f%%) - проверьте логи и стабильность БД", errorRate))
	}

	queueSize, ok := stats["queue_size"].(int64)
	if ok && queueSize > 1000 {
		recommendations = append(recommendations, "Большая очередь задач - рассмотрите оптимизацию или масштабирование")
	}

	workerUtilization := h.calculateWorkerUtilization(metrics, stats)
	if workerUtilization > 80 {
		recommendations = append(recommendations, "Высокая загрузка воркеров - рассмотрите увеличение pool размера")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Планировщик работает оптимально")
	}

	return recommendations
}

func (h *SchedulerHandler) RegisterRoutes(router *gin.RouterGroup) {
	scheduler := router.Group("/scheduler")
	{
		scheduler.GET("/stats", h.GetStats)
		scheduler.GET("/metrics", h.GetSchedulerMetrics)
	}
}
