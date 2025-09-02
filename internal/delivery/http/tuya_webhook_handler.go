package http

import (
	"log"
	"net/http"

	"github.com/russo2642/renti_kz/internal/domain"

	"github.com/gin-gonic/gin"
)

type TuyaWebhookHandler struct {
	lockUseCase domain.LockUseCase
}

func NewTuyaWebhookHandler(lockUseCase domain.LockUseCase) *TuyaWebhookHandler {
	return &TuyaWebhookHandler{
		lockUseCase: lockUseCase,
	}
}

func (h *TuyaWebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/tuya", h.HandleTuyaWebhook)
	}
}

func (h *TuyaWebhookHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	locks := router.Group("/locks")
	{
		locks.GET("/unique/:uniqueId/auto-update/status", h.GetAutoUpdateStatus)
		locks.POST("/unique/:uniqueId/auto-update/enable", h.EnableAutoUpdate)
		locks.POST("/unique/:uniqueId/auto-update/disable", h.DisableAutoUpdate)
		locks.POST("/unique/:uniqueId/sync", h.SyncWithTuya)
		locks.POST("/unique/:uniqueId/configure-webhooks", h.ConfigureWebhooks)
	}
}

// @Summary Webhook для событий Tuya
// @Description Обрабатывает webhook события от Tuya API (онлайн/оффлайн статус, heartbeat, данные батареи)
// @Tags tuya-webhooks
// @Accept json
// @Produce json
// @Param event body domain.TuyaWebhookEvent true "Событие от Tuya"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Router /webhooks/tuya [post]
func (h *TuyaWebhookHandler) HandleTuyaWebhook(c *gin.Context) {
	var event domain.TuyaWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		log.Printf("Ошибка парсинга Tuya webhook: %v", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат данных",
		})
		return
	}

	log.Printf("Получен Tuya webhook: BizCode=%s, DevID=%s", event.BizCode, event.DevID)

	if err := h.lockUseCase.ProcessTuyaWebhookEvent(&event); err != nil {
		log.Printf("Ошибка обработки Tuya webhook: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Ошибка обработки события",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Событие успешно обработано",
	})
}

// @Summary Получить статус автообновления замка
// @Description Получает информацию о состоянии автоматического обновления для замка
// @Tags lock-auto-update
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} domain.ErrorResponse
// @Router /locks/unique/{uniqueId}/auto-update/status [get]
func (h *TuyaWebhookHandler) GetAutoUpdateStatus(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	lock, err := h.lockUseCase.GetLockByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Error: "Замок не найден",
		})
		return
	}

	status := map[string]interface{}{
		"auto_update_enabled": lock.AutoUpdateEnabled,
		"webhook_configured":  lock.WebhookConfigured,
		"last_tuya_sync":      lock.LastTuyaSync,
		"last_battery_check":  lock.LastBatteryCheck,
		"battery_type":        lock.BatteryType,
		"charging_status":     lock.ChargingStatus,
	}

	c.JSON(http.StatusOK, status)
}

// @Summary Включить автообновление для замка
// @Description Включает автоматическое обновление статуса и батареи для замка
// @Tags lock-auto-update
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /locks/unique/{uniqueId}/auto-update/enable [post]
func (h *TuyaWebhookHandler) EnableAutoUpdate(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	if err := h.lockUseCase.EnableAutoUpdate(uniqueID); err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Error: "Ошибка включения автообновления",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Автообновление включено",
	})
}

// @Summary Отключить автообновление для замка
// @Description Отключает автоматическое обновление статуса и батареи для замка
// @Tags lock-auto-update
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /locks/unique/{uniqueId}/auto-update/disable [post]
func (h *TuyaWebhookHandler) DisableAutoUpdate(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	if err := h.lockUseCase.DisableAutoUpdate(uniqueID); err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Error: "Ошибка отключения автообновления",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Автообновление отключено",
	})
}

// @Summary Синхронизировать замок с Tuya
// @Description Выполняет ручную синхронизацию данных замка с Tuya API
// @Tags lock-auto-update
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /locks/unique/{uniqueId}/sync [post]
func (h *TuyaWebhookHandler) SyncWithTuya(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	if err := h.lockUseCase.SyncLockWithTuya(uniqueID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Ошибка синхронизации с Tuya",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Синхронизация выполнена",
	})
}

// @Summary Настроить webhooks для замка
// @Description Настраивает Tuya webhooks для автоматического получения событий
// @Tags lock-auto-update
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /locks/unique/{uniqueId}/configure-webhooks [post]
func (h *TuyaWebhookHandler) ConfigureWebhooks(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	if err := h.lockUseCase.ConfigureTuyaWebhooks(uniqueID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Ошибка настройки webhooks",
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Message: "Webhooks настроены",
	})
}
