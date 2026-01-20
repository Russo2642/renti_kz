package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type NotificationHandler struct {
	notificationUseCase domain.NotificationUseCase
}

func NewNotificationHandler(notificationUseCase domain.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: notificationUseCase,
	}
}

func (h *NotificationHandler) RegisterRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	notifications.Use()
	{
		notifications.GET("/unread", h.GetUnreadNotifications)
		notifications.GET("", h.GetNotifications)
		notifications.GET("/count", h.GetUnreadCount)
		notifications.POST("/:id/read", h.MarkAsRead)
		notifications.POST("/read-multiple", h.MarkMultipleAsRead)
		notifications.POST("/read-all", h.MarkAllAsRead)
		notifications.DELETE("/:id", h.DeleteNotification)
		notifications.DELETE("/read", h.DeleteReadNotifications)
		notifications.DELETE("/all", h.DeleteAllNotifications)
	}

	devices := router.Group("/devices")
	devices.Use()
	{
		devices.POST("/register", h.RegisterDevice)
		devices.POST("/heartbeat", h.UpdateDeviceHeartbeat)
		devices.DELETE("/:token", h.RemoveDevice)
	}
}

// @Summary Получение непрочитанных уведомлений
// @Description Получает список непрочитанных уведомлений текущего пользователя
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/unread [get]
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	notifications, err := h.notificationUseCase.GetUnreadNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении уведомлений: "+err.Error()))
		return
	}

	responses := convertNotificationsToResponse(notifications)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("непрочитанные уведомления получены", responses))
}

// @Summary Получение уведомлений с пагинацией
// @Description Получает список уведомлений текущего пользователя с пагинацией
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	pageParam := c.DefaultQuery("page", "1")
	pageSizeParam := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeParam)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	notifications, err := h.notificationUseCase.GetUserNotifications(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении уведомлений: "+err.Error()))
		return
	}

	totalCount, err := h.notificationUseCase.GetUnreadCount(userID)
	if err != nil {
		totalCount = 0
	}

	responses := convertNotificationsToResponse(notifications)

	result := gin.H{
		"notifications": responses,
		"pagination": gin.H{
			"total":     totalCount,
			"page":      page,
			"page_size": pageSize,
			"pages":     (totalCount + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("уведомления получены", result))
}

// @Summary Получение количества непрочитанных уведомлений
// @Description Получает количество непрочитанных уведомлений текущего пользователя
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	count, err := h.notificationUseCase.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении количества: "+err.Error()))
		return
	}

	result := gin.H{
		"count": count,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("количество получено", result))
}

// @Summary Пометить уведомление как прочитанное
// @Description Помечает указанное уведомление как прочитанное
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID уведомления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	notificationID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.notificationUseCase.MarkAsRead(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("уведомление помечено как прочитанное", nil))
}

type MarkMultipleAsReadRequest struct {
	NotificationIDs []int `json:"notification_ids" validate:"required"`
}

// @Summary Пометить несколько уведомлений как прочитанные
// @Description Помечает указанные уведомления как прочитанные
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body MarkMultipleAsReadRequest true "Список ID уведомлений"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/read-multiple [post]
func (h *NotificationHandler) MarkMultipleAsRead(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var req MarkMultipleAsReadRequest
	if !utils.ShouldBindJSON(c, &req) {
		return
	}

	if !utils.ValidateStruct(c, req) {
		return
	}

	err := h.notificationUseCase.MarkMultipleAsRead(req.NotificationIDs, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("уведомления помечены как прочитанные", nil))
}

// @Summary Пометить все уведомления как прочитанные
// @Description Помечает все уведомления пользователя как прочитанные
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	err := h.notificationUseCase.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при пометке уведомлений: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("все уведомления помечены как прочитанные", nil))
}

type RegisterDeviceRequest struct {
	DeviceToken string            `json:"device_token" validate:"required"`
	DeviceType  domain.DeviceType `json:"device_type" validate:"required"`
	AppVersion  string            `json:"app_version"`
	OSVersion   string            `json:"os_version"`
}

// @Summary Регистрация устройства
// @Description Регистрирует устройство пользователя для получения push уведомлений
// @Tags devices
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body RegisterDeviceRequest true "Данные устройства"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /devices/register [post]
func (h *NotificationHandler) RegisterDevice(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var req RegisterDeviceRequest
	if !utils.ShouldBindJSON(c, &req) {
		return
	}

	if !utils.ValidateStruct(c, req) {
		return
	}

	if err := domain.ValidateExpoToken(req.DeviceToken); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("невалидный Expo токен: "+err.Error()))
		return
	}

	device := &domain.UserDevice{
		UserID:      userID,
		DeviceToken: req.DeviceToken,
		DeviceType:  req.DeviceType,
		IsActive:    true,
		AppVersion:  req.AppVersion,
		OSVersion:   req.OSVersion,
		LastSeenAt:  time.Now(),
	}

	err := h.notificationUseCase.RegisterDevice(device)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("устройство зарегистрировано", nil))
}

type DeviceHeartbeatRequest struct {
	DeviceToken string `json:"device_token" validate:"required"`
}

// @Summary Обновление активности устройства
// @Description Обновляет время последней активности устройства (heartbeat)
// @Tags devices
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body DeviceHeartbeatRequest true "Токен устройства"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /devices/heartbeat [post]
func (h *NotificationHandler) UpdateDeviceHeartbeat(c *gin.Context) {
	_, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var req DeviceHeartbeatRequest
	if !utils.ShouldBindJSON(c, &req) {
		return
	}

	if !utils.ValidateStruct(c, req) {
		return
	}

	err := h.notificationUseCase.UpdateDeviceHeartbeat(req.DeviceToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("активность устройства обновлена", nil))
}

// @Summary Удаление устройства
// @Description Деактивирует устройство в системе уведомлений
// @Tags devices
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param token path string true "Токен устройства"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /devices/{token} [delete]
func (h *NotificationHandler) RemoveDevice(c *gin.Context) {
	_, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("токен устройства обязателен"))
		return
	}

	err := h.notificationUseCase.DeactivateDevice(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("устройство удалено", nil))
}

// @Summary Удалить уведомление
// @Description Удаляет указанное уведомление пользователя
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID уведомления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	notificationID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.notificationUseCase.DeleteNotification(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("уведомление удалено", nil))
}

// @Summary Удалить прочитанные уведомления
// @Description Удаляет все прочитанные уведомления пользователя
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/read [delete]
func (h *NotificationHandler) DeleteReadNotifications(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	deletedCount, err := h.notificationUseCase.DeleteReadNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении уведомлений: "+err.Error()))
		return
	}

	result := gin.H{
		"deleted_count": deletedCount,
		"message":       fmt.Sprintf("Удалено %d прочитанных уведомлений", deletedCount),
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("прочитанные уведомления удалены", result))
}

// @Summary Удалить все уведомления
// @Description Удаляет все уведомления пользователя (прочитанные и непрочитанные)
// @Tags notifications
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /notifications/all [delete]
func (h *NotificationHandler) DeleteAllNotifications(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	deletedCount, err := h.notificationUseCase.DeleteAllNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении уведомлений: "+err.Error()))
		return
	}

	result := gin.H{
		"deleted_count": deletedCount,
		"message":       fmt.Sprintf("Удалено %d уведомлений", deletedCount),
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("все уведомления удалены", result))
}

func convertNotificationsToResponse(notifications []*domain.Notification) []*domain.NotificationResponse {
	responses := make([]*domain.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		var readAt *string
		if notification.ReadAt != nil {
			readAtStr := notification.ReadAt.Format(time.RFC3339)
			readAt = &readAtStr
		}

		responses[i] = &domain.NotificationResponse{
			ID:          notification.ID,
			Type:        notification.Type,
			Priority:    notification.Priority,
			Title:       notification.Title,
			Body:        notification.Message,
			Data:        notification.Data,
			IsRead:      notification.IsRead,
			BookingID:   notification.BookingID,
			ApartmentID: notification.ApartmentID,
			CreatedAt:   notification.CreatedAt.Format(time.RFC3339),
			ReadAt:      readAt,
		}
	}
	return responses
}
